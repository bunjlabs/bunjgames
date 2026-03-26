package server

import (
	"bunjgames/hub"
	"bunjgames/storage"
	"log"
	"time"
)

func StartGameLoop(store *storage.GameStore, hub *hub.Hub) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		lastTick := time.Now()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				delta := now.Sub(lastTick)
				lastTick = now

				store.Mutex.RLock()
				for _, game := range store.Games {
					responseCommand, err := game.Tick(delta)
					if err != nil {
						log.Printf("Error in game loop: %v", err)
					} else if responseCommand != nil {
						hub.Broadcast(game.GetToken(), responseCommand)
					}
				}
				store.Mutex.RUnlock()
			case <-stop:
				return
			}
		}
	}()

	return stop
}
