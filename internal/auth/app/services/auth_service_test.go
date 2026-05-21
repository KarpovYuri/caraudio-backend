package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
)

type fakeUserRepo struct {
	createUserFn     func(ctx context.Context, user *domain.User) error
	getUserByLoginFn func(ctx context.Context, login string) (*domain.User, error)
	getUserByIDFn    func(ctx context.Context, id string) (*domain.User, error)
	updateUserFn     func(ctx context.Context, user *domain.User) error
	deleteUserFn     func(ctx context.Context, id string) error
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	if f.createUserFn == nil {
		return nil
	}
	return f.createUserFn(ctx, user)
}

func (f *fakeUserRepo) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	if f.getUserByLoginFn == nil {
		return nil, errors.New("getUserByLoginFn is not set")
	}
	return f.getUserByLoginFn(ctx, login)
}

func (f *fakeUserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if f.getUserByIDFn == nil {
		return nil, errors.New("getUserByIDFn is not set")
	}
	return f.getUserByIDFn(ctx, id)
}

func (f *fakeUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	if f.updateUserFn == nil {
		return nil
	}
	return f.updateUserFn(ctx, user)
}

func (f *fakeUserRepo) DeleteUser(ctx context.Context, id string) error {
	if f.deleteUserFn == nil {
		return nil
	}
	return f.deleteUserFn(ctx, id)
}

type fakeRefreshTokenRepo struct {
	createFn         func(ctx context.Context, token *domain.RefreshToken) error
	replaceForUserFn func(ctx context.Context, token *domain.RefreshToken) error
	getByHashFn      func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	deleteByHashFn   func(ctx context.Context, hash string) error
	deleteByUserIDFn func(ctx context.Context, userID string) error
}

func (f *fakeRefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, token)
}

func (f *fakeRefreshTokenRepo) ReplaceForUser(ctx context.Context, token *domain.RefreshToken) error {
	if f.replaceForUserFn == nil {
		return nil
	}
	return f.replaceForUserFn(ctx, token)
}

func (f *fakeRefreshTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if f.getByHashFn == nil {
		return nil, errors.New("getByHashFn is not set")
	}
	return f.getByHashFn(ctx, hash)
}

func (f *fakeRefreshTokenRepo) DeleteByHash(ctx context.Context, hash string) error {
	if f.deleteByHashFn == nil {
		return nil
	}
	return f.deleteByHashFn(ctx, hash)
}

func (f *fakeRefreshTokenRepo) DeleteByUserId(ctx context.Context, userID string) error {
	if f.deleteByUserIDFn == nil {
		return nil
	}
	return f.deleteByUserIDFn(ctx, userID)
}

func (f *fakeRefreshTokenRepo) DeleteExpired(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	ctx := context.Background()
	hashedPassword, err := utils.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	userRepo := &fakeUserRepo{
		getUserByLoginFn: func(_ context.Context, login string) (*domain.User, error) {
			if login != "admin" {
				t.Fatalf("unexpected login: %s", login)
			}
			return &domain.User{
				ID:       "user-1",
				Login:    "admin",
				Password: hashedPassword,
				Role:     "admin",
			}, nil
		},
	}
	tokenRepo := &fakeRefreshTokenRepo{}

	svc := NewAuthService(userRepo, tokenRepo, "secret-key")
	user, accessToken, refreshToken, err := svc.Login(ctx, "admin", "password123", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil || user.ID != "user-1" {
		t.Fatalf("unexpected user returned: %+v", user)
	}
	if accessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	ctx := context.Background()
	hashedPassword, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	t.Run("user not found", func(t *testing.T) {
		userRepo := &fakeUserRepo{
			getUserByLoginFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		tokenRepo := &fakeRefreshTokenRepo{}
		svc := NewAuthService(userRepo, tokenRepo, "secret-key")

		_, _, _, err := svc.Login(ctx, "admin", "whatever", false)
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		userRepo := &fakeUserRepo{
			getUserByLoginFn: func(_ context.Context, _ string) (*domain.User, error) {
				return &domain.User{
					ID:       "user-1",
					Login:    "admin",
					Password: hashedPassword,
					Role:     "admin",
				}, nil
			},
		}
		tokenRepo := &fakeRefreshTokenRepo{}
		svc := NewAuthService(userRepo, tokenRepo, "secret-key")

		_, _, _, err := svc.Login(ctx, "admin", "wrong-password", false)
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestAuthServiceLoginRepoError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("database unavailable")

	userRepo := &fakeUserRepo{
		getUserByLoginFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, repoErr
		},
	}
	tokenRepo := &fakeRefreshTokenRepo{}
	svc := NewAuthService(userRepo, tokenRepo, "secret-key")

	_, _, _, err := svc.Login(ctx, "admin", "password", false)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestAuthServiceRefreshValid(t *testing.T) {
	ctx := context.Background()
	rawRefreshToken := "plain-refresh-token"

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			if id != "user-1" {
				t.Fatalf("unexpected user id: %s", id)
			}
			return &domain.User{
				ID:   "user-1",
				Role: "admin",
			}, nil
		},
	}
	tokenRepo := &fakeRefreshTokenRepo{
		getByHashFn: func(_ context.Context, _ string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				UserID:    "user-1",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	svc := NewAuthService(userRepo, tokenRepo, "secret-key")

	accessToken, err := svc.Refresh(ctx, rawRefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
}

func TestAuthServiceRefreshExpired(t *testing.T) {
	ctx := context.Background()
	rawRefreshToken := "expired-refresh-token"
	deleteCalled := false

	userRepo := &fakeUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			t.Fatalf("user lookup must not be called for expired token")
			return nil, nil
		},
	}
	tokenRepo := &fakeRefreshTokenRepo{
		getByHashFn: func(_ context.Context, _ string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				UserID:    "user-1",
				ExpiresAt: time.Now().Add(-time.Minute),
			}, nil
		},
		deleteByHashFn: func(_ context.Context, _ string) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewAuthService(userRepo, tokenRepo, "secret-key")

	_, err := svc.Refresh(ctx, rawRefreshToken)
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
	if !deleteCalled {
		t.Fatalf("expected DeleteByHash to be called for expired token")
	}
}

func TestAuthServiceRefreshNotFound(t *testing.T) {
	ctx := context.Background()
	rawRefreshToken := "missing-refresh-token"

	userRepo := &fakeUserRepo{}
	tokenRepo := &fakeRefreshTokenRepo{
		getByHashFn: func(_ context.Context, _ string) (*domain.RefreshToken, error) {
			return nil, errors.New("sql: no rows in result set")
		},
	}
	svc := NewAuthService(userRepo, tokenRepo, "secret-key")

	_, err := svc.Refresh(ctx, rawRefreshToken)
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthServiceValidateTokenInvalidSignature(t *testing.T) {
	ctx := context.Background()
	token, err := utils.GenerateJWT("user-1", "admin", "good-secret", time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, "wrong-secret")
	_, _, isValid, err := svc.ValidateToken(ctx, token)

	if isValid {
		t.Fatalf("expected token to be invalid")
	}
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthServiceValidateTokenExpired(t *testing.T) {
	ctx := context.Background()
	token, err := utils.GenerateJWT("user-1", "admin", "secret-key", -time.Minute)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, "secret-key")
	_, _, isValid, err := svc.ValidateToken(ctx, token)

	if isValid {
		t.Fatalf("expected token to be invalid")
	}
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
