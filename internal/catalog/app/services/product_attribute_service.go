package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListProductAttributes(
	ctx context.Context,
	productID string,
	adminAccess bool,
) ([]domain.ProductAttribute, error) {
	if err := s.ensureProductVisible(ctx, productID, adminAccess); err != nil {
		return nil, err
	}
	return s.productAttributes.ListByProductID(ctx, productID)
}

func (s *catalogService) GetProductAttribute(
	ctx context.Context,
	productID, attrID string,
	adminAccess bool,
) (*domain.ProductAttribute, error) {
	if err := s.ensureProductVisible(ctx, productID, adminAccess); err != nil {
		return nil, err
	}
	attr, err := s.productAttributes.GetByID(ctx, attrID)
	if err != nil {
		return nil, err
	}
	if attr.ProductID != productID {
		return nil, domain.ErrProductAttributeNotFound
	}
	return attr, nil
}

func (s *catalogService) CreateProductAttribute(
	ctx context.Context,
	productID string,
	input domain.ProductAttributeInput,
) (*domain.ProductAttribute, error) {
	if _, err := s.products.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	value := strings.TrimSpace(input.Value)
	if name == "" || value == "" {
		return nil, domain.ErrInvalidArgument
	}
	if input.SortOrder < 0 {
		return nil, domain.ErrInvalidArgument
	}

	now := time.Now()
	attr := &domain.ProductAttribute{
		ID:        uuid.NewString(),
		ProductID: productID,
		Name:      name,
		Value:     value,
		SortOrder: input.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.productAttributes.Create(ctx, attr); err != nil {
		return nil, err
	}
	return attr, nil
}

func (s *catalogService) UpdateProductAttribute(
	ctx context.Context,
	productID, attrID string,
	input domain.ProductAttributeInput,
) (*domain.ProductAttribute, error) {
	existing, err := s.productAttributes.GetByID(ctx, attrID)
	if err != nil {
		return nil, err
	}
	if existing.ProductID != productID {
		return nil, domain.ErrProductAttributeNotFound
	}

	name := strings.TrimSpace(input.Name)
	value := strings.TrimSpace(input.Value)
	if name == "" || value == "" {
		return nil, domain.ErrInvalidArgument
	}

	attr := &domain.ProductAttribute{
		ID:        attrID,
		ProductID: productID,
		Name:      name,
		Value:     value,
		SortOrder: input.SortOrder,
		UpdatedAt: time.Now(),
	}
	if err := s.productAttributes.Update(ctx, attr); err != nil {
		return nil, err
	}
	return s.productAttributes.GetByID(ctx, attrID)
}

func (s *catalogService) DeleteProductAttribute(ctx context.Context, productID, attrID string) error {
	attr, err := s.productAttributes.GetByID(ctx, attrID)
	if err != nil {
		return err
	}
	if attr.ProductID != productID {
		return domain.ErrProductAttributeNotFound
	}
	return s.productAttributes.Delete(ctx, attrID)
}
