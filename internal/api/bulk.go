package api

import (
	"net/http"
	"openshield-manager/internal/db"
	"openshield-manager/internal/events"
	"openshield-manager/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetBulkOperations returns all bulk operations
func GetBulkOperations(c *gin.Context) {
	var operations []models.BulkOperation
	if err := db.DB.Order("created_at desc").Find(&operations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch operations"})
		return
	}
	c.JSON(http.StatusOK, operations)
}

// CreateBulkOperation creates a new bulk operation
func CreateBulkOperation(c *gin.Context) {
	var op models.BulkOperation
	if err := c.ShouldBindJSON(&op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize progress
	op.Status = models.BulkOpPending
	op.Progress = models.BulkProgress{
		Total:     0,
		Completed: 0,
		Failed:    0,
		Pending:   0,
	}

	if err := db.DB.Create(&op).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation"})
		return
	}

	// Start processing in background
	go processBulkOperation(op.ID)

	c.JSON(http.StatusCreated, op)
}

// GetBulkOperation returns a specific operation
func GetBulkOperation(c *gin.Context) {
	id := c.Param("id")
	var op models.BulkOperation
	if err := db.DB.Where("id = ?", id).First(&op).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
		return
	}
	c.JSON(http.StatusOK, op)
}

// CancelBulkOperation cancels a running operation
func CancelBulkOperation(c *gin.Context) {
	id := c.Param("id")
	var op models.BulkOperation
	if err := db.DB.Where("id = ?", id).First(&op).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
		return
	}

	if op.Status != models.BulkOpRunning && op.Status != models.BulkOpPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel operation in current state"})
		return
	}

	op.Status = models.BulkOpCancelled
	now := time.Now()
	op.CompletedAt = &now
	db.DB.Save(&op)

	c.JSON(http.StatusOK, op)
}

// processBulkOperation handles the actual bulk processing
func processBulkOperation(opID uuid.UUID) {
	var op models.BulkOperation
	if err := db.DB.Where("id = ?", opID).First(&op).Error; err != nil {
		return
	}

	// Get target agents
	agentIDs := resolveBulkTarget(op.Target)
	total := len(agentIDs)
	
	if total == 0 {
		op.Status = models.BulkOpFailed
		now := time.Now()
		op.CompletedAt = &now
		db.DB.Save(&op)
		return
	}

	// Update status to running
	op.Status = models.BulkOpRunning
	op.Progress.Total = total
	op.Progress.Pending = total
	db.DB.Save(&op)

	events.PublishBulkOpStarted(opID.String(), map[string]interface{}{
		"type":      op.Type,
		"total":     total,
		"target":    op.Target,
	})

	// Process in batches of 100
	batchSize := 100
	results := make([]models.BulkResult, 0, total)
	
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		
		batch := agentIDs[i:end]
		batchResults := processBatch(&op, batch)
		results = append(results, batchResults...)
		
		// Update progress
		completed := 0
		failed := 0
		for _, r := range results {
			if r.Status == "SUCCESS" {
				completed++
			} else {
				failed++
			}
		}
		
		op.Progress.Completed = completed
		op.Progress.Failed = failed
		op.Progress.Pending = total - completed - failed
		op.Results = results
		db.DB.Save(&op)
		
		events.PublishBulkOpProgress(opID.String(), op.Progress)
		
		// Check if cancelled
		db.DB.Where("id = ?", opID).First(&op)
		if op.Status == models.BulkOpCancelled {
			return
		}
	}

	// Mark as completed
	if op.Status != models.BulkOpCancelled {
		if op.Progress.Failed == 0 {
			op.Status = models.BulkOpCompleted
		} else if op.Progress.Completed > 0 {
			op.Status = models.BulkOpPartial
		} else {
			op.Status = models.BulkOpFailed
		}
		now := time.Now()
		op.CompletedAt = &now
		db.DB.Save(&op)
		
		events.PublishBulkOpCompleted(opID.String(), results)
	}
}

// resolveBulkTarget resolves the target specification to agent IDs
func resolveBulkTarget(target models.BulkTarget) []string {
	agentIDs := make([]string, 0)
	
	if target.AllConnected {
		var agents []models.Agent
		db.DB.Where("state = ?", models.AgentStateConnected).Find(&agents)
		for _, a := range agents {
			agentIDs = append(agentIDs, a.ID.String())
		}
		return agentIDs
	}
	
	if target.GroupID != "" {
		var memberships []models.GroupMembership
		db.DB.Where("group_id = ?", target.GroupID).Find(&memberships)
		for _, m := range memberships {
			agentIDs = append(agentIDs, m.AgentID.String())
		}
		return agentIDs
	}
	
	return target.AgentIDs
}

// processBatch processes a batch of agents
func processBatch(op *models.BulkOperation, agentIDs []string) []models.BulkResult {
	results := make([]models.BulkResult, 0, len(agentIDs))
	
	for _, agentID := range agentIDs {
		result := models.BulkResult{
			AgentID:    agentID,
			ExecutedAt: time.Now(),
		}
		
		// Process based on operation type
		switch op.Type {
		case models.BulkOpTaskAssign:
			result = processTaskAssign(agentID, op.Payload)
		case models.BulkOpQueryRun:
			result = processQueryRun(agentID, op.Payload)
		case models.BulkOpToolExecute:
			result = processToolExecute(agentID, op.Payload)
		default:
			result.Status = "ERROR"
			result.Error = "Unknown operation type"
		}
		
		results = append(results, result)
	}
	
	return results
}

func processTaskAssign(agentID string, payload interface{}) models.BulkResult {
	result := models.BulkResult{
		AgentID:    agentID,
		ExecutedAt: time.Now(),
	}
	
	// TODO: Implement task assignment logic
	result.Status = "SUCCESS"
	result.Data = map[string]string{"message": "Task assigned"}
	
	return result
}

func processQueryRun(agentID string, payload interface{}) models.BulkResult {
	result := models.BulkResult{
		AgentID:    agentID,
		ExecutedAt: time.Now(),
	}
	
	// TODO: Implement query execution logic
	result.Status = "SUCCESS"
	result.Data = map[string]string{"message": "Query executed"}
	
	return result
}

func processToolExecute(agentID string, payload interface{}) models.BulkResult {
	result := models.BulkResult{
		AgentID:    agentID,
		ExecutedAt: time.Now(),
	}
	
	// TODO: Implement tool execution logic
	result.Status = "SUCCESS"
	result.Data = map[string]string{"message": "Tool executed"}
	
	return result
}
