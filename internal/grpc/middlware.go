package managergrpc

import (
	"context"
	"strings"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey is used for storing values in gRPC context.
type contextKey string

const (
	// ContextKeyAgentID stores the authenticated agent's ID in the context.
	ContextKeyAgentID contextKey = "agent_id"
	// ContextKeyOrgID stores the agent's organization ID in the context.
	ContextKeyOrgID contextKey = "org_id"
)

// AgentTokenInterceptor validates the agent token for all ManagerService RPCs
// except RegisterAgent (which uses registration_token instead).
// It also extracts the agent's organization ID and stores it in the context.
func AgentTokenInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Exclude RegisterAgent from token check (uses registration_token)
		if strings.HasSuffix(info.FullMethod, "RegisterAgent") {
			return handler(ctx, req)
		}

		// All other ManagerService RPCs require a valid agent token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md["agent-token"]
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing agent token")
		}
		token := tokens[0]

		// Validate token against DB and load agent with organization info
		var agent models.Agent
		if err := db.DB.Where("token = ?", token).First(&agent).Error; err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid agent token")
		}

		// Store agent ID and org ID in context for downstream handlers
		ctx = context.WithValue(ctx, ContextKeyAgentID, agent.ID.String())
		if agent.OrganizationID != nil {
			ctx = context.WithValue(ctx, ContextKeyOrgID, agent.OrganizationID.String())
		}

		return handler(ctx, req)
	}
}

// GetAgentIDFromContext extracts the authenticated agent's ID from the context.
func GetAgentIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ContextKeyAgentID).(string)
	return id, ok
}

// GetOrgIDFromContext extracts the agent's organization ID from the context.
func GetOrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(ContextKeyOrgID).(string)
	return orgID, ok
}

// ParseOrgID parses an org ID string into a uuid.UUID pointer.
// Returns nil if the string is empty.
func ParseOrgID(orgIDStr string) *uuid.UUID {
	if orgIDStr == "" {
		return nil
	}
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil
	}
	return &id
}
