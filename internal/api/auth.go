package api

import (
	"net/http"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
)

// AgentAuthMiddleware checks for a valid agent token in the request headers.
func AgentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Agent-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing agent token"})
			return
		}

		var agent models.Agent
		err := db.DB.Where("token = ?", token).First(&agent).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
			return
		}

		// Optionally, set agent info in context
		c.Set("agent", agent)
		c.Next()
	}
}
