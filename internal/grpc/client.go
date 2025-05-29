package managergrpc

import (
	"openshield-manager/internal/utils"
	"openshield-manager/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AgentClient wraps the gRPC client and connection.
type AgentClient struct {
	conn   *grpc.ClientConn
	client proto.AgentServiceClient
}

// NewAgentClient initializes and returns a new AgentClient.
func NewAgentClient(agentAddress string) (*AgentClient, error) {
	// Load TLS credentials
	tlsConfig, err := utils.LoadClientTLSCredentials()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		agentAddress+":50051",
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), // Use TLS in production
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewAgentServiceClient(conn)

	return &AgentClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close terminates the connection to the agent.
func (a *AgentClient) Close() {
	a.conn.Close()
}
