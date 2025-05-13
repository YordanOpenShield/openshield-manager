package grpcclient

import (
	"openshield-manager/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient wraps the gRPC client and connection.
type AgentClient struct {
	conn   *grpc.ClientConn
	client proto.AgentTaskServiceClient
}

// NewAgentClient initializes and returns a new AgentClient.
func NewAgentClient(agentAddress string /*, timeout time.Duration*/) (*AgentClient, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel()

	// Set default agentAddress if empty
	if agentAddress == "" {
		agentAddress = "localhost:50051"
	}
	// Set default timeout if zero
	// if timeout == 0 {
	// 	timeout = 60 * time.Second
	// }

	conn, err := grpc.NewClient(
		agentAddress,
		// ctx,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use TLS in production
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewAgentTaskServiceClient(conn)

	return &AgentClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close terminates the connection to the agent.
func (a *AgentClient) Close() {
	a.conn.Close()
}
