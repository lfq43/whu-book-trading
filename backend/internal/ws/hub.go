package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uint]*Client
}

var DefaultHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	if existing, ok := h.clients[client.UserID]; ok {
		existing.Conn.Close()
		close(existing.Send)
	}
	h.clients[client.UserID] = client
	h.mu.Unlock()
	go h.writePump(client)
}

func (h *Hub) Unregister(userID uint) {
	h.mu.Lock()
	client, ok := h.clients[userID]
	if ok {
		delete(h.clients, userID)
		close(client.Send)
		client.Conn.Close()
	}
	h.mu.Unlock()
}

func (h *Hub) SendToUser(userID uint, payload interface{}) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok || client == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
		h.Unregister(userID)
	}
}

func (h *Hub) writePump(client *Client) {
	defer h.Unregister(client.UserID)
	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
