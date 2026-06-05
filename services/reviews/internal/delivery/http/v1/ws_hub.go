package v1

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	UserID  int64
	GroupID int64
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[c.GroupID] == nil {
		h.rooms[c.GroupID] = make(map[*Client]bool)
	}
	h.rooms[c.GroupID][c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[c.GroupID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, c.GroupID)
		}
	}
	_ = c.Conn.Close()
}

func (h *Hub) Broadcast(groupID int64, event string, data any) {
	h.mu.RLock()
	clients := h.rooms[groupID]
	if len(clients) == 0 {
		h.mu.RUnlock()
		return
	}

	// Копируем список клиентов во избежание дедлоков/блокировок во время сетевой записи
	clientList := make([]*Client, 0, len(clients))
	for c := range clients {
		clientList = append(clientList, c)
	}
	h.mu.RUnlock()

	msg := map[string]any{
		"event": event,
		"data":  data,
	}

	for _, c := range clientList {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			h.Unregister(c)
		}
	}
}
