package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationApproved ApplicationStatus = "approved"
	ApplicationRejected ApplicationStatus = "rejected"
)

type PartnerApplication struct {
	ID                  uuid.UUID         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID              uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`
	BusinessName        string            `gorm:"not null" json:"business_name" validate:"required"`
	BusinessAddress     string            `gorm:"not null" json:"business_address" validate:"required"`
	BusinessPhone       *string           `json:"business_phone,omitempty"`
	BusinessDescription *string           `json:"business_description,omitempty"`
	Status              ApplicationStatus `gorm:"type:application_status;default:pending" json:"status"`
	RejectionReason     *string           `json:"rejection_reason,omitempty"`
	SubmittedAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"submitted_at"`
	DecidedAt           *time.Time        `json:"decided_at,omitempty"`
	DecidedBy           *uuid.UUID        `gorm:"type:uuid" json:"decided_by,omitempty"`

	// Relationships
	User    User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Decider *User `gorm:"foreignKey:DecidedBy" json:"decider,omitempty"`
}

func (PartnerApplication) TableName() string {
	return "partner_applications"
}

func (pa *PartnerApplication) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == uuid.Nil {
		pa.ID = uuid.New()
	}
	return nil
}
