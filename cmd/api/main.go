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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		authService,
		passwordResetService,
		userRepo,
		int(cfg.JWT.AccessTokenDuration.Seconds()),
	)

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
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
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

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
