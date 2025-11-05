package service

import (
	"errors"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
)

var (
	ErrApplicationNotFound       = errors.New("partner application not found")
	ErrApplicationAlreadyExists  = errors.New("user already has a partner application")
	ErrApplicationAlreadyDecided = errors.New("application already decided")
	ErrUserAlreadyPartner        = errors.New("user is already a partner")
)

type PartnerApplicationService interface {
	// Customer methods
	SubmitApplication(userID uint, applicationData *model.PartnerApplication) error
	GetMyApplication(userID uint) (*model.PartnerApplication, error)

	// Admin methods
	GetPendingApplications(requestorRole model.UserRole, limit, offset int) ([]*model.PartnerApplication, error)
	GetAllApplications(requestorRole model.UserRole, limit, offset int) ([]*model.PartnerApplication, error)
	ApproveApplication(adminID uint, requestorRole model.UserRole, applicationID uint) error
	RejectApplication(adminID uint, requestorRole model.UserRole, applicationID uint, reason string) error
	GetApplicationDetail(requestorRole model.UserRole, applicationID uint) (*model.PartnerApplication, error)
}

type partnerApplicationService struct {
	applicationRepo repository.PartnerApplicationRepository
	userRepo        repository.UserRepository
}

func NewPartnerApplicationService(applicationRepo repository.PartnerApplicationRepository, userRepo repository.UserRepository) PartnerApplicationService {
	return &partnerApplicationService{
		applicationRepo: applicationRepo,
		userRepo:        userRepo,
	}
}

func (s *partnerApplicationService) SubmitApplication(userID uint, applicationData *model.PartnerApplication) error {
	// Check if user exists and is not already a partner
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Role == model.RolePartner {
		return ErrUserAlreadyPartner
	}

	// Check if user already has an application
	existingApp, _ := s.applicationRepo.GetByUserID(userID)
	if existingApp != nil {
		return ErrApplicationAlreadyExists
	}

	// Set user ID and default status
	applicationData.UserID = userID
	applicationData.Status = model.ApplicationPending

	return s.applicationRepo.Create(applicationData)
}

func (s *partnerApplicationService) GetMyApplication(userID uint) (*model.PartnerApplication, error) {
	return s.applicationRepo.GetByUserID(userID)
}

func (s *partnerApplicationService) GetPendingApplications(requestorRole model.UserRole, limit, offset int) ([]*model.PartnerApplication, error) {
	if !s.canManageApplications(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.applicationRepo.GetPendingApplications(limit, offset)
}

func (s *partnerApplicationService) GetAllApplications(requestorRole model.UserRole, limit, offset int) ([]*model.PartnerApplication, error) {
	if !s.canManageApplications(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.applicationRepo.GetAllApplications(limit, offset)
}

func (s *partnerApplicationService) ApproveApplication(adminID uint, requestorRole model.UserRole, applicationID uint) error {
	if !s.canManageApplications(requestorRole) {
		return ErrInsufficientPermission
	}

	application, err := s.applicationRepo.GetByID(applicationID)
	if err != nil {
		return ErrApplicationNotFound
	}

	if application.Status != model.ApplicationPending {
		return ErrApplicationAlreadyDecided
	}

	// Update application status
	err = s.applicationRepo.UpdateApplicationStatus(applicationID, model.ApplicationApproved, adminID, nil)
	if err != nil {
		return err
	}

	// Update user role to partner
	return s.userRepo.UpdateRole(application.UserID, model.RolePartner)
}

func (s *partnerApplicationService) RejectApplication(adminID uint, requestorRole model.UserRole, applicationID uint, reason string) error {
	if !s.canManageApplications(requestorRole) {
		return ErrInsufficientPermission
	}

	application, err := s.applicationRepo.GetByID(applicationID)
	if err != nil {
		return ErrApplicationNotFound
	}

	if application.Status != model.ApplicationPending {
		return ErrApplicationAlreadyDecided
	}

	return s.applicationRepo.UpdateApplicationStatus(applicationID, model.ApplicationRejected, adminID, &reason)
}

func (s *partnerApplicationService) GetApplicationDetail(requestorRole model.UserRole, applicationID uint) (*model.PartnerApplication, error) {
	if !s.canManageApplications(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.applicationRepo.GetByIDWithRelations(applicationID)
}

func (s *partnerApplicationService) canManageApplications(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
