package services

import (
	"context"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/database/postgres"
)

type CatalogService interface {
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error)
	CreateSupplier(ctx context.Context, input domain.SupplierInput) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, id int64, input domain.SupplierInput) (*domain.Supplier, error)
	DeleteSupplier(ctx context.Context, id int64) error
}

type catalogService struct {
	suppliers postgres.SupplierRepository
}

func NewCatalogService(
	suppliers postgres.SupplierRepository,
) CatalogService {
	return &catalogService{
		suppliers: suppliers,
	}
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
