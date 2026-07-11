package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openshield-manager/internal/middleware"
	"openshield-manager/internal/models"
	"openshield-manager/internal/service"
	"openshield-manager/internal/utils"
)

// CreateRegistrationToken handles POST /api/registration-tokens
// Org admins can only create tokens for their own org.
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

	// Enforce org ownership: org_admin must create tokens for their own org
	callerOrgID, _ := utils.GetOrgID(c)
	callerRole, _ := utils.GetUserRole(c)
	if callerRole != models.UserRoleSuperAdmin && callerOrgID != nil && *callerOrgID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot create registration token for a different organization"})
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

// ListRegistrationTokens handles GET /api/registration-tokens, scoped to the caller's org.
func ListRegistrationTokens(c *gin.Context) {
	orgID, _ := utils.GetOrgID(c)
	tokens, err := service.ListRegistrationTokens(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tokens,
		"error": 0,
	})
}

// GetRegistrationToken handles GET /api/registration-tokens/:id, scoped to the caller's org.
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

	// Enforce org scoping: non-super-admins can only see their org's tokens
	callerOrgID, _ := utils.GetOrgID(c)
	callerRole, _ := utils.GetUserRole(c)
	if callerRole != models.UserRoleSuperAdmin && token.OrganizationID != uuid.Nil {
		if callerOrgID == nil || *callerOrgID != token.OrganizationID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registration token not found"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  token,
		"error": 0,
	})
}

// RevokeRegistrationToken handles DELETE /api/registration-tokens/:id, scoped to the caller's org.
func RevokeRegistrationToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// First load the token to verify org ownership
	token, err := service.GetRegistrationTokenByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration token not found"})
		return
	}

	callerOrgID, _ := utils.GetOrgID(c)
	callerRole, _ := utils.GetUserRole(c)
	if callerRole != models.UserRoleSuperAdmin && token.OrganizationID != uuid.Nil {
		if callerOrgID == nil || *callerOrgID != token.OrganizationID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registration token not found"})
			return
		}
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
