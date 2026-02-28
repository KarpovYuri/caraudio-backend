package grpc

import (
	"context"
	"errors"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	authv1.UnimplementedAuthServiceServer
	authService services.AuthService
}

func NewAuthGRPCServer(authService services.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService: authService,
	}
}

func (s *AuthGRPCServer) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {

	if req.Login == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "login and password are required")
	}

	user, accessToken, refreshToken, err :=
		s.authService.Login(ctx, req.Login, req.Password, req.RememberMe)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.LoginResponse{
		UserId:       user.ID,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthGRPCServer) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.RefreshResponse, error) {

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, err := s.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(
	ctx context.Context,
	req *authv1.ValidateTokenRequest,
) (*authv1.ValidateTokenResponse, error) {

	if req.AccessToken == "" {
		return &authv1.ValidateTokenResponse{
			IsValid: false,
		}, nil
	}

	userID, role, isValid, err :=
		s.authService.ValidateToken(ctx, req.AccessToken)

	if err != nil || !isValid {
		return &authv1.ValidateTokenResponse{
			IsValid: false,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		UserId:  userID,
		Role:    role,
		IsValid: true,
	}, nil
}

func (s *AuthGRPCServer) Logout(
	ctx context.Context,
	req *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	if err := s.authService.Logout(ctx, req.RefreshToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authv1.LogoutResponse{
		Success: true,
	}, nil
}
