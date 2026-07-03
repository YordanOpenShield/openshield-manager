package managergrpc

import (
	"context"
	"log"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/proto"
	"time"

	"github.com/google/uuid"
)

// ExecuteQueryOnAgents executes a query on multiple agents
func ExecuteQueryOnAgents(executionID uuid.UUID, query models.Query, agents []models.Agent) {
	// Update execution status to running
	db.DB.Model(&models.QueryExecution{}).Where("id = ?", executionID).Update("status", models.QueryStatusRunning)

	// Execute on each agent concurrently
	for _, agent := range agents {
		go func(agent models.Agent) {
			if err := executeQueryOnAgent(executionID, query, agent); err != nil {
				log.Printf("[QUERY] Failed to execute query on agent %s: %v", agent.ID, err)
				// Update result to failed
				db.DB.Model(&models.QueryExecutionResult{}).
					Where("execution_id = ? AND agent_id = ?", executionID, agent.ID).
					Updates(map[string]interface{}{
						"status": models.QueryStatusFailed,
						"error":  err.Error(),
					})
			}
		}(agent)
	}

	// Check if all results are complete
	go monitorQueryExecution(executionID, len(agents))
}

// executeQueryOnAgent executes a query on a single agent
func executeQueryOnAgent(executionID uuid.UUID, query models.Query, agent models.Agent) error {
	// Update result status to running
	now := time.Now()
	db.DB.Model(&models.QueryExecutionResult{}).
		Where("execution_id = ? AND agent_id = ?", executionID, agent.ID).
		Updates(map[string]interface{}{
			"status":     models.QueryStatusRunning,
			"started_at": &now,
		})

	// Create gRPC client
	client, err := NewAgentClient(agent.Address)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Send query to agent
	req := &proto.QueryExecutionRequest{
		ExecutionId: executionID.String(),
		Query: &proto.Query{
			Id:       query.ID.String(),
			Name:     query.Name,
			Sql:      query.SQL,
			Platform: query.Platform,
		},
	}

	resp, err := client.client.ExecuteQuery(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Accepted {
		return err
	}

	log.Printf("[QUERY] Query accepted by agent %s", agent.ID)
	return nil
}

// ReportQueryResult handles query results reported by agents
func (s *ManagerServer) ReportQueryResult(ctx context.Context, req *proto.QueryResultRequest) (*proto.QueryResultResponse, error) {
	executionID, err := uuid.Parse(req.ExecutionId)
	if err != nil {
		return nil, err
	}

	// Find the execution result for this agent
	var result models.QueryExecutionResult
	if err := db.DB.Where("execution_id = ?", executionID).First(&result).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.QueryStatusCompleted,
		"completed_at": &now,
	}

	if req.Success {
		updates["result_json"] = req.ResultJson
	} else {
		updates["status"] = models.QueryStatusFailed
		updates["error"] = req.Error
	}

	db.DB.Model(&result).Updates(updates)

	return &proto.QueryResultResponse{Received: true}, nil
}

// monitorQueryExecution monitors the execution and updates status when complete
func monitorQueryExecution(executionID uuid.UUID, totalAgents int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-ticker.C:
			var completed, failed int64
			db.DB.Model(&models.QueryExecutionResult{}).
				Where("execution_id = ? AND status = ?", executionID, models.QueryStatusCompleted).
				Count(&completed)
			db.DB.Model(&models.QueryExecutionResult{}).
				Where("execution_id = ? AND status = ?", executionID, models.QueryStatusFailed).
				Count(&failed)

			if completed+failed >= int64(totalAgents) {
				// All agents completed
				status := models.QueryStatusCompleted
				if failed == int64(totalAgents) {
					status = models.QueryStatusFailed
				}
				now := time.Now()
				db.DB.Model(&models.QueryExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
					"status":       status,
					"completed_at": &now,
				})
				log.Printf("[QUERY] Execution %s completed", executionID)
				return
			}
		case <-timeout:
			// Timeout - mark as failed
			now := time.Now()
			db.DB.Model(&models.QueryExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
				"status":       models.QueryStatusFailed,
				"completed_at": &now,
			})
			log.Printf("[QUERY] Execution %s timed out", executionID)
			return
		}
	}
}

// SubmitQueryExecution handles query execution submissions from agents (for distributed queries)
func (s *ManagerServer) SubmitQueryExecution(ctx context.Context, req *proto.QueryExecutionRequest) (*proto.QueryExecutionResponse, error) {
	// This is used by agents to report they're ready to execute
	return &proto.QueryExecutionResponse{
		Accepted: true,
		Message:  "Query execution acknowledged",
	}, nil
}
