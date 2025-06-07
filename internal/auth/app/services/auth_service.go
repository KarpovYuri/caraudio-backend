package services

import (
	"context"
	"errors"
	"fmt"
	utils2 "github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
	"log"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	ValidateToken(ctx context.Context, token string) (userID, role string, isValid bool, err error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	userRepo        postgres.UserRepository
	jwtSecret       string
	jwtExpirationHs int
}

func NewAuthService(userRepo postgres.UserRepository, jwtSecret string, jwtExpirationHs int) AuthService {
	return &authService{
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		jwtExpirationHs: jwtExpirationHs,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*domain.User, string, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, "", domain.ErrUserAlreadyExists
	}

	hashedPassword, err := utils2.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := utils2.GenerateJWT(newUser.ID, newUser.Role, s.jwtSecret, s.jwtExpirationHs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT for new user: %w", err)
	}

	return newUser, token, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to retrieve user: %w", err)
	}

	if !utils2.CheckPasswordHash(password, user.Password) {
		return nil, "", domain.ErrInvalidCredentials
	}

	token, err := utils2.GenerateJWT(user.ID, user.Role, s.jwtSecret, s.jwtExpirationHs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return user, token, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (userID, role string, isValid bool, err error) {
	claims, err := utils2.ParseJWT(token, s.jwtSecret)
	if err != nil {
		return "", "", false, fmt.Errorf("token validation failed: %w", err)
	}

	return claims.UserID, claims.Role, true, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	log.Printf("User logout simulated for token: %s", token)
	return nil
}
