package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openshield-manager/internal/middleware"
	"openshield-manager/internal/service"
	"openshield-manager/internal/utils"
)

// CreateRegistrationToken handles POST /api/registration-tokens
func CreateRegistrationToken(c *gin.Context) {
	var req struct {
		OrganizationID string `json:"organization_id" binding:"required"`
		MaxUses        int    `json:"max_uses"`
		ExpiresInHours int    `json:"expires_in_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}

	userID, _ := utils.GetUserID(c)

	var expiresIn *time.Duration
	if req.ExpiresInHours > 0 {
		d := time.Duration(req.ExpiresInHours) * time.Hour
		expiresIn = &d
	}

	token, err := service.CreateRegistrationToken(orgID, req.MaxUses, expiresIn, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"token":           token.Token,
			"organization_id": token.OrganizationID,
			"max_uses":        token.MaxUses,
			"expires_at":      token.ExpiresAt,
		},
		"error": 0,
	})
}

// ListRegistrationTokens handles GET /api/registration-tokens
func ListRegistrationTokens(c *gin.Context) {
	tokens, err := service.ListRegistrationTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tokens,
		"error": 0,
	})
}

// GetRegistrationToken handles GET /api/registration-tokens/:id
func GetRegistrationToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	token, err := service.GetRegistrationTokenByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  token,
		"error": 0,
	})
}

// RevokeRegistrationToken handles DELETE /api/registration-tokens/:id
func RevokeRegistrationToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	if err := service.RevokeRegistrationToken(id); err != nil {
		if err == service.ErrRegTokenNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registration token not found or already revoked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  gin.H{"message": "Registration token revoked"},
		"error": 0,
	})
}

// RegisterRegTokenRoutes adds registration token management endpoints to the given API group.
func RegisterRegTokenRoutes(apiGroup *gin.RouterGroup) {
	tokens := apiGroup.Group("/registration-tokens")
	tokens.Use(middleware.AuthMiddleware(), middleware.RequireOrgAccess())
	{
		tokens.POST("", middleware.RequireRole(middleware.RoleAdmin), CreateRegistrationToken)
		tokens.GET("", middleware.RequireRole(middleware.RoleViewer), ListRegistrationTokens)
		tokens.GET("/:id", middleware.RequireRole(middleware.RoleViewer), GetRegistrationToken)
		tokens.DELETE("/:id", middleware.RequireRole(middleware.RoleAdmin), RevokeRegistrationToken)
	}
}
