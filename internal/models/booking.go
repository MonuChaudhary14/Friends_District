package models

import "time"

type Booking struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	UserID            uint      `gorm:"not null;index" json:"user_id"`
	BookedForID       *uint     `json:"booked_for_id,omitempty"` // ID of the friend this ticket was bought for (optional)
	ExternalEventID   string    `gorm:"not null" json:"external_event_id"`
	ExternalEventType string    `gorm:"not null" json:"external_event_type"` // e.g., "movie" or "concert"
	Quantity          int       `gorm:"not null;default:1" json:"quantity"`
	TotalPrice        float64   `gorm:"not null;default:0.0" json:"total_price"`
	Status            string    `gorm:"not null;default:'CONFIRMED'" json:"status"`
	StartTime         *string   `json:"start_time,omitempty"`
	EndTime           *string   `json:"end_time,omitempty"`
	CreatedAt         time.Time `json:"created_at"`

	User      User  `gorm:"foreignKey:UserID" json:"user"`
	BookedFor *User `gorm:"foreignKey:BookedForID" json:"booked_for,omitempty"`
}
