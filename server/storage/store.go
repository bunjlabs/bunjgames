package storage

import (
	"bunjgames/game/abstract"
	"fmt"
	"sync"
)

type GameStore struct {
	mutex sync.RWMutex
	games map[string]abstract.Game
}

func NewGameStore() *GameStore {
	return &GameStore{games: make(map[string]abstract.Game)}
}

func (store *GameStore) Get(token string) abstract.Game {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	game, _ := store.games[token]
	return game
}

func (store *GameStore) Add(game abstract.Game) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, tokenAlreadyExists := store.games[game.GetToken()]; tokenAlreadyExists {
		return fmt.Errorf("game with token %s already exists", game.GetToken())
	}

	store.games[game.GetToken()] = game
	return nil
}
