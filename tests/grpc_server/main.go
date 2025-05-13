package main

import (
	"context"
	"log"
	"net"
	"openshield-manager/proto" // Update to match your proto package path

	"google.golang.org/grpc"
)

type agentServer struct {
	proto.UnimplementedAgentTaskServiceServer
}

func (s *agentServer) AssignTask(ctx context.Context, req *proto.AssignTaskRequest) (*proto.AssignTaskResponse, error) {
	log.Printf("[Mock Agent] Received Task ID: %s, Command: %s", req.Task.Id, req.Job.Command)

	// Simulate accepting the task
	return &proto.AssignTaskResponse{
		Accepted: true,
		Message:  "Task received and queued for execution",
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051") // use the port your client expects
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAgentTaskServiceServer(grpcServer, &agentServer{})

	log.Println("Mock Agent gRPC server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
