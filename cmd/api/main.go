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
	performanceSnapshotRepo := repository.NewPerformanceSnapshotRepository(db)

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
	holdingService := services.NewHoldingService(holdingRepo, portfolioRepo)

	// Initialize market data service
	var marketDataService services.MarketDataService
	if cfg.MarketData.APIKey != "" {
		alphaVantageProvider := services.NewAlphaVantageProvider(cfg.MarketData.APIKey)
		marketDataService = services.NewMarketDataService(alphaVantageProvider, 15*time.Minute)
		log.Println("Market data service initialized with Alpha Vantage provider")
	} else {
		log.Println("Warning: Market data service not initialized (no API key provided)")
	}

	// Initialize performance snapshot service
	performanceSnapshotService := services.NewPerformanceSnapshotService(
		performanceSnapshotRepo,
		portfolioRepo,
		holdingRepo,
	)

	// Initialize performance analytics service (only if market data is available)
	var performanceAnalyticsService services.PerformanceAnalyticsService
	if marketDataService != nil {
		performanceAnalyticsService = services.NewPerformanceAnalyticsService(
			portfolioRepo,
			transactionRepo,
			performanceSnapshotRepo,
			marketDataService,
		)
		log.Println("Performance analytics service initialized")
	} else {
		log.Println("Warning: Performance analytics service not initialized (requires market data)")
	}

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

	// Initialize CSV import service
	csvImportService := services.NewCSVImportService(
		transactionRepo,
		portfolioRepo,
		holdingRepo,
	)

	// Initialize background job scheduler
	scheduler := jobs.NewScheduler()

	// Add corporate action detection job
	corporateActionJob := jobs.NewCorporateActionDetectionJob(corporateActionMonitor)
	scheduler.AddJob(corporateActionJob)

	// Add market data jobs (only if market data service is available)
	if marketDataService != nil {
		// Price update job - refreshes market data cache
		priceUpdateJob := jobs.NewPriceUpdateJob(marketDataService)
		scheduler.AddJob(priceUpdateJob)

		// Performance snapshot job - generates daily snapshots
		// Note: Simplified version - full implementation requires repository enhancements
		snapshotJob := jobs.NewSnapshotGenerationJob()
		scheduler.AddJob(snapshotJob)

		// Cleanup job - cleans up stale data
		cleanupJob := jobs.NewCleanupJob(marketDataService, 365)
		scheduler.AddJob(cleanupJob)

		log.Println("Market data background jobs initialized")
	}

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
	holdingHandler := handlers.NewHoldingHandler(holdingService)
	portfolioActionHandler := handlers.NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo, corporateActionService)

	// Initialize performance handlers (only if analytics service is available)
	var performanceAnalyticsHandler *handlers.PerformanceAnalyticsHandler
	if performanceAnalyticsService != nil {
		performanceAnalyticsHandler = handlers.NewPerformanceAnalyticsHandler(performanceAnalyticsService)
	}

	// Initialize market data handler (only if market data service is available)
	var marketDataHandler *handlers.MarketDataHandler
	if marketDataService != nil {
		marketDataHandler = handlers.NewMarketDataHandler(marketDataService)
	}

	// Initialize performance snapshot handler
	performanceSnapshotHandler := handlers.NewPerformanceSnapshotHandler(performanceSnapshotService)

	// Initialize CSV import handler
	importHandler := handlers.NewImportHandler(csvImportService)

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

				// CSV import routes
				portfolios.POST("/:id/transactions/import/csv", importHandler.ImportCSV)
				portfolios.POST("/:id/transactions/import/bulk", importHandler.ImportBulk)
				portfolios.GET("/:id/imports/batches", importHandler.GetImportBatches)
				portfolios.DELETE("/:id/imports/batches/:batch_id", importHandler.DeleteImportBatch)

				// Holding routes under portfolio
				portfolios.GET("/:id/holdings", holdingHandler.GetAll)
				portfolios.GET("/:id/holdings/:symbol", holdingHandler.GetBySymbol)

				// Performance analytics routes (if available)
				if performanceAnalyticsHandler != nil {
					portfolios.GET("/:id/performance/metrics", performanceAnalyticsHandler.GetPerformanceMetrics)
					portfolios.GET("/:id/performance/twr", performanceAnalyticsHandler.GetTWR)
					portfolios.GET("/:id/performance/mwr", performanceAnalyticsHandler.GetMWR)
					portfolios.GET("/:id/performance/annualized", performanceAnalyticsHandler.GetAnnualizedReturn)
					portfolios.GET("/:id/performance/benchmark", performanceAnalyticsHandler.GetBenchmarkComparison)
				}

				// Performance snapshot routes
				portfolios.GET("/:id/snapshots", performanceSnapshotHandler.GetSnapshots)
				portfolios.GET("/:id/snapshots/range", performanceSnapshotHandler.GetSnapshotsByDateRange)
				portfolios.GET("/:id/snapshots/latest", performanceSnapshotHandler.GetLatestSnapshot)
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

			// Market data routes (if available)
			if marketDataHandler != nil {
				market := v1.Group("/market")
				{
					market.GET("/quote/:symbol", marketDataHandler.GetQuote)
					market.POST("/quotes", marketDataHandler.GetQuotes)
					market.GET("/history/:symbol", marketDataHandler.GetHistoricalPrices)
					market.GET("/exchange", marketDataHandler.GetExchangeRate)
					market.POST("/cache/clear", marketDataHandler.ClearCache)
				}
			}
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
