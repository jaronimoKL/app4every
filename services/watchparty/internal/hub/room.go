package hub

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Room struct {
	ID          string
	VideoURL    string
	VideoType   string
	IsPlaying   bool
	CurrentTime float64
	UpdatedAt   time.Time
	OwnerID     int64
	ShikimoriID string
	AnilibertyAlias string
	EmptySince  *time.Time

	mu      sync.RWMutex
	Clients map[int64]*Client
	Pending map[int64]*Client // users waiting for approval
}

type ParticipantInfo struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	IsOwner  bool   `json:"is_owner"`
}

type Message struct {
	Type        string            `json:"type"`
	RoomID      string            `json:"room_id,omitempty"`
	YourID      int64             `json:"your_id,omitempty"`
	FromID      int64             `json:"from_id,omitempty"`
	TargetID    int64             `json:"target_id,omitempty"`
	UserID      int64             `json:"user_id,omitempty"`
	Username    string            `json:"username,omitempty"`
	Code        string            `json:"code,omitempty"`
	Message     string            `json:"message,omitempty"`
	VideoURL    string            `json:"video_url,omitempty"`
	VideoType   string            `json:"video_type,omitempty"`
	ShikimoriID string            `json:"shikimori_id,omitempty"`
	AnilibertyAlias string        `json:"aniliberty_alias,omitempty"`
	IsPlaying   *bool             `json:"is_playing,omitempty"`
	CurrentTime *float64          `json:"current_time,omitempty"`
	IsOwner     *bool             `json:"is_owner,omitempty"`
	Participants []ParticipantInfo `json:"participants,omitempty"`
}

func NewRoom(id string, ownerID int64) *Room {
	now := time.Now()
	return &Room{
		ID:        id,
		OwnerID:   ownerID,
		UpdatedAt: now,
		EmptySince: &now,
		Clients:   make(map[int64]*Client),
		Pending:   make(map[int64]*Client),
	}
}

func (r *Room) EstimatedCurrentTime() float64 {
	if !r.IsPlaying {
		return r.CurrentTime
	}
	return r.CurrentTime + time.Since(r.UpdatedAt).Seconds()
}

func (r *Room) UpdateState(isPlaying bool, currentTime float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.IsPlaying = isPlaying
	r.CurrentTime = currentTime
	r.UpdatedAt = time.Now()
}

func (r *Room) ChangeVideo(url, videoType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.VideoURL = url
	r.VideoType = videoType
	r.IsPlaying = false
	r.CurrentTime = 0
	r.UpdatedAt = time.Now()
}

func (r *Room) Join(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Clients) >= 16 {
		return errors.New("room_full")
	}

	r.Clients[c.UserID] = c
	r.EmptySince = nil
	return nil
}

func (r *Room) Leave(userID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, userID)
	delete(r.Pending, userID)

	if len(r.Clients) == 0 && len(r.Pending) == 0 {
		now := time.Now()
		r.EmptySince = &now
	}
}

func (r *Room) AddPending(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Pending[c.UserID] = c
	r.EmptySince = nil
}

func (r *Room) PopPending(userID int64) (*Client, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.Pending[userID]
	if ok {
		delete(r.Pending, userID)
	}
	return c, ok
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
			}
		}
	}
}

func (r *Room) SendTo(targetUserID int64, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.Clients[targetUserID]
	if !ok {
		// Also check pending for direct messages (e.g., knock_rejected)
		client, ok = r.Pending[targetUserID]
		if !ok {
			return errors.New("client_not_found")
		}
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
			IsOwner:  client.UserID == r.OwnerID,
		})
	}
	return list
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Clients) == 0 && len(r.Pending) == 0
}
