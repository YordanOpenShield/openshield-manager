package api

import (
	"net/http"
	"openshield-manager/internal/db"
	managergrpc "openshield-manager/internal/grpc"
	"openshield-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RunQueryRequest represents a request to run a query
type RunQueryRequest struct {
	QueryID  string   `json:"query_id" binding:"required"`
	AgentIDs []string `json:"agent_ids"` // empty means all online agents
}

// RunQuery executes a query on specified agents
func RunQuery(c *gin.Context) {
	var req RunQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get the query
	var query models.Query
	if err := db.DB.Where("id = ?", req.QueryID).First(&query).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		return
	}

	// Create execution record
	execution := models.QueryExecution{
		ID:      uuid.New(),
		QueryID: query.ID,
		Status:  models.QueryStatusPending,
	}
	if err := db.DB.Create(&execution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create execution"})
		return
	}

	// Get target agents
	var agents []models.Agent
	if len(req.AgentIDs) == 0 {
		// All online agents
		if err := db.DB.Where("state = ?", models.AgentStateConnected).Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
			return
		}
	} else {
		// Specific agents
		if err := db.DB.Where("id IN ?", req.AgentIDs).Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
			return
		}
	}

	// Create execution results for each agent
	for _, agent := range agents {
		result := models.QueryExecutionResult{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			AgentID:     agent.ID,
			Status:      models.QueryStatusPending,
		}
		db.DB.Create(&result)
	}

	// Start execution in background
	go managergrpc.ExecuteQueryOnAgents(execution.ID, query, agents)

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": execution.ID,
		"status":       "pending",
		"agent_count":  len(agents),
	})
}

// RunLiveQuery runs a query without saving it (ad-hoc query)
type RunLiveQueryRequest struct {
	SQL      string   `json:"sql" binding:"required"`
	Platform string   `json:"platform"`
	AgentIDs []string `json:"agent_ids"` // empty means all online agents
}

func RunLiveQuery(c *gin.Context) {
	var req RunLiveQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create a temporary query (saved to DB for FK constraint)
	query := models.Query{
		ID:       uuid.New(),
		Name:     "Live Query",
		SQL:      req.SQL,
		Platform: req.Platform,
	}
	if err := db.DB.Create(&query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create query"})
		return
	}

	// Create execution record
	execution := models.QueryExecution{
		ID:      uuid.New(),
		QueryID: query.ID,
		Status:  models.QueryStatusPending,
	}
	if err := db.DB.Create(&execution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create execution"})
		return
	}

	// Get target agents
	var agents []models.Agent
	if len(req.AgentIDs) == 0 {
		// All online agents
		if err := db.DB.Where("state = ?", models.AgentStateConnected).Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
			return
		}
	} else {
		// Specific agents
		if err := db.DB.Where("id IN ?", req.AgentIDs).Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
			return
		}
	}

	// Filter by platform if specified
	if req.Platform != "" {
		// Note: In a real implementation, you'd check agent OS
		// For now, we'll run on all selected agents
	}

	// Create execution results for each agent
	for _, agent := range agents {
		result := models.QueryExecutionResult{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			AgentID:     agent.ID,
			Status:      models.QueryStatusPending,
		}
		db.DB.Create(&result)
	}

	// Start execution in background
	go managergrpc.ExecuteQueryOnAgents(execution.ID, query, agents)

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": execution.ID,
		"status":       "pending",
		"agent_count":  len(agents),
	})
}
