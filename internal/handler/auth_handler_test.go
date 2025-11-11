package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yoockh/go-game-rental-api/internal/model"
)

// ============= MOCK USER SERVICE =============
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(userID uint) (*model.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(userID uint, updateData interface{}) error {
	args := m.Called(userID, updateData)
	return args.Error(0)
}

func (m *MockUserService) Register(registerData interface{}) (*model.User, error) {
	args := m.Called(registerData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Login(loginData interface{}, jwtSecret string) (interface{}, error) {
	args := m.Called(loginData, jwtSecret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockUserService) GetAllUsers(requestorRole model.UserRole, limit, offset int) ([]*model.User, int64, error) {
	args := m.Called(requestorRole, limit, offset)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) GetUserDetail(requestorRole model.UserRole, userID uint) (*model.User, error) {
	args := m.Called(requestorRole, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateUserRole(requestorRole model.UserRole, userID uint, newRole model.UserRole) error {
	args := m.Called(requestorRole, userID, newRole)
	return args.Error(0)
}

func (m *MockUserService) ToggleUserStatus(requestorRole model.UserRole, userID uint) error {
	args := m.Called(requestorRole, userID)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(requestorID uint, requestorRole model.UserRole, targetUserID uint) error {
	args := m.Called(requestorID, requestorRole, targetUserID)
	return args.Error(0)
}

// ============= MOCK EMAIL REPO =============
type MockEmailRepository struct {
	mock.Mock
}

func (m *MockEmailRepository) SendEmail(ctx context.Context, to, subject, plainText, htmlContent string) error {
	args := m.Called(ctx, to, subject, plainText, htmlContent)
	return args.Error(0)
}

func (m *MockEmailRepository) SendWithTemplate(ctx context.Context, to, templateName string, data map[string]interface{}) error {
	args := m.Called(ctx, to, templateName, data)
	return args.Error(0)
}

// ============= TEST REGISTER SUCCESS =============
func TestRegister_Success(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	// Mock data
	reqBody := `{
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User",
        "phone": "081234567890",
        "address": "Test Address"
    }`

	expectedUser := &model.User{
		ID:       1,
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     model.RoleCustomer,
		IsActive: true,
	}

	// Mock expectations
	mockUserService.On("Register", mock.Anything).Return(expectedUser, nil)
	mockEmailRepo.On("SendEmail", mock.Anything, "test@example.com", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "User registered successfully")
		assert.Contains(t, rec.Body.String(), "test@example.com")
	}

	mockUserService.AssertExpectations(t)
}

// ============= TEST REGISTER VALIDATION ERROR =============
func TestRegister_ValidationError(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	// Invalid email
	reqBody := `{
        "email": "invalid-email",
        "password": "pass",
        "full_name": "Test User"
    }`

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Register(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Validation error")
	}
}

// ============= TEST REGISTER EMAIL EXISTS =============
func TestRegister_EmailExists(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	reqBody := `{
        "email": "existing@example.com",
        "password": "password123",
        "full_name": "Test User",
        "phone": "081234567890",
        "address": "Test Address"
    }`

	// Mock: service returns error
	mockUserService.On("Register", mock.Anything).Return(nil, errors.New("email already exists"))

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Register(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "email already exists")
	}

	mockUserService.AssertExpectations(t)
}

// ============= TEST LOGIN SUCCESS =============
func TestLogin_Success(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	reqBody := `{
        "email": "test@example.com",
        "password": "password123"
    }`

	// Mock response (interface{} sesuai method Login)
	expectedResponse := map[string]interface{}{
		"access_token": "mock-jwt-token",
		"user": model.User{
			ID:       1,
			Email:    "test@example.com",
			FullName: "Test User",
			Role:     model.RoleCustomer,
			IsActive: true,
		},
	}

	// Mock expectations
	mockUserService.On("Login", mock.Anything, "test-secret").Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Login successful")
		assert.Contains(t, rec.Body.String(), "mock-jwt-token")
	}

	mockUserService.AssertExpectations(t)
}

// ============= TEST LOGIN INVALID CREDENTIALS =============
func TestLogin_InvalidCredentials(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	reqBody := `{
        "email": "test@example.com",
        "password": "wrongpassword"
    }`

	// Mock: service returns error
	mockUserService.On("Login", mock.Anything, "test-secret").Return(nil, errors.New("invalid credentials"))

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid credentials")
	}

	mockUserService.AssertExpectations(t)
}

// ============= TEST LOGIN VALIDATION ERROR =============
func TestLogin_ValidationError(t *testing.T) {
	// Setup
	mockUserService := new(MockUserService)
	mockEmailRepo := new(MockEmailRepository)
	handler := NewAuthHandler(mockUserService, "test-secret", mockEmailRepo)
	e := echo.New()

	// Missing password
	reqBody := `{
        "email": "test@example.com"
    }`

	// Execute
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assert
	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Validation error")
	}
}
