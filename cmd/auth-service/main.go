package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	authgrpc "github.com/KarpovYuri/caraudio-backend/internal/auth/adapters/grpc"
	authservice "github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	authconfig "github.com/KarpovYuri/caraudio-backend/internal/auth/config"
	authdb "github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	go func() {
		log.Printf("Auth gRPC Service listening on %s", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost"+cfg.GRPCPort, opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	log.Printf("Auth HTTP Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", allowCORS(mux, cfg.AllowedOrigin)); err != nil {
		log.Fatalf("Failed to serve Gateway: %v", err)
	}
}

func allowCORS(h http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}
