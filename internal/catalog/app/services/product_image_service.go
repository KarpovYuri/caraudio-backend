package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListProductImages(
	ctx context.Context,
	productID string,
	adminAccess bool,
) ([]domain.ProductImage, error) {
	if err := s.ensureProductVisible(ctx, productID, adminAccess); err != nil {
		return nil, err
	}
	return s.productImages.ListByProductID(ctx, productID)
}

func (s *catalogService) GetProductImage(
	ctx context.Context,
	productID, imageID string,
	adminAccess bool,
) (*domain.ProductImage, error) {
	if err := s.ensureProductVisible(ctx, productID, adminAccess); err != nil {
		return nil, err
	}
	image, err := s.productImages.GetByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.ProductID != productID {
		return nil, domain.ErrProductImageNotFound
	}
	return image, nil
}

func (s *catalogService) CreateProductImage(
	ctx context.Context,
	productID string,
	input domain.ProductImageInput,
) (*domain.ProductImage, error) {
	if _, err := s.products.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(input.URL)
	if url == "" {
		return nil, domain.ErrInvalidArgument
	}
	if input.SortOrder < 0 {
		return nil, domain.ErrInvalidArgument
	}

	now := time.Now()
	image := &domain.ProductImage{
		ID:        uuid.NewString(),
		ProductID: productID,
		URL:       url,
		AltText:   strings.TrimSpace(input.AltText),
		SortOrder: input.SortOrder,
		IsPrimary: input.IsPrimary,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if image.IsPrimary {
		if err := s.productImages.UnsetPrimaryForProduct(ctx, productID, ""); err != nil {
			return nil, err
		}
	}

	if err := s.productImages.Create(ctx, image); err != nil {
		return nil, err
	}
	return image, nil
}

func (s *catalogService) UpdateProductImage(
	ctx context.Context,
	productID, imageID string,
	input domain.ProductImageInput,
) (*domain.ProductImage, error) {
	existing, err := s.productImages.GetByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if existing.ProductID != productID {
		return nil, domain.ErrProductImageNotFound
	}

	url := strings.TrimSpace(input.URL)
	if url == "" {
		return nil, domain.ErrInvalidArgument
	}
	if input.SortOrder < 0 {
		return nil, domain.ErrInvalidArgument
	}

	if input.IsPrimary {
		if err := s.productImages.UnsetPrimaryForProduct(ctx, productID, imageID); err != nil {
			return nil, err
		}
	}

	image := &domain.ProductImage{
		ID:        imageID,
		ProductID: productID,
		URL:       url,
		AltText:   strings.TrimSpace(input.AltText),
		SortOrder: input.SortOrder,
		IsPrimary: input.IsPrimary,
		UpdatedAt: time.Now(),
	}

	if err := s.productImages.Update(ctx, image); err != nil {
		return nil, err
	}
	return s.productImages.GetByID(ctx, imageID)
}

func (s *catalogService) DeleteProductImage(ctx context.Context, productID, imageID string) error {
	image, err := s.productImages.GetByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image.ProductID != productID {
		return domain.ErrProductImageNotFound
	}
	return s.productImages.Delete(ctx, imageID)
}

func (s *catalogService) ensureProductVisible(
	ctx context.Context,
	productID string,
	adminAccess bool,
) error {
	product, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if !adminAccess && !product.IsActive {
		return domain.ErrProductNotFound
	}
	return nil
}
