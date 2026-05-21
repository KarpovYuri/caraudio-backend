package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/utils"
)

type UserService interface {
	CreateUser(ctx context.Context, login, password, role string) (*domain.User, error)
	UpdateUser(ctx context.Context, id, login, password, role string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUser(ctx context.Context, id string) (*domain.User, error)
}

type userService struct {
	userRepo postgres.UserRepository
}

func NewUserService(userRepo postgres.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(
	ctx context.Context,
	login, password, role string,
) (*domain.User, error) {
	if role == "" {
		role = domain.RoleUser
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &domain.User{
		ID:        uuid.NewString(),
		Login:     login,
		Password:  hashedPassword,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUser(
	ctx context.Context,
	id, login, password, role string,
) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if login == "" && password == "" && role == "" {
		return nil, domain.ErrInvalidArgument
	}

	if login != "" {
		user.Login = login
	}
	if role != "" {
		user.Role = role
	}
	if password != "" {
		hashedPassword, hashErr := utils.HashPassword(password)
		if hashErr != nil {
			return nil, hashErr
		}
		user.Password = hashedPassword
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.DeleteUser(ctx, id)
}

func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}
