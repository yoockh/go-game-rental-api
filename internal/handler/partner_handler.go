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

type PartnerHandler struct {
	partnerService service.PartnerApplicationService
	bookingService service.BookingService
	validate       *validator.Validate
}

func NewPartnerHandler(partnerService service.PartnerApplicationService, bookingService service.BookingService) *PartnerHandler {
	return &PartnerHandler{
		partnerService: partnerService,
		bookingService: bookingService,
		validate:       validator.New(),
	}
}

// ApplyPartner godoc
// @Summary Apply for partner
// @Description Submit application to become a game rental partner
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePartnerApplicationRequest true "Partner application details"
// @Success 201 {object} map[string]interface{} "Partner application submitted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /partner/apply [post]
func (h *PartnerHandler) ApplyPartner(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	var req dto.CreatePartnerApplicationRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	applicationData := &model.PartnerApplication{
		BusinessName:        req.BusinessName,
		BusinessAddress:     req.BusinessAddress,
		BusinessPhone:       &req.BusinessPhone,
		BusinessDescription: &req.BusinessDescription,
	}

	err := h.partnerService.SubmitApplication(userID, applicationData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Partner application submitted successfully", nil)
}

// GetPartnerBookings godoc
// @Summary Get partner bookings
// @Description Get list of bookings for partner's games
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Partner bookings retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Partner role required"
// @Router /partner/bookings [get]
func (h *PartnerHandler) GetPartnerBookings(c echo.Context) error {
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

	bookings, err := h.bookingService.GetPartnerBookings(userID, limit, (page-1)*limit)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve bookings")
	}

	totalCount := int64(len(bookings))
	meta := map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	return myResponse.Paginated(c, "Partner bookings retrieved successfully", bookings, meta)
}

// ConfirmHandover godoc
// @Summary Confirm game handover
// @Description Confirm that game has been handed over to customer
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Handover confirmed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Partner role required"
// @Router /partner/bookings/{id}/confirm-handover [patch]
func (h *PartnerHandler) ConfirmHandover(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	err := h.bookingService.ConfirmHandover(userID, bookingID)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Handover confirmed successfully", nil)
}

// ConfirmReturn godoc
// @Summary Confirm game return
// @Description Confirm that game has been returned by customer
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]interface{} "Return confirmed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Partner role required"
// @Router /partner/bookings/{id}/confirm-return [patch]
func (h *PartnerHandler) ConfirmReturn(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	err := h.bookingService.ConfirmReturn(userID, bookingID)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Return confirmed successfully", nil)
}
