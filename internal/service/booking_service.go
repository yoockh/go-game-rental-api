package service

import (
	"errors"
	"time"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
)

var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingNotOwned      = errors.New("you don't own this booking")
	ErrBookingDateConflict  = errors.New("booking dates conflict with existing bookings")
	ErrBookingInvalidDate   = errors.New("invalid booking dates")
	ErrBookingCannotCancel  = errors.New("cannot cancel booking in current status")
	ErrBookingCannotConfirm = errors.New("cannot confirm booking in current status")
	ErrBookingGameNotActive = errors.New("game is not active for booking")
	ErrGameStockInsufficient = errors.New("insufficient stock")
)

type BookingService interface {
	// Customer methods
	CreateBooking(userID uint, bookingData *model.Booking) error
	GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, error)
	GetBookingDetail(userID uint, bookingID uint) (*model.Booking, error)
	CancelBooking(userID uint, bookingID uint) error

	// Partner methods
	GetPartnerBookings(partnerID uint, limit, offset int) ([]*model.Booking, error)
	ConfirmHandover(partnerID uint, bookingID uint) error
	ConfirmReturn(partnerID uint, bookingID uint) error

	// Admin methods
	GetAllBookings(requestorRole model.UserRole, limit, offset int) ([]*model.Booking, error)
	GetBookingsByStatus(requestorRole model.UserRole, status model.BookingStatus, limit, offset int) ([]*model.Booking, error)
	ForceUpdateBookingStatus(requestorRole model.UserRole, bookingID uint, status model.BookingStatus) error

	// System methods (for payment service)
	ConfirmPayment(bookingID uint) error
	FailPayment(bookingID uint) error
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	gameRepo    repository.GameRepository
	userRepo    repository.UserRepository
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		gameRepo:    gameRepo,
		userRepo:    userRepo,
	}
}

// Customer methods
func (s *bookingService) CreateBooking(userID uint, bookingData *model.Booking) error {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Validate game exists and is active
	game, err := s.gameRepo.GetByID(bookingData.GameID)
	if err != nil {
		return ErrGameNotFound
	}

	if !game.IsActive || game.ApprovalStatus != model.ApprovalApproved {
		return ErrBookingGameNotActive
	}

	// Validate dates
	if bookingData.StartDate.After(bookingData.EndDate) || bookingData.StartDate.Before(time.Now().Truncate(24*time.Hour)) {
		return ErrBookingInvalidDate
	}

	// Check for date conflicts
	hasConflict, err := s.bookingRepo.CheckDateConflicts(game.ID, bookingData.StartDate, bookingData.EndDate, nil)
	if err != nil {
		return err
	}
	if hasConflict {
		return ErrBookingDateConflict
	}

	// Check stock availability
	available, err := s.gameRepo.CheckAvailability(game.ID, 1)
	if err != nil {
		return err
	}
	if !available {
		return ErrGameStockInsufficient
	}

	// Calculate rental details
	rentalDays := int(bookingData.EndDate.Sub(bookingData.StartDate).Hours()/24) + 1
	totalRentalPrice := float64(rentalDays) * game.RentalPricePerDay
	totalAmount := totalRentalPrice + game.SecurityDeposit

	// Set booking details
	bookingData.UserID = userID
	bookingData.PartnerID = game.PartnerID
	bookingData.RentalDays = rentalDays
	bookingData.DailyPrice = game.RentalPricePerDay
	bookingData.TotalRentalPrice = totalRentalPrice
	bookingData.SecurityDeposit = game.SecurityDeposit
	bookingData.TotalAmount = totalAmount
	bookingData.Status = model.BookingPendingPayment

	// Reserve stock
	err = s.gameRepo.ReserveStock(game.ID, 1)
	if err != nil {
		return err
	}

	return s.bookingRepo.Create(bookingData)
}

func (s *bookingService) GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, error) {
	return s.bookingRepo.GetUserBookings(userID, limit, offset)
}

func (s *bookingService) GetBookingDetail(userID uint, bookingID uint) (*model.Booking, error) {
	booking, err := s.bookingRepo.GetByIDWithRelations(bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	// Check ownership
	if booking.UserID != userID {
		return nil, ErrBookingNotOwned
	}

	return booking, nil
}

func (s *bookingService) CancelBooking(userID uint, bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	// Check ownership
	if booking.UserID != userID {
		return ErrBookingNotOwned
	}

	// Can only cancel pending payment or confirmed bookings
	if booking.Status != model.BookingPendingPayment && booking.Status != model.BookingConfirmed {
		return ErrBookingCannotCancel
	}

	// Release reserved stock
	err = s.gameRepo.ReleaseStock(booking.GameID, 1)
	if err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(bookingID, model.BookingCancelled)
}

// Partner methods
func (s *bookingService) GetPartnerBookings(partnerID uint, limit, offset int) ([]*model.Booking, error) {
	return s.bookingRepo.GetPartnerBookings(partnerID, limit, offset)
}

func (s *bookingService) ConfirmHandover(partnerID uint, bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	// Check ownership
	if booking.PartnerID != partnerID {
		return ErrBookingNotOwned
	}

	// Can only confirm handover for confirmed bookings
	if booking.Status != model.BookingConfirmed {
		return ErrBookingCannotConfirm
	}

	return s.bookingRepo.UpdateHandoverConfirmation(bookingID)
}

func (s *bookingService) ConfirmReturn(partnerID uint, bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	// Check ownership
	if booking.PartnerID != partnerID {
		return ErrBookingNotOwned
	}

	// Can only confirm return for active bookings
	if booking.Status != model.BookingActive {
		return ErrBookingCannotConfirm
	}

	// Release stock back
	err = s.gameRepo.ReleaseStock(booking.GameID, 1)
	if err != nil {
		return err
	}

	return s.bookingRepo.UpdateReturnConfirmation(bookingID)
}

// Admin methods
func (s *bookingService) GetAllBookings(requestorRole model.UserRole, limit, offset int) ([]*model.Booking, error) {
	if !s.canManageBookings(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	// Use GetBookingsByStatus with empty status to get all bookings
	return s.bookingRepo.GetBookingsByStatus(model.BookingStatus(""), limit, offset)
}

func (s *bookingService) GetBookingsByStatus(requestorRole model.UserRole, status model.BookingStatus, limit, offset int) ([]*model.Booking, error) {
	if !s.canManageBookings(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.bookingRepo.GetBookingsByStatus(status, limit, offset)
}

func (s *bookingService) ForceUpdateBookingStatus(requestorRole model.UserRole, bookingID uint, status model.BookingStatus) error {
	if !s.canManageBookings(requestorRole) {
		return ErrInsufficientPermission
	}

	return s.bookingRepo.UpdateStatus(bookingID, status)
}

// System methods (for payment service)
func (s *bookingService) ConfirmPayment(bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.Status != model.BookingPendingPayment {
		return errors.New("booking is not in pending payment status")
	}

	return s.bookingRepo.UpdateStatus(bookingID, model.BookingConfirmed)
}

func (s *bookingService) FailPayment(bookingID uint) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	// Release reserved stock
	err = s.gameRepo.ReleaseStock(booking.GameID, 1)
	if err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(bookingID, model.BookingCancelled)
}

// Helper methods
func (s *bookingService) canManageBookings(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
