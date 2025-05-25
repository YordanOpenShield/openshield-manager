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

	if err := db.DB.Create(&agent).Error; err != nil {
		log.Printf("[REGISTER] Failed to create agent: %v", err)
		return nil, err
	}

	log.Printf("[REGISTER] New agent registered with ID: %s", agentID)

	return &proto.RegisterAgentResponse{
		Id:    agentID.String(),
		Token: agent.Token,
	}, nil
}

func (c *AgentClient) UnregisterAgentAsk(ctx context.Context) error {
	_, err := c.client.UnregisterAgentAsk(ctx, &emptypb.Empty{})
	return err
}

func (s *ManagerServer) UnregisterAgent(ctx context.Context, req *proto.UnregisterAgentRequest) (*emptypb.Empty, error) {
	agentID := req.Id
	log.Printf("[UNREGISTER] Unregistering agent with ID: %s", agentID)

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

// Heartbeat handles the heartbeat from the agent
func (s *ManagerServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	log.Printf("[HEARTBEAT] Received heartbeat from agent %s", req.AgentId)

	// Check if the agent exists
	var agent models.Agent
	if err := db.DB.Where("id = ?", req.AgentId).First(&agent).Error; err != nil {
		log.Printf("[HEARTBEAT] Agent not found: %s.", req.AgentId)
		return nil, err
	}

	// Retrieve JSON data from the request
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

	addr, err := TryAgentAddresses(agent.ID.String())
	if err != nil {
		log.Printf("Failed to connect to agent addresses: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	// Update the agent's last seen time
	agent.LastSeen = time.Now()
	agent.Address = addr.Address
	agent.State = "CONNECTED"
	if err := db.DB.Save(&agent).Error; err != nil {
		log.Printf("Failed to update agent last seen time: %v", err)
		return &proto.HeartbeatResponse{Ok: false}, err
	}

	return &proto.HeartbeatResponse{Ok: true}, nil
}

// TryAgentAddresses attempts to connect to the agent using all its addresses.
// On a successful connection, updates the agent's address in the database.
func TryAgentAddresses(agentID string) (*models.AgentAddress, error) {
	var agent models.Agent
	if err := db.DB.Where("id = ?", agentID).First(&agent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	var addresses []models.AgentAddress
	if err := db.DB.Where("agent_id = ?", agentID).Find(&addresses).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent addresses: %w", err)
	}

	for _, addr := range addresses {
		// Create a new gRPC client for each address
		c, err := NewAgentClient(addr.Address)
		if err != nil {
			log.Fatalf("Failed to create gRPC client: %v", err)
		}

		// Attempt to connect to the agent using the address
		log.Printf("Trying agent address: %s", addr.Address)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = c.client.TryAgentAddress(ctx, &emptypb.Empty{})
		if err == nil {
			// Set this address as the primary one in the database
			log.Printf("Connected to agent at address: %s", addr.Address)
			return &addr, nil
		}
		log.Printf("Failed to connect to agent address %s: %v", addr.Address, err)
	}
	return nil, fmt.Errorf("could not connect to any agent address")
}
