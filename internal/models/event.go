package models

import (
	"time"
)

type Event struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Type        string    `gorm:"not null" json:"type"` // e.g., Movie, Concert, Dining, Laughter Show
	Location    string    `json:"location"`
	Date        time.Time `json:"date"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
