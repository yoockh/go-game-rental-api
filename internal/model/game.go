package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending_approval"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type Game struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	PartnerID         uuid.UUID      `gorm:"type:uuid;not null" json:"partner_id"`
	CategoryID        uuid.UUID      `gorm:"type:uuid;not null" json:"category_id"`
	Name              string         `gorm:"not null" json:"name" validate:"required"`
	Description       *string        `json:"description,omitempty"`
	Platform          *string        `json:"platform,omitempty"`
	Stock             int            `gorm:"default:1" json:"stock" validate:"min=1"`
	AvailableStock    int            `gorm:"default:1" json:"available_stock"`
	RentalPricePerDay float64        `gorm:"type:decimal(10,2);not null" json:"rental_price_per_day" validate:"required,gt=0"`
	SecurityDeposit   float64        `gorm:"type:decimal(10,2);default:0" json:"security_deposit"`
	Condition         string         `gorm:"default:excellent" json:"condition"`
	Images            pq.StringArray `gorm:"type:text[]" json:"images,omitempty"`
	IsActive          bool           `gorm:"default:false" json:"is_active"`
	ApprovalStatus    ApprovalStatus `gorm:"type:approval_status;default:pending_approval" json:"approval_status"`
	ApprovedBy        *uuid.UUID     `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt        *time.Time     `json:"approved_at,omitempty"`
	RejectionReason   *string        `json:"rejection_reason,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Partner  User      `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
	Category Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Approver *User     `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Bookings []Booking `gorm:"foreignKey:GameID" json:"-"`
	Reviews  []Review  `gorm:"foreignKey:GameID" json:"-"`
}

func (Game) TableName() string {
	return "games"
}

func (g *Game) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}
