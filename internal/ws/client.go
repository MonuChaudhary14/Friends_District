package ws

import (
	"log"
	"time"

	"district-friends/internal/models"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The user details
	User models.User

	// The room ID
	RoomID uint

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Database connection for saving messages
	DB *gorm.DB
}

func NewClient(hub *Hub, user models.User, roomID uint, conn *websocket.Conn, db *gorm.DB) *Client {
	return &Client{
		Hub:    hub,
		User:   user,
		RoomID: roomID,
		Conn:   conn,
		send:   make(chan []byte, 256),
		DB:     db,
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Save message to DB
		msg := models.Message{
			RoomID:    c.RoomID,
			SenderID:  c.User.ID,
			Content:   string(message),
			CreatedAt: time.Now(),
		}

		if err := c.DB.Create(&msg).Error; err != nil {
			log.Println("Database error:", err)
			continue
		}

		// Attach sender for JSON broadcast
		msg.Sender = c.User

		broadcastMsg := BroadcastMessage{
			RoomID:  c.RoomID,
			Message: msg,
		}

		c.Hub.Broadcast <- broadcastMsg
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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
