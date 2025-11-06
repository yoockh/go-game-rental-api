package model

import (
	"time"
)

type UserRole string

const (
	RoleCustomer   UserRole = "customer"
	RolePartner    UserRole = "partner"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	Password  string         `gorm:"not null" json:"-"`
	FullName  string         `gorm:"not null" json:"full_name" validate:"required"`
	Phone     *string        `json:"phone,omitempty"`
	Address   *string        `json:"address,omitempty"`
	Role      UserRole       `gorm:"type:user_role;default:customer" json:"role"`
	IsActive  bool           `gorm:"default:false" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	PartnerApplications []PartnerApplication `gorm:"foreignKey:UserID" json:"-"`
	Games               []Game               `gorm:"foreignKey:PartnerID" json:"-"`
	Bookings            []Booking            `gorm:"foreignKey:UserID" json:"-"`
	Reviews             []Review             `gorm:"foreignKey:UserID" json:"-"`

	EmailVerificationTokens []EmailVerificationToken `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}
