package grpc

import (
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/app/services"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
)

type CatalogGRPCServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	catalogService services.CatalogService
	jwtSecret      string
}

func NewCatalogGRPCServer(
	catalogService services.CatalogService,
	jwtSecret string,
) *CatalogGRPCServer {
	return &CatalogGRPCServer{
		catalogService: catalogService,
		jwtSecret:      jwtSecret,
	}
}
