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

	// Get agent details
	var agent models.Agent
	if err := db.DB.Where("id = ?", id).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Get agent addresses
	var addresses []models.AgentAddress
	if err := db.DB.Where("agent_id = ?", agent.ID).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent addresses"})
		return
	}

	// Get agent services
	var services []models.AgentService
	if err := db.DB.Where("agent_id = ?", agent.ID).Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent services"})
		return
	}

	// Generate response
	agentDetailsResponse := struct {
		Agent     models.Agent          `json:"agent"`
		Addresses []models.AgentAddress `json:"addresses"`
		Services  []models.AgentService `json:"services"`
	}{
		Agent:     agent,
		Addresses: addresses,
		Services:  services,
	}

	c.JSON(http.StatusOK, agentDetailsResponse)
}
