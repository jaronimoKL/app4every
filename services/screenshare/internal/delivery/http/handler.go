package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"app4every/services/screenshare/internal/hub"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for signaling handshakes
	},
}

type ScreenshareHandler struct {
	hub *hub.Hub
}

func NewScreenshareHandler(h *hub.Hub) *ScreenshareHandler {
	return &ScreenshareHandler{
		hub: h,
	}
}

func (h *ScreenshareHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad_request","message":"room_id is required"}`))
		return
	}

	userID, ok1 := r.Context().Value(UserIDKey).(int64)
	username, ok2 := r.Context().Value(UsernameKey).(string)
	if !ok1 || !ok2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}

	client := hub.NewClient(userID, username, roomID, conn, h.hub)
	room := h.hub.GetOrCreate(roomID)

	if err := room.Join(client); err != nil {
		payload, _ := json.Marshal(hub.Message{
			Type:    "error",
			Code:    err.Error(),
			Message: "Room is full (max 8 participants)",
		})
		_ = conn.WriteMessage(websocket.TextMessage, payload)
		conn.Close()
		return
	}

	go client.WritePump()

	// Welcome user, send details of existing room participants
	participants := room.Participants()
	joinedPayload := hub.Message{
		Type:         "joined",
		RoomID:       roomID,
		YourID:       userID,
		Participants: participants,
	}
	client.Send <- serializeMessage(joinedPayload)

	// Broadcast that a new user has joined to existing participants
	room.Broadcast(hub.Message{
		Type:     "user_joined",
		UserID:   userID,
		Username: username,
	}, userID)

	client.ReadPump(
		func(rawMsg []byte) {
			var msg hub.Message
			if err := json.Unmarshal(rawMsg, &msg); err != nil {
				return
			}

			switch msg.Type {
			case "offer", "answer", "ice_candidate":
				if msg.TargetID != 0 {
					msg.FromID = userID
					_ = room.SendTo(msg.TargetID, msg)
				}
			case "leave":
				conn.Close()
			case "ping":
				client.Send <- serializeMessage(hub.Message{Type: "pong"})
			}
		},
		func() {
			room.Leave(userID)
			room.Broadcast(hub.Message{
				Type:   "user_left",
				UserID: userID,
			}, userID)

			if room.IsEmpty() {
				h.hub.Delete(roomID)
			}
		},
	)
}

func (h *ScreenshareHandler) HandleRoomInfo(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/api/v1/screenshare/rooms/")
	parts := strings.Split(p, "/")
	if len(parts) == 0 || parts[0] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad_request","message":"invalid path"}`))
		return
	}
	roomID := parts[0]

	room := h.hub.GetOrCreate(roomID)
	participants := room.Participants()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"room_id":      roomID,
		"participants": participants,
	})
}

func serializeMessage(msg hub.Message) []byte {
	b, _ := json.Marshal(msg)
	return b
}
