package models

import (
	"time"
)

type FriendshipStatus string

const (
	StatusPending  FriendshipStatus = "pending"
	StatusAccepted FriendshipStatus = "accepted"
	StatusDeclined FriendshipStatus = "declined"
)

type Friendship struct {
	ID        uint             `gorm:"primarykey" json:"id"`
	UserID    uint             `gorm:"not null" json:"user_id"`
	FriendID  uint             `gorm:"not null" json:"friend_id"`
	Status    FriendshipStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	User   User `gorm:"foreignKey:UserID" json:"-"`
	Friend User `gorm:"foreignKey:FriendID" json:"-"`
}
