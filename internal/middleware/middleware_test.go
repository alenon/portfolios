package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/logger"
	"github.com/lenon/portfolios/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real token service for testing
	tokenService := services.NewTokenService("test-secret-key-for-testing")
	userID := uuid.New().String()

	// Generate a valid token
	token, err := tokenService.GenerateAccessToken(userID, 15*time.Minute)
	assert.NoError(t, err)

	handler := AuthRequired(tokenService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	// Create a test handler to check context
	finalHandler := func(c *gin.Context) {
		contextUserID := c.GetString(UserIDContextKey)
		assert.Equal(t, userID, contextUserID)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

	// Chain the middleware and final handler
	handler(c)
	if !c.IsAborted() {
		finalHandler(c)
	}

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenService := services.NewTokenService("test-secret-key-for-testing")

	handler := AuthRequired(tokenService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenService := services.NewTokenService("test-secret-key-for-testing")

	handler := AuthRequired(tokenService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat token")

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenService := services.NewTokenService("test-secret-key-for-testing")

	handler := AuthRequired(tokenService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestGetUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDContextKey, "user-123")

	userID := GetUserID(c)

	assert.Equal(t, "user-123", userID)
}

func TestGetUserID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No user ID set

	userID := GetUserID(c)

	assert.Equal(t, "", userID)
}

func TestCORS_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedOrigins := []string{"http://localhost:3000"}
	handler := CORS(allowedOrigins)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:3000")

	handler(c)

	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedOrigins := []string{"http://localhost:3000"}
	handler := CORS(allowedOrigins)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Origin", "http://evil.com")

	handler(c)

	assert.NotEqual(t, "http://evil.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedOrigins := []string{"http://localhost:3000"}
	handler := CORS(allowedOrigins)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("OPTIONS", "/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:3000")

	handler(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestErrorHandler_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := ErrorHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	handler(c)
	c.Next()

	// Should not interfere with normal request
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_AllowedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Allow 10 requests per second
	handler := RateLimit(10, 1*time.Second)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// First request should be allowed
	handler(c)

	assert.False(t, c.IsAborted())
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Allow only 1 request per second
	handler := RateLimit(1, 1*time.Second)

	// First request
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "192.168.1.1:1234"
	handler(c1)
	assert.False(t, c1.IsAborted())

	// Second request immediately after (should be rate limited)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.1:1234"
	handler(c2)
	assert.True(t, c2.IsAborted())
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Allow only 1 request per second
	handler := RateLimit(1, 1*time.Second)

	// Request from first IP
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "192.168.1.1:1234"
	handler(c1)
	assert.False(t, c1.IsAborted())

	// Request from different IP (should not be rate limited)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.2:1234"
	handler(c2)
	assert.False(t, c2.IsAborted())
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appLogger := logger.NewLogger(logger.Config{
		Level:  "info",
		Format: "json",
	})
	handler := LoggingMiddleware(appLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	handler(c)
	c.Next()

	// Should complete without error
	assert.NotNil(t, handler)
}

func TestErrorLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appLogger := logger.NewLogger(logger.Config{
		Level:  "info",
		Format: "json",
	})
	handler := ErrorLoggingMiddleware(appLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	handler(c)
	c.Next()

	// Should complete without error
	assert.NotNil(t, handler)
}

func TestRecoveryLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appLogger := logger.NewLogger(logger.Config{
		Level:  "info",
		Format: "json",
	})
	handler := RecoveryLoggingMiddleware(appLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Test that recovery middleware is created without error
	handler(c)
	c.Next()

	// Should complete without crashing
	assert.NotNil(t, handler)
}
