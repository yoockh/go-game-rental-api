package handler

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	myResponse "github.com/yoockh/go-api-utils/pkg-echo/response"
	"github.com/yoockh/go-game-rental-api/internal/dto"
	"github.com/yoockh/go-game-rental-api/internal/repository/email"
	"github.com/yoockh/go-game-rental-api/internal/service"
	"github.com/yoockh/go-game-rental-api/internal/utils"
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

	// Send welcome email (no verification needed)
	go func() {
		subject := "Welcome to Game Rental Platform"
		htmlContent := fmt.Sprintf(`
			<h1>Welcome %s!</h1>
			<p>Thank you for registering at Game Rental Platform.</p>
			<p>Your account is now active and ready to use.</p>
			<p>Start browsing our game collection and make your first rental!</p>
		`, user.FullName)
		plainText := fmt.Sprintf("Welcome %s! Your account is now active.", user.FullName)

		if err := h.emailRepo.SendEmail(c.Request().Context(), user.Email, subject, plainText, htmlContent); err != nil {
			logrus.WithError(err).Error("Failed to send welcome email")
		}
	}()

	return myResponse.Created(c, "User registered successfully. You can now login.", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
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
