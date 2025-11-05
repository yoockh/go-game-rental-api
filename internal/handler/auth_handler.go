package handler

import (
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
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
}

func NewAuthHandler(userService service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		validate:    utils.GetValidator(),
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

	user, err := h.userService.Register(&req)
	if err != nil {
		logrus.WithError(err).Error("Registration failed")
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "User registered successfully", user)
}

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
