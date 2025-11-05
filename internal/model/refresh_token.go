package model

import (
	"time"
)

type RefreshToken struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	UserID    uint       `gorm:"not null" json:"user_id"`
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
