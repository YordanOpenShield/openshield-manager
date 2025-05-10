package api

import (
	"net/http"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
)

func GetAvailableJobs(c *gin.Context) {
	var jobs []models.Job
	db.DB.Find(&jobs)
	c.JSON(http.StatusOK, jobs)
}

type AssignJobToAgentRequest struct {
	AgentID string `json:"id" binding:"required"`
}

func AssignJobToAgent(c *gin.Context) {
	var req AssignJobToAgentRequest

	// Parse the request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.AgentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// TODO: Publish job to agent's Redis channel
}
