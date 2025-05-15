package grpcclient

import (
	"openshield-manager/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient wraps the gRPC client and connection.
type AgentClient struct {
	conn   *grpc.ClientConn
	client proto.AgentServiceClient
}

// NewAgentClient initializes and returns a new AgentClient.
func NewAgentClient(agentAddress string) (*AgentClient, error) {

	conn, err := grpc.NewClient(
		agentAddress+":50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use TLS in production
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
