package whirligig

import (
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
		TempDirPrefix: "whirligig-*",
		FileExtension: ".zip",
		MediaSubdir:   "whirligig",
		NeedsUnzip:    true,
	},
)

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

			if err := game.ProcessCommand(method, params); err != nil {
				return nil, nil, err
			}

			return game.Serialize(), nil, nil
		},
	}
}
