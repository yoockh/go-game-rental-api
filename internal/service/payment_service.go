package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Yoochan45/go-game-rental-api/internal/integration/payment"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists for this booking")
	ErrPaymentAlreadyPaid   = errors.New("payment already paid")
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
)

type PaymentService interface {
	// Customer methods
	CreatePayment(userID uint, bookingID uint, provider model.PaymentProvider, paymentType string) (*model.Payment, error)
	GetPaymentByBooking(userID uint, bookingID uint) (*model.Payment, error)

	// Admin methods
	GetAllPayments(requestorRole model.UserRole, limit, offset int) ([]*model.Payment, error)
	GetPaymentsByStatus(requestorRole model.UserRole, status model.PaymentStatus, limit, offset int) ([]*model.Payment, error)
	GetPaymentDetail(requestorRole model.UserRole, paymentID uint) (*model.Payment, error)

	// Webhook/System methods
	ProcessWebhook(providerPaymentID string, status string, paymentMethod string, failureReason *string) error
}

type paymentService struct {
	paymentRepo     repository.PaymentRepository
	bookingRepo     repository.BookingRepository
	userRepo        repository.UserRepository
	bookingService  BookingService
	paymentGateway  payment.PaymentGateway
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	userRepo repository.UserRepository,
	bookingService BookingService,
	paymentGateway payment.PaymentGateway,
) PaymentService {
	return &paymentService{
		paymentRepo:     paymentRepo,
		bookingRepo:     bookingRepo,
		userRepo:        userRepo,
		bookingService:  bookingService,
		paymentGateway:  paymentGateway,
	}
}

func (s *paymentService) CreatePayment(userID uint, bookingID uint, provider model.PaymentProvider, paymentType string) (*model.Payment, error) {
	// Get booking and validate ownership
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	if booking.UserID != userID {
		return nil, ErrBookingNotOwned
	}

	// Check if booking is in pending payment status
	if booking.Status != model.BookingPendingPayment {
		return nil, errors.New("booking is not in pending payment status")
	}

	// Check if payment already exists
	existingPayment, _ := s.paymentRepo.GetByBookingID(bookingID)
	if existingPayment != nil {
		return existingPayment, nil // Return existing payment
	}

	// Create payment record
	payment := &model.Payment{
		BookingID: bookingID,
		Provider:  provider,
		Amount:    booking.TotalAmount,
		Status:    model.PaymentPending,
	}

	err = s.paymentRepo.Create(payment)
	if err != nil {
		return nil, err
	}

	// Create charge with payment gateway
	orderID := fmt.Sprintf("booking-%d", bookingID)
	// Set default payment type if not provided
	if paymentType == "" {
		if provider == model.ProviderMidtrans {
			paymentType = "bank_transfer"
		} else {
			paymentType = "credit_card"
		}
	}

	txID, _, err := s.paymentGateway.CreateCharge(
		context.Background(),
		orderID,
		int64(payment.Amount), // Rupiah penuh sesuai API Midtrans
		paymentType,
		nil,
	)
	if err != nil {
		// Payment gateway failed, but payment record exists
		return payment, fmt.Errorf("payment gateway error: %w", err)
	}

	// Update payment with provider transaction ID
	payment.ProviderPaymentID = &txID
	s.paymentRepo.Update(payment)

	return payment, nil
}

func (s *paymentService) GetPaymentByBooking(userID uint, bookingID uint) (*model.Payment, error) {
	// Validate booking ownership
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	if booking.UserID != userID {
		return nil, ErrBookingNotOwned
	}

	return s.paymentRepo.GetByBookingID(bookingID)
}

func (s *paymentService) GetAllPayments(requestorRole model.UserRole, limit, offset int) ([]*model.Payment, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.paymentRepo.GetPaymentsByStatus(model.PaymentStatus(""), limit, offset) // Get all
}

func (s *paymentService) GetPaymentsByStatus(requestorRole model.UserRole, status model.PaymentStatus, limit, offset int) ([]*model.Payment, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.paymentRepo.GetPaymentsByStatus(status, limit, offset)
}

func (s *paymentService) GetPaymentDetail(requestorRole model.UserRole, paymentID uint) (*model.Payment, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.paymentRepo.GetByIDWithRelations(paymentID)
}

func (s *paymentService) ProcessWebhook(providerPaymentID string, status string, paymentMethod string, failureReason *string) error {
	// Find payment by provider payment ID
	payment, err := s.paymentRepo.GetByProviderPaymentID(providerPaymentID)
	if err != nil {
		return ErrPaymentNotFound
	}

	// Process based on status
	switch status {
	case "paid", "success", "completed":
		if payment.Status == model.PaymentPaid {
			return nil // Already processed
		}

		// Mark payment as paid
		err = s.paymentRepo.MarkAsPaid(payment.ID, providerPaymentID, paymentMethod)
		if err != nil {
			return err
		}

		// Update booking status to confirmed
		return s.bookingService.ConfirmPayment(payment.BookingID)

	case "failed", "error", "cancelled":
		if payment.Status == model.PaymentFailed {
			return nil // Already processed
		}

		reason := "Payment failed"
		if failureReason != nil {
			reason = *failureReason
		}

		// Mark payment as failed
		err = s.paymentRepo.MarkAsFailed(payment.ID, reason)
		if err != nil {
			return err
		}

		// Update booking status and release stock
		return s.bookingService.FailPayment(payment.BookingID)

	default:
		return fmt.Errorf("unknown payment status: %s", status)
	}
}

func (s *paymentService) canManagePayments(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
