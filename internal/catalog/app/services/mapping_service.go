package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListSupplierCategoryMappings(
	ctx context.Context,
	filter domain.SupplierCategoryMappingFilter,
) ([]domain.SupplierCategoryMapping, error) {
	return s.categoryMappings.List(ctx, filter)
}

func (s *catalogService) GetSupplierCategoryMapping(
	ctx context.Context,
	id string,
) (*domain.SupplierCategoryMapping, error) {
	return s.categoryMappings.GetByID(ctx, id)
}

func (s *catalogService) CreateSupplierCategoryMapping(
	ctx context.Context,
	input domain.SupplierCategoryMappingInput,
) (*domain.SupplierCategoryMapping, error) {
	if err := s.validateCategoryMappingInput(ctx, input); err != nil {
		return nil, err
	}

	now := time.Now()
	m := &domain.SupplierCategoryMapping{
		ID:           uuid.NewString(),
		CategoryID:   input.CategoryID,
		SupplierID:   input.SupplierID,
		ExternalID:   strings.TrimSpace(input.ExternalID),
		ExternalName: strings.TrimSpace(input.ExternalName),
		Notes:        strings.TrimSpace(input.Notes),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.categoryMappings.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *catalogService) UpdateSupplierCategoryMapping(
	ctx context.Context,
	id string,
	input domain.SupplierCategoryMappingInput,
) (*domain.SupplierCategoryMapping, error) {
	if _, err := s.categoryMappings.GetByID(ctx, id); err != nil {
		return nil, err
	}
	if err := s.validateCategoryMappingInput(ctx, input); err != nil {
		return nil, err
	}

	m := &domain.SupplierCategoryMapping{
		ID:           id,
		CategoryID:   input.CategoryID,
		SupplierID:   input.SupplierID,
		ExternalID:   strings.TrimSpace(input.ExternalID),
		ExternalName: strings.TrimSpace(input.ExternalName),
		Notes:        strings.TrimSpace(input.Notes),
		UpdatedAt:    time.Now(),
	}
	if err := s.categoryMappings.Update(ctx, m); err != nil {
		return nil, err
	}
	return s.categoryMappings.GetByID(ctx, id)
}

func (s *catalogService) DeleteSupplierCategoryMapping(ctx context.Context, id string) error {
	return s.categoryMappings.Delete(ctx, id)
}

func (s *catalogService) ListSupplierProductMappings(
	ctx context.Context,
	filter domain.SupplierProductMappingFilter,
) ([]domain.SupplierProductMapping, error) {
	return s.productMappings.List(ctx, filter)
}

func (s *catalogService) GetSupplierProductMapping(
	ctx context.Context,
	id string,
) (*domain.SupplierProductMapping, error) {
	return s.productMappings.GetByID(ctx, id)
}

func (s *catalogService) CreateSupplierProductMapping(
	ctx context.Context,
	input domain.SupplierProductMappingInput,
) (*domain.SupplierProductMapping, error) {
	if err := s.validateProductMappingInput(ctx, input); err != nil {
		return nil, err
	}

	now := time.Now()
	m := &domain.SupplierProductMapping{
		ID:           uuid.NewString(),
		ProductID:    input.ProductID,
		SupplierID:   input.SupplierID,
		ExternalID:   strings.TrimSpace(input.ExternalID),
		ExternalSKU:  strings.TrimSpace(input.ExternalSKU),
		ExternalName: strings.TrimSpace(input.ExternalName),
		Notes:        strings.TrimSpace(input.Notes),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.productMappings.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *catalogService) UpdateSupplierProductMapping(
	ctx context.Context,
	id string,
	input domain.SupplierProductMappingInput,
) (*domain.SupplierProductMapping, error) {
	if _, err := s.productMappings.GetByID(ctx, id); err != nil {
		return nil, err
	}
	if err := s.validateProductMappingInput(ctx, input); err != nil {
		return nil, err
	}

	m := &domain.SupplierProductMapping{
		ID:           id,
		ProductID:    input.ProductID,
		SupplierID:   input.SupplierID,
		ExternalID:   strings.TrimSpace(input.ExternalID),
		ExternalSKU:  strings.TrimSpace(input.ExternalSKU),
		ExternalName: strings.TrimSpace(input.ExternalName),
		Notes:        strings.TrimSpace(input.Notes),
		UpdatedAt:    time.Now(),
	}
	if err := s.productMappings.Update(ctx, m); err != nil {
		return nil, err
	}
	return s.productMappings.GetByID(ctx, id)
}

func (s *catalogService) DeleteSupplierProductMapping(ctx context.Context, id string) error {
	return s.productMappings.Delete(ctx, id)
}

func (s *catalogService) validateCategoryMappingInput(
	ctx context.Context,
	input domain.SupplierCategoryMappingInput,
) error {
	if input.CategoryID == "" || input.SupplierID == 0 || strings.TrimSpace(input.ExternalID) == "" {
		return domain.ErrInvalidArgument
	}
	if _, err := s.categories.GetByID(ctx, input.CategoryID); err != nil {
		return err
	}
	if _, err := s.suppliers.GetByID(ctx, input.SupplierID); err != nil {
		return err
	}
	return nil
}

func (s *catalogService) validateProductMappingInput(
	ctx context.Context,
	input domain.SupplierProductMappingInput,
) error {
	if input.ProductID == "" || input.SupplierID == 0 || strings.TrimSpace(input.ExternalID) == "" {
		return domain.ErrInvalidArgument
	}
	if _, err := s.products.GetByID(ctx, input.ProductID); err != nil {
		return err
	}
	if _, err := s.suppliers.GetByID(ctx, input.SupplierID); err != nil {
		return err
	}
	return nil
}
