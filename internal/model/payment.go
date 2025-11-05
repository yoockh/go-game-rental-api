package model

import (
	"time"
)

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentPaid     PaymentStatus = "paid"
	PaymentFailed   PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"
)

type PaymentProvider string

const (
	ProviderStripe   PaymentProvider = "stripe"
	ProviderMidtrans PaymentProvider = "midtrans"
)

type Payment struct {
	ID                uint            `gorm:"primarykey" json:"id"`
	BookingID         uint            `gorm:"not null" json:"booking_id"`
	Provider          PaymentProvider `gorm:"type:payment_provider;not null" json:"provider"`
	ProviderPaymentID *string         `json:"provider_payment_id,omitempty"`
	Amount            float64         `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status            PaymentStatus   `gorm:"type:payment_status;default:pending" json:"status"`
	PaymentMethod     *string         `json:"payment_method,omitempty"`
	PaidAt            *time.Time      `json:"paid_at,omitempty"`
	FailedAt          *time.Time      `json:"failed_at,omitempty"`
	FailureReason     *string         `json:"failure_reason,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`

	// Relationships
	Booking Booking `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}
