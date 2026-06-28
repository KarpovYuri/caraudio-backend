package services

import (
	"context"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/database/postgres"
	"github.com/google/uuid"
)

type CatalogService interface {
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error)
	CreateSupplier(ctx context.Context, input domain.SupplierInput) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, id int64, input domain.SupplierInput) (*domain.Supplier, error)
	DeleteSupplier(ctx context.Context, id int64) error

	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id string) (*domain.Category, error)
	CreateCategory(ctx context.Context, name, slug, parentID string) (*domain.Category, error)
	UpdateCategory(ctx context.Context, id, name, slug, parentID string) (*domain.Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

type catalogService struct {
	suppliers  postgres.SupplierRepository
	categories postgres.CategoryRepository
}

func NewCatalogService(
	suppliers postgres.SupplierRepository,
	categories postgres.CategoryRepository,
) CatalogService {
	return &catalogService{
		suppliers:  suppliers,
		categories: categories,
	}
}

func (s *catalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

func (s *catalogService) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.categories.GetByID(ctx, id)
}

func (s *catalogService) CreateCategory(
	ctx context.Context,
	name, slug, parentID string,
) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" || slug == "" {
		return nil, domain.ErrInvalidArgument
	}

	if parentID != "" {
		if _, err := s.categories.GetByID(ctx, parentID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	category := &domain.Category{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		ParentID:  stringPtrOrNil(parentID),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *catalogService) UpdateCategory(
	ctx context.Context,
	id, name, slug, parentID string,
) (*domain.Category, error) {
	category, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" && slug == "" && parentID == "" {
		return nil, domain.ErrInvalidArgument
	}

	if name != "" {
		category.Name = strings.TrimSpace(name)
	}
	if slug != "" {
		category.Slug = strings.TrimSpace(slug)
	}
	if parentID != "" {
		if parentID == id {
			return nil, domain.ErrInvalidArgument
		}
		if _, err := s.categories.GetByID(ctx, parentID); err != nil {
			return nil, err
		}
		category.ParentID = stringPtrOrNil(parentID)
	}

	category.UpdatedAt = time.Now()
	if err := s.categories.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *catalogService) DeleteCategory(ctx context.Context, id string) error {
	count, err := s.categories.CountProducts(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrCategoryHasProducts
	}
	return s.categories.Delete(ctx, id)
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
