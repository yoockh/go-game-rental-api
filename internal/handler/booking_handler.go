package handler

import (
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/Yoochan45/go-game-rental-api/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type BookingHandler struct {
	bookingService service.BookingService
	validate       *validator.Validate
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		validate:       utils.GetValidator(),
	}
}

// CreateBooking godoc
// @Summary Create booking
// @Description Create a new game rental booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBookingRequest true "Booking details"
// @Success 201 {object} map[string]interface{} "Booking created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	var req dto.CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	bookingData := &model.Booking{
		UserID:    userID,
		GameID:    req.GameID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Notes:     utils.PtrOrNil(req.Notes),
	}

	err := h.bookingService.CreateBooking(userID, bookingData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Booking created successfully", nil)
}

// GetMyBookings godoc
// @Summary Get my bookings
// @Description Get list of current user's bookings
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Bookings retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings/my [get]
func (h *BookingHandler) GetMyBookings(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	params := utils.ParsePagination(c)

	bookings, err := h.bookingService.GetUserBookings(userID, params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve bookings")
	}

	meta := utils.CreateMeta(params, int64(len(bookings)))
	return myResponse.Paginated(c, "Bookings retrieved successfully", bookings, meta)
}

// GetBookingDetail godoc
// @Summary Get booking detail
// @Description Get detailed information about a specific booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} dto.BookingDTO "Booking retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Booking not found"
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBookingDetail(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	booking, err := h.bookingService.GetBookingDetail(userID, bookingID)
	if err != nil {
		return myResponse.NotFound(c, err.Error())
	}

	return myResponse.Success(c, "Booking retrieved successfully", booking)
}

// CancelBooking godoc
// @Summary Cancel booking
// @Description Cancel a pending booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Booking cancelled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID or booking cannot be cancelled"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings/{id}/cancel [patch]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	err := h.bookingService.CancelBooking(userID, bookingID)
	if err != nil {
		switch err {
		case service.ErrBookingNotFound:
			return myResponse.NotFound(c, err.Error())
		case service.ErrBookingNotOwned:
			return myResponse.Forbidden(c, err.Error())
		case service.ErrBookingCannotCancel:
			return myResponse.BadRequest(c, err.Error()) // 409 would be better but keeping 400
		default:
			return myResponse.BadRequest(c, err.Error())
		}
	}

	return myResponse.Success(c, "Booking cancelled successfully", nil)
}

// Admin endpoints
// GetAllBookings godoc
// @Summary Get all bookings
// @Description Get list of all bookings (Admin only)
// @Tags Admin - Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Bookings retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/bookings [get]
func (h *BookingHandler) GetAllBookings(c echo.Context) error {
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	bookings, err := h.bookingService.GetAllBookings(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(bookings)))
	return myResponse.Paginated(c, "Bookings retrieved successfully", bookings, meta)
}

// GetBookingDetailAdmin godoc
// @Summary Get booking detail (Admin)
// @Description Get detailed information about a specific booking (Admin only)
// @Tags Admin - Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} dto.BookingDTO "Booking retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Booking not found"
// @Router /admin/bookings/{id} [get]
func (h *BookingHandler) GetBookingDetailAdmin(c echo.Context) error {
	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	booking, err := h.bookingService.GetBookingDetail(0, bookingID)
	if err != nil {
		return myResponse.NotFound(c, err.Error())
	}

	return myResponse.Success(c, "Booking retrieved successfully", booking)
}
