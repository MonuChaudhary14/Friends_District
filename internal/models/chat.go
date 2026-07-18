package models

import (
	"time"
)

type ChatRoom struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Members  []ChatRoomMember `gorm:"foreignKey:RoomID" json:"-"`
	Messages []Message        `gorm:"foreignKey:RoomID" json:"-"`
}

type ChatRoomMember struct {
	ID       uint      `gorm:"primarykey" json:"id"`
	RoomID   uint      `gorm:"not null;index" json:"room_id"`
	UserID   uint      `gorm:"not null;index" json:"user_id"`
	Status   string    `gorm:"default:'joined'" json:"status"` // "pending" or "joined"
	JoinedAt time.Time `json:"joined_at"`

	Room ChatRoom `gorm:"foreignKey:RoomID" json:"-"`
	User User     `gorm:"foreignKey:UserID" json:"-"`
}

type Message struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	RoomID            uint      `gorm:"not null;index" json:"room_id"`
	SenderID          uint      `gorm:"not null;index" json:"sender_id"`
	Content           string    `gorm:"type:text;not null" json:"content"`
	ExternalEventID   string    `json:"external_event_id,omitempty"`   // String ID from external API (TMDB/Ticketmaster)
	ExternalEventType string    `json:"external_event_type,omitempty"` // "movie" or "concert"
	CreatedAt         time.Time `json:"created_at"`

	Room   ChatRoom `gorm:"foreignKey:RoomID" json:"-"`
	Sender User     `gorm:"foreignKey:SenderID" json:"-"`
}
