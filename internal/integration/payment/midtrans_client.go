package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	midtrans "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type MidtransClient struct {
	core      *coreapi.Client
	serverKey string
}

var validPaymentTypes = map[string]bool{
	"credit_card":   true,
	"bank_transfer": true,
	"echannel":      true,
	"gopay":         true,
	"shopeepay":     true,
	"qris":          true,
}

func NewMidtransClient() *MidtransClient {
	key := os.Getenv("MIDTRANS_SERVER_KEY")
	env := os.Getenv("MIDTRANS_ENV")

	if key == "" {
		log.Println("WARN: MIDTRANS_SERVER_KEY not set")
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
	}
}

// CreateCharge creates payment charge. Note: ctx is accepted but Midtrans SDK ignores it.
// Consider implementing timeout/rate-limit before calling SDK if needed.
func (m *MidtransClient) CreateCharge(ctx context.Context, orderID string, grossAmount int64, paymentType string, params map[string]interface{}) (string, string, error) {
	if m.serverKey == "" {
		return "", "", fmt.Errorf("midtrans not configured")
	}
	if !validPaymentTypes[paymentType] {
		return "", "", fmt.Errorf("invalid payment type: %s", paymentType)
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
		log.Printf("ERROR: Midtrans charge failed for order %s: %v", orderID, err)
		return "", "", fmt.Errorf("payment gateway error: %w", err)
	}

	log.Printf("INFO: Midtrans charge created: order=%s tx=%s status=%s fraud=%s",
		orderID, resp.TransactionID, resp.TransactionStatus, resp.FraudStatus)

	var redirect string
	if resp.RedirectURL != "" {
		redirect = resp.RedirectURL
	}

	return resp.TransactionID, redirect, nil
}

func (m *MidtransClient) GetStatus(ctx context.Context, transactionID string) (*coreapi.TransactionStatusResponse, error) {
	if m.serverKey == "" {
		return nil, fmt.Errorf("midtrans not configured")
	}
	resp, err := m.core.CheckTransaction(transactionID)
	if err != nil {
		log.Printf("ERROR: Midtrans status check failed for tx %s: %v", transactionID, err)
		return nil, fmt.Errorf("failed to check payment status: %w", err)
	}
	return resp, nil
}

// VerifyNotification: signature = sha512(order_id + status_code + gross_amount + server_key)
func (m *MidtransClient) VerifyNotification(orderID, statusCode, grossAmount, signatureKey string) bool {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + m.serverKey))
	expected := hex.EncodeToString(sum[:])
	return strings.EqualFold(expected, signatureKey)
}
