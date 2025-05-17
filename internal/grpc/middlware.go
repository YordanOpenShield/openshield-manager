package managergrpc

import (
	"context"
	"strings"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AgentTokenInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Exclude RegisterAgent from token check
		if strings.HasSuffix(info.FullMethod, "RegisterAgent") {
			return handler(ctx, req)
		}
		// Only check token for certain RPCs
		if strings.HasSuffix(info.FullMethod, "AssignTask") || strings.HasSuffix(info.FullMethod, "ReportTaskStatus") {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "missing metadata")
			}
			tokens := md["agent-token"]
			if len(tokens) == 0 {
				return nil, status.Error(codes.Unauthenticated, "missing agent token")
			}
			token := tokens[0]
			// Validate token against DB
			var agent models.Agent
			if err := db.DB.Where("token = ?", token).First(&agent).Error; err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid agent token")
			}
			// Optionally, set agent info in context for handler use
		}
		// Continue to handler
		return handler(ctx, req)
	}
}
