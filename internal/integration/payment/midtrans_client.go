package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	midtrans "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/sirupsen/logrus"
)

type MidtransClient struct {
	core      *coreapi.Client
	serverKey string
}



func NewMidtransClient() (*MidtransClient, error) {
	key := os.Getenv("MIDTRANS_SERVER_KEY")
	env := os.Getenv("MIDTRANS_ENV")

	if key == "" {
		return nil, fmt.Errorf("midtrans not configured: missing SERVER_KEY")
	}

	c := coreapi.Client{}
	if env == "production" {
		c.New(key, midtrans.Production)
	} else {
		c.New(key, midtrans.Sandbox)
	}

	return &MidtransClient{
		core:      &c,
		serverKey: key,
	}, nil
}

// CreateCharge creates payment charge. Note: ctx is accepted but Midtrans SDK ignores it.
func (m *MidtransClient) CreateCharge(ctx context.Context, orderID string, grossAmount int64, paymentType string, params map[string]interface{}) (string, string, error) {
	_ = ctx // ctx unused - Midtrans SDK doesn't support context
	// Log unknown payment types but allow them
	knownTypes := map[string]bool{
		"credit_card": true, "bank_transfer": true, "echannel": true,
		"gopay": true, "shopeepay": true, "qris": true,
	}
	if !knownTypes[paymentType] {
		logrus.WithField("payment_type", paymentType).Warn("Unknown payment type, proceeding anyway")
	}

	req := &coreapi.ChargeReq{
		PaymentType: coreapi.CoreapiPaymentType(paymentType),
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: grossAmount,
		},
	}

	resp, err := m.core.ChargeTransaction(req)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Midtrans charge failed")
		return "", "", fmt.Errorf("payment gateway error: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"order_id":           orderID,
		"transaction_id":     resp.TransactionID,
		"transaction_status": resp.TransactionStatus,
		"fraud_status":       resp.FraudStatus,
	}).Info("Midtrans charge created")

	var redirect string
	if resp.RedirectURL != "" {
		redirect = resp.RedirectURL
	}

	return resp.TransactionID, redirect, nil
}

// MapStatusToInternal maps Midtrans status to internal status
func MapStatusToInternal(midtransStatus string) string {
	switch midtransStatus {
	case "capture", "settlement":
		return "paid"
	case "pending":
		return "pending"
	case "deny", "cancel", "expire", "failure":
		return "failed"
	default:
		return midtransStatus
	}
}

func (m *MidtransClient) GetStatus(ctx context.Context, transactionID string) (string, error) {
	_ = ctx // ctx unused - Midtrans SDK doesn't support context
	resp, err := m.core.CheckTransaction(transactionID)
	if err != nil {
		logrus.WithError(err).WithField("transaction_id", transactionID).Error("Midtrans status check failed")
		return "", fmt.Errorf("failed to check payment status: %w", err)
	}
	return resp.TransactionStatus, nil
}

// VerifyNotification: signature = sha512(order_id + status_code + gross_amount + server_key)
func (m *MidtransClient) VerifyNotification(orderID, statusCode, grossAmount, signatureKey string) bool {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + m.serverKey))
	expected := hex.EncodeToString(sum[:])
	return strings.EqualFold(expected, signatureKey)
}
