package handler

import (
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/model/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type DisputeHandler struct {
	disputeService service.DisputeService
	validate       *validator.Validate
}

func NewDisputeHandler(disputeService service.DisputeService) *DisputeHandler {
	return &DisputeHandler{
		disputeService: disputeService,
		validate:       validator.New(),
	}
}

// CreateDispute godoc
// @Summary Create dispute
// @Description Create a dispute for a booking
// @Tags Disputes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking_id path int true "Booking ID"
// @Param request body dto.CreateDisputeRequest true "Dispute details"
// @Success 201 {object} map[string]interface{} "Dispute created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings/{booking_id}/disputes [post]
func (h *DisputeHandler) CreateDispute(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	// BookingID dari path parameter, bukan dari body
	bookingID := myRequest.PathParamUint(c, "booking_id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	var req dto.CreateDisputeRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	disputeData := &model.Dispute{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
	}

	err := h.disputeService.CreateDispute(userID, bookingID, disputeData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Dispute created successfully", nil)
}

// GetMyDisputes godoc
// @Summary Get my disputes
// @Description Get list of current user's disputes
// @Tags Disputes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Disputes retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /disputes/my [get]
func (h *DisputeHandler) GetMyDisputes(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	page := myRequest.QueryInt(c, "page", 1)
	limit := myRequest.QueryInt(c, "limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	role := echomw.CurrentRole(c)
	disputes, err := h.disputeService.GetAllDisputes(model.UserRole(role), limit, (page-1)*limit)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve disputes")
	}

	// Filter by current user
	var userDisputes []*model.Dispute
	for _, dispute := range disputes {
		// Booking is embedded struct, not pointer
		if dispute.Booking.UserID == userID {
			userDisputes = append(userDisputes, dispute)
		}
	}

	totalCount := int64(len(userDisputes))
	meta := map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	return myResponse.Paginated(c, "Disputes retrieved successfully", userDisputes, meta)
}

// GetAllDisputes godoc
// @Summary Get all disputes
// @Description Get list of all disputes (Admin only)
// @Tags Admin - Disputes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Disputes retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/disputes [get]
func (h *DisputeHandler) GetAllDisputes(c echo.Context) error {
	page := myRequest.QueryInt(c, "page", 1)
	limit := myRequest.QueryInt(c, "limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	role := echomw.CurrentRole(c)
	disputes, err := h.disputeService.GetAllDisputes(model.UserRole(role), limit, (page-1)*limit)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	totalCount := int64(len(disputes))
	meta := map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	return myResponse.Paginated(c, "Disputes retrieved successfully", disputes, meta)
}
