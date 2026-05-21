package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
)

func TestUserServiceCreateUserSuccess(t *testing.T) {
	ctx := context.Background()
	var saved *domain.User

	userRepo := &fakeUserRepo{
		createUserFn: func(_ context.Context, user *domain.User) error {
			saved = user
			return nil
		},
	}
	svc := NewUserService(userRepo)

	user, err := svc.CreateUser(ctx, "new-user", "password123", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Role != domain.RoleUser {
		t.Fatalf("expected default role %q, got %q", domain.RoleUser, user.Role)
	}
	if saved == nil || saved.Login != "new-user" {
		t.Fatalf("unexpected saved user: %+v", saved)
	}
	if !utils.CheckPasswordHash("password123", saved.Password) {
		t.Fatalf("expected hashed password to be stored")
	}
}

func TestUserServiceCreateUserAlreadyExists(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		createUserFn: func(_ context.Context, _ *domain.User) error {
			return domain.ErrUserAlreadyExists
		},
	}
	svc := NewUserService(userRepo)

	_, err := svc.CreateUser(ctx, "existing", "password123", domain.RoleAdmin)
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserServiceUpdateUserPartial(t *testing.T) {
	ctx := context.Background()
	existing := &domain.User{
		ID:        "user-1",
		Login:     "old-login",
		Password:  "hash",
		Role:      domain.RoleUser,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			if id != "user-1" {
				t.Fatalf("unexpected id: %s", id)
			}
			copy := *existing
			return &copy, nil
		},
		updateUserFn: func(_ context.Context, user *domain.User) error {
			if user.Role != domain.RoleAdmin {
				t.Fatalf("expected role to be updated, got %q", user.Role)
			}
			return nil
		},
	}
	svc := NewUserService(userRepo)

	user, err := svc.UpdateUser(ctx, "user-1", "", "", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Role != domain.RoleAdmin {
		t.Fatalf("expected updated role in response")
	}
}

func TestUserServiceUpdateUserNoFields(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1"}, nil
		},
	}
	svc := NewUserService(userRepo)

	_, err := svc.UpdateUser(ctx, "user-1", "", "", "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUserServiceDeleteUserNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		deleteUserFn: func(_ context.Context, _ string) error {
			return domain.ErrUserNotFound
		},
	}
	svc := NewUserService(userRepo)

	err := svc.DeleteUser(ctx, "missing-id")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
