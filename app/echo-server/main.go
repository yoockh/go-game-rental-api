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
	"time"

	myOrm "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
	myConfig "github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-game-rental-api/app/echo-server/router"
	_ "github.com/Yoochan45/go-game-rental-api/docs"
	"github.com/Yoochan45/go-game-rental-api/internal/handler"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/email"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/storage"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/transaction"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
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

	// Add parameters to avoid prepared statement cache in pooler
	dbURL := cfg.DatabaseURL + "&statement_cache_mode=describe&prefer_simple_protocol=true"
	db, err := myOrm.Init(dbURL)
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}

	// Disable prepared statements completely
	db = db.Session(&gorm.Session{PrepareStmt: false})

	// Configure connection pool with shorter lifetime to force reconnect
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatal("Failed to get underlying sql.DB:", err)
	}
	sqlDB.SetMaxOpenConns(1)                  // Limit to 1 connection
	sqlDB.SetMaxIdleConns(0)                  // No idle connections
	sqlDB.SetConnMaxLifetime(1 * time.Second) // Force reconnect every second

	// Auto migrate all models
	err = db.AutoMigrate(
		&model.User{},
		&model.EmailVerificationToken{},
		&model.Category{},
		&model.Game{},
		&model.Booking{},
		&model.Payment{},
		&model.Review{},
		&model.PartnerApplication{},
	)
	if err != nil {
		logrus.Warn("Migration warning:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	verificationRepo := repository.NewEmailVerificationRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	gameRepo := repository.NewGameRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	partnerRepo := repository.NewPartnerApplicationRepository(db)

	// Initialize 3rd party repositories with fallback to mock
	var emailRepo email.EmailRepository
	var storageRepo storage.StorageRepository
	var transactionRepo transaction.TransactionRepository

	// Try real repositories, fallback to mock on error
	if repo, err := email.NewSendGridRepository(); err != nil {
		logrus.Warn("SendGrid failed, using mock:", err)
		emailRepo = &email.MockEmailRepository{}
	} else {
		emailRepo = repo
	}

	if repo, err := storage.NewSupabaseRepository(); err != nil {
		logrus.Warn("Supabase failed, using mock:", err)
		storageRepo = &storage.MockStorageRepository{}
	} else {
		storageRepo = repo
	}

	if repo, err := transaction.NewMidtransRepository(); err != nil {
		logrus.Warn("Midtrans failed, using mock:", err)
		transactionRepo = &transaction.MockTransactionRepository{}
	} else {
		transactionRepo = repo
	}

	// Initialize services
	userService := service.NewUserService(userRepo, verificationRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	gameService := service.NewGameService(gameRepo, userRepo)
	bookingService := service.NewBookingService(bookingRepo, gameRepo, userRepo)
	paymentService := service.NewPaymentService(paymentRepo, bookingRepo, userRepo, bookingService, transactionRepo)
	reviewService := service.NewReviewService(reviewRepo, bookingRepo)
	partnerService := service.NewPartnerApplicationService(partnerRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, JwtSecret, emailRepo)
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	gameHandler := handler.NewGameHandler(gameService, storageRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	partnerHandler := handler.NewPartnerHandler(partnerService, bookingService)
	adminHandler := handler.NewAdminHandler(partnerService, gameService)

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

		JwtSecret,
	)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logrus.Infof("Server starting on :%s", port)
	logrus.Fatal(e.Start(":" + port))
}
