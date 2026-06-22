package hub

import (
	"log"
	"sync"
	"time"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub() *Hub {
	h := &Hub{
		rooms: make(map[string]*Room),
	}
	go h.startCleanupTicker()
	return h
}

func (h *Hub) GetOrCreate(roomID string, ownerID int64) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[roomID]
	if !ok {
		room = NewRoom(roomID, ownerID)
		h.rooms[roomID] = room
		log.Printf("Room created: %s, Owner: %d", roomID, ownerID)
	}
	return room
}

func (h *Hub) Get(roomID string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.rooms[roomID]
	return room, ok
}

func (h *Hub) Delete(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.rooms, roomID)
	log.Printf("Room deleted: %s", roomID)
}

func (h *Hub) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for id, room := range h.rooms {
		if room.IsEmpty() {
			delete(h.rooms, id)
			log.Printf("Room cleaned up (empty): %s", id)
		}
	}
}

func (h *Hub) startCleanupTicker() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		h.Cleanup()
	}
}
