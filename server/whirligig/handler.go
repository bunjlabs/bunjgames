package whirligig

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bunjgames-server/common"
)

var GameStore = common.NewStore[Game]()

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("game")
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, "Missing game file")
		return
	}
	defer file.Close()

	game := NewGame()
	token := GameStore.GenerateUniqueToken()
	game.Token = token

	tmpDir, err := os.MkdirTemp("", "whirligig-*")
	if err != nil {
		common.ErrorResponse(w, http.StatusInternalServerError, "Failed to create temp dir")
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "game.zip")
	out, err := os.Create(tmpFile)
	if err != nil {
		common.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	io.Copy(out, file)
	out.Close()

	gamePath := filepath.Join("media", "whirligig", token)
	os.MkdirAll(gamePath, 0755)

	if err := common.Unzip(tmpFile, gamePath); err != nil {
		os.RemoveAll(gamePath)
		common.ErrorResponse(w, http.StatusBadRequest, "Bad game file")
		return
	}

	contentFilePath := filepath.Join(gamePath, "content.yaml")
	contentData, err := os.ReadFile(contentFilePath)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, "Cannot read file")
	}
	log.Printf("Parsing game file %s", contentFilePath)
	if err := game.Parse(contentData); err != nil {
		os.RemoveAll(gamePath)
		if bfe, ok := err.(*common.BadFormatError); ok {
			common.ErrorResponse(w, http.StatusBadRequest, bfe.Msg)
		} else {
			log.Printf("Parse error: %v", err)
			common.ErrorResponse(w, http.StatusBadRequest, "Bad game file")
		}
		return
	}
	os.Remove(contentFilePath)

	entries, _ := os.ReadDir(gamePath)
	if len(entries) == 0 {
		os.Remove(gamePath)
	}

	GameStore.Set(token, game)
	common.JSONResponse(w, game.Serialize())
}

func NewConsumer(hub *common.Hub) *common.ConsumerHandler {
	return &common.ConsumerHandler{
		Hub:      hub,
		GameName: "whirligig",
		GetState: func(token string) (any, error) {
			game, ok := GameStore.Get(token)
			if !ok {
				return nil, &common.BadStateError{Msg: "Game not found"}
			}
			game.Mu.Lock()
			defer game.Mu.Unlock()
			return game.Serialize(), nil
		},
		Process: func(token string, method string, params map[string]any) (any, []string, error) {
			game, ok := GameStore.Get(token)
			if !ok {
				return nil, nil, &common.BadStateError{Msg: "Game not found"}
			}
			game.Mu.Lock()
			defer game.Mu.Unlock()

			var err = game.ProcessCommand(method, params)

			if err != nil {
				return nil, nil, err
			}
			return game.Serialize(), nil, nil
		},
	}
}
