package hub

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	hubInstance *Hub
	once        sync.Once
)

type client struct {
	mu sync.Mutex
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]*client
}

func GetHub() *Hub {
	once.Do(func() {
		hubInstance = &Hub{
			clients: make(map[string]map[*websocket.Conn]*client),
		}
		fmt.Println("Creating single instance now.")
	})
	return hubInstance
}

func (hub *Hub) Register(token string, conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if hub.clients[token] == nil {
		hub.clients[token] = make(map[*websocket.Conn]*client)
	}
	hub.clients[token][conn] = &client{}
}

func (hub *Hub) Unregister(token string, conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if conns, ok := hub.clients[token]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(hub.clients, token)
		}
	}
}

func (hub *Hub) Send(token string, conn *websocket.Conn, message any) error {
	hub.mu.RLock()
	cl, ok := hub.clients[token][conn]
	hub.mu.RUnlock()
	if !ok {
		return fmt.Errorf("connection not registered for token %s", token)
	}
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return conn.WriteJSON(message)
}

type target struct {
	conn   *websocket.Conn
	client *client
}

func (hub *Hub) Broadcast(token string, message any) {
	hub.mu.RLock()

	if _, ok := hub.clients[token]; !ok {
		hub.mu.RUnlock()
		return
	}

	targets := make([]target, 0, len(hub.clients[token]))
	for conn, cl := range hub.clients[token] {
		targets = append(targets, target{conn, cl})
	}
	hub.mu.RUnlock()

	for _, t := range targets {
		go func(c *websocket.Conn, cl *client) {
			cl.mu.Lock()
			defer cl.mu.Unlock()
			_ = c.WriteJSON(message)
		}(t.conn, t.client)
	}
}
