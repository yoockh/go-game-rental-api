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
	e.POST("/auth/refresh", authH.RefreshToken)

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
		UseCustomToken: true,
	}

	// Protected routes - require authentication
	protected := e.Group("")
	protected.Use(myMiddleware.JWTMiddleware(jwtConfig))

	// User profile routes (all authenticated users can access)
	protected.GET("/users/me", userH.GetMyProfile)
	protected.PUT("/users/me", userH.UpdateMyProfile)

	// Customer routes (all authenticated users can access)
	protected.POST("/bookings", bookingH.CreateBooking)
	protected.GET("/bookings/my", bookingH.GetMyBookings)
	protected.GET("/bookings/:id", bookingH.GetBookingDetail)
	protected.PATCH("/bookings/:id/cancel", bookingH.CancelBooking)

	protected.POST("/bookings/:booking_id/payments", paymentH.CreatePayment)
	protected.GET("/bookings/:booking_id/payments", paymentH.GetPaymentByBooking)
	protected.GET("/payments/:id", paymentH.GetPaymentDetail)

	protected.POST("/bookings/:booking_id/reviews", reviewH.CreateReview)



	// Partner routes (requires partner, admin, or super_admin role)
	partner := protected.Group("/partner")
	partner.Use(myMiddleware.RequireRoles("partner", "admin", "super_admin"))

	partner.POST("/apply", partnerH.ApplyPartner)
	partner.GET("/bookings", partnerH.GetPartnerBookings)
	partner.PATCH("/bookings/:id/confirm-handover", partnerH.ConfirmHandover)
	partner.PATCH("/bookings/:id/confirm-return", partnerH.ConfirmReturn)

	// Partner game management
	partner.POST("/games", gameH.CreateGame)
	partner.PUT("/games/:id", gameH.UpdateGame)
	partner.GET("/games", gameH.GetPartnerGames)
	partner.POST("/games/:id/upload-image", gameH.UploadGameImage)

	// Admin routes (requires admin or super_admin role)
	admin := protected.Group("/admin")
	admin.Use(myMiddleware.RequireRoles("admin", "super_admin"))

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

	// LATER
}
