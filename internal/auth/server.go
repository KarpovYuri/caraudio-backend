package auth

import (
	"context"
	"fmt"
	"log"

	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
}

func NewAuthServer() *Server {
	return &Server{}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	log.Printf("Received Register request for email: %s", req.GetEmail())

	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	userID := fmt.Sprintf("user-%s", req.GetEmail())
	accessToken := "simulated_access_token_for_" + userID
	refreshToken := "simulated_refresh_token_for_" + userID

	log.Printf("User %s registered successfully.", userID)

	return &authv1.RegisterResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	log.Printf("Received Login request for email: %s", req.GetEmail())

	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	userID := fmt.Sprintf("user-%s", req.GetEmail())
	accessToken := "simulated_access_token_for_" + userID
	refreshToken := "simulated_refresh_token_for_" + userID

	log.Printf("User %s logged in successfully.", userID)

	return &authv1.LoginResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	log.Printf("Received ValidateToken request for token: %s", req.GetAccessToken())

	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}

	isValid := req.GetAccessToken() == "simulated_access_token_for_user-test@example.com"
	var userID string
	if isValid {
		userID = "user-test@example.com"
	} else {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	log.Printf("Token valid: %t, User ID: %s", isValid, userID)

	return &authv1.ValidateTokenResponse{
		UserId:  userID,
		IsValid: isValid,
	}, nil
}

func RegisterServer(grpcServer *grpc.Server) {
	authv1.RegisterAuthServiceServer(grpcServer, NewAuthServer())
}
