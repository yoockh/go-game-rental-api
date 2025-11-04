package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DisputeStatus string

const (
	DisputeOpen          DisputeStatus = "open"
	DisputeInvestigating DisputeStatus = "investigating"
	DisputeResolved      DisputeStatus = "resolved"
	DisputeClosed        DisputeStatus = "closed"
)

type DisputeType string

const (
	DisputePayment       DisputeType = "payment"
	DisputeItemCondition DisputeType = "item_condition"
	DisputeLateReturn    DisputeType = "late_return"
	DisputeNoShow        DisputeType = "no_show"
	DisputeOther         DisputeType = "other"
)

type Dispute struct {
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	BookingID   uuid.UUID     `gorm:"type:uuid;not null" json:"booking_id"`
	ReporterID  uuid.UUID     `gorm:"type:uuid;not null" json:"reporter_id"`
	Type        DisputeType   `gorm:"type:dispute_type;not null" json:"type" validate:"required"`
	Title       string        `gorm:"not null" json:"title" validate:"required"`
	Description string        `gorm:"not null" json:"description" validate:"required"`
	Status      DisputeStatus `gorm:"type:dispute_status;default:open" json:"status"`
	Resolution  *string       `json:"resolution,omitempty"`
	ResolvedBy  *uuid.UUID    `gorm:"type:uuid" json:"resolved_by,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`

	// Relationships
	Booking  Booking `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	Reporter User    `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Resolver *User   `gorm:"foreignKey:ResolvedBy" json:"resolver,omitempty"`
}

func (Dispute) TableName() string {
	return "disputes"
}

func (d *Dispute) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
