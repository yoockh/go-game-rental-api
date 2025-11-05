package handler

import (
	"time"

	"github.com/Yoochan45/go-api-utils/pkg-echo/auth"
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/model/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	userService service.UserService
	validate    *validator.Validate
	jwtSecret   string
}

func NewAuthHandler(userService service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validate:    validator.New(),
		jwtSecret:   jwtSecret,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}
	if !myRequest.ValidateEmail(c, req.Email) {
		return nil
	}

	// Hash password using go-api-utils
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to process password")
	}

	// Create user data directly using model.User
	userData := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Phone:    &req.Phone,
		Address:  &req.Address,
		Role:     "customer",
		IsActive: true,
	}

	// CreateUser only returns error, not (user, error)
	err = h.userService.CreateUser(userData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Use GetUserByID instead of GetUserByEmail (since it doesn't exist)
	// userData should have ID after creation, or we get by email from repo directly
	user, err := h.userService.GetProfile(userData.ID)
	if err != nil {
		return myResponse.InternalServerError(c, "User created but failed to retrieve")
	}

	response := dto.ToUserDTO(user)
	return myResponse.Created(c, "User registered successfully", response)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}
	if !myRequest.ValidateEmail(c, req.Email) {
		return nil
	}

	user, err := h.userService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		return myResponse.Unauthorized(c, "Invalid credentials")
	}

	// Generate JWT token using go-api-utils
	token, err := auth.GenerateCustomToken(map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    string(user.Role),
	}, h.jwtSecret, 24*time.Hour)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to generate token")
	}

	// Generate refresh token (longer expiry)
	refreshToken, err := auth.GenerateCustomToken(map[string]any{
		"user_id": user.ID,
		"type":    "refresh",
	}, h.jwtSecret, 7*24*time.Hour) // 7 days
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to generate refresh token")
	}

	response := &dto.LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         dto.ToUserDTO(user),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	return myResponse.Success(c, "Login successful", response)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req dto.RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	// Validate refresh token
	claims, err := auth.ValidateCustomToken(req.RefreshToken, h.jwtSecret)
	if err != nil {
		return myResponse.Unauthorized(c, "Invalid refresh token")
	}

	userID := myRequest.GetUint(claims, "user_id")
	tokenType := myRequest.GetString(claims, "type")

	if tokenType != "refresh" || userID == 0 {
		return myResponse.Unauthorized(c, "Invalid refresh token")
	}

	// Get user data - use GetProfile instead of GetUserByID
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return myResponse.Unauthorized(c, "User not found")
	}

	// Generate new access token
	newToken, err := auth.GenerateCustomToken(map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    string(user.Role),
	}, h.jwtSecret, 24*time.Hour)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to generate token")
	}

	response := map[string]any{
		"access_token": newToken,
		"expires_at":   time.Now().Add(24 * time.Hour),
	}

	return myResponse.Success(c, "Token refreshed successfully", response)
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return myResponse.NotFound(c, "User not found")
	}

	response := dto.ToUserDTO(user)
	return myResponse.Success(c, "Profile retrieved successfully", response)
}

func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	// Create update data using model.User
	updateData := &model.User{
		FullName: req.FullName,
		Phone:    &req.Phone,
		Address:  &req.Address,
	}

	// UpdateProfile only returns error, not (user, error)
	err := h.userService.UpdateProfile(userID, updateData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated user for response
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Profile updated but failed to retrieve")
	}

	response := dto.ToUserDTO(user)
	return myResponse.Success(c, "Profile updated successfully", response)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to process password")
	}

	err = h.userService.ChangePassword(userID, req.CurrentPassword, hashedPassword)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Password changed successfully", nil)
}
