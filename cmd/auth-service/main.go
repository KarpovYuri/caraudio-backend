package main

import (
	"log"
	"net"
	"os"

	authgrpc "github.com/KarpovYuri/caraudio-backend/internal/auth/adapters/grpc"
	authservice "github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	authconfig "github.com/KarpovYuri/caraudio-backend/internal/auth/config"
	authdb "github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"

	"google.golang.org/grpc"
)

func main() {
	log.SetOutput(os.Stdout)

	cfg, err := authconfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := authdb.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	userRepo := authdb.NewPostgresUserRepository(db)
	tokenRepo := authdb.NewPgRefreshTokenRepository(db)

	authService := authservice.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWTSecret,
	)

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
