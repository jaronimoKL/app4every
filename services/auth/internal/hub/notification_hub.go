package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type NotificationMessage struct {
	Type     string                 `json:"type"`     // "friend_request", "group_invite", "watchparty_room_created", etc.
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

type NotificationHub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]bool
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[int64]map[*Client]bool),
	}
}

func (h *NotificationHub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true
	log.Printf("[Notifications] User %d connected", client.UserID)
}

func (h *NotificationHub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[client.UserID]; ok {
		if _, ok := conns[client]; ok {
			delete(conns, client)
			close(client.Send)
			if len(conns) == 0 {
				delete(h.clients, client.UserID)
			}
		}
	}
	log.Printf("[Notifications] User %d disconnected", client.UserID)
}

func (h *NotificationHub) SendToUser(userID int64, msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[userID]
	if !ok || len(conns) == 0 {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Notifications] Error marshaling message: %v", err)
		return
	}

	for client := range conns {
		select {
		case client.Send <- payload:
		default:
			// channel is full or blocked
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
