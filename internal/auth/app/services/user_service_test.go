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
	svc := NewUserService(userRepo, nil)

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
	svc := NewUserService(userRepo, nil)

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
	svc := NewUserService(userRepo, nil)

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
	svc := NewUserService(userRepo, nil)

	_, err := svc.UpdateUser(ctx, "user-1", "", "", "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUserServiceDeleteUserNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := NewUserService(userRepo, nil)

	err := svc.DeleteUser(ctx, "missing-id")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserServiceUpdateUserAvatarSuccess(t *testing.T) {
	ctx := context.Background()
	existing := &domain.User{
		ID:     "user-1",
		Login:  "admin",
		Role:   domain.RoleAdmin,
		Avatar: "",
	}
	var saved *domain.User

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			if id != "user-1" {
				t.Fatalf("unexpected id: %s", id)
			}
			copy := *existing
			return &copy, nil
		},
		updateUserFn: func(_ context.Context, user *domain.User) error {
			saved = user
			return nil
		},
	}
	svc := NewUserService(userRepo, nil)

	user, err := svc.UpdateUserAvatar(ctx, "user-1", "  http://localhost:8080/static/avatars/a.webp  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Avatar != "http://localhost:8080/static/avatars/a.webp" {
		t.Fatalf("unexpected avatar: %q", user.Avatar)
	}
	if saved == nil || saved.Avatar != user.Avatar {
		t.Fatalf("expected avatar to be persisted, got %+v", saved)
	}
}

func TestUserServiceUpdateUserAvatarInvalidArgument(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(&fakeUserRepo{}, nil)

	_, err := svc.UpdateUserAvatar(ctx, "", "http://example/avatar.webp")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for empty id, got %v", err)
	}

	_, err = svc.UpdateUserAvatar(ctx, "user-1", "   ")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for empty avatar, got %v", err)
	}
}

func TestUserServiceUpdateUserAvatarNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := NewUserService(userRepo, nil)

	_, err := svc.UpdateUserAvatar(ctx, "missing", "http://example/avatar.webp")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

type fakePublicURLDeleter struct {
	deleted []string
	err     error
}

func (f *fakePublicURLDeleter) DeletePublicURL(publicURL string) error {
	f.deleted = append(f.deleted, publicURL)
	return f.err
}

func TestUserServiceDeleteUserRemovesAvatarFile(t *testing.T) {
	ctx := context.Background()
	avatarURL := "http://localhost:8080/static/avatars/old.webp"
	files := &fakePublicURLDeleter{}

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Avatar: avatarURL}, nil
		},
		deleteUserFn: func(_ context.Context, id string) error {
			if id != "user-1" {
				t.Fatalf("unexpected id: %s", id)
			}
			return nil
		},
	}
	svc := NewUserService(userRepo, files)

	if err := svc.DeleteUser(ctx, "user-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files.deleted) != 1 || files.deleted[0] != avatarURL {
		t.Fatalf("expected avatar file delete, got %#v", files.deleted)
	}
}

func TestUserServiceDeleteUserSkipsEmptyAvatar(t *testing.T) {
	ctx := context.Background()
	files := &fakePublicURLDeleter{}

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Avatar: ""}, nil
		},
	}
	svc := NewUserService(userRepo, files)

	if err := svc.DeleteUser(ctx, "user-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files.deleted) != 0 {
		t.Fatalf("expected no file delete for empty avatar, got %#v", files.deleted)
	}
}

func TestUserServiceGetUser(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, Login: "admin", Role: domain.RoleAdmin}, nil
		},
	}
	svc := NewUserService(userRepo, nil)

	user, err := svc.GetUser(ctx, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Login != "admin" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserServiceListUsersTrimsFilters(t *testing.T) {
	ctx := context.Background()
	var got domain.UserListFilter

	userRepo := &fakeUserRepo{
		listUsersFn: func(_ context.Context, filter domain.UserListFilter) (*domain.UserListResult, error) {
			got = filter
			return &domain.UserListResult{}, nil
		},
	}
	svc := NewUserService(userRepo, nil)

	_, err := svc.ListUsers(ctx, domain.UserListFilter{
		Page:     1,
		PageSize: 10,
		Search:   "  admin  ",
		Role:     "  user  ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Search != "admin" || got.Role != "user" {
		t.Fatalf("expected trimmed filters, got search=%q role=%q", got.Search, got.Role)
	}
}
