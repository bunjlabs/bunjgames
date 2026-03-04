package abstract

import (
	"crypto/rand"
	"math/big"
)

func (game *BaseGame) GenerateToken() *BaseGame {
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
	return game
}

func (game *BaseGame) GetToken() string {
	return game.Token
}

func (game *BaseGame) Serialize() any {
	return game
}
