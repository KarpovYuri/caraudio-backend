package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListBrands(ctx context.Context, activeOnly bool) ([]domain.Brand, error) {
	return s.brands.List(ctx, activeOnly)
}

func (s *catalogService) GetBrand(ctx context.Context, id string) (*domain.Brand, error) {
	brand, err := s.brands.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !brand.IsActive {
		return nil, domain.ErrBrandNotFound
	}
	return brand, nil
}

func (s *catalogService) GetBrandByID(ctx context.Context, id string) (*domain.Brand, error) {
	return s.brands.GetByID(ctx, id)
}

func (s *catalogService) CreateBrand(ctx context.Context, input domain.BrandInput) (*domain.Brand, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	if name == "" || slug == "" {
		return nil, domain.ErrInvalidArgument
	}

	now := time.Now()
	brand := &domain.Brand{
		ID:          uuid.NewString(),
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.brands.Create(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

func (s *catalogService) UpdateBrand(ctx context.Context, id string, input domain.BrandInput) (*domain.Brand, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	if name == "" || slug == "" {
		return nil, domain.ErrInvalidArgument
	}

	brand := &domain.Brand{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		IsActive:    input.IsActive,
		UpdatedAt:   time.Now(),
	}
	if err := s.brands.Update(ctx, brand); err != nil {
		return nil, err
	}
	return s.brands.GetByID(ctx, id)
}

func (s *catalogService) DeleteBrand(ctx context.Context, id string) error {
	count, err := s.products.CountByBrand(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrBrandHasProducts
	}
	return s.brands.Delete(ctx, id)
}

func (s *catalogService) validateBrandID(ctx context.Context, brandID string) error {
	if brandID == "" {
		return nil
	}
	_, err := s.brands.GetByID(ctx, brandID)
	return err
}
