package managergrpc

import (
	"fmt"
	"log"
	"net"
	"openshield-manager/proto"

	"google.golang.org/grpc"
)

// AgentServer implements proto.AgentServiceServer
type ManagerServer struct {
	proto.UnimplementedManagerServiceServer
}

func StartGRPCServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Register the gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterManagerServiceServer(grpcServer, &ManagerServer{})

	log.Printf("[MANAGER] gRPC server listening on port %d", port)
	return grpcServer.Serve(lis)
}
