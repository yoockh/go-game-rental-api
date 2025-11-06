package dto

import (
	"time"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=10"`
	Address  string `json:"address,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        *model.User `json:"user"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}


