package model

import (
	"time"
)

type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BookingID uint      `gorm:"not null" json:"booking_id"`
	Booking   *Booking  `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	GameID    uint      `gorm:"not null" json:"game_id"`
	Game      *Game     `gorm:"foreignKey:GameID" json:"-"` // Omit dari JSON
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   *string   `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Review) TableName() string {
	return "reviews"
}
