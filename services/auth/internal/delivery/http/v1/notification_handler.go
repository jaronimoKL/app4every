package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	auth_http "app4every/services/auth/internal/delivery/http"
	"app4every/services/auth/internal/hub"
	"app4every/services/auth/internal/service"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type NotificationHandler struct {
	svc service.NotificationService
	hub *hub.NotificationHub
}

func NewNotificationHandler(svc service.NotificationService, h *hub.NotificationHub) *NotificationHandler {
	return &NotificationHandler{
		svc: svc,
		hub: h,
	}
}

// GET /api/v1/auth/notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth_http.UserIDKey).(int64)
	
	notifs, err := h.svc.GetUserNotifications(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifs)
}

// POST /api/v1/auth/notifications/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth_http.UserIDKey).(int64)

	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), userID, req.IDs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/v1/auth/notifications/{id}
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth_http.UserIDKey).(int64)
	
	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteNotification(r.Context(), userID, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /api/v1/auth/ws/notifications
func (h *NotificationHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth_http.UserIDKey).(int64)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := &hub.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)

	// Keep connection alive until closed
	go client.WritePump()
	go func() {
		defer func() {
			h.hub.Unregister(client)
			client.Conn.Close()
		}()
		for {
			_, _, err := client.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
		}
	}()
}

// POST /internal/notifications
func (h *NotificationHandler) InternalSendNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   int64                  `json:"user_id"`
		Type     string                 `json:"type"`
		Message  string                 `json:"message"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.SendNotification(r.Context(), req.UserID, req.Type, req.Message, req.Metadata); err != nil {
		log.Printf("InternalSendNotification error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
