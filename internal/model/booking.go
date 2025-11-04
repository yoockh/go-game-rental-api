package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingStatus string

const (
	BookingPendingPayment BookingStatus = "pending_payment"
	BookingConfirmed      BookingStatus = "confirmed"
	BookingActive         BookingStatus = "active"
	BookingCompleted      BookingStatus = "completed"
	BookingCancelled      BookingStatus = "cancelled"
	BookingDisputed       BookingStatus = "disputed"
)

type Booking struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID              uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	GameID              uuid.UUID      `gorm:"type:uuid;not null" json:"game_id"`
	PartnerID           uuid.UUID      `gorm:"type:uuid;not null" json:"partner_id"`
	StartDate           time.Time      `gorm:"type:date;not null" json:"start_date" validate:"required"`
	EndDate             time.Time      `gorm:"type:date;not null" json:"end_date" validate:"required"`
	RentalDays          int            `gorm:"not null" json:"rental_days"`
	DailyPrice          float64        `gorm:"type:decimal(10,2);not null" json:"daily_price"`
	TotalRentalPrice    float64        `gorm:"type:decimal(10,2);not null" json:"total_rental_price"`
	SecurityDeposit     float64        `gorm:"type:decimal(10,2);default:0" json:"security_deposit"`
	TotalAmount         float64        `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status              BookingStatus  `gorm:"type:booking_status;default:pending_payment" json:"status"`
	Notes               *string        `json:"notes,omitempty"`
	HandoverConfirmedAt *time.Time     `json:"handover_confirmed_at,omitempty"`
	ReturnConfirmedAt   *time.Time     `json:"return_confirmed_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Game     Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Partner  User      `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
	Payment  *Payment  `gorm:"foreignKey:BookingID" json:"payment,omitempty"`
	Review   *Review   `gorm:"foreignKey:BookingID" json:"review,omitempty"`
	Disputes []Dispute `gorm:"foreignKey:BookingID" json:"-"`
}

func (Booking) TableName() string {
	return "bookings"
}

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
