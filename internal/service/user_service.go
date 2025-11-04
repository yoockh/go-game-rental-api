package service

import (
	"errors"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInsufficientPermission = errors.New("insufficient permission")
	ErrCannotDeleteSuperAdmin = errors.New("cannot delete super admin")
	ErrCannotDeleteSelf       = errors.New("cannot delete yourself")
)

type UserService interface {
	// Public methods
	GetProfile(userID uuid.UUID) (*model.User, error)
	UpdateProfile(userID uuid.UUID, updateData *model.User) error

	// Admin methods
	GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, error)
	BanUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error
	UnbanUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error

	// Super Admin methods
	CreateAdmin(requestorRole model.UserRole, userData *model.User) error
	DeleteUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error
	PromoteToAdmin(requestorRole model.UserRole, userID uuid.UUID) error
	DemoteFromAdmin(requestorRole model.UserRole, userID uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetProfile(userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *userService) UpdateProfile(userID uuid.UUID, updateData *model.User) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// User just can update their own profile
	user.FullName = updateData.FullName
	user.Phone = updateData.Phone
	user.Address = updateData.Address

	return s.userRepo.Update(user)
}

func (s *userService) GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, error) {
	if !s.canManageUsers(requestorRole) {
		return nil, ErrInsufficientPermission
	}

	return s.userRepo.GetAll(limit, offset)
}

func (s *userService) BanUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error {
	if !s.canManageUsers(requestorRole) {
		return ErrInsufficientPermission
	}

	if requestorID == targetUserID {
		return ErrCannotDeleteSelf
	}

	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Admin can't ban other admins or super admins
	if requestorRole == model.RoleAdmin && (targetUser.Role == model.RoleSuperAdmin || targetUser.Role == model.RoleAdmin) {
		return ErrInsufficientPermission
	}

	return s.userRepo.UpdateActiveStatus(targetUserID, false)
}

func (s *userService) UnbanUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error {
	if !s.canManageUsers(requestorRole) {
		return ErrInsufficientPermission
	}

	return s.userRepo.UpdateActiveStatus(targetUserID, true)
}

func (s *userService) CreateAdmin(requestorRole model.UserRole, userData *model.User) error {
	if requestorRole != model.RoleSuperAdmin {
		return ErrInsufficientPermission
	}

	userData.Role = model.RoleAdmin
	return s.userRepo.Create(userData)
}

func (s *userService) DeleteUser(requestorID uuid.UUID, requestorRole model.UserRole, targetUserID uuid.UUID) error {
	if requestorRole != model.RoleSuperAdmin {
		return ErrInsufficientPermission
	}

	if requestorID == targetUserID {
		return ErrCannotDeleteSelf
	}

	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Super admin cannot be deleted
	if targetUser.Role == model.RoleSuperAdmin {
		return ErrCannotDeleteSuperAdmin
	}

	return s.userRepo.Delete(targetUserID)
}

func (s *userService) PromoteToAdmin(requestorRole model.UserRole, userID uuid.UUID) error {
	if requestorRole != model.RoleSuperAdmin {
		return ErrInsufficientPermission
	}

	return s.userRepo.UpdateRole(userID, model.RoleAdmin)
}

func (s *userService) DemoteFromAdmin(requestorRole model.UserRole, userID uuid.UUID) error {
	if requestorRole != model.RoleSuperAdmin {
		return ErrInsufficientPermission
	}

	return s.userRepo.UpdateRole(userID, model.RoleCustomer)
}

// Helper methods
func (s *userService) canManageUsers(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
