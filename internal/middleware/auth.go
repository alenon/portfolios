package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/services"
)

const (
	// AuthorizationHeader is the header name for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// UserIDContextKey is the context key for user ID
	UserIDContextKey = "user_id"
)

// AuthRequired is a middleware that validates JWT tokens and attaches user ID to context
func AuthRequired(tokenService *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must start with 'Bearer '",
				"code":  "INVALID_AUTH_HEADER_FORMAT",
			})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// Validate token
		token, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Extract user ID
		userID, err := tokenService.ExtractUserID(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to extract user information from token",
				"code":  "INVALID_TOKEN_CLAIMS",
			})
			c.Abort()
			return
		}

		// Attach user ID to context
		c.Set(UserIDContextKey, userID)

		// Continue to next handler
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return ""
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return ""
	}

	return userIDStr
}
