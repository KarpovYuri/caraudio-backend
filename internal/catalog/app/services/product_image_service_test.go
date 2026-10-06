package services

import (
	"context"
	"errors"
	"testing"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func TestListProductImagesInactiveHiddenFromPublic(t *testing.T) {
	ctx := context.Background()
	products := &fakeProductRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Product, error) {
			return &domain.Product{ID: id, IsActive: false}, nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, nil)

	_, err := svc.ListProductImages(ctx, "prod-1", false)
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound for public access, got %v", err)
	}
}

func TestListProductImagesInactiveVisibleForAdmin(t *testing.T) {
	ctx := context.Background()
	products := &fakeProductRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Product, error) {
			return &domain.Product{ID: id, IsActive: false}, nil
		},
	}
	images := &fakeProductImageRepo{
		listByProductFn: func(_ context.Context, productID string) ([]domain.ProductImage, error) {
			return []domain.ProductImage{{ID: "img-1", ProductID: productID}}, nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, images)

	got, err := svc.ListProductImages(ctx, "prod-1", true)
	if err != nil {
		t.Fatalf("expected no error for admin, got %v", err)
	}
	if len(got) != 1 || got[0].ID != "img-1" {
		t.Fatalf("unexpected images: %+v", got)
	}
}

func TestCreateProductImageRequiresURL(t *testing.T) {
	ctx := context.Background()
	products := &fakeProductRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Product, error) {
			return &domain.Product{ID: id, IsActive: true}, nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, nil)

	_, err := svc.CreateProductImage(ctx, "prod-1", domain.ProductImageInput{URL: "  "})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestCreateProductImagePrimaryUnsetsOthers(t *testing.T) {
	ctx := context.Background()
	unsetCalled := false
	products := &fakeProductRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Product, error) {
			return &domain.Product{ID: id, IsActive: true}, nil
		},
	}
	images := &fakeProductImageRepo{
		unsetPrimaryFn: func(_ context.Context, productID, exceptID string) error {
			unsetCalled = true
			if productID != "prod-1" || exceptID != "" {
				t.Fatalf("unexpected unset args: %s %q", productID, exceptID)
			}
			return nil
		},
		createFn: func(_ context.Context, image *domain.ProductImage) error {
			if !image.IsPrimary || image.URL != "https://cdn/a.webp" {
				t.Fatalf("unexpected image: %+v", image)
			}
			return nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, images)

	_, err := svc.CreateProductImage(ctx, "prod-1", domain.ProductImageInput{
		URL:       " https://cdn/a.webp ",
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !unsetCalled {
		t.Fatalf("expected UnsetPrimaryForProduct")
	}
}

func TestGetProductImageWrongProduct(t *testing.T) {
	ctx := context.Background()
	products := &fakeProductRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Product, error) {
			return &domain.Product{ID: id, IsActive: true}, nil
		},
	}
	images := &fakeProductImageRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.ProductImage, error) {
			return &domain.ProductImage{ID: id, ProductID: "other-prod"}, nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, images)

	_, err := svc.GetProductImage(ctx, "prod-1", "img-1", true)
	if !errors.Is(err, domain.ErrProductImageNotFound) {
		t.Fatalf("expected ErrProductImageNotFound, got %v", err)
	}
}
