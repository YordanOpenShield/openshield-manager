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
)

// SendTask sends a task and job to the agent via gRPC.
func (a *AgentClient) SendTask(ctx context.Context, task *proto.Task, job *proto.Job) error {
	req := &proto.AssignTaskRequest{
		Task: task,
		Job:  job,
	}

	res, err := a.client.AssignTask(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("Task sent to agent. Accepted: %v, Message: %s", res.GetAccepted(), res.GetMessage())
	return nil
}

// ReportTaskStatus asks the agent for the status of a given job ID.
func (a *AgentClient) ReportTaskStatus(ctx context.Context, jobID string) (*proto.JobStatusResponse, error) {
	req := &proto.JobStatusRequest{
		JobId: jobID,
	}

	res, err := a.client.ReportTaskStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TrackTaskStatus periodically checks the task status from the agent and updates the DB.
// It stops automatically when task status is COMPLETED or FAILED.
func TrackTaskStatus(agentAddr string, taskID string, jobID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	errorCount := 0
	const maxErrors = 5

	for {
		select {
		case <-ticker.C:
			status, result, err := checkTaskStatus(agentAddr, jobID)
			if err != nil {
				errorCount++
				log.Printf("[TRACKER] Error tracking task %s: %v (attempt %d)", taskID, err, errorCount)
				if errorCount >= maxErrors {
					log.Printf("[TRACKER] Giving up on tracking task %s after %d failed attempts", taskID, errorCount)
					_ = db.DB.Model(&models.Task{}).
						Where("id = ?", taskID).
						Update("status", proto.TaskStatus_FAILED.String()).Error
					return
				}
				continue
			}

			err = db.DB.Model(&models.Task{}).
				Where("id = ?", taskID).
				Updates(map[string]interface{}{
					"status": status.String(),
					"result": result,
				}).Error
			if err != nil {
				log.Printf("[TRACKER] DB update failed for task %s: %v", taskID, err)
				continue
			}

			if status == proto.TaskStatus_COMPLETED || status == proto.TaskStatus_FAILED {
				log.Printf("[TRACKER] Task %s finished with status: %s", taskID, status.String())
				return // Stop polling
			}
		}
	}
}

func checkTaskStatus(agentAddr, jobID string) (proto.TaskStatus, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewAgentClient(agentAddr)
	if err != nil {
		return proto.TaskStatus_PENDING, "", err
	}

	resp, err := client.ReportTaskStatus(ctx, jobID)
	if err != nil {
		return proto.TaskStatus_PENDING, "", err
	}
	// Decode the base64-encoded result field
	decodedBytes, err := base64.StdEncoding.DecodeString(resp.Result)
	if err != nil {
		log.Printf("[TRACKER] Failed to decode base64 result for job %s: %q", jobID, resp.Result)
		return proto.TaskStatus_FAILED, "", fmt.Errorf("failed to decode base64 result: %w", err)
	}
	result := utils.SanitizeString(string(decodedBytes))
	if !utf8.ValidString(result) {
		log.Printf("[TRACKER] Invalid UTF-8 in decoded result field for job %s: %q", jobID, result)
		return proto.TaskStatus_FAILED, "", fmt.Errorf("invalid UTF-8 in decoded result field")
	}

	return resp.Status, result, nil
}
