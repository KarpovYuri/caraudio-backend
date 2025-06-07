package main

import (
	"github.com/KarpovYuri/caraudio-backend/internal/auth"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	grpcPort := ":50051"

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	auth.RegisterServer(grpcServer)

	log.Printf("Auth Service starting gRPC server on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
