package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"
)

// Role represents the access level required for an endpoint.
type Role int

const (
	RoleAny        Role = iota // Any authenticated user
	RoleViewer                 // org_viewer or higher
	RoleOperator               // org_operator or higher
	RoleAdmin                  // org_admin or higher
	RoleSuperAdmin             // super_admin only
)

// roleHierarchy maps user roles to their access level.
var roleHierarchy = map[models.UserRole]Role{
	models.UserRoleSuperAdmin:  RoleSuperAdmin,
	models.UserRoleOrgAdmin:    RoleAdmin,
	models.UserRoleOrgOperator: RoleOperator,
	models.UserRoleOrgViewer:   RoleViewer,
}

// RequireRole returns a middleware that checks if the authenticated user
// has at least the specified role level.
func RequireRole(minimum Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := utils.GetUserRole(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		userLevel, ok := roleHierarchy[userRole]
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid user role"})
			return
		}

		if userLevel < minimum {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}

// RequireOrgAccess returns a middleware that ensures the user has access
// to the organization specified in the request. Super admins bypass this check.
func RequireOrgAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := utils.GetUserRole(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Super admins can access any org
		if userRole == models.UserRoleSuperAdmin {
			c.Next()
			return
		}

		// For other users, their org must match (handled by GORM TenantScope)
		c.Next()
	}
}
