package common

import (
	"archive/zip"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// --- Errors ---

type BadFormatError struct{ Msg string }

func (e *BadFormatError) Error() string { return e.Msg }

type BadStateError struct{ Msg string }

func (e *BadStateError) Error() string { return e.Msg }

type NothingToDoError struct{}

func (e *NothingToDoError) Error() string { return "nothing to do" }

var ErrNothingToDo = &NothingToDoError{}

// --- ID Generation ---

var globalID atomic.Int64

func NextID() int {
	return int(globalID.Add(1))
}

// --- Token Generation ---

func GenerateToken() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// --- Store ---

type Store[T any] struct {
	mu    sync.RWMutex
	games map[string]*T
}

func NewStore[T any]() *Store[T] {
	return &Store[T]{games: make(map[string]*T)}
}

func (s *Store[T]) Get(token string) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.games[token]
	return g, ok
}

func (s *Store[T]) Set(token string, game *T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[token] = game
}

func (s *Store[T]) Exists(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.games[token]
	return ok
}

func (s *Store[T]) GenerateUniqueToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := 0; i < 3; i++ {
		token := GenerateToken()
		if _, exists := s.games[token]; !exists {
			return token, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique token after 3 attempts")
}

// --- Param Helpers ---

func IntParam(params map[string]any, key string) (int, error) {
	v, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing param: %s", key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case json.Number:
		i, err := n.Int64()
		return int(i), err
	}
	return 0, fmt.Errorf("invalid param type for %s", key)
}

func BoolParam(params map[string]any, key string) (bool, error) {
	v, ok := params[key]
	if !ok {
		return false, fmt.Errorf("missing param: %s", key)
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("invalid param type for %s", key)
	}
	return b, nil
}

func StringParam(params map[string]any, key string) (string, error) {
	v, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing param: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("invalid param type for %s", key)
	}
	return s, nil
}

func OptStringParam(params map[string]any, key string) *string {
	v, ok := params[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

func IntSliceParam(params map[string]any, key string) ([]int, error) {
	v, ok := params[key]
	if !ok {
		return nil, fmt.Errorf("missing param: %s", key)
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid param type for %s", key)
	}
	result := make([]int, len(arr))
	for i, item := range arr {
		f, ok := item.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid array element type")
		}
		result[i] = int(f)
	}
	return result, nil
}

func OptIntParam(params map[string]any, key string) int {
	v, ok := params[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

// --- HTTP Helpers ---

func JSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"detail": message})
}

// --- Unzip ---

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name, _ := url.QueryUnescape(f.Name)
		if strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
			continue
		}

		target := filepath.Join(dest, filepath.FromSlash(name))
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// --- Intercom Queue (embeddable) ---

type IntercomQueue struct {
	pending []string
}

func (q *IntercomQueue) QueueIntercom(msg string) {
	q.pending = append(q.pending, msg)
}

func (q *IntercomQueue) DrainIntercoms() []string {
	msgs := q.pending
	q.pending = nil
	return msgs
}

// --- Generic Create Game Handler ---

type CreateGameConfig struct {
	TempDirPrefix string
	FileExtension string
	MediaSubdir   string
	NeedsUnzip    bool
}

func CreateGameHandler[T any](
	store *Store[T],
	newGame func(token string) *T,
	parseGame func(*T, any) error,
	serializeGame func(*T) any,
	config CreateGameConfig,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("game")
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, "Missing game file")
			return
		}
		defer file.Close()

		token, err := store.GenerateUniqueToken()
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Failed to generate unique token")
			return
		}
		game := newGame(token)

		tmpDir, err := os.MkdirTemp("", config.TempDirPrefix)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Failed to create temp dir")
			return
		}
		defer os.RemoveAll(tmpDir)

		tmpFile := filepath.Join(tmpDir, "game"+config.FileExtension)
		out, err := os.Create(tmpFile)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			ErrorResponse(w, http.StatusInternalServerError, "Failed to write file")
			return
		}
		out.Close()

		var parseData any = tmpFile
		var gamePath string

		if config.NeedsUnzip {
			gamePath = filepath.Join("media", config.MediaSubdir, token)
			os.MkdirAll(gamePath, 0755)

			if err := Unzip(tmpFile, gamePath); err != nil {
				os.RemoveAll(gamePath)
				ErrorResponse(w, http.StatusBadRequest, "Bad game file")
				return
			}

			yamlFile := filepath.Join(gamePath, "content.yaml")
			xmlFile := filepath.Join(gamePath, "content.xml")

			contentData, err := os.ReadFile(yamlFile)
			if err != nil {
				contentData, err = os.ReadFile(xmlFile)
				if err != nil {
					os.RemoveAll(gamePath)
					ErrorResponse(w, http.StatusBadRequest, "Cannot read content file (tried yaml and xml)")
					return
				}
			}
			parseData = contentData
		} else {
			// Read XML file directly
			contentData, err := os.ReadFile(tmpFile)
			if err != nil {
				ErrorResponse(w, http.StatusInternalServerError, "Failed to read file")
				return
			}
			parseData = contentData
		}

		if err := parseGame(game, parseData); err != nil {
			if config.NeedsUnzip {
				os.RemoveAll(gamePath)
			}
			if bfe, ok := err.(*BadFormatError); ok {
				ErrorResponse(w, http.StatusBadRequest, bfe.Msg)
			} else {
				log.Printf("Parse error: %v", err)
				ErrorResponse(w, http.StatusBadRequest, "Bad game file")
			}
			return
		}

		if config.NeedsUnzip && gamePath != "" {
			entries, _ := os.ReadDir(gamePath)
			if len(entries) == 0 {
				os.Remove(gamePath)
			}
		}

		store.Set(token, game)
		JSONResponse(w, serializeGame(game))
	}
}
