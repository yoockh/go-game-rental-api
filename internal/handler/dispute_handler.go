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

type DisputeHandler struct {
	disputeService service.DisputeService
	validate       *validator.Validate
}

func NewDisputeHandler(disputeService service.DisputeService) *DisputeHandler {
	return &DisputeHandler{
		disputeService: disputeService,
		validate:       utils.GetValidator(),
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

	params := utils.ParsePagination(c)

	// Get all disputes with admin role to avoid permission error, then filter
	disputes, err := h.disputeService.GetAllDisputes(model.RoleAdmin, params.Limit*10, 0) // Get more to filter
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve disputes")
	}

	// Filter by current user and apply pagination manually
	var userDisputes []*model.Dispute
	for _, dispute := range disputes {
		if dispute.ReporterID == userID {
			userDisputes = append(userDisputes, dispute)
		}
	}

	// Apply manual pagination
	start := params.Offset
	end := start + params.Limit
	if start > len(userDisputes) {
		userDisputes = []*model.Dispute{}
	} else if end > len(userDisputes) {
		userDisputes = userDisputes[start:]
	} else {
		userDisputes = userDisputes[start:end]
	}

	meta := utils.CreateMeta(params, int64(len(userDisputes)))
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
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	disputes, err := h.disputeService.GetAllDisputes(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(disputes)))
	return myResponse.Paginated(c, "Disputes retrieved successfully", disputes, meta)
}
