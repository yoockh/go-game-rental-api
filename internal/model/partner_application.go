package model

import (
	"time"
)

type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationApproved ApplicationStatus = "approved"
	ApplicationRejected ApplicationStatus = "rejected"
)

type PartnerApplication struct {
	ID                  uint              `gorm:"primarykey" json:"id"`
	UserID              uint              `gorm:"not null" json:"user_id"`
	BusinessName        string            `gorm:"not null" json:"business_name" validate:"required"`
	BusinessAddress     string            `gorm:"not null" json:"business_address" validate:"required"`
	BusinessPhone       *string           `json:"business_phone,omitempty"`
	BusinessDescription *string           `json:"business_description,omitempty"`
	Status              ApplicationStatus `gorm:"type:application_status;default:pending" json:"status"`
	RejectionReason     *string           `json:"rejection_reason,omitempty"`
	SubmittedAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"submitted_at"`
	DecidedAt           *time.Time        `json:"decided_at,omitempty"`
	DecidedBy           *uint             `json:"decided_by,omitempty"`

	// Relationships
	User    User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Decider *User `gorm:"foreignKey:DecidedBy" json:"decider,omitempty"`
}

func (PartnerApplication) TableName() string {
	return "partner_applications"
}
