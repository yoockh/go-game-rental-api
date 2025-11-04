package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	BookingID uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"booking_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	GameID    uuid.UUID      `gorm:"type:uuid;not null" json:"game_id"`
	Rating    int            `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating" validate:"required,min=1,max=5"`
	Comment   *string        `json:"comment,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Booking Booking `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Game    Game    `gorm:"foreignKey:GameID" json:"game,omitempty"`
}

func (Review) TableName() string {
	return "reviews"
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
