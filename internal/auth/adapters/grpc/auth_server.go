package grpc

import (
	"context"
	"log"

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
	return &AuthGRPCServer{authService: authService}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	user, token, err := s.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
		}
		log.Printf("Error registering user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &authv1.RegisterResponse{
		UserId:       user.ID,
		AccessToken:  token,
		RefreshToken: "",
	}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, token, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
		}
		log.Printf("Error logging in user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	return &authv1.LoginResponse{
		UserId:       user.ID,
		AccessToken:  token,
		RefreshToken: "",
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, role, isValid, err := s.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil || !isValid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token: %v", err)
	}

	return &authv1.ValidateTokenResponse{
		UserId:  userID,
		IsValid: isValid,
		Role:    role,
	}, nil
}

func (s *AuthGRPCServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := s.authService.Logout(ctx, req.AccessToken)
	if err != nil {
		log.Printf("Error during logout: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to logout: %v", err)
	}
	log.Printf("Logout successful for token: %s", req.AccessToken)
	return &authv1.LogoutResponse{Success: true}, nil
}
