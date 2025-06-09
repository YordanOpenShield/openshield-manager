package managergrpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"
	"openshield-manager/proto"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetTools handles the GetTools RPC.
func (c *AgentClient) GetTools(ctx context.Context) (*proto.GetToolsResponse, error) {
	res, err := c.client.GetTools(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ExecuteTool handles the ExecuteTool RPC.
func (c *AgentClient) ExecuteTool(ctx context.Context, name string, action string, options []string) (*proto.ExecuteToolResponse, error) {
	req := &proto.ExecuteToolRequest{
		Name:    name,
		Action:  action,
		Options: options,
	}

	res, err := c.client.ExecuteTool(ctx, req)
	if err != nil {
		return nil, err
	}

	return &proto.ExecuteToolResponse{
		Name:     res.Name,
		Action:   res.Action,
		Accepted: res.Accepted,
		Message:  res.Message,
	}, nil
}

// ReportToolExecutionStatus handles the ReportToolExecutionStatus RPC.
func (c *AgentClient) ReportToolExecutionStatus(ctx context.Context, name string, action string) (*proto.ToolExecutionStatusResponse, error) {
	req := &proto.ToolExecutionStatusRequest{
		Name:   name,
		Action: action,
	}

	res, err := c.client.ReportToolExecutionStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TrackTaskStatus periodically checks the task status from the agent and updates the DB.
// It stops automatically when task status is COMPLETED or FAILED.
func TrackToolActionStatus(agentAddr string, tool string, action string, execId uuid.UUID, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	errorCount := 0
	const maxErrors = 5

	for {
		select {
		case <-ticker.C:
			status, result, err := checkToolActionStatus(agentAddr, tool, action)
			if err != nil {
				errorCount++
				log.Printf("[TRACKER] Error tracking tool %s action %s: %v (attempt %d)", tool, action, err, errorCount)
				if errorCount >= maxErrors {
					log.Printf("[TRACKER] Giving up on tracking tool %s action %s after %d failed attempts", tool, action, errorCount)
					_ = db.DB.Model(&models.ToolActionExecution{}).
						Where("id = ?", execId).
						Update("status", proto.TaskStatus_FAILED.String()).Error
					return
				}
				continue
			}

			err = db.DB.Model(&models.ToolActionExecution{}).
				Where("id = ?", execId).
				Updates(map[string]interface{}{
					"status": status.String(),
					"result": result,
				}).Error
			if err != nil {
				log.Printf("[TRACKER] DB update failed for tool %s action %s: %v", tool, action, err)
				continue
			}

			if status == proto.TaskStatus_COMPLETED || status == proto.TaskStatus_FAILED {
				log.Printf("[TRACKER] Tool %s action %s finished with status: %s", tool, action, status.String())
				return // Stop polling
			}
		}
	}
}

func checkToolActionStatus(agentAddr, tool string, action string) (proto.TaskStatus, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewAgentClient(agentAddr)
	if err != nil {
		return proto.TaskStatus_PENDING, "", err
	}

	resp, err := client.ReportToolExecutionStatus(ctx, tool, action)
	if err != nil {
		return proto.TaskStatus_PENDING, "", err
	}
	// Decode the base64-encoded result field
	decodedBytes, err := base64.StdEncoding.DecodeString(resp.Result)
	if err != nil {
		log.Printf("[TRACKER] Failed to decode base64 result for tool %s action %s: %q", tool, action, resp.Result)
		return proto.TaskStatus_FAILED, "", fmt.Errorf("failed to decode base64 result: %w", err)
	}
	result := utils.SanitizeString(string(decodedBytes))
	if !utf8.ValidString(result) {
		log.Printf("[TRACKER] Invalid UTF-8 in decoded result field for tool %s action %s: %q", tool, action, result)
		return proto.TaskStatus_FAILED, "", fmt.Errorf("invalid UTF-8 in decoded result field")
	}

	return resp.Status, result, nil
}
