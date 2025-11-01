package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that catches panics and converts them to 500 errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				fmt.Printf("Panic recovered: %v\n", err)

				// Return standardized error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_SERVER_ERROR",
				})

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()

		// Check if there were any errors during request processing
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Log the error
			fmt.Printf("Request error: %v\n", err.Err)

			// If response wasn't already sent, send error response
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
					"code":  "REQUEST_ERROR",
				})
			}
		}
	}
}
