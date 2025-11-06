package model

import (
	"time"
)

type Review struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	BookingID uint      `gorm:"uniqueIndex;not null" json:"booking_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	GameID    uint      `gorm:"not null" json:"game_id"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating" validate:"required,min=1,max=5"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Booking Booking `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Game    Game    `gorm:"foreignKey:GameID" json:"game,omitempty"`
}

func (Review) TableName() string {
	return "reviews"
}
