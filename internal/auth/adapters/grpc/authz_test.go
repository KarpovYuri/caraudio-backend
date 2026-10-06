package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
	"google.golang.org/grpc/metadata"
)

type fakeAuthService struct {
	validateFn func(ctx context.Context, accessToken string) (userID, role string, isValid bool, err error)
}

func (f *fakeAuthService) Login(
	context.Context,
	string,
	string,
	bool,
) (*domain.User, string, string, error) {
	return nil, "", "", errors.New("not implemented")
}

func (f *fakeAuthService) Refresh(context.Context, string) (string, error) {
	return "", errors.New("not implemented")
}

func (f *fakeAuthService) ValidateToken(
	ctx context.Context,
	accessToken string,
) (string, string, bool, error) {
	if f.validateFn == nil {
		return "", "", false, errors.New("validateFn is not set")
	}
	return f.validateFn(ctx, accessToken)
}

func (f *fakeAuthService) Logout(context.Context, string) error {
	return errors.New("not implemented")
}

func TestRequireAdminSuccess(t *testing.T) {
	token, err := utils.GenerateJWT("user-1", domain.RoleAdmin, "secret", time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	auth := &fakeAuthService{
		validateFn: func(_ context.Context, accessToken string) (string, string, bool, error) {
			if accessToken != token {
				t.Fatalf("unexpected token")
			}
			return "user-1", domain.RoleAdmin, true, nil
		},
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+token,
	))

	if err := requireAdmin(ctx, auth); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRequireAdminMissingToken(t *testing.T) {
	err := requireAdmin(context.Background(), &fakeAuthService{})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestRequireAdminForbidden(t *testing.T) {
	token, err := utils.GenerateJWT("user-1", domain.RoleUser, "secret", time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	auth := &fakeAuthService{
		validateFn: func(_ context.Context, _ string) (string, string, bool, error) {
			return "user-1", domain.RoleUser, true, nil
		},
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+token,
	))

	err = requireAdmin(ctx, auth)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRequireAdminInvalidToken(t *testing.T) {
	auth := &fakeAuthService{
		validateFn: func(_ context.Context, _ string) (string, string, bool, error) {
			return "", "", false, domain.ErrInvalidToken
		},
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer bad-token",
	))

	err := requireAdmin(ctx, auth)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
