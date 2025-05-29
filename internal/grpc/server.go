package managergrpc

import (
	"fmt"
	"log"
	"net"
	"openshield-manager/internal/utils"
	"openshield-manager/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ManagerServer struct {
	proto.UnimplementedManagerServiceServer
}

func StartGRPCServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Load TLS credentials for the server
	tlsConfig, err := utils.LoadServerTLSCredentials()
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	// Register the gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(AgentTokenInterceptor()),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)
	proto.RegisterManagerServiceServer(grpcServer, &ManagerServer{})

	log.Printf("[MANAGER] gRPC server listening on port %d", port)
	return grpcServer.Serve(lis)
}

type ManagerRegistrationServer struct {
	proto.UnimplementedManagerServiceServer
}

func StartManagerRegistrationServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Register the gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterManagerServiceServer(grpcServer, &ManagerRegistrationServer{})

	log.Printf("[MANAGER] Registration gRPC server listening on port %d", port)
	return grpcServer.Serve(lis)
}
