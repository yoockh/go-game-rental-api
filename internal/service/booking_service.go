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
)

var (
	ErrBookingNotFound       = errors.New("booking not found")
	ErrBookingNotOwned       = errors.New("you don't own this booking")
	ErrBookingInvalidDate    = errors.New("invalid booking dates")
	ErrBookingCannotCancel   = errors.New("cannot cancel booking in current status")
	ErrGameStockInsufficient = errors.New("insufficient stock")
)

type BookingService interface {
	// Customer
	Create(userID uint, bookingData *model.Booking) error
	GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, int64, error)
	GetByID(userID uint, bookingID uint) (*model.Booking, error)
	Cancel(userID uint, bookingID uint) error

	// Admin
	GetAll(requestorRole model.UserRole, limit, offset int) ([]*model.Booking, int64, error)
	UpdateStatus(requestorRole model.UserRole, bookingID uint, status model.BookingStatus) error

	// System (for payment)
	ConfirmPayment(bookingID uint) error
	FailPayment(bookingID uint) error
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	gameRepo    repository.GameRepository
	userRepo    repository.UserRepository
	emailRepo   email.EmailRepository
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	emailRepo email.EmailRepository,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		gameRepo:    gameRepo,
		userRepo:    userRepo,
		emailRepo:   emailRepo,
	}
}

func (s *bookingService) Create(userID uint, bookingData *model.Booking) error {
	game, err := s.gameRepo.GetByID(bookingData.GameID)
	if err != nil {
		return ErrGameNotFound
	}

	if !game.IsActive {
		return errors.New("game is not available for booking")
	}

	if bookingData.StartDate.After(bookingData.EndDate) || bookingData.StartDate.Before(time.Now().Truncate(24*time.Hour)) {
		return ErrBookingInvalidDate
	}

	available, err := s.gameRepo.CheckAvailability(game.ID)
	if err != nil {
		return err
	}
	if !available {
		return ErrGameStockInsufficient
	}

	rentalDays := int(bookingData.EndDate.Sub(bookingData.StartDate).Hours()/24) + 1
	totalRentalPrice := float64(rentalDays) * game.RentalPricePerDay
	totalAmount := totalRentalPrice + game.SecurityDeposit

	bookingData.UserID = userID
	bookingData.RentalDays = rentalDays
	bookingData.DailyPrice = game.RentalPricePerDay
	bookingData.TotalRentalPrice = totalRentalPrice
	bookingData.SecurityDeposit = game.SecurityDeposit
	bookingData.TotalAmount = totalAmount
	bookingData.Status = model.BookingPending

	err = s.gameRepo.ReserveStock(game.ID)
	if err != nil {
		return err
	}

	if err := s.bookingRepo.Create(bookingData); err != nil {
		return err
	}

	// SEND EMAIL: Booking confirmation
	user, _ := s.userRepo.GetByID(userID)
	if user != nil {
		go func() {
			subject := "Booking Confirmation - Game Rental"
			platform := "Unknown"
			if game.Platform != nil {
				platform = *game.Platform
			}
			htmlContent := fmt.Sprintf(`
				<h1>Booking Confirmation</h1>
				<p>Hi %s,</p>
				<p>Your booking has been created successfully!</p>
				<h3>Details:</h3>
				<ul>
					<li><strong>Game:</strong> %s</li>
					<li><strong>Platform:</strong> %s</li>
					<li><strong>Period:</strong> %s to %s (%d days)</li>
					<li><strong>Total:</strong> Rp %.0f</li>
				</ul>
				<p><strong>Next:</strong> Please complete the payment.</p>
			`, user.FullName, game.Name, platform, bookingData.StartDate.Format("2006-01-02"), bookingData.EndDate.Format("2006-01-02"), rentalDays, totalAmount)

			plainText := fmt.Sprintf("Booking confirmed for %s. Total: Rp %.0f", game.Name, totalAmount)

			if err := s.emailRepo.SendEmail(context.Background(), user.Email, subject, plainText, htmlContent); err != nil {
				logrus.WithError(err).Error("Failed to send booking email")
			}
		}()
	}

	return nil
}

func (s *bookingService) GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, int64, error) {
	bookings, err := s.bookingRepo.GetUserBookings(userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.bookingRepo.CountUserBookings(userID)
	return bookings, count, err
}

func (s *bookingService) GetByID(userID uint, bookingID uint) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	if booking.UserID != userID {
		return nil, ErrBookingNotOwned
	}

	return booking, nil
}

func (s *bookingService) Cancel(userID uint, bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.UserID != userID {
		return ErrBookingNotOwned
	}

	if booking.Status != model.BookingPending && booking.Status != model.BookingConfirmed {
		return ErrBookingCannotCancel
	}

	err = s.gameRepo.ReleaseStock(booking.GameID)
	if err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(bookingID, model.BookingCancelled)
}

func (s *bookingService) GetAll(requestorRole model.UserRole, limit, offset int) ([]*model.Booking, int64, error) {
	if !s.canManageBookings(requestorRole) {
		return nil, 0, ErrInsufficientPermission
	}

	bookings, err := s.bookingRepo.GetAllBookings(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.bookingRepo.Count()
	return bookings, count, err
}

func (s *bookingService) UpdateStatus(requestorRole model.UserRole, bookingID uint, status model.BookingStatus) error {
	if !s.canManageBookings(requestorRole) {
		return ErrInsufficientPermission
	}

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if err := s.bookingRepo.UpdateStatus(bookingID, status); err != nil {
		return err
	}

	// SEND EMAIL: Status update
	user, _ := s.userRepo.GetByID(booking.UserID)
	game, _ := s.gameRepo.GetByID(booking.GameID)
	if user != nil && game != nil {
		go func() {
			subject := "Booking Status Updated - Game Rental"
			statusMsg := ""
			switch status {
			case model.BookingActive:
				statusMsg = "Your game is ready!"
			case model.BookingCompleted:
				statusMsg = "Thank you!"
			}

			htmlContent := fmt.Sprintf(`
				<h1>Status Updated</h1>
				<p>Hi %s,</p>
				<p>Booking status: <strong>%s</strong></p>
				<p>Game: %s</p>
				<p>%s</p>
			`, user.FullName, status, game.Name, statusMsg)

			plainText := fmt.Sprintf("Booking status: %s for %s", status, game.Name)

			if err := s.emailRepo.SendEmail(context.Background(), user.Email, subject, plainText, htmlContent); err != nil {
				logrus.WithError(err).Error("Failed to send status update email")
			}
		}()
	}

	return nil
}

func (s *bookingService) ConfirmPayment(bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.Status != model.BookingPending {
		return errors.New("booking is not in pending status")
	}

	if err := s.bookingRepo.UpdateStatus(bookingID, model.BookingConfirmed); err != nil {
		return err
	}

	// SEND EMAIL: Payment confirmed
	user, _ := s.userRepo.GetByID(booking.UserID)
	game, _ := s.gameRepo.GetByID(booking.GameID)
	if user != nil && game != nil {
		go func() {
			subject := "Payment Confirmed - Game Rental"
			platform := "Unknown"
			if game.Platform != nil {
				platform = *game.Platform
			}
			htmlContent := fmt.Sprintf(`
				<h1>Payment Successful!</h1>
				<p>Hi %s,</p>
				<p>Your payment has been confirmed!</p>
				<h3>Details:</h3>
				<ul>
					<li><strong>Game:</strong> %s</li>
					<li><strong>Platform:</strong> %s</li>
					<li><strong>Period:</strong> %s to %s</li>
					<li><strong>Amount:</strong> Rp %.0f</li>
				</ul>
			`, user.FullName, game.Name, platform, booking.StartDate.Format("2006-01-02"), booking.EndDate.Format("2006-01-02"), booking.TotalAmount)

			plainText := fmt.Sprintf("Payment confirmed for %s", game.Name)

			if err := s.emailRepo.SendEmail(context.Background(), user.Email, subject, plainText, htmlContent); err != nil {
				logrus.WithError(err).Error("Failed to send payment confirmation email")
			}
		}()
	}

	return nil
}

func (s *bookingService) FailPayment(bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	err = s.gameRepo.ReleaseStock(booking.GameID)
	if err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(bookingID, model.BookingCancelled)
}

func (s *bookingService) canManageBookings(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
