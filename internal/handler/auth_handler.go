package handler

import (
	"fmt"

	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/email"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/Yoochan45/go-game-rental-api/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	userService service.UserService
	jwtSecret   string
	validate    *validator.Validate
	emailRepo   email.EmailRepository
}

func NewAuthHandler(userService service.UserService, jwtSecret string, emailRepo email.EmailRepository) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		validate:    utils.GetValidator(),
		emailRepo:   emailRepo,
	}
}

// Register godoc
// @Summary Register user
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input or validation error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		logrus.WithError(err).Error("Registration failed")
		return myResponse.BadRequest(c, err.Error())
	}

	// Send email verification with actual token
	go func() {
		// Get the verification token that was created in Register
		token, err := h.userService.GetVerificationToken(user.ID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get verification token")
			return
		}

		verificationURL := fmt.Sprintf("http://localhost:8080/auth/verify?token=%s", token)
		subject := "Verify Your Email - Game Rental"
		htmlContent := fmt.Sprintf(`
			<h1>Welcome %s!</h1>
			<p>Thank you for registering at Game Rental Platform.</p>
			<p>Please click the link below to verify your email:</p>
			<a href="%s">Verify Email</a>
			<p>This link will expire in 24 hours.</p>
		`, user.FullName, verificationURL)
		plainText := "Welcome " + user.FullName + "! Please verify your email by visiting: " + verificationURL

		if err := h.emailRepo.SendEmail(c.Request().Context(), user.Email, subject, plainText, htmlContent); err != nil {
			logrus.WithError(err).Error("Failed to send verification email")
		}
	}()

	return myResponse.Created(c, "User registered successfully. Please check your email to verify your account.", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"message": "Verification email sent",
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	response, err := h.userService.Login(&req, h.jwtSecret)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Error("Login failed")
		return myResponse.Unauthorized(c, err.Error())
	}

	return myResponse.Success(c, "Login successful", response)
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user email with verification token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]interface{} "Email verified successfully"
// @Failure 400 {object} map[string]interface{} "Invalid or expired token"
// @Router /auth/verify [get]
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return myResponse.BadRequest(c, "Verification token is required")
	}

	user, err := h.userService.VerifyEmail(token)
	if err != nil {
		logrus.WithError(err).Error("Email verification failed")
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Email verified successfully. You can now login.", map[string]interface{}{
		"user_id":  user.ID,
		"email":    user.Email,
		"verified": true,
	})
}

// ResendVerification godoc
// @Summary Resend email verification
// @Description Resend verification email to user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "Email to resend verification"
// @Success 200 {object} map[string]interface{} "Verification email sent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input or user already verified"
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c echo.Context) error {
	var req dto.ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	err := h.userService.ResendVerification(req.Email)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Verification email sent successfully", nil)
}
