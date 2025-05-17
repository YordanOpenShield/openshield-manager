package api

import (
	"net/http"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
)

// GetAgentsList returns a list of all agents
func GetAgentsList(c *gin.Context) {
	var agents []models.Agent
	if err := db.DB.Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}
	c.JSON(http.StatusOK, agents)
}

// GetAgentDetails returns details for a specific agent by ID
func GetAgentDetails(c *gin.Context) {
	id := c.Param("id")
	var agent models.Agent
	if err := db.DB.Where("id = ?", id).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}
