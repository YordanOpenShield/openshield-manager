package api

import (
	"context"
	"net/http"
	"time"

	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
)

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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	client, err := managergrpc.NewAgentClient(agent.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC client: " + err.Error()})
		return
	}
	err = client.UnregisterAgentAsk(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered successfully"})
}
