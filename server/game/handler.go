package game

import (
	"bunjgames/game/abstract"
	"bunjgames/game/whirligig"
	"bunjgames/storage"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

func newGame(name string) (abstract.Game, bool) {
	switch name {
	case "whirligig":
		return whirligig.NewGame(), true
	}
	return nil, false
}

func HandleGameCreation(
	context *echo.Context,
	store *storage.GameStore,
) error {
	gameName := context.FormValue("game")
	game, unzip := newGame(gameName)
	if game == nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Invalid game type"})
	}

	file, err := context.FormFile("file")
	if err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Missing game file"})
	}

	gameFile, err := storage.UploadFile(file, filepath.Join("media", game.GetToken()), unzip)
	if err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Missing game file"})
	}
	defer gameFile.Close()

	err = game.Parse(gameFile)
	if err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": err})
	}

	err = store.Add(game)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, map[string]any{"detail": err})
	}

	return context.JSON(http.StatusOK, game)
}

var (
	upgrader = websocket.Upgrader{}
)

type CommandMessage struct {
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Message any            `json:"message"`
}

func HandleGameConnection(
	context *echo.Context,
	game abstract.Game,
) error {
	if game == nil {
		return context.JSON(http.StatusNotFound, map[string]any{"detail": "Game not found"})
	}

	ws, err := upgrader.Upgrade(context.Response(), context.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	err = ws.WriteJSON(game)
	if err != nil {
		context.Logger().Error("failed to write WS message", "error", err)
		return err
	}

	for {
		var message CommandMessage
		err := ws.ReadJSON(message)
		if err != nil {
			context.Logger().Error("failed to read WS message", "error", err)
		}
		payload, err := game.ProcessCommand(message.Method, message.Params)
		if err != nil {
			err := ws.WriteJSON(map[string]any{"detail": "Game not found"})
			if err != nil {
				context.Logger().Error("failed to write WS message", "error", err)
			}
		}
		if payload != nil {
			err := ws.WriteJSON(payload)
			if err != nil {
				context.Logger().Error("failed to write WS message", "error", err)
			}
		} else {
			err := ws.WriteJSON(game)
			if err != nil {
				context.Logger().Error("failed to write WS message", "error", err)
			}
		}
	}
}
