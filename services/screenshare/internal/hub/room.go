package hub

import (
	"encoding/json"
	"errors"
	"sync"
)

type Room struct {
	ID      string
	mu      sync.RWMutex
	Clients map[int64]*Client // userID -> Client
}

type ParticipantInfo struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type Message struct {
	Type         string            `json:"type"`
	RoomID       string            `json:"room_id,omitempty"`
	YourID       int64             `json:"your_id,omitempty"`
	FromID       int64             `json:"from_id,omitempty"`
	TargetID     int64             `json:"target_id,omitempty"`
	UserID       int64             `json:"user_id,omitempty"`
	Username     string            `json:"username,omitempty"`
	Code         string            `json:"code,omitempty"`
	Message      string            `json:"message,omitempty"`
	Sdp          string            `json:"sdp,omitempty"`
	Candidate    any               `json:"candidate,omitempty"`
	Participants []ParticipantInfo `json:"participants,omitempty"`
}

func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		Clients: make(map[int64]*Client),
	}
}

func (r *Room) Join(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Limit room capacity to 8 participants
	if len(r.Clients) >= 8 {
		return errors.New("room_full")
	}

	r.Clients[c.UserID] = c
	return nil
}

func (r *Room) Leave(userID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, userID)
}

func (r *Room) Broadcast(msg Message, exceptUserID int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for id, client := range r.Clients {
		if id != exceptUserID {
			select {
			case client.Send <- payload:
			default:
				// Avoid blocking if a client's send queue is full
			}
		}
	}
}

func (r *Room) SendTo(targetUserID int64, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.Clients[targetUserID]
	if !ok {
		return errors.New("client_not_found")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case client.Send <- payload:
		return nil
	default:
		return errors.New("send_channel_full")
	}
}

func (r *Room) Participants() []ParticipantInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]ParticipantInfo, 0, len(r.Clients))
	for _, client := range r.Clients {
		list = append(list, ParticipantInfo{
			UserID:   client.UserID,
			Username: client.Username,
		})
	}
	return list
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Clients) == 0
}
