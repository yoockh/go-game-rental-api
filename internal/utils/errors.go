package utils

import (
	"errors"

	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/labstack/echo/v4"
)

// MapServiceError maps service errors to appropriate HTTP responses
func MapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrBookingNotFound),
		errors.Is(err, service.ErrGameNotFound),
		errors.Is(err, service.ErrUserNotFound):
		return myResponse.NotFound(c, err.Error())
	
	case errors.Is(err, service.ErrBookingNotOwned),
		errors.Is(err, service.ErrGameNotOwned),
		errors.Is(err, service.ErrInsufficientPermission),
		errors.Is(err, service.ErrGameInsufficientPermission):
		return myResponse.Forbidden(c, err.Error())
	
	case errors.Is(err, service.ErrBookingCannotCancel),
		errors.Is(err, service.ErrBookingCannotConfirm):
		return myResponse.BadRequest(c, err.Error()) // Could be 409 Conflict
	
	default:
		return myResponse.BadRequest(c, err.Error())
	}
}