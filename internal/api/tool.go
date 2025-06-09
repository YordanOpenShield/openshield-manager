package api

import (
	"context"
	"net/http"
	"openshield-manager/internal/db"
	agentgrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetToolsByAgent returns all tools for a given agentID.
func GetToolsByAgent(c *gin.Context) {
	agentID := c.Param("id")

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", agentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Fetch all tools from the gRPC client
	client, err := agentgrpc.NewAgentClient(agent.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC client: " + err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // use background context
	defer cancel()

	res, err := client.GetTools(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tools: " + err.Error()})
		return
	}

	var result []models.Tool
	for _, tool := range res.Tools {
		// Convert gRPC tool to database model
		var actions []models.ToolAction
		for _, action := range tool.Actions {
			actions = append(actions, models.ToolAction{
				Name: action.Name,
				Opts: action.Options,
			})
		}
		result = append(result, models.Tool{
			Name:    tool.Name,
			Actions: actions,
			OS:      tool.Os,
		})
	}

	c.JSON(http.StatusOK, result)
}

type ExecuteToolRequest struct {
	AgentID     string   `json:"agent_id" binding:"required"`
	ToolName    string   `json:"tool_name" binding:"required"`
	ToolAction  string   `json:"tool_action" binding:"required"`
	ToolOptions []string `json:"tool_options" binding:"required"`
}

func ExecuteTool(c *gin.Context) {
	var req ExecuteToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.AgentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Create gRPC client
	client, err := agentgrpc.NewAgentClient(agent.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC client: " + err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.ExecuteTool(ctx, req.ToolName, req.ToolAction, req.ToolOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute tool: " + err.Error()})
		return
	}

	// Track the tool action status
	toolActionExecution := models.ToolActionExecution{
		AgentID:     agent.ID,
		ToolName:    req.ToolName,
		ToolAction:  req.ToolAction,
		ToolOptions: req.ToolOptions,
	}
	if err := db.DB.Create(&toolActionExecution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tool action execution: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)

	go agentgrpc.TrackToolActionStatus(agent.Address, req.ToolName, req.ToolAction, toolActionExecution.ID, 1*time.Second)
}
