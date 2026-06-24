package grpc

import (
	"context"
	"strings"

	"github.com/KarpovYuri/caraudio-backend/pkg/jwt"
	"google.golang.org/grpc/metadata"
)

func extractBearerToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}
	value := strings.TrimSpace(values[0])
	if len(value) > 7 && strings.EqualFold(value[:7], "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}

func requireAdmin(ctx context.Context, jwtSecret string) error {
	return jwt.ValidateAdmin(extractBearerToken(ctx), jwtSecret)
}

func isAdmin(ctx context.Context, jwtSecret string) bool {
	return requireAdmin(ctx, jwtSecret) == nil
}
