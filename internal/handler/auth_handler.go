package handler

import (
	"fmt"

	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/integration/email"
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
	emailSender email.EmailSender
}

func NewAuthHandler(userService service.UserService, jwtSecret string, emailSender email.EmailSender) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		validate:    utils.GetValidator(),
		emailSender: emailSender,
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

	// send email notification welcome new user
	go func() {
		subject := "Welcome to Game Rental!"
		htmlContent := fmt.Sprintf(`
			<h1>Welcome %s!</h1>
			<p>Thank you for registering at Game Rental Platform.</p>
			<p>You can now browse and rent games from our partners.</p>
		`, user.FullName)
		plainText := fmt.Sprintf("Welcome %s! Thank you for registering at Game Rental Platform.", user.FullName)

		if err := h.emailSender.SendEmail(c.Request().Context(), user.Email, subject, plainText, htmlContent); err != nil {
			logrus.WithError(err).Error("Failed to send welcome email")
		}
	}()

	return myResponse.Created(c, "User registered successfully", user)
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

	// send email notification login alert
	go func() {
		subject := "Login Notification"
		htmlContent := `
			<h2>Login Alert</h2>
			<p>Someone just logged into your Game Rental account.</p>
			<p>If this wasn't you, please contact support immediately.</p>
		`
		plainText := "Someone just logged into your Game Rental account. If this wasn't you, please contact support."

		if err := h.emailSender.SendEmail(c.Request().Context(), req.Email, subject, plainText, htmlContent); err != nil {
			logrus.WithError(err).Error("Failed to send login notification email")
		}
	}()

	return myResponse.Success(c, "Login successful", response)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh expired JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]interface{} "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	response, err := h.userService.RefreshToken(&req, h.jwtSecret)
	if err != nil {
		return myResponse.Unauthorized(c, err.Error())
	}

	return myResponse.Success(c, "Token refreshed successfully", response)
}
