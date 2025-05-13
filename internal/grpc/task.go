package grpcclient

import (
	"context"
	"log"
	"openshield-manager/proto"
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
