package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type AuthService interface {
	Login(
		ctx context.Context,
		login, password string,
	) (*domain.User, string, string, error)

	Refresh(ctx context.Context, refreshToken string) (string, error)

	ValidateToken(
		ctx context.Context,
		accessToken string,
	) (userID, role string, isValid bool, err error)

	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	userRepo  postgres.UserRepository
	tokenRepo postgres.RefreshTokenRepository
	jwtSecret string
}

func NewAuthService(
	userRepo postgres.UserRepository,
	tokenRepo postgres.RefreshTokenRepository,
	jwtSecret string,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Login(
	ctx context.Context,
	login, password string,
) (*domain.User, string, string, error) {

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", "", domain.ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateJWT(
		user.ID,
		user.Role,
		s.jwtSecret,
		accessTokenTTL,
	)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken := uuid.NewString()
	refreshTokenHash := utils.HashString(refreshToken)

	err = s.tokenRepo.Create(ctx, &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Refresh(
	ctx context.Context,
	refreshToken string,
) (string, error) {

	hash := utils.HashString(refreshToken)

	rt, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return "", domain.ErrInvalidToken
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.DeleteByHash(ctx, hash)
		return "", domain.ErrInvalidToken
	}

	user, err := s.userRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return "", err
	}

	return utils.GenerateJWT(
		user.ID,
		user.Role,
		s.jwtSecret,
		accessTokenTTL,
	)
}

func (s *authService) ValidateToken(
	_ context.Context,
	accessToken string,
) (string, string, bool, error) {

	claims, err := utils.ParseJWT(accessToken, s.jwtSecret)
	if err != nil {
		return "", "", false, nil
	}

	return claims.UserID, claims.Role, true, nil
}

func (s *authService) Logout(
	ctx context.Context,
	refreshToken string,
) error {

	hash := utils.HashString(refreshToken)
	return s.tokenRepo.DeleteByHash(ctx, hash)
}
