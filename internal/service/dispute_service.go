package service

import (
	"errors"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
)

var (
	ErrDisputeNotFound        = errors.New("dispute not found")
	ErrDisputeAlreadyResolved = errors.New("dispute already resolved")
	ErrCannotCreateDispute    = errors.New("cannot create dispute for this booking")
)

type DisputeService interface {
	// Customer/Partner methods
	CreateDispute(reporterID uint, bookingID uint, disputeData *model.Dispute) error
	GetMyDisputes(userID uint, limit, offset int) ([]*model.Dispute, error)
	GetDisputeDetail(userID uint, disputeID uint) (*model.Dispute, error)

	// Admin methods
	GetPendingDisputes(requestorRole model.UserRole, limit, offset int) ([]*model.Dispute, error)
	GetAllDisputes(requestorRole model.UserRole, limit, offset int) ([]*model.Dispute, error)
	GetDisputesByStatus(requestorRole model.UserRole, status model.DisputeStatus, limit, offset int) ([]*model.Dispute, error)
	InvestigateDispute(adminID uint, requestorRole model.UserRole, disputeID uint) error
	ResolveDispute(adminID uint, requestorRole model.UserRole, disputeID uint, resolution string) error
	CloseDispute(adminID uint, requestorRole model.UserRole, disputeID uint, resolution string) error
	GetDisputeDetailAdmin(requestorRole model.UserRole, disputeID uint) (*model.Dispute, error)
}

type disputeService struct {
	disputeRepo repository.DisputeRepository
	bookingRepo repository.BookingRepository
}

func NewDisputeService(disputeRepo repository.DisputeRepository, bookingRepo repository.BookingRepository) DisputeService {
	return &disputeService{
		disputeRepo: disputeRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *disputeService) CreateDispute(reporterID uint, bookingID uint, disputeData *model.Dispute) error {
	// Validate booking exists
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	// Check if user is involved in the booking (either customer or partner)
	if booking.UserID != reporterID && booking.PartnerID != reporterID {
		return ErrCannotCreateDispute
	}

	// Can only create dispute for confirmed, active, or completed bookings
	validStatuses := []model.BookingStatus{
		model.BookingConfirmed,
		model.BookingActive,
		model.BookingCompleted,
	}

	isValidStatus := false
	for _, status := range validStatuses {
		if booking.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		return ErrCannotCreateDispute
	}

	// Set dispute details
	disputeData.BookingID = bookingID
	disputeData.ReporterID = reporterID
	disputeData.Status = model.DisputeOpen

	// Update booking status to disputed
	err = s.bookingRepo.UpdateStatus(bookingID, model.BookingDisputed)
	if err != nil {
		return err
	}

	return s.disputeRepo.Create(disputeData)
}

func (s *disputeService) GetMyDisputes(userID uint, limit, offset int) ([]*model.Dispute, error) {
	return s.disputeRepo.GetUserDisputes(userID, limit, offset)
}

func (s *disputeService) GetDisputeDetail(userID uint, disputeID uint) (*model.Dispute, error) {
	dispute, err := s.disputeRepo.GetByIDWithRelations(disputeID)
	if err != nil {
		return nil, ErrDisputeNotFound
	}

	// Check if user is involved (reporter or affected party)
	booking := dispute.Booking
	if dispute.ReporterID != userID && booking.UserID != userID && booking.PartnerID != userID {
		return nil, ErrInsufficientPermission
	}

	return dispute, nil
}

func (s *disputeService) GetPendingDisputes(requestorRole model.UserRole, limit, offset int) ([]*model.Dispute, error) {
	if !s.canManageDisputes(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.disputeRepo.GetPendingDisputes(limit, offset)
}

func (s *disputeService) GetAllDisputes(requestorRole model.UserRole, limit, offset int) ([]*model.Dispute, error) {
	if !s.canManageDisputes(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.disputeRepo.GetAllDisputes(limit, offset)
}

func (s *disputeService) GetDisputesByStatus(requestorRole model.UserRole, status model.DisputeStatus, limit, offset int) ([]*model.Dispute, error) {
	if !s.canManageDisputes(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.disputeRepo.GetDisputesByStatus(status, limit, offset)
}

func (s *disputeService) InvestigateDispute(adminID uint, requestorRole model.UserRole, disputeID uint) error {
	if !s.canManageDisputes(requestorRole) {
		return ErrInsufficientPermission
	}

	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status != model.DisputeOpen {
		return errors.New("can only investigate open disputes")
	}

	return s.disputeRepo.UpdateDisputeStatus(disputeID, model.DisputeInvestigating, &adminID, nil)
}

func (s *disputeService) ResolveDispute(adminID uint, requestorRole model.UserRole, disputeID uint, resolution string) error {
	if !s.canManageDisputes(requestorRole) {
		return ErrInsufficientPermission
	}

	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status == model.DisputeResolved || dispute.Status == model.DisputeClosed {
		return ErrDisputeAlreadyResolved
	}

	return s.disputeRepo.UpdateDisputeStatus(disputeID, model.DisputeResolved, &adminID, &resolution)
}

func (s *disputeService) CloseDispute(adminID uint, requestorRole model.UserRole, disputeID uint, resolution string) error {
	if !s.canManageDisputes(requestorRole) {
		return ErrInsufficientPermission
	}

	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status == model.DisputeClosed {
		return errors.New("dispute already closed")
	}

	// Close dispute and potentially restore booking status
	err = s.disputeRepo.UpdateDisputeStatus(disputeID, model.DisputeClosed, &adminID, &resolution)
	if err != nil {
		return err
	}

	// Restore booking status to completed if dispute is resolved
	return s.bookingRepo.UpdateStatus(dispute.BookingID, model.BookingCompleted)
}

func (s *disputeService) GetDisputeDetailAdmin(requestorRole model.UserRole, disputeID uint) (*model.Dispute, error) {
	if !s.canManageDisputes(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.disputeRepo.GetByIDWithRelations(disputeID)
}

func (s *disputeService) canManageDisputes(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
