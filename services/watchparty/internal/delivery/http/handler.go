package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"app4every/services/watchparty/internal/config"
	"app4every/services/watchparty/internal/hub"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type Handler struct {
	Hub *hub.Hub
	Cfg *config.Config
}

func NewHandler(h *hub.Hub, cfg *config.Config) *Handler {
	return &Handler{
		Hub: h,
		Cfg: cfg,
	}
}

func (h *Handler) GetRoomState(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/api/v1/watchparty/rooms/")
	room, ok := h.Hub.Get(roomID)
	if !ok {
		http.Error(w, `{"error":"room_not_found"}`, http.StatusNotFound)
		return
	}

	state := map[string]interface{}{
		"room_id":      room.ID,
		"video_url":    room.VideoURL,
		"video_type":   room.VideoType,
		"is_playing":   room.IsPlaying,
		"current_time": room.EstimatedCurrentTime(),
		"participants": room.Participants(),
		"owner_id":     room.OwnerID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *Handler) GetActiveRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userMap := make(map[int64]bool)
	for _, id := range req.UserIDs {
		userMap[id] = true
	}

	activeRooms := h.Hub.GetRoomsForUsers(userMap)
	if activeRooms == nil {
		activeRooms = make([]map[string]interface{}, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeRooms)
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "token is required", http.StatusUnauthorized)
		return
	}

	claims, err := ValidateToken(tokenString, h.Cfg.JWTSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = claims.Username
	}

	client := hub.NewClient(claims.UserID, username, roomID, conn, h.Hub)
	room := h.Hub.GetOrCreate(roomID, claims.UserID)

	// Authorization logic
	isAuthorized := false
	if claims.UserID == room.OwnerID {
		isAuthorized = true
	} else if _, ok := room.Clients[claims.UserID]; ok {
		isAuthorized = true
	} else {
		// Check friends list via internal auth-service call
		isAuthorized = h.checkIfFriend(claims.UserID, room.OwnerID)
	}

	if !isAuthorized {
		room.AddPending(client)
		
		// Notify owner
		room.SendTo(room.OwnerID, hub.Message{
			Type:     "knock_request",
			UserID:   claims.UserID,
			Username: claims.Username,
		})

		// Send waiting msg
		client.Send <- []byte(`{"type":"error","code":"waiting_for_approval"}`)
	} else {
		err = room.Join(client)
		if err != nil {
			client.Send <- []byte(`{"type":"error","code":"room_full"}`)
			conn.Close()
			return
		}
		
		h.sendJoinedMessage(client, room)
		room.Broadcast(hub.Message{
			Type:     "user_joined",
			UserID:   client.UserID,
			Username: client.Username,
		}, client.UserID)
	}

	// Start pump goroutines
	go client.WritePump()
	go client.ReadPump(func(msg []byte) {
		h.handleClientMessage(client, room, msg)
	}, func() {
		// On disconnect
		room.Leave(client.UserID)
		room.Broadcast(hub.Message{
			Type:   "user_left",
			UserID: client.UserID,
		}, client.UserID)
	})
}

func (h *Handler) checkIfFriend(userID, ownerID int64) bool {
	url := fmt.Sprintf("%s/internal/users/%d/friends", h.Cfg.AuthServiceURL, ownerID)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	var friendIDs []int64
	if err := json.NewDecoder(resp.Body).Decode(&friendIDs); err != nil {
		return false
	}

	for _, id := range friendIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (h *Handler) sendJoinedMessage(client *hub.Client, room *hub.Room) {
	isPlaying := room.IsPlaying
	currentTime := room.EstimatedCurrentTime()
	isOwner := client.UserID == room.OwnerID

	msg := hub.Message{
		Type:         "joined",
		RoomID:       room.ID,
		YourID:       client.UserID,
		IsOwner:      &isOwner,
		VideoURL:     room.VideoURL,
		VideoType:    room.VideoType,
		ShikimoriID:  room.ShikimoriID,
		AnilibertyAlias: room.AnilibertyAlias,
		IsPlaying:    &isPlaying,
		CurrentTime:  &currentTime,
		Participants: room.Participants(),
	}

	b, _ := json.Marshal(msg)
	client.Send <- b
}

func (h *Handler) handleClientMessage(client *hub.Client, room *hub.Room, msg []byte) {
	var parsed hub.Message
	if err := json.Unmarshal(msg, &parsed); err != nil {
		return
	}

	switch parsed.Type {
	case "play", "pause", "seek":
		if parsed.CurrentTime != nil {
			isPlaying := parsed.Type == "play"
			if parsed.Type == "seek" {
				isPlaying = room.IsPlaying
			}
			room.UpdateState(isPlaying, *parsed.CurrentTime)
			
			// Broadcast sync
			syncMsg := hub.Message{
				Type:        parsed.Type,
				FromID:      client.UserID,
				CurrentTime: parsed.CurrentTime,
			}
			room.Broadcast(syncMsg, client.UserID)
		}

	case "change_video":
		if client.UserID == room.OwnerID && parsed.VideoURL != "" {
			videoType := parsed.VideoType
			if videoType == "" {
				videoType = "direct"
			}
			room.ChangeVideo(parsed.VideoURL, videoType)
			
			room.Broadcast(hub.Message{
				Type:      "video_changed",
				FromID:    client.UserID,
				VideoURL:  parsed.VideoURL,
				VideoType: videoType,
			}, 0) // broadcast to all
		}

	case "update_metadata":
		if client.UserID == room.OwnerID {
			room.UpdateMetadata(parsed.ShikimoriID, parsed.AnilibertyAlias)
			
			room.Broadcast(hub.Message{
				Type:            "metadata_updated",
				ShikimoriID:     room.ShikimoriID,
				AnilibertyAlias: room.AnilibertyAlias,
			}, 0)
		}

	case "admit":
		if client.UserID == room.OwnerID && parsed.TargetID != 0 {
			if pendingClient, ok := room.PopPending(parsed.TargetID); ok {
				if err := room.Join(pendingClient); err == nil {
					h.sendJoinedMessage(pendingClient, room)
					
					room.Broadcast(hub.Message{
						Type:     "user_joined",
						UserID:   pendingClient.UserID,
						Username: pendingClient.Username,
					}, pendingClient.UserID)
				}
			}
		}

	case "reject":
		if client.UserID == room.OwnerID && parsed.TargetID != 0 {
			if pendingClient, ok := room.PopPending(parsed.TargetID); ok {
				pendingClient.Send <- []byte(`{"type":"knock_rejected"}`)
				// wait a bit before close
				go func(c *hub.Client) {
					time.Sleep(1 * time.Second)
					c.Conn.Close()
				}(pendingClient)
			}
		}

	case "kick":
		if client.UserID == room.OwnerID && parsed.TargetID != 0 {
			if parsed.TargetID != room.OwnerID {
				room.SendTo(parsed.TargetID, hub.Message{Type: "kicked"})
				room.Leave(parsed.TargetID)
				
				room.Broadcast(hub.Message{
					Type:   "user_left",
					UserID: parsed.TargetID,
				}, 0)
			}
		}

	case "ping":
		client.Send <- []byte(`{"type":"pong"}`)
	}
}
