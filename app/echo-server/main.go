// @title Video Game Rental API
// @version 1.0
// @description Video Game Rental API adalah sistem backend untuk platform penyewaan game fisik
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host go-game-rental-3beef3913ef8.herokuapp.com
// @BasePath /
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"os"
	"strings"
	"time"

	myConfig "github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-game-rental-api/app/echo-server/router"
	_ "github.com/Yoochan45/go-game-rental-api/docs"
	"github.com/Yoochan45/go-game-rental-api/internal/handler"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/email"
	"github.com/Yoochan45/go-game-rental-api/internal/repository/transaction"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	// Database connection WITHOUT prepared statements
	dbURL := cfg.DatabaseURL

	// Parse and modify DSN to disable prepared statements
	if !strings.Contains(dbURL, "statement_cache_mode") {
		separator := "?"
		if strings.Contains(dbURL, "?") {
			separator = "&"
		}
		dbURL = dbURL + separator + "statement_cache_mode=describe&prefer_simple_protocol=true"
	}

	logrus.Info("Connecting to database with disabled prepared statements...")

	// Use custom GORM config
	dsn := dbURL
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables prepared statements
	}), &gorm.Config{
		PrepareStmt:            false, // globally disable prepared statements
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	logrus.Info("GORM connected to PostgreSQL (PrepareStmt disabled)")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatal("Failed to get underlying sql.DB:", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(0)
	sqlDB.SetConnMaxLifetime(500 * time.Millisecond)

	logrus.Info("Database configured: PrepareStmt=false, MaxOpenConns=1")

	// COMMENT OUT AutoMigrate - pakai DDL manual saja
	/*
		err = db.AutoMigrate(
			&model.User{},
			&model.Category{},
			&model.Game{},
			&model.Booking{},
			&model.Payment{},
			&model.Review{},
		)
		if err != nil {
			logrus.Warn("Migration warning:", err)
		}
	*/
	logrus.Info("Skipping AutoMigrate - using manual DDL from migrations/ddl.sql")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	gameRepo := repository.NewGameRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// Initialize 3rd party repositories with fallback to mock
	var emailRepo email.EmailRepository
	var transactionRepo transaction.TransactionRepository

	if repo, err := email.NewSendGridRepository(); err != nil {
		logrus.Warn("SendGrid failed, using mock:", err)
		emailRepo = &email.MockEmailRepository{}
	} else {
		emailRepo = repo
	}

	if repo, err := transaction.NewMidtransRepository(); err != nil {
		logrus.Warn("Midtrans failed, using mock:", err)
		transactionRepo = &transaction.MockTransactionRepository{}
	} else {
		transactionRepo = repo
	}

	// Initialize services
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	gameService := service.NewGameService(gameRepo)
	bookingService := service.NewBookingService(bookingRepo, gameRepo, userRepo, emailRepo)
	paymentService := service.NewPaymentService(paymentRepo, bookingRepo, userRepo, gameRepo, bookingService, transactionRepo, emailRepo)
	reviewService := service.NewReviewService(reviewRepo, bookingRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, JwtSecret, emailRepo)
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	gameHandler := handler.NewGameHandler(gameService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	reviewHandler := handler.NewReviewHandler(reviewService)

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
		JwtSecret,
	)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logrus.Infof("Server starting on :%s", port)
	logrus.Fatal(e.Start(":" + port))
}
