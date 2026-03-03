package common

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
	}
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*Client]bool)}
}

func (h *Hub) Register(room string, conn *websocket.Conn) *Client {
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	h.mu.Lock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	h.mu.Unlock()
	go client.writePump()
	return client
}

func (h *Hub) Unregister(room string, client *Client) {
	h.mu.Lock()
	delete(h.rooms[room], client)
	if len(h.rooms[room]) == 0 {
		delete(h.rooms, room)
	}
	h.mu.Unlock()
	close(client.send)
}

func (h *Hub) Broadcast(room string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.rooms[room] {
		client.Send(message)
	}
}

// ConsumerHandler provides the generic WebSocket consumer pattern shared across all games.
type ConsumerHandler struct {
	Hub      *Hub
	GameName string
	GetState func(token string) (any, error)
	Process  func(token string, method string, params map[string]any) (state any, intercoms []string, err error)
}

func (ch *ConsumerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.ToUpper(strings.TrimSpace(r.PathValue("token")))
	roomName := ch.GameName + "_" + token

	state, err := ch.GetState(token)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := ch.Hub.Register(roomName, conn)
	defer ch.Hub.Unregister(roomName, client)

	stateJSON, _ := json.Marshal(map[string]any{"type": "game", "message": state})
	client.Send(stateJSON)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var data struct {
			Method  string         `json:"method"`
			Params  map[string]any `json:"params"`
			Message any            `json:"message"`
		}
		if err := json.Unmarshal(msg, &data); err != nil {
			errJSON, _ := json.Marshal(map[string]any{"type": "error", "message": "invalid JSON"})
			client.Send(errJSON)
			continue
		}

		if data.Method == "intercom" {
			intercomJSON, _ := json.Marshal(map[string]any{"type": "intercom", "message": data.Message})
			ch.Hub.Broadcast(roomName, intercomJSON)
			continue
		}

		newState, intercoms, err := ch.Process(token, data.Method, data.Params)
		if err != nil {
			if _, ok := err.(*NothingToDoError); ok {
				continue
			}
			errJSON, _ := json.Marshal(map[string]any{"type": "error", "message": err.Error()})
			client.Send(errJSON)
			log.Printf("Bad request: %v", err)
			continue
		}

		for _, intercom := range intercoms {
			intercomJSON, _ := json.Marshal(map[string]any{"type": "intercom", "message": intercom})
			ch.Hub.Broadcast(roomName, intercomJSON)
		}

		newStateJSON, _ := json.Marshal(map[string]any{"type": "game", "message": newState})
		ch.Hub.Broadcast(roomName, newStateJSON)
	}
}
