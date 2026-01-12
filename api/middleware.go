package api

import (
	"net/http"

	"github.com/PRPO-skupina-02/auth/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireRole middleware checks if the authenticated user has one of the required roles
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role information not found"})
			return
		}

		role := userRole.(models.UserRole)

		// Check if user has one of the required roles
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}

// RequireAdmin is a convenience middleware for admin-only endpoints
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// GetContextUserID retrieves the user ID from the context
func GetContextUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("user_id").(uuid.UUID)
}

// GetContextUserRole retrieves the user role from the context
func GetContextUserRole(c *gin.Context) models.UserRole {
	return c.MustGet("user_role").(models.UserRole)
}
