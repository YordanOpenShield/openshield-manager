package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UnregisterRequest struct {
	ID    string `json:"id" binding:"required"`
	Force bool   `json:"force"` // If true, skip agent notification and force-remove
}

// deleteAgentFromDB removes an agent and all its related records from the database, scoped to org.
func deleteAgentFromDB(agentID string, orgID *uuid.UUID) error {
	// Delete agent addresses
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.AgentAddress{}).Error; err != nil {
		return err
	}
	// Delete agent services
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.AgentService{}).Error; err != nil {
		return err
	}
	// Delete group memberships
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.GroupMembership{}).Error; err != nil {
		return err
	}
	// Delete tool action executions
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.ToolActionExecution{}).Error; err != nil {
		return err
	}
	// Delete query execution results
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.QueryExecutionResult{}).Error; err != nil {
		return err
	}
	// Delete tasks
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("agent_id = ?", agentID).Delete(&models.Task{}).Error; err != nil {
		return err
	}
	// Delete the agent itself
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", agentID).Delete(&models.Agent{}).Error; err != nil {
		return err
	}
	return nil
}

func UnregisterAgent(c *gin.Context) {
	var req UnregisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	orgID, _ := utils.GetOrgID(c)

	var agent models.Agent
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", req.ID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	if req.Force {
		// Force unregister: skip agent notification, remove from DB directly
		log.Printf("[UNREGISTER] Force-unregistering agent %s (%s)", agent.ID, agent.Address)
		if err := deleteAgentFromDB(req.ID, orgID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove agent: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Agent force-unregistered successfully"})
		return
	}

	// Try to notify the agent first
	if agent.Address != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		client, err := managergrpc.NewAgentClient(agent.Address)
		if err != nil {
			// Can't reach agent, offer force option
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Agent is unreachable. Use force=true to unregister anyway",
				"message": "Failed to connect to agent: " + err.Error(),
			})
			return
		}
		err = client.UnregisterAgentAsk(ctx)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Agent did not acknowledge. Use force=true to unregister anyway",
				"message": "Failed to unregister agent: " + err.Error(),
			})
			return
		}
	}

	// Agent acknowledged, now remove from manager's DB
	if err := deleteAgentFromDB(req.ID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent notified but failed to remove from database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered successfully"})
}
