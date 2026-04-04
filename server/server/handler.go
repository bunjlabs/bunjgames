package server

import (
	"bunjgames/game/abstract"
	"bunjgames/game/feud"
	"bunjgames/game/jeopardy"
	"bunjgames/game/weakest"
	"bunjgames/game/whirligig"
	"bunjgames/hub"
	"bunjgames/storage"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

func newGame(name string) (abstract.Game, storage.ArchiveType) {
	switch name {
	case "whirligig":
		return whirligig.NewGame(), storage.ArchiveTarGz
	case "jeopardy":
		return jeopardy.NewGame(), storage.ArchiveZip
	case "weakest":
		return weakest.NewGame(), storage.ArchiveNone
	case "feud":
		return feud.NewGame(), storage.ArchiveNone
	}
	return nil, storage.ArchiveNone
}

func HandleGameCreation(
	context *echo.Context,
	store *storage.GameStore,
) error {
	gameName := context.FormValue("game")
	game, archiveType := newGame(gameName)
	if game == nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Invalid game type"})
	}

	file, err := context.FormFile("file")
	if err != nil {
		context.Logger().Error("failed to get game file", "error", err)
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Missing game file"})
	}

	gameFile, err := storage.UploadFile(file, filepath.Join("media", game.GetToken()), archiveType)
	if err != nil {
		context.Logger().Error("failed to upload game file", "error", err)
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Failed to process game file"})
	}

	err = game.Parse(gameFile)
	defer gameFile.Close()
	if err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": err.Error()})
	}

	err = store.Add(game)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, map[string]any{"detail": err.Error()})
	}

	return context.JSON(http.StatusOK, game)
}

func HandlePlayerRegistration(
	context *echo.Context,
	store *storage.GameStore,
	hub *hub.Hub,
) error {
	var body struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}
	if err := context.Bind(&body); err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Invalid request"})
	}

	token := strings.ToUpper(strings.TrimSpace(body.Token))
	name := strings.ToUpper(strings.TrimSpace(body.Name))
	if token == "" || name == "" {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Token and name are required"})
	}

	game := store.Get(token)
	if game == nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": "Game not found"})
	}

	err := game.RegisterPlayer(name)
	if err != nil {
		return context.JSON(http.StatusBadRequest, map[string]any{"detail": err.Error()})
	}

	hub.Broadcast(game.GetToken(), &abstract.Command{"game", game})
	return context.JSON(http.StatusOK, map[string]any{"player": name, "game": game})
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // allow all origins (or add specific origin checks)
		},
	}
)

func HandleGameConnection(
	context *echo.Context,
	game abstract.Game,
	hub *hub.Hub,
) error {
	if game == nil {
		return context.JSON(http.StatusNotFound, map[string]any{"type": "error", "message": "Game not found"})
	}

	ws, err := upgrader.Upgrade(context.Response(), context.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	hub.Register(game.GetToken(), ws)
	defer hub.Unregister(game.GetToken(), ws)

	token := game.GetToken()

	err = hub.Send(token, ws, abstract.Command{Type: "game", Message: game})
	if err != nil {
		context.Logger().Error("failed to write WS message", "error", err)
		return nil
	}

	for {
		var command abstract.Command
		err := ws.ReadJSON(&command)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				context.Logger().Info("WS connection closed", "error", err)
				return nil
			}
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				context.Logger().Info("WS connection closed", "error", err)
				return nil
			}
			if websocket.IsUnexpectedCloseError(err) {
				context.Logger().Error("unexpected WS close", "error", err)
				return nil
			}
			_ = hub.Send(token, ws, abstract.Command{Type: "error", Message: "Invalid command format"})
			context.Logger().Error("failed to read WS message", "error", err)
			continue
		} else {
			context.Logger().Info("received command", "command", command)
		}

		if command.Type == "intercom" {
			hub.Broadcast(token, command)
			continue
		}

		commandMessage, ok := command.Message.(map[string]any)
		if !ok {
			continue
		}

		outputCommand, err := game.ProcessCommand(command.Type, commandMessage)
		if err != nil {
			if sendErr := hub.Send(token, ws, abstract.Command{Type: "error", Message: err.Error()}); sendErr != nil {
				context.Logger().Error("failed to write WS message", "error", sendErr)
			}
		} else if outputCommand != nil {
			hub.Broadcast(token, outputCommand)
		}
	}
}
