package services

import (
	"context"
	"errors"
	"testing"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func TestCreateBrandSuccess(t *testing.T) {
	ctx := context.Background()
	var saved *domain.Brand
	brands := &fakeBrandRepo{
		createFn: func(_ context.Context, brand *domain.Brand) error {
			saved = brand
			return nil
		},
	}
	svc := newTestCatalogService(nil, nil, nil, brands, nil)

	brand, err := svc.CreateBrand(ctx, domain.BrandInput{
		Name:        "  Focal  ",
		Slug:        " focal ",
		Description: " speakers ",
		IsActive:    true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if brand.Name != "Focal" || brand.Slug != "focal" || brand.Description != "speakers" {
		t.Fatalf("unexpected brand: %+v", brand)
	}
	if saved == nil || saved.ID == "" {
		t.Fatalf("expected brand to be persisted with id")
	}
}

func TestCreateBrandInvalid(t *testing.T) {
	ctx := context.Background()
	svc := newTestCatalogService(nil, nil, nil, nil, nil)

	_, err := svc.CreateBrand(ctx, domain.BrandInput{Name: "", Slug: "x"})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	_, err = svc.CreateBrand(ctx, domain.BrandInput{Name: "X", Slug: ""})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetBrandHidesInactive(t *testing.T) {
	ctx := context.Background()
	brands := &fakeBrandRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Brand, error) {
			return &domain.Brand{ID: id, Name: "Hidden", IsActive: false}, nil
		},
	}
	svc := newTestCatalogService(nil, nil, nil, brands, nil)

	_, err := svc.GetBrand(ctx, "brand-1")
	if !errors.Is(err, domain.ErrBrandNotFound) {
		t.Fatalf("expected ErrBrandNotFound, got %v", err)
	}

	brand, err := svc.GetBrandByID(ctx, "brand-1")
	if err != nil {
		t.Fatalf("GetBrandByID should return inactive brand, got %v", err)
	}
	if brand.IsActive {
		t.Fatalf("expected inactive brand")
	}
}

func TestDeleteBrandBlockedWhenHasProducts(t *testing.T) {
	ctx := context.Background()
	products := &fakeProductRepo{
		countByBrandFn: func(_ context.Context, brandID string) (int64, error) {
			if brandID != "brand-1" {
				t.Fatalf("unexpected brand id: %s", brandID)
			}
			return 3, nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, nil, nil)

	err := svc.DeleteBrand(ctx, "brand-1")
	if !errors.Is(err, domain.ErrBrandHasProducts) {
		t.Fatalf("expected ErrBrandHasProducts, got %v", err)
	}
}

func TestDeleteBrandSuccess(t *testing.T) {
	ctx := context.Background()
	deleted := false
	products := &fakeProductRepo{
		countByBrandFn: func(_ context.Context, _ string) (int64, error) {
			return 0, nil
		},
	}
	brands := &fakeBrandRepo{
		deleteFn: func(_ context.Context, id string) error {
			deleted = id == "brand-1"
			return nil
		},
	}
	svc := newTestCatalogService(nil, nil, products, brands, nil)

	if err := svc.DeleteBrand(ctx, "brand-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !deleted {
		t.Fatalf("expected brand delete")
	}
}
