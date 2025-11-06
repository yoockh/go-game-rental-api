package model

import (
	"time"
)

type EmailVerificationToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	TokenHash string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}
