package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/repository"
	"github.com/yoockh/go-game-rental-api/internal/repository/email"
	"github.com/yoockh/go-game-rental-api/internal/repository/transaction"
)

var (
	ErrPaymentNotFound               = errors.New("payment not found")
	ErrPaymentAlreadyExists          = errors.New("payment already exists for this booking")
	ErrPaymentBookingNotFound        = errors.New("booking not found")
	ErrPaymentInvalidStatus          = errors.New("invalid payment status transition")
	ErrPaymentInsufficientPermission = errors.New("insufficient permission")
)

type PaymentService interface {
	// Customer methods
	CreatePayment(userID uint, bookingID uint, provider model.PaymentProvider, paymentType string) (*model.Payment, error)
	GetPaymentByBooking(userID uint, bookingID uint) (*model.Payment, error)

	// Admin methods
	GetAllPayments(requestorRole model.UserRole, limit, offset int) ([]*model.Payment, int64, error)
	GetPaymentsByStatus(requestorRole model.UserRole, status model.PaymentStatus, limit, offset int) ([]*model.Payment, int64, error)
	GetPaymentDetail(requestorRole model.UserRole, paymentID uint) (*model.Payment, error)

	// Webhook/System methods
	ProcessWebhook(data interface{}) error
}

type paymentService struct {
	paymentRepo     repository.PaymentRepository
	bookingRepo     repository.BookingRepository
	userRepo        repository.UserRepository
	gameRepo        repository.GameRepository
	bookingService  BookingService
	transactionRepo transaction.TransactionRepository
	emailRepo       email.EmailRepository
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	userRepo repository.UserRepository,
	gameRepo repository.GameRepository,
	bookingService BookingService,
	transactionRepo transaction.TransactionRepository,
	emailRepo email.EmailRepository,
) PaymentService {
	return &paymentService{
		paymentRepo:     paymentRepo,
		bookingRepo:     bookingRepo,
		userRepo:        userRepo,
		gameRepo:        gameRepo,
		bookingService:  bookingService,
		transactionRepo: transactionRepo,
		emailRepo:       emailRepo,
	}
}

func (s *paymentService) CreatePayment(userID uint, bookingID uint, provider model.PaymentProvider, paymentType string) (*model.Payment, error) {
	// Get booking and validate ownership
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrPaymentBookingNotFound
	}

	if booking.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if booking.Status != model.BookingPending {
		return nil, errors.New("booking must be in pending status")
	}

	// Check if payment already exists
	existingPayment, _ := s.paymentRepo.GetByBookingID(bookingID)
	if existingPayment != nil {
		return nil, ErrPaymentAlreadyExists
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

	// Create charge with payment gateway based on provider
	orderID := fmt.Sprintf("booking-%d", bookingID)

	switch provider {
	case model.ProviderMidtrans:
		// Set default payment type if not provided
		if paymentType == "" {
			paymentType = "bank_transfer"
		}

		txID, _, err := s.transactionRepo.CreateCharge(
			context.Background(),
			orderID,
			int64(payment.Amount),
			paymentType,
			nil,
		)
		if err != nil {
			return payment, fmt.Errorf("midtrans payment gateway error: %w", err)
		}

		// Update payment with provider transaction ID
		payment.ProviderPaymentID = &txID
		s.paymentRepo.Update(payment)

	case model.ProviderStripe:
		return payment, errors.New("stripe payment provider not implemented yet")

	default:
		return payment, errors.New("unsupported payment provider")
	}

	// SEND EMAIL: Payment instruction
	user, _ := s.userRepo.GetByID(userID)
	game, _ := s.gameRepo.GetByID(booking.GameID)
	if user != nil && game != nil {
		go func() {
			subject := "Payment Instruction - Game Rental"
			orderIDStr := "N/A"
			if payment.ProviderPaymentID != nil {
				orderIDStr = *payment.ProviderPaymentID
			}
			htmlContent := fmt.Sprintf(`
				<h1>Complete Your Payment</h1>
				<p>Hi %s,</p>
				<p>Please complete payment to confirm your booking.</p>
				<h3>Payment Details:</h3>
				<ul>
					<li><strong>Order ID:</strong> %s</li>
					<li><strong>Amount:</strong> Rp %.0f</li>
					<li><strong>Game:</strong> %s</li>
				</ul>
				<p>Complete within 24 hours.</p>
			`, user.FullName, orderIDStr, payment.Amount, game.Name)

			plainText := fmt.Sprintf("Payment instruction. Order ID: %s, Amount: Rp %.0f", orderIDStr, payment.Amount)

			if err := s.emailRepo.SendEmail(context.Background(), user.Email, subject, plainText, htmlContent); err != nil {
				logrus.WithError(err).Error("Failed to send payment instruction email")
			}
		}()
	}

	return payment, nil
}

func (s *paymentService) GetPaymentByBooking(userID uint, bookingID uint) (*model.Payment, error) {
	// Validate booking ownership
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrPaymentBookingNotFound
	}

	if booking.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.paymentRepo.GetByBookingID(bookingID)
}

func (s *paymentService) GetAllPayments(requestorRole model.UserRole, limit, offset int) ([]*model.Payment, int64, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, 0, ErrPaymentInsufficientPermission
	}

	payments, err := s.paymentRepo.GetAllPayments(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.paymentRepo.CountAllPayments()
	return payments, count, err
}

func (s *paymentService) GetPaymentsByStatus(requestorRole model.UserRole, status model.PaymentStatus, limit, offset int) ([]*model.Payment, int64, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, 0, ErrPaymentInsufficientPermission
	}

	payments, err := s.paymentRepo.GetPaymentsByStatus(status, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.paymentRepo.CountByStatus(status)
	return payments, count, err
}

func (s *paymentService) GetPaymentDetail(requestorRole model.UserRole, paymentID uint) (*model.Payment, error) {
	if !s.canManagePayments(requestorRole) {
		return nil, ErrPaymentInsufficientPermission
	}

	return s.paymentRepo.GetByIDWithRelations(paymentID)
}

func (s *paymentService) ProcessWebhook(data interface{}) error {
	webhookData, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("invalid webhook data")
	}

	providerPaymentID, ok := webhookData["order_id"].(string)
	if !ok {
		return errors.New("missing order_id in webhook")
	}

	transactionStatus, ok := webhookData["transaction_status"].(string)
	if !ok {
		return errors.New("missing transaction_status in webhook")
	}

	payment, err := s.paymentRepo.GetByProviderPaymentID(providerPaymentID)
	if err != nil {
		return ErrPaymentNotFound
	}

	var newStatus model.PaymentStatus
	switch transactionStatus {
	case "capture", "settlement":
		newStatus = model.PaymentPaid
	case "pending":
		newStatus = model.PaymentPending
	case "deny", "expire", "cancel":
		newStatus = model.PaymentFailed
	default:
		return errors.New("unknown transaction status")
	}

	now := time.Now()
	switch newStatus {
	case model.PaymentPaid:
		payment.PaidAt = &now
		if err := s.bookingService.ConfirmPayment(payment.BookingID); err != nil {
			return err
		}
	case model.PaymentFailed:
		if err := s.bookingService.FailPayment(payment.BookingID); err != nil {
			return err
		}
	}

	payment.Status = newStatus
	return s.paymentRepo.Update(payment)
}

func (s *paymentService) canManagePayments(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
