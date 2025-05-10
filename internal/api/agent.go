package api

import (
	"net/http"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterAgent(c *gin.Context) {
	agentID := uuid.New()
	agent := models.Agent{
		ID:       agentID,
		Token:    uuid.New().String(),
		LastSeen: time.Now(),
	}

	db.DB.Create(&agent)
	c.JSON(http.StatusOK, gin.H{"id": agent.ID, "token": agent.Token})
}

type UnregisterRequest struct {
	ID string `json:"id" binding:"required"`
}

func UnregisterAgent(c *gin.Context) {
	var req UnregisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var agent models.Agent
	if err := db.DB.Where("id = ?", req.ID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	if err := db.DB.Delete(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered successfully"})
}

type HeartbeatRequest struct {
	ID string `json:"id" binding:"required"`
}

func AgentHeartbeat(c *gin.Context) {
	var req HeartbeatRequest

	// Parse the request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.ID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Update the agent’s last heartbeat timestamp
	agent.LastSeen = time.Now()
	db.DB.Save(&agent)

	// Respond with a success message
	// c.JSON(http.StatusOK, gin.H{"status": "heartbeat received"})
	c.Status(http.StatusOK)
}
