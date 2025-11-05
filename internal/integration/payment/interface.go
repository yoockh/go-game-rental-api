package integration

import "context"

// PaymentGateway interface for payment service abstraction
type PaymentGateway interface {
	CreateCharge(ctx context.Context, orderID string, grossAmount int64, paymentType string, params map[string]interface{}) (string, string, error)
	VerifyNotification(orderID, statusCode, grossAmount, signatureKey string) bool
}
