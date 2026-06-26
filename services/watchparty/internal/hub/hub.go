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
		room.mu.RLock()
		isEmpty := len(room.Clients) == 0 && len(room.Pending) == 0
		emptySince := room.EmptySince
		room.mu.RUnlock()

		if isEmpty && emptySince != nil && time.Since(*emptySince) > 10*time.Minute {
			delete(h.rooms, id)
			log.Printf("Room cleaned up (empty > 10m): %s", id)
		}
	}
}

func (h *Hub) GetRoomsForUsers(userMap map[int64]bool) []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var activeRooms []map[string]interface{}

	for _, room := range h.rooms {
		room.mu.RLock()
		
		hasMatch := false
		if userMap[room.OwnerID] {
			hasMatch = true
		} else {
			for clientID := range room.Clients {
				if userMap[clientID] {
					hasMatch = true
					break
				}
			}
		}

		if hasMatch && (len(room.Clients) > 0 || len(room.Pending) > 0) {
			state := map[string]interface{}{
				"room_id":          room.ID,
				"video_url":        room.VideoURL,
				"video_type":       room.VideoType,
				"shikimori_id":     room.ShikimoriID,
				"aniliberty_alias": room.AnilibertyAlias,
				"is_playing":       room.IsPlaying,
				"owner_id":         room.OwnerID,
				"participants":     room.Participants(),
			}
			activeRooms = append(activeRooms, state)
		}
		
		room.mu.RUnlock()
	}

	return activeRooms
}

func (h *Hub) startCleanupTicker() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		h.Cleanup()
	}
}
