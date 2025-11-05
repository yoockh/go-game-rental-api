// @title Video Game Rental API
// @version 1.0
// @description Video Game Rental API adalah sistem backend untuk platform penyewaan game fisik
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"os"

	myOrm "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
	myConfig "github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-game-rental-api/app/echo-server/router"
	"github.com/Yoochan45/go-game-rental-api/internal/handler"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "github.com/Yoochan45/go-game-rental-api/docs"
)

func main() {
	// Setup logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	cfg := myConfig.LoadEnv()
	JwtSecret := os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		JwtSecret = "dev-secret"
		logrus.Warn("Using default JWT secret for development")
	}

	db, err := myOrm.Init(cfg.DatabaseURL)
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.Category{},
		&model.Game{},
		&model.Booking{},
		&model.Payment{},
		&model.Review{},
		&model.PartnerApplication{},
		&model.Dispute{},
	)
	if err != nil {
		logrus.Warn("Migration warning:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	gameRepo := repository.NewGameRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	partnerRepo := repository.NewPartnerApplicationRepository(db)
	disputeRepo := repository.NewDisputeRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	gameService := service.NewGameService(gameRepo, userRepo)
	bookingService := service.NewBookingService(bookingRepo, gameRepo, userRepo)
	paymentService := service.NewPaymentService(paymentRepo, bookingRepo, userRepo, bookingService)
	reviewService := service.NewReviewService(reviewRepo, bookingRepo)
	partnerService := service.NewPartnerApplicationService(partnerRepo, userRepo)
	disputeService := service.NewDisputeService(disputeRepo, bookingRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, JwtSecret)
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	gameHandler := handler.NewGameHandler(gameService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	partnerHandler := handler.NewPartnerHandler(partnerService, bookingService)
	adminHandler := handler.NewAdminHandler(partnerService, gameService)
	disputeHandler := handler.NewDisputeHandler(disputeService)

	// Setup Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Register routes
	router.RegisterRoutes(
		e,
		authHandler,
		userHandler,
		categoryHandler,
		gameHandler,
		bookingHandler,
		paymentHandler,
		reviewHandler,
		partnerHandler,
		adminHandler,
		disputeHandler,
		JwtSecret,
	)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logrus.Infof("Server starting on :%s", port)
	logrus.Fatal(e.Start(":" + port))
}
