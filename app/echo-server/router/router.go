package router

import (
	myMiddleware "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	"github.com/Yoochan45/go-game-rental-api/internal/handler"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(
	e *echo.Echo,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	categoryH *handler.CategoryHandler,
	gameH *handler.GameHandler,
	bookingH *handler.BookingHandler,
	paymentH *handler.PaymentHandler,
	reviewH *handler.ReviewHandler,
	partnerH *handler.PartnerHandler,
	adminH *handler.AdminHandler,

	jwtSecret string,
) {
	// Public endpoints (no authentication required)
	e.POST("/auth/register", authH.Register)
	e.POST("/auth/login", authH.Login)
	e.GET("/auth/verify", authH.VerifyEmail)

	// Public game catalog
	e.GET("/games", gameH.GetAllGames)
	e.GET("/games/:id", gameH.GetGameDetail)
	e.GET("/games/search", gameH.SearchGames)

	// Public categories
	e.GET("/categories", categoryH.GetAllCategories)
	e.GET("/categories/:id", categoryH.GetCategoryDetail)

	// Public game reviews
	e.GET("/games/:game_id/reviews", reviewH.GetGameReviews)

	// Payment webhook (public but validated by provider)
	e.POST("/webhooks/payments", paymentH.PaymentWebhook)

	// JWT Middleware Configuration
	jwtConfig := myMiddleware.JWTConfig{
		SecretKey:      jwtSecret,
		UseCustomToken: false, // ubah ke false, kecuali middleware custom kamu butuh true
	}

	// Protected routes - require authentication
	protected := e.Group("")
	protected.Use(myMiddleware.JWTMiddleware(jwtConfig))

	// User profile
	protected.GET("/users/me", userH.GetMyProfile)
	protected.PUT("/users/me", userH.UpdateMyProfile)

	// Partner application (customer harus bisa akses)
	protected.POST("/partner/apply", partnerH.ApplyPartner)

	// Customer bookings
	protected.POST("/bookings", bookingH.CreateBooking)
	protected.GET("/bookings/my", bookingH.GetMyBookings)
	protected.GET("/bookings/:booking_id", bookingH.GetBookingDetail)
	protected.PATCH("/bookings/:booking_id/cancel", bookingH.CancelBooking)

	// Payments (create & get by booking)
	protected.POST("/bookings/:booking_id/payments", paymentH.CreatePayment)
	protected.GET("/bookings/:booking_id/payments", paymentH.GetPaymentByBooking)
	// protected.GET("/payments/:id", paymentH.GetPaymentDetail) // moved to admin scope

	// Reviews
	protected.POST("/bookings/:booking_id/reviews", reviewH.CreateReview)

	// Partner routes (requires partner/admin/super_admin)
	partner := protected.Group("/partner")
	partner.Use(myMiddleware.RequireRoles("partner", "admin", "super_admin"))
	partner.GET("/bookings", partnerH.GetPartnerBookings)
	partner.PATCH("/bookings/:booking_id/confirm-handover", partnerH.ConfirmHandover)
	partner.PATCH("/bookings/:booking_id/confirm-return", partnerH.ConfirmReturn)
	partner.POST("/games", gameH.CreateGame)
	partner.PUT("/games/:id", gameH.UpdateGame)
	partner.GET("/games", gameH.GetPartnerGames)
	partner.POST("/games/:id/upload-image", gameH.UploadGameImage)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(myMiddleware.RequireRoles("admin", "super_admin"))

	// Payment detail (admin only)
	admin.GET("/payments/:id", paymentH.GetPaymentDetail)

	// User management
	admin.GET("/users", userH.GetAllUsers)
	admin.GET("/users/:id", userH.GetUserDetail)
	admin.PATCH("/users/:id/role", userH.UpdateUserRole)
	admin.PATCH("/users/:id/status", userH.ToggleUserStatus)
	admin.DELETE("/users/:id", userH.DeleteUser)

	// Partner application management
	admin.GET("/partner-applications", adminH.GetPartnerApplications)
	admin.PATCH("/partner-applications/:id/approve", adminH.ApprovePartnerApplication)
	admin.PATCH("/partner-applications/:id/reject", adminH.RejectPartnerApplication)

	// Game listing management
	admin.GET("/listings", adminH.GetGameListings)
	admin.PATCH("/listings/:id/approve", adminH.ApproveGameListing)

	// Category management
	admin.POST("/categories", categoryH.CreateCategory)
	admin.PUT("/categories/:id", categoryH.UpdateCategory)
	admin.DELETE("/categories/:id", categoryH.DeleteCategory)

	// Booking management
	admin.GET("/bookings", bookingH.GetAllBookings)

	// Payment management
	admin.GET("/payments", paymentH.GetAllPayments)
	admin.GET("/payments/status", paymentH.GetPaymentsByStatus)

	// Super Admin routes (requires super_admin role only)
	superAdmin := protected.Group("/superadmin")
	superAdmin.Use(myMiddleware.RequireRoles("super_admin"))
}
