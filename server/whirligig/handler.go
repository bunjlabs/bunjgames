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

	contentXML := filepath.Join(gamePath, "content.xml")
	if err := game.Parse(contentXML); err != nil {
		os.RemoveAll(gamePath)
		if bfe, ok := err.(*common.BadFormatError); ok {
			common.ErrorResponse(w, http.StatusBadRequest, bfe.Msg)
		} else {
			log.Printf("Parse error: %v", err)
			common.ErrorResponse(w, http.StatusBadRequest, "Bad game file")
		}
		return
	}
	os.Remove(contentXML)

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

			var err error
			switch method {
			case "next_state":
				fromState := common.OptStringParam(params, "from_state")
				err = game.NextState(fromState)
			case "change_score":
				cs, e1 := common.IntParam(params, "connoisseurs_score")
				vs, e2 := common.IntParam(params, "viewers_score")
				if e1 != nil || e2 != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				game.ChangeScore(cs, vs)
			case "change_timer":
				paused, e := common.BoolParam(params, "paused")
				if e != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.ChangeTimer(paused)
			case "answer_correct":
				isCorrect, e := common.BoolParam(params, "is_correct")
				if e != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.AnswerCorrect(isCorrect)
			case "extra_time":
				err = game.ExtraTime()
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
