package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/handlers"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockEmailService implements EmailService interface for testing
type mockEmailService struct{}

func (m *mockEmailService) SendPasswordResetEmail(to, resetToken string) error {
	return nil
}

// setupSecurityTestServer creates a test server with rate limiting
func setupSecurityTestServer(t *testing.T) (*gin.Engine, *gorm.DB, services.AuthService) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.PasswordResetToken{})
	require.NoError(t, err)

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)

	// Create services
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")
	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Mock email service
	mockEmail := &mockEmailService{}

	passwordResetService := services.NewPasswordResetService(
		userRepo,
		passwordResetRepo,
		mockEmail,
		1*time.Hour,
	)

	// Create handler
	authHandler := handlers.NewAuthHandler(authService, passwordResetService, userRepo, 3600)

	// Create router with rate limiting
	router := gin.New()
	router.Use(gin.Recovery())

	// Create rate limiter: 5 requests per minute
	rateLimiter := middleware.NewRateLimiter(5, time.Minute)

	// Auth routes with rate limiting
	authGroup := router.Group("/api/auth")
	authGroup.Use(rateLimiter.Middleware())
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	}

	return router, db, authService
}

// Test 1: Rate limiting enforcement (make 6 requests in rapid succession, verify 6th fails with 429)
func TestRateLimitingEnforcement(t *testing.T) {
	router, db, _ := setupSecurityTestServer(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Prepare login request
	loginReq := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)

	// Make 5 requests (should all succeed or fail with 401, but NOT 429)
	successfulRequests := 0
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.100") // Same IP for all requests

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not be rate limited (status should be 401 for invalid credentials or 200 for success)
		if w.Code != http.StatusTooManyRequests {
			successfulRequests++
		}

		assert.NotEqual(t, http.StatusTooManyRequests, w.Code,
			"Request %d should not be rate limited", i+1)
	}

	assert.Equal(t, 5, successfulRequests, "First 5 requests should not be rate limited")

	// Make 6th request (should be rate limited with 429)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.100") // Same IP

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be rate limited
	assert.Equal(t, http.StatusTooManyRequests, w.Code,
		"6th request should be rate limited with 429 status")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	code, codeExists := response["code"].(string)
	assert.True(t, codeExists, "Response should contain code field")
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", code,
		"Error code should indicate rate limit exceeded")

	errorMsg, errorExists := response["error"].(string)
	assert.True(t, errorExists, "Response should contain error field")
	assert.Contains(t, strings.ToLower(errorMsg), "rate limit",
		"Error message should mention rate limiting")
}

// Test 2: SQL injection prevention (attempt SQL injection in email/password fields, verify no DB compromise)
func TestSQLInjectionPrevention(t *testing.T) {
	router, db, authService := setupSecurityTestServer(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// First, create a legitimate user
	legitimateEmail := "legitimate@example.com"
	legitimatePassword := "SecurePass123"
	_, _, _, err := authService.Register(legitimateEmail, legitimatePassword)
	require.NoError(t, err, "Legitimate user registration should succeed")

	// Common SQL injection attempts
	sqlInjectionAttempts := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "SQL injection in email with single quote",
			email:    "admin'--",
			password: "anything",
		},
		{
			name:     "SQL injection with OR clause",
			email:    "admin' OR '1'='1",
			password: "password",
		},
		{
			name:     "SQL injection with UNION SELECT",
			email:    "admin' UNION SELECT * FROM users--",
			password: "password",
		},
		{
			name:     "SQL injection in password field",
			email:    "test@example.com",
			password: "password' OR '1'='1",
		},
		{
			name:     "SQL injection with DROP TABLE",
			email:    "test@example.com'; DROP TABLE users; --",
			password: "password",
		},
	}

	for _, attempt := range sqlInjectionAttempts {
		t.Run(attempt.name, func(t *testing.T) {
			loginReq := dto.LoginRequest{
				Email:    attempt.email,
				Password: attempt.password,
			}
			reqBody, _ := json.Marshal(loginReq)

			// Use different IPs for each attempt to avoid rate limiting in tests
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "192.168.2." + attempt.email[0:1] + ":12345" // Different IP per attempt

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// SQL injection should fail with 401 Unauthorized (not succeed)
			assert.NotEqual(t, http.StatusOK, w.Code,
				"SQL injection attempt should not succeed")

			// Most likely result is 401 (invalid credentials) or 400 (bad request)
			assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest,
				"SQL injection should return 401 or 400, got %d", w.Code)
		})
	}

	// Verify database integrity - legitimate user should still exist
	var userCount int64
	err = db.Model(&models.User{}).Count(&userCount).Error
	require.NoError(t, err, "Should be able to count users")
	assert.Equal(t, int64(1), userCount, "Should have exactly 1 user (SQL injection should not have deleted or modified data)")

	// Verify legitimate login still works (using different IP to avoid rate limiting)
	loginReq := dto.LoginRequest{
		Email:    legitimateEmail,
		Password: legitimatePassword,
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.99.99:12345" // Different IP from injection attempts

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code,
		"Legitimate login should still work after SQL injection attempts")
}

// Test 3: Verify bcrypt password hashing (password never stored in plain text)
func TestBcryptPasswordHashing(t *testing.T) {
	_, db, authService := setupSecurityTestServer(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "hashtest@example.com"
	plainPassword := "MySecurePassword123"

	// Register user
	_, _, _, err := authService.Register(email, plainPassword)
	require.NoError(t, err, "User registration should succeed")

	// Fetch user from database
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	require.NoError(t, err, "Should fetch user from database")

	// Verify password is hashed, not plain text
	assert.NotEqual(t, plainPassword, user.PasswordHash,
		"Password should NOT be stored as plain text")

	// Verify hash starts with bcrypt prefix ($2a$, $2b$, or $2y$)
	assert.True(t,
		strings.HasPrefix(user.PasswordHash, "$2a$") ||
			strings.HasPrefix(user.PasswordHash, "$2b$") ||
			strings.HasPrefix(user.PasswordHash, "$2y$"),
		"Password hash should use bcrypt format (starts with $2a$, $2b$, or $2y$)")

	// Verify hash length is appropriate for bcrypt (should be 60 characters)
	assert.Equal(t, 60, len(user.PasswordHash),
		"Bcrypt hash should be 60 characters long")

	// Verify password verification works
	assert.True(t, user.CheckPassword(plainPassword),
		"Should be able to verify correct password")

	assert.False(t, user.CheckPassword("WrongPassword"),
		"Should reject incorrect password")

	// Verify that same password produces different hashes (bcrypt uses salt)
	user2 := &models.User{
		ID:    uuid.New(),
		Email: "another@example.com",
	}
	err = user2.SetPassword(plainPassword)
	require.NoError(t, err)

	assert.NotEqual(t, user.PasswordHash, user2.PasswordHash,
		"Same password should produce different hashes due to salt")
}
