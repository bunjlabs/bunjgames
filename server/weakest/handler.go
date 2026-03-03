package weakest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	tmpDir, err := os.MkdirTemp("", "weakest-*")
	if err != nil {
		common.ErrorResponse(w, http.StatusInternalServerError, "Failed to create temp dir")
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "game.xml")
	out, err := os.Create(tmpFile)
	if err != nil {
		common.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	io.Copy(out, file)
	out.Close()

	if err := game.Parse(tmpFile); err != nil {
		if bfe, ok := err.(*common.BadFormatError); ok {
			common.ErrorResponse(w, http.StatusBadRequest, bfe.Msg)
		} else {
			log.Printf("Parse error: %v", err)
			common.ErrorResponse(w, http.StatusBadRequest, "Bad game file")
		}
		return
	}

	GameStore.Set(token, game)
	common.JSONResponse(w, game.Serialize())
}

func RegisterPlayerHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	token := strings.ToUpper(strings.TrimSpace(body.Token))
	name := strings.ToUpper(strings.TrimSpace(body.Name))

	game, ok := GameStore.Get(token)
	if !ok {
		common.ErrorResponse(w, http.StatusBadRequest, "Game not found")
		return
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	existing := game.RegisterPlayer(name)
	if existing != nil {
		common.JSONResponse(w, map[string]any{
			"player_id": existing.ID,
			"game":      game.Serialize(),
		})
		return
	}

	if game.State != StateWaitingForPlayers {
		common.ErrorResponse(w, http.StatusBadRequest, "Game already started")
		return
	}

	player := game.AddPlayer(name)
	common.JSONResponse(w, map[string]any{
		"player_id": player.ID,
		"game":      game.Serialize(),
	})
}

func NewConsumer(hub *common.Hub) *common.ConsumerHandler {
	return &common.ConsumerHandler{
		Hub:      hub,
		GameName: "weakest",
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

			var err error
			switch method {
			case "next_state":
				fromState := common.OptStringParam(params, "from_state")
				err = game.NextState(fromState)
			case "save_bank":
				err = game.SaveBank(false)
			case "answer_correct":
				isCorrect, e := common.BoolParam(params, "is_correct")
				if e != nil {
					return nil, nil, e
				}
				err = game.AnswerCorrect(isCorrect)
			case "select_weakest":
				pid, e1 := common.IntParam(params, "player_id")
				wid, e2 := common.IntParam(params, "weakest_id")
				if e1 != nil || e2 != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.SelectWeakest(pid, wid)
			case "select_final_answerer":
				pid, e := common.IntParam(params, "player_id")
				if e != nil {
					return nil, nil, e
				}
				err = game.SelectFinalAnswerer(pid)
			default:
				return nil, nil, &common.BadFormatError{Msg: "Unknown method"}
			}

			if err != nil {
				return nil, nil, err
			}
			return game.Serialize(), nil, nil
		},
	}
}
