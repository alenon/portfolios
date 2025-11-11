package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/config"
	"github.com/lenon/portfolios/internal/database"
	"github.com/lenon/portfolios/internal/handlers"
	"github.com/lenon/portfolios/internal/jobs"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.Connect(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)
	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	// Initialize services
	tokenService := services.NewTokenService(cfg.JWT.Secret)
	emailService := services.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)
	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
		cfg.JWT.RememberMeAccessDuration,
		cfg.JWT.RememberMeRefreshDuration,
	)
	passwordResetService := services.NewPasswordResetService(
		userRepo,
		passwordResetRepo,
		emailService,
		1*time.Hour, // Password reset token validity duration
	)
	portfolioService := services.NewPortfolioService(portfolioRepo, userRepo)
	transactionService := services.NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)
	taxLotService := services.NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo, transactionRepo)

	// Initialize corporate action service
	corporateActionService := services.NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	// Initialize corporate action monitor
	corporateActionMonitor := services.NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	// Initialize background job scheduler
	scheduler := jobs.NewScheduler()
	corporateActionJob := jobs.NewCorporateActionDetectionJob(corporateActionMonitor)
	scheduler.AddJob(corporateActionJob)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		authService,
		passwordResetService,
		userRepo,
		int(cfg.JWT.AccessTokenDuration.Seconds()),
	)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	taxLotHandler := handlers.NewTaxLotHandler(taxLotService)
	portfolioActionHandler := handlers.NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo, corporateActionService)

	// Set up Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS(cfg.Server.CORSOrigins))

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.Security.RateLimitRequests, cfg.Security.RateLimitDuration)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes with rate limiting
		auth := api.Group("/auth")
		auth.Use(rateLimiter.Middleware())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)

			// Protected routes
			authenticated := auth.Group("")
			authenticated.Use(middleware.AuthRequired(tokenService))
			{
				authenticated.POST("/logout", authHandler.Logout)
				authenticated.GET("/me", authHandler.GetCurrentUser)
			}
		}

		// API v1 routes (protected)
		v1 := api.Group("/v1")
		v1.Use(middleware.AuthRequired(tokenService))
		{
			// Portfolio routes
			portfolios := v1.Group("/portfolios")
			{
				portfolios.POST("", portfolioHandler.Create)
				portfolios.GET("", portfolioHandler.GetAll)
				portfolios.GET("/:id", portfolioHandler.GetByID)
				portfolios.PUT("/:id", portfolioHandler.Update)
				portfolios.DELETE("/:id", portfolioHandler.Delete)

				// Transaction routes under portfolio
				portfolios.POST("/:portfolio_id/transactions", transactionHandler.Create)
				portfolios.GET("/:portfolio_id/transactions", transactionHandler.GetAll)
			}

			// Transaction routes
			transactions := v1.Group("/transactions")
			{
				transactions.GET("/:id", transactionHandler.GetByID)
				transactions.PUT("/:id", transactionHandler.Update)
				transactions.DELETE("/:id", transactionHandler.Delete)
			}

			// Tax lot routes
			taxLots := v1.Group("/tax-lots")
			{
				taxLots.GET("/:id", taxLotHandler.GetByID)
			}

			// Portfolio-specific tax lot routes
			v1.GET("/portfolios/:portfolio_id/tax-lots", taxLotHandler.GetAll)
			v1.POST("/portfolios/:portfolio_id/tax-lots/allocate", taxLotHandler.AllocateSale)
			v1.GET("/portfolios/:portfolio_id/tax-lots/harvest", taxLotHandler.IdentifyTaxLossOpportunities)
			v1.POST("/portfolios/:portfolio_id/tax-lots/report", taxLotHandler.GenerateTaxReport)

			// Portfolio action routes (pending corporate actions)
			v1.GET("/portfolios/:portfolio_id/actions", portfolioActionHandler.GetAllActions)
			v1.GET("/portfolios/:portfolio_id/actions/pending", portfolioActionHandler.GetPendingActions)
			v1.GET("/portfolios/:portfolio_id/actions/:action_id", portfolioActionHandler.GetActionByID)
			v1.POST("/portfolios/:portfolio_id/actions/:action_id/approve", portfolioActionHandler.ApproveAction)
			v1.POST("/portfolios/:portfolio_id/actions/:action_id/reject", portfolioActionHandler.RejectAction)
		}
	}

	// Start background job scheduler
	log.Println("Starting background job scheduler...")
	scheduler.Start()

	// Create HTTP server with timeouts to prevent slowloris attacks
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop background job scheduler
	log.Println("Stopping background job scheduler...")
	scheduler.Stop()

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
