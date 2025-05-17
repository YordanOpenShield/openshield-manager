package managergrpc

import (
	"context"
	"encoding/json"
	"log"
	"openshield-manager/proto"
)

// Heartbeat handles the heartbeat from the agent
func (s *ManagerServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	log.Printf("[HEARTBEAT] Received heartbeat from agent %s", req.AgentId)

	// Check if the response is valid JSON
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(req.Message), &message); err != nil {
		log.Printf("Failed to unmarshal heartbeat response: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	return &proto.HeartbeatResponse{Ok: true}, nil
}
