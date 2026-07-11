package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openshield-manager/internal/middleware"
	"openshield-manager/internal/models"
	"openshield-manager/internal/service"
)

// Authenticate handles POST /api/auth/authenticate (Basic Auth login).
// Wazuh-style: returns JWT token on successful authentication.
func Authenticate(c *gin.Context) {
	token, exists := c.Get("auth_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	user, _ := c.Get("auth_user")

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token": token.(string),
			"user":  user,
		},
		"error": 0,
	})
}

// Logout handles DELETE /api/auth/authenticate (revoke tokens).
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "User was successfully logged out",
		},
		"error": 0,
	})
}

// RegisterUser handles POST /api/auth/register (create a new user).
func RegisterUser(c *gin.Context) {
	var req struct {
		Email    string          `json:"email" binding:"required"`
		Password string          `json:"password" binding:"required,min=6"`
		Name     string          `json:"name" binding:"required"`
		Role     models.UserRole `json:"role"`
		OrgID    *string         `json:"organization_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Default role if not specified
	if req.Role == "" {
		req.Role = models.UserRoleOrgViewer
	}

	// Parse organization_id from request (optional)
	var orgID *uuid.UUID
	if req.OrgID != nil && *req.OrgID != "" {
		parsed, err := uuid.Parse(*req.OrgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization_id format"})
			return
		}
		orgID = &parsed
	}

	user, err := service.RegisterUser(req.Email, req.Password, req.Name, req.Role, orgID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"user": user,
		},
		"error": 0,
	})
}

// GetCurrentUser handles GET /api/auth/me (current user info).
func GetCurrentUser(c *gin.Context) {
	// AuthMiddleware already validated the token and set context
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user_id":         c.MustGet("user_id"),
			"email":           c.MustGet("user_email"),
			"role":            c.MustGet("user_role"),
			"organization_id": c.MustGet("organization_id"),
		},
		"error": 0,
	})
}

// RegisterAuthRoutes adds auth endpoints to the given API group.
func RegisterAuthRoutes(apiGroup *gin.RouterGroup) {
	auth := apiGroup.Group("/auth")
	{
		auth.POST("/authenticate", middleware.BasicAuthMiddleware(), Authenticate)
		auth.DELETE("/authenticate", middleware.AuthMiddleware(), Logout)
		auth.POST("/register", middleware.AuthMiddleware(), middleware.RequireRole(middleware.RoleAdmin), RegisterUser)
		auth.GET("/me", middleware.AuthMiddleware(), GetCurrentUser)
	}
}
