package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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

	cookieValue := fmt.Sprintf("refresh_token=%s; Path=/; HttpOnly; SameSite=Lax; MaxAge=%d",
		refreshToken, 30*24*60*60)

	header := metadata.Pairs("Set-Cookie", cookieValue)

	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Error(codes.Internal, "failed to send response headers")
	}

	return &authv1.LoginResponse{
		UserId:      user.ID,
		Role:        user.Role,
		AccessToken: accessToken,
	}, nil
}

func (s *AuthGRPCServer) Refresh(
	ctx context.Context,
	_ *authv1.RefreshRequest,
) (*authv1.RefreshResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	var refreshToken string
	cookieHeader := md.Get("grpcgateway-cookie")

	if len(cookieHeader) > 0 {
		refreshToken = extractToken(cookieHeader[0], "refresh_token")
	}

	if refreshToken == "" {
		return nil, status.Error(codes.Unauthenticated, "refresh token is missing in cookies")
	}

	accessToken, err := s.authService.Refresh(ctx, refreshToken)
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
	_ *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	var refreshToken string

	if ok {
		cookieHeader := md.Get("grpcgateway-cookie")
		if len(cookieHeader) > 0 {
			refreshToken = extractToken(cookieHeader[0], "refresh_token")
		}
	}

	if refreshToken != "" {
		_ = s.authService.Logout(ctx, refreshToken)
	}

	deleteCookie := "refresh_token=; Path=/; HttpOnly; SameSite=Lax; MaxAge=-1"

	header := metadata.Pairs("Set-Cookie", deleteCookie)

	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Error(codes.Internal, "failed to send logout headers")
	}

	return &authv1.LogoutResponse{
		Success: true,
	}, nil
}

func extractToken(cookieStr, name string) string {
	parts := strings.Split(cookieStr, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, name+"=") {
			return strings.TrimPrefix(p, name+"=")
		}
	}
	return ""
}
