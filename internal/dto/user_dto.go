package dto

import "github.com/yoockh/go-game-rental-api/internal/model"

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required,min=2"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=10"`
	Address  string `json:"address,omitempty"`
}

type UpdateUserRoleRequest struct {
	Role model.UserRole `json:"role" validate:"required,oneof=customer partner admin"`
}
