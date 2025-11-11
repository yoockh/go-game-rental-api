package dto

import "github.com/yoockh/go-game-rental-api/internal/model"

type CreatePaymentRequest struct {
	Provider    model.PaymentProvider `json:"provider" validate:"required,oneof=stripe midtrans"`
	PaymentType string                `json:"payment_type,omitempty"`
}

type PaymentWebhookRequest struct {
	ProviderPaymentID string  `json:"provider_payment_id" validate:"required"`
	Status            string  `json:"status" validate:"required"`
	PaymentMethod     *string `json:"payment_method,omitempty"`
	FailureReason     *string `json:"failure_reason,omitempty"`
}
