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

type ReviewHandler struct {
	reviewService service.ReviewService
	validate      *validator.Validate
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		validate:      validator.New(),
	}
}

// CreateReview godoc
// @Summary Create review
// @Description Create a review for a completed booking
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking_id path int true "Booking ID"
// @Param request body dto.CreateReviewRequest true "Review details"
// @Success 201 {object} map[string]interface{} "Review created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings/{booking_id}/reviews [post]
func (h *ReviewHandler) CreateReview(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "booking_id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	var req dto.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	reviewData := &model.Review{
		Rating:  req.Rating,
		Comment: &req.Comment,
	}

	err := h.reviewService.CreateReview(userID, bookingID, reviewData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Review created successfully", nil)
}

// GetGameReviews godoc
// @Summary Get game reviews
// @Description Get list of reviews for a specific game
// @Tags Reviews
// @Accept json
// @Produce json
// @Param game_id path int true "Game ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Reviews retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid game ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /games/{game_id}/reviews [get]
func (h *ReviewHandler) GetGameReviews(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "game_id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	page := myRequest.QueryInt(c, "page", 1)
	limit := myRequest.QueryInt(c, "limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	reviews, err := h.reviewService.GetGameReviews(gameID, limit, (page-1)*limit)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve reviews")
	}

	totalCount := int64(len(reviews))
	meta := map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	return myResponse.Paginated(c, "Reviews retrieved successfully", reviews, meta)
}
