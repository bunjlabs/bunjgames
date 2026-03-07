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
		// Message sent successfully
	default:
		// Channel full - drop message to avoid blocking                                                                             │
		log.Printf("Warning: Dropped message for slow client (buffer full)")
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

type wsMessage struct {
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Message any            `json:"message"`
}

func (ch *ConsumerHandler) handleConnection(w http.ResponseWriter, r *http.Request, token string) (*Client, string, any, error) {
	roomName := ch.GameName + "_" + token

	state, err := ch.GetState(token)
	if err != nil {
		return nil, "", nil, err
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, "", nil, err
	}

	client := ch.Hub.Register(roomName, conn)
	return client, roomName, state, nil
}

func (ch *ConsumerHandler) sendInitialState(client *Client, state any) {
	stateJSON, _ := json.Marshal(map[string]any{"type": "game", "message": state})
	client.Send(stateJSON)
}

func (ch *ConsumerHandler) handleIntercom(client *Client, roomName string, message any) {
	intercomJSON, _ := json.Marshal(map[string]any{"type": "intercom", "message": message})
	ch.Hub.Broadcast(roomName, intercomJSON)
}

func (ch *ConsumerHandler) handleGameCommand(client *Client, roomName, token string, data wsMessage) {
	newState, intercoms, err := ch.Process(token, data.Method, data.Params)
	if err != nil {
		if _, ok := err.(*NothingToDoError); ok {
			return
		}
		ch.sendError(client, err)
		return
	}

	ch.broadcastUpdate(roomName, newState, intercoms)
}

func (ch *ConsumerHandler) sendError(client *Client, err error) {
	errJSON, _ := json.Marshal(map[string]any{"type": "error", "message": err.Error()})
	client.Send(errJSON)
	log.Printf("Bad request: %v", err)
}

func (ch *ConsumerHandler) broadcastUpdate(roomName string, state any, intercoms []string) {
	for _, intercom := range intercoms {
		intercomJSON, _ := json.Marshal(map[string]any{"type": "intercom", "message": intercom})
		ch.Hub.Broadcast(roomName, intercomJSON)
	}

	stateJSON, _ := json.Marshal(map[string]any{"type": "game", "message": state})
	ch.Hub.Broadcast(roomName, stateJSON)
}

func (ch *ConsumerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.ToUpper(strings.TrimSpace(r.PathValue("token")))

	client, roomName, state, err := ch.handleConnection(w, r, token)
	if err != nil {
		if client == nil {
			http.Error(w, "Game not found", http.StatusNotFound)
		} else {
			log.Printf("WebSocket upgrade error: %v", err)
		}
		return
	}
	defer ch.Hub.Unregister(roomName, client)

	ch.sendInitialState(client, state)

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		var data wsMessage
		if err := json.Unmarshal(msg, &data); err != nil {
			ch.sendError(client, &BadFormatError{Msg: "invalid JSON"})
			continue
		}

		if data.Method == "intercom" {
			ch.handleIntercom(client, roomName, data.Message)
			continue
		}

		ch.handleGameCommand(client, roomName, token, data)
	}
}
