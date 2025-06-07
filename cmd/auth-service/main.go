package main

import (
	"log"
	"net"

	authgrpc "github.com/KarpovYuri/caraudio-backend/internal/auth/adapters/grpc"
	authservice "github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	authconfig "github.com/KarpovYuri/caraudio-backend/internal/auth/config"
	authdb "github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := authconfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := authdb.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := authdb.NewPostgresUserRepository(db)
	authService := authservice.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	authGRPCServer := authgrpc.NewAuthGRPCServer(authService)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	authv1.RegisterAuthServiceServer(s, authGRPCServer)
	log.Printf("Auth Service listening on %s", cfg.GRPCPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
