package grpc

import (
	"context"
	"strings"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
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

func requireAdmin(ctx context.Context, authService services.AuthService) error {
	token := extractBearerToken(ctx)
	if token == "" {
		return domain.ErrUnauthorized
	}

	_, role, isValid, err := authService.ValidateToken(ctx, token)
	if err != nil || !isValid {
		return domain.ErrUnauthorized
	}
	if role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	return nil
}
