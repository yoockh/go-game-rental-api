package handler

import (
	"time"

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
	var req dto.CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	// Parse date strings
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return myResponse.BadRequest(c, "Invalid start_date format (use YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return myResponse.BadRequest(c, "Invalid end_date format (use YYYY-MM-DD)")
	}

	userID := echomw.CurrentUserID(c)
	bookingData := &model.Booking{
		UserID:    userID,
		GameID:    req.GameID,
		StartDate: startDate,
		EndDate:   endDate,
		Notes:     utils.PtrOrNil(req.Notes),
	}

	err = h.bookingService.Create(userID, bookingData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Booking created successfully", bookingData)
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

	bookings, total, err := h.bookingService.GetUserBookings(userID, params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve bookings")
	}

	meta := utils.CreateMeta(params, total)
	return myResponse.Paginated(c, "Bookings retrieved successfully", bookings, meta)
}

// GetBookingDetail godoc
// @Summary Get booking detail
// @Description Get detailed information about a specific booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Booking retrieved successfully"
// @Router /bookings/{booking_id} [get]
func (h *BookingHandler) GetBookingDetail(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	bookingID := myRequest.PathParamUint(c, "booking_id")

	booking, err := h.bookingService.GetByID(userID, bookingID)
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
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Booking cancelled successfully"
// @Router /bookings/{booking_id}/cancel [patch]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	bookingID := myRequest.PathParamUint(c, "booking_id")

	err := h.bookingService.Cancel(userID, bookingID)
	if err != nil {
		return utils.MapServiceError(c, err)
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

	bookings, total, err := h.bookingService.GetAll(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return utils.MapServiceError(c, err)
	}

	meta := utils.CreateMeta(params, total)
	return myResponse.Paginated(c, "Bookings retrieved successfully", bookings, meta)
}

// UpdateBookingStatus godoc
// @Summary Update booking status
// @Description Update booking status (Admin only)
// @Tags Admin - Bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Param request body object{status=string} true "New status"
// @Success 200 {object} map[string]interface{} "Booking status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/bookings/{id}/status [patch]
func (h *BookingHandler) UpdateBookingStatus(c echo.Context) error {
	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	var req struct {
		Status model.BookingStatus `json:"status" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}

	role := echomw.CurrentRole(c)
	err := h.bookingService.UpdateStatus(model.UserRole(role), bookingID, req.Status)
	if err != nil {
		return utils.MapServiceError(c, err)
	}

	return myResponse.Success(c, "Booking status updated successfully", nil)
}
