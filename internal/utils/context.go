package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openshield-manager/internal/models"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyOrgID     = "organization_id"
	ContextKeyUserRole  = "user_role"
	ContextKeyUserEmail = "user_email"
	ContextKeyAgent     = "agent"
	ContextKeyAgentID   = "agent_id"
)

// GetUserID extracts the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	idStr, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(idStr.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// GetOrgID extracts the organization ID from the Gin context.
// Returns nil for super admins (cross-org access).
func GetOrgID(c *gin.Context) (*uuid.UUID, bool) {
	orgIDVal, exists := c.Get(ContextKeyOrgID)
	if !exists {
		return nil, false
	}
	// Handle nil interface value (super admin with no org)
	// Must check reflect-based nil since *string(nil) interface != nil
	if orgIDVal == nil {
		return nil, true
	}
	// The value is stored as *string (pointer), convert to string safely
	var orgIDStr string
	switch v := orgIDVal.(type) {
	case string:
		orgIDStr = v
	case *string:
		if v == nil {
			return nil, true
		}
		orgIDStr = *v
	default:
		return nil, false
	}
	if orgIDStr == "" {
		return nil, true
	}
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil, false
	}
	return &id, true
}

// GetUserRole extracts the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) (models.UserRole, bool) {
	role, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return "", false
	}
	return models.UserRole(role.(string)), true
}

// GetUserEmail extracts the authenticated user's email from the Gin context.
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(ContextKeyUserEmail)
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetAgent extracts the authenticated agent from the Gin context.
func GetAgent(c *gin.Context) (*models.Agent, bool) {
	agent, exists := c.Get(ContextKeyAgent)
	if !exists {
		return nil, false
	}
	return agent.(*models.Agent), true
}

// RequireOrgID is a helper that gets the OrgID and returns nil if not found.
// Useful for handlers that need to enforce org scoping.
func RequireOrgID(c *gin.Context) *uuid.UUID {
	orgID, exists := GetOrgID(c)
	if !exists || orgID == nil {
		return nil
	}
	return orgID
}
