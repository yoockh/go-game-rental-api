package handler

import (
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/Yoochan45/go-game-rental-api/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	partnerService service.PartnerApplicationService
	gameService    service.GameService
	validate       *validator.Validate
}

func NewAdminHandler(partnerService service.PartnerApplicationService, gameService service.GameService) *AdminHandler {
	return &AdminHandler{
		partnerService: partnerService,
		gameService:    gameService,
		validate:       utils.GetValidator(),
	}
}

// GetPartnerApplications godoc
// @Summary Get partner applications
// @Description Get list of all partner applications (Admin only)
// @Tags Admin - Partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Partner applications retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/partner-applications [get]
func (h *AdminHandler) GetPartnerApplications(c echo.Context) error {
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	applications, err := h.partnerService.GetAllApplications(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(applications)))
	return myResponse.Paginated(c, "Partner applications retrieved successfully", applications, meta)
}

// ApprovePartnerApplication godoc
// @Summary Approve partner application
// @Description Approve a pending partner application (Admin only)
// @Tags Admin - Partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Success 200 {object} map[string]interface{} "Partner application approved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid application ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/partner-applications/{id}/approve [patch]
func (h *AdminHandler) ApprovePartnerApplication(c echo.Context) error {
	applicationID := myRequest.PathParamUint(c, "id")
	if applicationID == 0 {
		return myResponse.BadRequest(c, "Invalid application ID")
	}

	adminID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	err := h.partnerService.ApproveApplication(adminID, model.UserRole(role), applicationID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Partner application approved successfully", nil)
}

// RejectPartnerApplication godoc
// @Summary Reject partner application
// @Description Reject a pending partner application (Admin only)
// @Tags Admin - Partners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Param request body object true "Rejection reason"
// @Success 200 {object} map[string]interface{} "Partner application rejected successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/partner-applications/{id}/reject [patch]
func (h *AdminHandler) RejectPartnerApplication(c echo.Context) error {
	applicationID := myRequest.PathParamUint(c, "id")
	if applicationID == 0 {
		return myResponse.BadRequest(c, "Invalid application ID")
	}

	var req struct {
		RejectionReason string `json:"rejection_reason" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	adminID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	err := h.partnerService.RejectApplication(adminID, model.UserRole(role), applicationID, req.RejectionReason)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Partner application rejected successfully", nil)
}

// GetGameListings godoc
// @Summary Get game listings
// @Description Get list of all game listings for approval (Admin only)
// @Tags Admin - Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Game listings retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/listings [get]
func (h *AdminHandler) GetGameListings(c echo.Context) error {
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	games, err := h.gameService.GetAllGames(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(games)))
	return myResponse.Paginated(c, "Game listings retrieved successfully", games, meta)
}

// ApproveGameListing godoc
// @Summary Approve game listing
// @Description Approve a pending game listing (Admin only)
// @Tags Admin - Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} map[string]interface{} "Game listing approved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid game ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/listings/{id}/approve [patch]
func (h *AdminHandler) ApproveGameListing(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	adminID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	err := h.gameService.ApproveGame(adminID, model.UserRole(role), gameID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Game listing approved successfully", nil)
}
