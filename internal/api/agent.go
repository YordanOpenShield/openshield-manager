package api

import (
	"net/http"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func RegisterAgent(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check if an agent with this DeviceID already exists
	var existing models.Agent
	if err := db.DB.Where("device_id = ?", req.DeviceID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Agent already registered on this device"})
		return
	}
	// Create a new agent
	agentID := uuid.New()
	agent := models.Agent{
		ID:       agentID,
		Token:    uuid.New().String(),
		LastSeen: time.Now(),
		DeviceID: req.DeviceID,
		State:    "DISCONNECTED",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var agent models.Agent
	if err := db.DB.Where("id = ?", req.ID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	agent.State = "UNREGISTERED"
	if err := db.DB.Save(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered successfully"})
}

type HeartbeatRequest struct {
	ID        string   `json:"id" binding:"required"`
	Addresses []string `json:"addresses" binding:"required"`
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

	// Update the agent’s last heartbeat timestamp and state
	agent.LastSeen = time.Now()
	agent.State = "CONNECTED"
	if err := db.DB.Save(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	// Remove old addresses for this agent
	if err := db.DB.Where("agent_id = ?", agent.ID).Delete(&models.AgentAddress{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old addresses"})
		return
	}

	// Insert new addresses
	for _, addr := range req.Addresses {
		address := models.AgentAddress{
			AgentID: agent.ID,
			Address: addr,
		}
		if err := db.DB.Create(&address).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save address"})
			return
		}
	}

	// Respond with a success message
	c.Status(http.StatusOK)
}
