package service

import (
	"errors"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	VerifyEmail(token string) (*model.User, error)
	ResendVerification(email string) error

	// Admin methods
	GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, int64, error)
	GetUserDetail(requestorRole model.UserRole, userID uint) (*model.User, error)
	UpdateUserRole(requestorRole model.UserRole, userID uint, newRole model.UserRole) error
	ToggleUserStatus(requestorRole model.UserRole, userID uint) error

	// Super Admin methods
	DeleteUser(requestorID uint, requestorRole model.UserRole, targetUserID uint) error
}

type userService struct {
	userRepo repository.UserRepository
	verificationRepo repository.EmailVerificationRepository
}

func NewUserService(userRepo repository.UserRepository, verificationRepo repository.EmailVerificationRepository) UserService {
	return &userService{
		userRepo: userRepo,
		verificationRepo: verificationRepo,
	}
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
	user.Phone = &req.Phone
	user.Address = &req.Address

	return s.userRepo.Update(user)
}

func (s *userService) Register(registerData interface{}) (*model.User, error) {
	req := registerData.(*dto.RegisterRequest)

	// Check if user exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    req.Email,
		Password: string(hashed),
		FullName: req.FullName,
		Phone:    &req.Phone,
		Address:  &req.Address,
		Role:     model.RoleCustomer,
		IsActive: false,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Create verification token
	_, err = s.createVerificationToken(user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(loginData interface{}, jwtSecret string) (interface{}, error) {
	req := loginData.(*dto.LoginRequest)

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is verified
	if !user.IsActive {
		return nil, errors.New("please verify your email before logging in")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		User:        user,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *userService) VerifyEmail(token string) (*model.User, error) {
	// Hash the token to match database
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Get verification token
	verificationToken, err := s.verificationRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired verification token")
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		return nil, errors.New("verification token has expired")
	}

	// Activate user
	user := &verificationToken.User
	user.IsActive = true
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	// Mark token as used
	err = s.verificationRepo.MarkAsUsed(verificationToken.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ResendVerification(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	if user.IsActive {
		return errors.New("user is already verified")
	}

	// Generate new verification token
	_, err = s.createVerificationToken(user.ID)
	return err
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
	return s.userRepo.UpdateRole(userID, newRole)
}

func (s *userService) ToggleUserStatus(requestorRole model.UserRole, userID uint) error {
	if !s.canManageUsers(requestorRole) {
		return ErrInsufficientPermission
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.UpdateActiveStatus(userID, !user.IsActive)
}

func (s *userService) DeleteUser(requestorID uint, requestorRole model.UserRole, targetUserID uint) error {
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

// Helper methods
func (s *userService) canManageUsers(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}

func (s *userService) createVerificationToken(userID uint) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Create verification token record
	verificationToken := &model.EmailVerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
		IsUsed:    false,
	}

	err = s.verificationRepo.Create(verificationToken)
	if err != nil {
		return "", err
	}

	return token, nil
}
