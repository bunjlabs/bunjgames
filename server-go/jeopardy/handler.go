package jeopardy

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bunjgames/server-go/common"
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

	tmpDir, err := os.MkdirTemp("", "jeopardy-*")
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

	gamePath := filepath.Join("media", "jeopardy", token)
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
		GameName: "jeopardy",
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
			case "choose_question":
				qid, e := common.IntParam(params, "question_id")
				if e != nil {
					return nil, nil, e
				}
				err = game.ChooseQuestion(qid)
			case "set_answerer_and_bet":
				pid, e1 := common.IntParam(params, "player_id")
				bet, e2 := common.IntParam(params, "bet")
				if e1 != nil || e2 != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.SetAnswererAndBet(pid, bet)
			case "skip_question":
				err = game.SkipQuestion()
			case "button_click":
				pid, e := common.IntParam(params, "player_id")
				if e != nil {
					return nil, nil, e
				}
				err = game.ButtonClick(pid)
			case "answer":
				isRight, e := common.BoolParam(params, "is_right")
				if e != nil {
					return nil, nil, e
				}
				err = game.AnswerQuestion(isRight)
			case "remove_final_theme":
				tid, e := common.IntParam(params, "theme_id")
				if e != nil {
					return nil, nil, e
				}
				err = game.RemoveFinalTheme(tid)
			case "final_bet":
				pid, e1 := common.IntParam(params, "player_id")
				bet, e2 := common.IntParam(params, "bet")
				if e1 != nil || e2 != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.FinalBet(pid, bet)
			case "final_answer":
				pid, e1 := common.IntParam(params, "player_id")
				answer, e2 := common.StringParam(params, "answer")
				if e1 != nil || e2 != nil {
					return nil, nil, &common.BadFormatError{Msg: "Invalid params"}
				}
				err = game.FinalAnswerPlayer(pid, answer)
			case "final_player_answer":
				isRight, e := common.BoolParam(params, "is_right")
				if e != nil {
					return nil, nil, e
				}
				err = game.FinalPlayerAnswer(isRight)
			case "set_balance":
				bl, e := common.IntSliceParam(params, "balance_list")
				if e != nil {
					return nil, nil, e
				}
				game.SetBalance(bl)
			case "set_round":
				round, e := common.IntParam(params, "round")
				if e != nil {
					return nil, nil, e
				}
				game.SetRound(round)
			default:
				return nil, nil, &common.BadFormatError{Msg: "Unknown method"}
			}

			if err != nil {
				return nil, nil, err
			}
			intercoms := game.DrainIntercoms()
			return game.Serialize(), intercoms, nil
		},
	}
}
