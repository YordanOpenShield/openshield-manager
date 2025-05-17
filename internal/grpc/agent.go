package managergrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/proto"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ManagerServer) RegisterAgent(ctx context.Context, req *proto.RegisterAgentRequest) (*proto.RegisterAgentResponse, error) {
	// Create a new agent
	agentID := uuid.New()
	agent := models.Agent{
		ID:       agentID,
		Token:    uuid.New().String(),
		LastSeen: time.Now(),
		DeviceID: req.DeviceId,
		State:    "DISCONNECTED",
	}

	db.DB.Create(&agent)

	return &proto.RegisterAgentResponse{
		Id:    agentID.String(),
		Token: agent.Token,
	}, nil
}

func (s *ManagerServer) UnregisterAgent(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	// Extract agent ID from context (e.g., from metadata or authentication)
	agentID, err := extractAgentIDFromContext(ctx)
	if err != nil {
		log.Printf("[UNREGISTER] Failed to extract agent ID: %v", err)
		return nil, err
	}

	// Delete agent and its addresses
	if err := db.DB.Where("agent_id = ?", agentID).Delete(&models.AgentAddress{}).Error; err != nil {
		log.Printf("[UNREGISTER] Failed to delete agent addresses: %v", err)
		return nil, err
	}
	if err := db.DB.Where("id = ?", agentID).Delete(&models.Agent{}).Error; err != nil {
		log.Printf("[UNREGISTER] Failed to delete agent: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Helper function to extract agent ID from context (stub, replace with real implementation)
func extractAgentIDFromContext(ctx context.Context) (string, error) {
	// TODO: Extract agent ID from context metadata or authentication
	return "", fmt.Errorf("not implemented")
}

// TryAgentAddresses attempts to connect to the agent using all its addresses.
// On a successful connection, updates the agent's address in the database.
func (c *AgentClient) TryAgentAddresses(agentID string) error {
	var agent models.Agent
	if err := db.DB.Where("id = ?", agentID).First(&agent).Error; err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	var addresses []models.AgentAddress
	if err := db.DB.Where("agent_id = ?", agentID).Find(&addresses).Error; err != nil {
		return fmt.Errorf("failed to get agent addresses: %w", err)
	}

	for _, addr := range addresses {
		_, err := c.client.TryAgentAddress(context.Background(), &emptypb.Empty{})
		if err == nil {
			// Set this address as the primary one in the database
			agent.Address = addr.Address
			if err := db.DB.Save(&agent).Error; err != nil {
				return fmt.Errorf("failed to update agent primary address: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("could not connect to any agent address")
}

// Heartbeat handles the heartbeat from the agent
func (s *ManagerServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	log.Printf("[HEARTBEAT] Received heartbeat from agent %s", req.AgentId)

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.AgentId).First(&agent).Error; err != nil {
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	// Check if the response is valid JSON
	var message struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.Unmarshal([]byte(req.Message), &message); err != nil {
		log.Printf("Failed to unmarshal heartbeat response: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	// Remove existing addresses for the agent
	if err := db.DB.Where("agent_id = ?", agent.ID).Delete(&models.AgentAddress{}).Error; err != nil {
		log.Printf("Failed to delete old agent addresses: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}
	// Save new addresses from the request
	for _, addr := range message.Addresses {
		address := models.AgentAddress{
			AgentID: agent.ID,
			Address: addr,
		}
		if err := db.DB.Create(&address).Error; err != nil {
			log.Printf("Failed to save agent address: %v", err)
			return &proto.HeartbeatResponse{Ok: false}, err
		}
	}

	// Update the agent's last seen time
	agent.LastSeen = time.Now()
	agent.State = "CONNECTED"
	if err := db.DB.Save(&agent).Error; err != nil {
		log.Printf("Failed to update agent last seen time: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	return &proto.HeartbeatResponse{Ok: true}, nil
}
