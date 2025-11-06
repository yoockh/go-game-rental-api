package payment

import "context"

// MockPaymentGateway for testing
type MockPaymentGateway struct {
	Charges []MockCharge
}

type MockCharge struct {
	OrderID     string
	Amount      int64
	PaymentType string
	Params      map[string]interface{}
}

func (m *MockPaymentGateway) CreateCharge(ctx context.Context, orderID string, grossAmount int64, paymentType string, params map[string]interface{}) (string, string, error) {
	m.Charges = append(m.Charges, MockCharge{
		OrderID:     orderID,
		Amount:      grossAmount,
		PaymentType: paymentType,
		Params:      params,
	})
	return "mock-tx-" + orderID, "https://mock-payment.com/redirect", nil
}

func (m *MockPaymentGateway) GetStatus(ctx context.Context, transactionID string) (string, error) {
	return "paid", nil // Always paid for testing
}

func (m *MockPaymentGateway) VerifyNotification(orderID, statusCode, grossAmount, signatureKey string) bool {
	return true // Always valid for testing
}
