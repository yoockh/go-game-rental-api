package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	TokenHash string     `gorm:"not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	IsRevoked bool       `gorm:"default:false" json:"is_revoked"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}
