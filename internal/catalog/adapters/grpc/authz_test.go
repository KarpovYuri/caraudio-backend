package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
	"google.golang.org/grpc/metadata"
)

func TestRequireAdmin(t *testing.T) {
	const secret = "catalog-secret"

	adminToken, err := pkgjwt.GenerateToken("admin-1", pkgjwt.RoleAdmin, secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	userToken, err := pkgjwt.GenerateToken("user-1", "user", secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	adminCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+adminToken,
	))
	if err := requireAdmin(adminCtx, secret); err != nil {
		t.Fatalf("admin expected nil, got %v", err)
	}
	if !isAdmin(adminCtx, secret) {
		t.Fatalf("expected isAdmin true")
	}

	userCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+userToken,
	))
	if err := requireAdmin(userCtx, secret); !errors.Is(err, pkgjwt.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if err := requireAdmin(context.Background(), secret); !errors.Is(err, pkgjwt.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
