package services

import (
	"context"
	"errors"
	"testing"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func TestNormalizePagination(t *testing.T) {
	svc := newTestCatalogService(nil, nil, nil, nil, nil)

	page, size := svc.NormalizePagination(0, 0, 10, 100)
	if page != 1 || size != 10 {
		t.Fatalf("expected defaults, got page=%d size=%d", page, size)
	}

	page, size = svc.NormalizePagination(2, 500, 10, 100)
	if page != 2 || size != 100 {
		t.Fatalf("expected clamped size, got page=%d size=%d", page, size)
	}
}

func TestCreateCategorySuccess(t *testing.T) {
	ctx := context.Background()
	var saved *domain.Category
	categories := &fakeCategoryRepo{
		createFn: func(_ context.Context, category *domain.Category) error {
			saved = category
			return nil
		},
	}
	svc := newTestCatalogService(nil, categories, nil, nil, nil)

	category, err := svc.CreateCategory(ctx, "  Speakers  ", " speakers ", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if category.Name != "Speakers" || category.Slug != "speakers" {
		t.Fatalf("unexpected category: %+v", category)
	}
	if saved == nil || saved.ID == "" || saved.ParentID != nil {
		t.Fatalf("unexpected saved category: %+v", saved)
	}
}

func TestCreateCategoryInvalidParent(t *testing.T) {
	ctx := context.Background()
	categories := &fakeCategoryRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Category, error) {
			return nil, domain.ErrCategoryNotFound
		},
	}
	svc := newTestCatalogService(nil, categories, nil, nil, nil)

	_, err := svc.CreateCategory(ctx, "Child", "child", "missing-parent")
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestDeleteCategoryBlockedWhenHasProducts(t *testing.T) {
	ctx := context.Background()
	categories := &fakeCategoryRepo{
		countProductsFn: func(_ context.Context, id string) (int64, error) {
			if id != "cat-1" {
				t.Fatalf("unexpected id: %s", id)
			}
			return 2, nil
		},
	}
	svc := newTestCatalogService(nil, categories, nil, nil, nil)

	err := svc.DeleteCategory(ctx, "cat-1")
	if !errors.Is(err, domain.ErrCategoryHasProducts) {
		t.Fatalf("expected ErrCategoryHasProducts, got %v", err)
	}
}

func TestUpdateCategoryRejectsSelfParent(t *testing.T) {
	ctx := context.Background()
	categories := &fakeCategoryRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Category, error) {
			return &domain.Category{ID: id, Name: "Root", Slug: "root"}, nil
		},
	}
	svc := newTestCatalogService(nil, categories, nil, nil, nil)

	_, err := svc.UpdateCategory(ctx, "cat-1", "", "", "cat-1")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
