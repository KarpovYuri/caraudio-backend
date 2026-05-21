package grpc

import (
	"context"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	authv1.UnimplementedUserServiceServer
	userService services.UserService
	authService services.AuthService
}

func NewUserGRPCServer(
	userService services.UserService,
	authService services.AuthService,
) *UserGRPCServer {
	return &UserGRPCServer{
		userService: userService,
		authService: authService,
	}
}

func (s *UserGRPCServer) CreateUser(
	ctx context.Context,
	req *authv1.CreateUserRequest,
) (*authv1.CreateUserResponse, error) {
	if err := requireAdmin(ctx, s.authService); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Login == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "login and password are required")
	}

	user, err := s.userService.CreateUser(ctx, req.Login, req.Password, req.Role)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.CreateUserResponse{User: toProtoUser(user)}, nil
}

func (s *UserGRPCServer) UpdateUser(
	ctx context.Context,
	req *authv1.UpdateUserRequest,
) (*authv1.UpdateUserResponse, error) {
	if err := requireAdmin(ctx, s.authService); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := s.userService.UpdateUser(ctx, req.Id, req.Login, req.Password, req.Role)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.UpdateUserResponse{User: toProtoUser(user)}, nil
}

func (s *UserGRPCServer) DeleteUser(
	ctx context.Context,
	req *authv1.DeleteUserRequest,
) (*authv1.DeleteUserResponse, error) {
	if err := requireAdmin(ctx, s.authService); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if err := s.userService.DeleteUser(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.DeleteUserResponse{Success: true}, nil
}

func (s *UserGRPCServer) GetUser(
	ctx context.Context,
	req *authv1.GetUserRequest,
) (*authv1.GetUserResponse, error) {
	if err := requireAdmin(ctx, s.authService); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &authv1.GetUserResponse{User: toProtoUser(user)}, nil
}

func toProtoUser(user *domain.User) *authv1.User {
	return &authv1.User{
		Id:        user.ID,
		Login:     user.Login,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
