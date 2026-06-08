package hub

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB (large enough for SDP offers/answers)
)

type Client struct {
	UserID   int64           `json:"user_id"`
	Username string          `json:"username"`
	RoomID   string          `json:"room_id"`
	Conn     *websocket.Conn `json:"-"`
	Send     chan []byte     `json:"-"`
	Hub      *Hub            `json:"-"`
}

func NewClient(userID int64, username string, roomID string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID:   userID,
		Username: username,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}
}

// ReadPump reads messages from the websocket connection and dispatches them.
func (c *Client) ReadPump(onMessage func(msg []byte), onDisconnect func()) {
	defer func() {
		onDisconnect()
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS client read error (user %d): %v", c.UserID, err)
			}
			break
		}
		onMessage(message)
	}
}

// WritePump writes messages from the send channel to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
