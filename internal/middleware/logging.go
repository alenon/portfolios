package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/logger"
)

// LoggingMiddleware logs all HTTP requests with details
func LoggingMiddleware(log *logger.AppLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request details
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response details
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// Get user ID if authenticated
		userID, exists := c.Get("user_id")
		userIDStr := ""
		if exists {
			userIDStr = userID.(string)
		}

		// Log request
		event := log.Info().
			Str("method", method).
			Str("path", path).
			Int("status_code", statusCode).
			Dur("duration_ms", duration).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Int("response_size", responseSize)

		if userIDStr != "" {
			event = event.Str("user_id", userIDStr)
		}

		// Add error details if request failed
		if statusCode >= 400 {
			errors := c.Errors.String()
			if errors != "" {
				event = event.Str("errors", errors)
			}
		}

		event.Msg("HTTP request")
	}
}

// ErrorLoggingMiddleware logs detailed error information
func ErrorLoggingMiddleware(log *logger.AppLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log errors if any occurred
		for _, err := range c.Errors {
			event := log.Error().
				Err(err.Err).
				Str("type", fmt.Sprintf("%v", err.Type)).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method)

			if err.Meta != nil {
				event = event.Interface("meta", err.Meta)
			}

			event.Msg("Request error")
		}
	}
}

// RecoveryLoggingMiddleware logs panic recovery with stack trace
func RecoveryLoggingMiddleware(log *logger.AppLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("Panic recovered")

				c.JSON(500, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
