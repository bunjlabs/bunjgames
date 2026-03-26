package abstract

import (
	"bunjgames/hub"
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"time"
)

func (game *BaseGame) generateToken() {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	game.Token = string(b)
}

func (game *BaseGame) Initialise() {
	game.generateToken()
}

func (game *BaseGame) GetToken() string {
	return game.Token
}

func (game *BaseGame) Tick(time.Duration) (*Command, error) {
	return nil, nil
}

func (game *BaseGame) RegisterPlayer(string) error {
	return errors.New("this game does not support players")
}

func (game *BaseGame) Parse(io.Reader) error {
	return errors.New("this game does not support parsing")
}

func (game *BaseGame) ProcessCommand(string, map[string]any) (*Command, error) {
	return nil, errors.New("this game does not support commands")
}

func (game *BaseGame) Intercom(message string) {
	hub.GetHub().Broadcast(
		game.Token,
		&Command{Type: "intercom", Message: message},
	)
}
