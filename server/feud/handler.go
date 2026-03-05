package feud

import (
	"encoding/json"
	"net/http"
	"strings"

	"bunjgames-server/common"
)

var GameStore = common.NewStore[Game]()

var CreateHandler = common.CreateGameHandler(
	GameStore,
	func(token string) *Game {
		game := NewGame()
		game.Token = token
		return game
	},
	func(game *Game, data any) error {
		return game.Parse(data.([]byte))
	},
	func(game *Game) any {
		return game.Serialize()
	},
	common.CreateGameConfig{
		TempDirPrefix: "feud-*",
		FileExtension: ".xml",
		MediaSubdir:   "",
		NeedsUnzip:    false,
	},
)

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

	if len(game.Players) >= 2 {
		common.ErrorResponse(w, http.StatusBadRequest, "Game already have 2 teams")
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
		GameName: "feud",
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

			if err := game.ProcessCommand(method, params); err != nil {
				return nil, nil, err
			}

			intercoms := game.DrainIntercoms()
			return game.Serialize(), intercoms, nil
		},
	}
}
