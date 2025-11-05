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

type PaymentHandler struct {
	paymentService service.PaymentService
	validate       *validator.Validate
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		validate:       utils.GetValidator(),
	}
}

// CreatePayment godoc
// @Summary Create payment
// @Description Create payment for a booking
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking_id path int true "Booking ID"
// @Param request body dto.CreatePaymentRequest true "Payment details"
// @Success 201 {object} model.Payment "Payment created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /bookings/{booking_id}/payments [post]
func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "booking_id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	var req dto.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	payment, err := h.paymentService.CreatePayment(userID, bookingID, req.Provider)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Created(c, "Payment created successfully", payment)
}

// GetPaymentByBooking godoc
// @Summary Get payment by booking
// @Description Get payment information for a specific booking
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking_id path int true "Booking ID"
// @Success 200 {object} model.Payment "Payment retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid booking ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Payment not found"
// @Router /bookings/{booking_id}/payments [get]
func (h *PaymentHandler) GetPaymentByBooking(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	bookingID := myRequest.PathParamUint(c, "booking_id")
	if bookingID == 0 {
		return myResponse.BadRequest(c, "Invalid booking ID")
	}

	payment, err := h.paymentService.GetPaymentByBooking(userID, bookingID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Payment retrieved successfully", payment)
}

// GetPaymentDetail godoc
// @Summary Get payment detail
// @Description Get detailed payment information
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Success 200 {object} model.Payment "Payment retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid payment ID"
// @Failure 404 {object} map[string]interface{} "Payment not found"
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPaymentDetail(c echo.Context) error {
	paymentID := myRequest.PathParamUint(c, "id")
	if paymentID == 0 {
		return myResponse.BadRequest(c, "Invalid payment ID")
	}

	role := echomw.CurrentRole(c)
	payment, err := h.paymentService.GetPaymentDetail(model.UserRole(role), paymentID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Payment retrieved successfully", payment)
}

// PaymentWebhook godoc
// @Summary Payment webhook
// @Description Receive payment status updates from payment provider
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param request body dto.PaymentWebhookRequest true "Webhook payload"
// @Success 200 {object} map[string]interface{} "Webhook processed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid webhook payload"
// @Router /webhooks/payments [post]
func (h *PaymentHandler) PaymentWebhook(c echo.Context) error {
	var req dto.PaymentWebhookRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid webhook payload: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Webhook validation error: "+err.Error())
	}

	// Handle pointer fields properly
	var paymentMethod string
	if req.PaymentMethod != nil {
		paymentMethod = *req.PaymentMethod
	}

	// ProcessWebhook takes individual parameters based on actual interface
	err := h.paymentService.ProcessWebhook(req.ProviderPaymentID, req.Status, paymentMethod, req.FailureReason)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Webhook processed successfully", nil)
}

// Admin endpoints
// GetAllPayments godoc
// @Summary Get all payments
// @Description Get list of all payments (Admin only)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Payments retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/payments [get]
func (h *PaymentHandler) GetAllPayments(c echo.Context) error {
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	payments, err := h.paymentService.GetAllPayments(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(payments)))
	return myResponse.Paginated(c, "Payments retrieved successfully", payments, meta)
}

// GetPaymentsByStatus godoc
// @Summary Get payments by status
// @Description Get list of payments filtered by status (Admin only)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string true "Payment status" Enums(pending, completed, failed, refunded)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Payments retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/payments/status [get]
func (h *PaymentHandler) GetPaymentsByStatus(c echo.Context) error {
	params := utils.ParsePagination(c)
	status := c.QueryParam("status")
	role := echomw.CurrentRole(c)

	payments, err := h.paymentService.GetPaymentsByStatus(model.UserRole(role), model.PaymentStatus(status), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, int64(len(payments)))
	return myResponse.Paginated(c, "Payments retrieved successfully", payments, meta)
}
