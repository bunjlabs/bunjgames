package storage

import (
	"bunjgames/game/abstract"
	"fmt"
	"sync"
)

type GameStore struct {
	Mutex sync.RWMutex
	Games map[string]abstract.Game
}

func NewGameStore() *GameStore {
	return &GameStore{Games: make(map[string]abstract.Game)}
}

func (store *GameStore) Get(token string) abstract.Game {
	store.Mutex.RLock()
	defer store.Mutex.RUnlock()
	game, _ := store.Games[token]
	return game
}

func (store *GameStore) Add(game abstract.Game) error {
	store.Mutex.Lock()
	defer store.Mutex.Unlock()

	if _, tokenAlreadyExists := store.Games[game.GetToken()]; tokenAlreadyExists {
		return fmt.Errorf("game with token %s already exists", game.GetToken())
	}

	store.Games[game.GetToken()] = game
	return nil
}
