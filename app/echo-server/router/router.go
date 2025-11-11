package router

import (
	"github.com/labstack/echo/v4"
	myMiddleware "github.com/yoockh/go-api-utils/pkg-echo/middleware"
	"github.com/yoockh/go-game-rental-api/internal/handler"
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
	jwtSecret string,
) {
	// Public endpoints
	e.POST("/auth/register", authH.Register)
	e.POST("/auth/login", authH.Login)
	e.GET("/games", gameH.GetAllGames)
	e.GET("/games/:id", gameH.GetGameDetail)
	e.GET("/games/search", gameH.SearchGames)
	e.GET("/categories", categoryH.GetAllCategories)
	e.GET("/categories/:id", categoryH.GetCategoryDetail)
	e.GET("/games/:game_id/reviews", reviewH.GetGameReviews)
	e.POST("/webhooks/payments", paymentH.PaymentWebhook)

	// Protected routes
	jwtConfig := myMiddleware.JWTConfig{
		SecretKey:      jwtSecret,
		UseCustomToken: false,
	}

	protected := e.Group("")
	protected.Use(myMiddleware.JWTMiddleware(jwtConfig))

	protected.GET("/users/me", userH.GetMyProfile)
	protected.PUT("/users/me", userH.UpdateMyProfile)

	protected.POST("/bookings", bookingH.CreateBooking)
	protected.GET("/bookings/my", bookingH.GetMyBookings)
	protected.GET("/bookings/:booking_id", bookingH.GetBookingDetail)
	protected.PATCH("/bookings/:booking_id/cancel", bookingH.CancelBooking)

	protected.POST("/bookings/:booking_id/payments", paymentH.CreatePayment)
	protected.GET("/bookings/:booking_id/payments", paymentH.GetPaymentByBooking)

	protected.POST("/bookings/:booking_id/reviews", reviewH.CreateReview)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(myMiddleware.RequireRoles("admin", "super_admin")) // BALIK PAKAI INI

	admin.POST("/games", gameH.CreateGame)
	admin.PUT("/games/:id", gameH.UpdateGame)
	admin.DELETE("/games/:id", gameH.DeleteGame)

	admin.POST("/categories", categoryH.CreateCategory)
	admin.PUT("/categories/:id", categoryH.UpdateCategory)
	admin.DELETE("/categories/:id", categoryH.DeleteCategory)

	admin.GET("/bookings", bookingH.GetAllBookings)
	admin.PATCH("/bookings/:id/status", bookingH.UpdateBookingStatus)

	admin.GET("/payments", paymentH.GetAllPayments)
	admin.GET("/payments/:id", paymentH.GetPaymentDetail)
	admin.GET("/payments/status", paymentH.GetPaymentsByStatus)

	admin.GET("/users", userH.GetAllUsers)
	admin.GET("/users/:id", userH.GetUserDetail)
	admin.PATCH("/users/:id/role", userH.UpdateUserRole)
	admin.PATCH("/users/:id/status", userH.ToggleUserStatus)
	admin.DELETE("/users/:id", userH.DeleteUser)
}
