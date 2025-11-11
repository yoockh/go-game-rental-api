package service

import (
	"errors"
	"log"
	"time"

	"github.com/yoockh/go-api-utils/pkg-echo/auth"
	"github.com/yoockh/go-game-rental-api/internal/dto"
	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/repository"
	"github.com/yoockh/go-game-rental-api/internal/utils"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInsufficientPermission = errors.New("insufficient permission")
	ErrCannotDeleteSuperAdmin = errors.New("cannot delete super admin")
	ErrCannotDeleteSelf       = errors.New("cannot delete yourself")
)

type UserService interface {
	// Public methods
	GetProfile(userID uint) (*model.User, error)
	UpdateProfile(userID uint, updateData interface{}) error

	// Auth methods
	Register(registerData interface{}) (*model.User, error)
	Login(loginData interface{}, jwtSecret string) (interface{}, error)

	// Admin methods
	GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, int64, error)
	GetUserDetail(requestorRole model.UserRole, userID uint) (*model.User, error)
	UpdateUserRole(requestorRole model.UserRole, userID uint, newRole model.UserRole) error
	ToggleUserStatus(requestorRole model.UserRole, userID uint) error
	DeleteUser(requestorID uint, requestorRole model.UserRole, targetUserID uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetProfile(userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *userService) UpdateProfile(userID uint, updateData interface{}) error {
	req := updateData.(*dto.UpdateProfileRequest)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	user.FullName = req.FullName
	user.Phone = utils.PtrOrNil(req.Phone)
	user.Address = utils.PtrOrNil(req.Address)

	return s.userRepo.Update(user)
}

func (s *userService) Register(registerData interface{}) (*model.User, error) {
	req := registerData.(*dto.RegisterRequest)

	// Check if user exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// Use our own HashPassword
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    req.Email,
		Password: hashed,
		FullName: req.FullName,
		Phone:    &req.Phone,
		Address:  &req.Address,
		Role:     model.RoleCustomer,
		IsActive: true, // Auto-active (no email verification)
	}

	return user, s.userRepo.Create(user)
}

func (s *userService) Login(loginData interface{}, jwtSecret string) (interface{}, error) {
	req := loginData.(*dto.LoginRequest)
	log.Printf("DEBUG: Login attempt for email: %s", req.Email)

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Printf("ERROR: GetByEmail failed: %v", err)
		return nil, errors.New("invalid credentials")
	}
	log.Printf("DEBUG: User found, hash: %s..., active: %v", user.Password[:30], user.IsActive)

	// Use our own CheckPassword
	if !utils.CheckPassword(user.Password, req.Password) {
		log.Printf("ERROR: Password mismatch for email %s", req.Email)
		return nil, errors.New("invalid credentials")
	}
	log.Printf("DEBUG: Password OK for %s", req.Email)

	if !user.IsActive {
		log.Printf("ERROR: User not active: %s", req.Email)
		return nil, errors.New("account is inactive")
	}

	// Still use go-api-utils for JWT generation
	accessToken, err := auth.GenerateToken(
		int(user.ID),
		user.Email,
		string(user.Role),
		jwtSecret,
		24*time.Hour,
	)
	if err != nil {
		log.Printf("ERROR: GenerateToken failed: %v", err)
		return nil, err
	}
	log.Printf("DEBUG: Token generated successfully for %s", req.Email)

	return &dto.LoginResponse{
		AccessToken: accessToken,
		User:        user,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *userService) GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, int64, error) {
	if !s.canManageUsers(requestorRole) {
		return nil, 0, ErrInsufficientPermission
	}

	users, err := s.userRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.userRepo.Count()
	return users, count, err
}

func (s *userService) GetUserDetail(requestorRole model.UserRole, userID uint) (*model.User, error) {
	if !s.canManageUsers(requestorRole) {
		return nil, ErrInsufficientPermission
	}
	return s.userRepo.GetByID(userID)
}

func (s *userService) UpdateUserRole(requestorRole model.UserRole, userID uint, newRole model.UserRole) error {
	if !s.canManageUsers(requestorRole) {
		return ErrInsufficientPermission
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// FIX 1: Validate target role
	validRoles := []model.UserRole{model.RoleCustomer, model.RoleAdmin, model.RoleSuperAdmin}
	isValidRole := false
	for _, validRole := range validRoles {
		if newRole == validRole {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		return errors.New("invalid role")
	}

	// FIX 2: Admin cannot modify super_admin
	if requestorRole == model.RoleAdmin && targetUser.Role == model.RoleSuperAdmin {
		return errors.New("admin cannot modify super admin role")
	}

	// FIX 3: Admin cannot promote to super_admin
	if requestorRole == model.RoleAdmin && newRole == model.RoleSuperAdmin {
		return errors.New("admin cannot promote user to super admin")
	}

	// FIX 4: Only super_admin can create/modify super_admin
	if newRole == model.RoleSuperAdmin && requestorRole != model.RoleSuperAdmin {
		return errors.New("only super admin can assign super admin role")
	}

	return s.userRepo.UpdateRole(userID, newRole)
}

func (s *userService) ToggleUserStatus(requestorRole model.UserRole, userID uint) error {
	if !s.canManageUsers(requestorRole) {
		return ErrInsufficientPermission
	}

	targetUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// FIX 1: Admin cannot disable super_admin
	if requestorRole == model.RoleAdmin && targetUser.Role == model.RoleSuperAdmin {
		return errors.New("admin cannot modify super admin status")
	}

	// FIX 2: Super admin cannot be disabled (extra protection)
	if targetUser.Role == model.RoleSuperAdmin && requestorRole != model.RoleSuperAdmin {
		return ErrCannotDeleteSuperAdmin
	}

	return s.userRepo.UpdateActiveStatus(userID, !targetUser.IsActive)
}

func (s *userService) DeleteUser(requestorID uint, requestorRole model.UserRole, targetUserID uint) error {
	// FIX: Allow both admin and super_admin
	if requestorRole != model.RoleAdmin && requestorRole != model.RoleSuperAdmin {
		return ErrInsufficientPermission
	}

	// Prevent self-delete
	if requestorID == targetUserID {
		return ErrCannotDeleteSelf
	}

	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Admin cannot delete super_admin
	if requestorRole == model.RoleAdmin && targetUser.Role == model.RoleSuperAdmin {
		return errors.New("admin cannot delete super admin")
	}

	// Super admin cannot be deleted (extra safety)
	if targetUser.Role == model.RoleSuperAdmin && requestorRole != model.RoleSuperAdmin {
		return ErrCannotDeleteSuperAdmin
	}

	return s.userRepo.Delete(targetUserID)
}

// Helper methods
func (s *userService) canManageUsers(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
