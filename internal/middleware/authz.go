package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireOwnership is a middleware that ensures the authenticated user owns the resource
// resourceUserIDParam is the name of the parameter (path, query, or form) that contains the resource owner's user ID
func RequireOwnership(resourceUserIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user ID from context
		authenticatedUserID := GetUserID(c)
		if authenticatedUserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Extract resource owner ID from request
		// Try path parameter first
		resourceUserID := c.Param(resourceUserIDParam)

		// If not in path, try query parameter
		if resourceUserID == "" {
			resourceUserID = c.Query(resourceUserIDParam)
		}

		// If not in query, try form data
		if resourceUserID == "" {
			resourceUserID = c.PostForm(resourceUserIDParam)
		}

		// If still empty, try JSON body
		if resourceUserID == "" {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err == nil {
				if val, ok := body[resourceUserIDParam]; ok {
					if strVal, ok := val.(string); ok {
						resourceUserID = strVal
					}
				}
			}
		}

		// If resource user ID is empty, allow access (user accessing their own profile)
		if resourceUserID == "" {
			c.Next()
			return
		}

		// Compare user IDs
		if authenticatedUserID != resourceUserID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to access this resource",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		// User owns the resource, continue
		c.Next()
	}
}
