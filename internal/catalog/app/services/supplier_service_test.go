package services

import (
	"context"
	"errors"
	"testing"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func TestCreateSupplierSuccess(t *testing.T) {
	ctx := context.Background()
	var saved *domain.Supplier

	suppliers := &fakeSupplierRepo{
		createFn: func(_ context.Context, supplier *domain.Supplier) error {
			supplier.ID = 42
			saved = supplier
			return nil
		},
	}
	svc := newTestCatalogService(suppliers, nil, nil, nil, nil)

	supplier, err := svc.CreateSupplier(ctx, domain.SupplierInput{
		Name:     "  Acme  ",
		Code:     " ACME ",
		ApiUrl:   "https://api.example.com/v1",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if supplier.ID != 42 || supplier.Name != "Acme" {
		t.Fatalf("unexpected supplier: %+v", supplier)
	}
	if saved == nil || saved.ApiUrl != "https://api.example.com/v1" {
		t.Fatalf("unexpected saved supplier: %+v", saved)
	}
	if saved.Code == nil || *saved.Code != "ACME" {
		t.Fatalf("expected trimmed code, got %#v", saved.Code)
	}
}

func TestCreateSupplierInvalidFields(t *testing.T) {
	ctx := context.Background()
	svc := newTestCatalogService(nil, nil, nil, nil, nil)

	cases := []domain.SupplierInput{
		{Name: "", ApiUrl: "https://example.com"},
		{Name: "Acme", ApiUrl: ""},
		{Name: "Acme", ApiUrl: "not-a-url"},
		{Name: "Acme", ApiUrl: "ftp://example.com"},
	}
	for _, input := range cases {
		_, err := svc.CreateSupplier(ctx, input)
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("input=%+v: expected ErrInvalidArgument, got %v", input, err)
		}
	}
}

func TestListSuppliersTrimsSearch(t *testing.T) {
	ctx := context.Background()
	var got domain.SupplierListFilter
	suppliers := &fakeSupplierRepo{
		listFn: func(_ context.Context, filter domain.SupplierListFilter) (*domain.SupplierListResult, error) {
			got = filter
			return &domain.SupplierListResult{}, nil
		},
	}
	svc := newTestCatalogService(suppliers, nil, nil, nil, nil)

	_, err := svc.ListSuppliers(ctx, domain.SupplierListFilter{Search: "  acme  "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Search != "acme" {
		t.Fatalf("expected trimmed search, got %q", got.Search)
	}
}

func TestUpdateSupplierLogoSuccess(t *testing.T) {
	ctx := context.Background()
	existing := &domain.Supplier{ID: 7, Name: "Acme", Logo: "", ApiUrl: "https://example.com"}
	var updated *domain.Supplier

	suppliers := &fakeSupplierRepo{
		getByIDFn: func(_ context.Context, id int64) (*domain.Supplier, error) {
			if id != 7 {
				t.Fatalf("unexpected id: %d", id)
			}
			copy := *existing
			if updated != nil {
				copy = *updated
			}
			return &copy, nil
		},
		updateFn: func(_ context.Context, supplier *domain.Supplier) error {
			updated = supplier
			return nil
		},
	}
	svc := newTestCatalogService(suppliers, nil, nil, nil, nil)

	supplier, err := svc.UpdateSupplierLogo(ctx, 7, "  /static/suppliers/a.webp  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if supplier.Logo != "/static/suppliers/a.webp" {
		t.Fatalf("unexpected logo: %q", supplier.Logo)
	}
}

func TestUpdateSupplierLogoInvalid(t *testing.T) {
	ctx := context.Background()
	svc := newTestCatalogService(nil, nil, nil, nil, nil)

	_, err := svc.UpdateSupplierLogo(ctx, 1, "   ")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUpdateSupplierNotFound(t *testing.T) {
	ctx := context.Background()
	suppliers := &fakeSupplierRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Supplier, error) {
			return nil, domain.ErrSupplierNotFound
		},
	}
	svc := newTestCatalogService(suppliers, nil, nil, nil, nil)

	_, err := svc.UpdateSupplier(ctx, 99, domain.SupplierInput{
		Name:   "Acme",
		ApiUrl: "https://example.com",
	})
	if !errors.Is(err, domain.ErrSupplierNotFound) {
		t.Fatalf("expected ErrSupplierNotFound, got %v", err)
	}
}

func TestDeleteSupplier(t *testing.T) {
	ctx := context.Background()
	deleted := false
	suppliers := &fakeSupplierRepo{
		deleteFn: func(_ context.Context, id int64) error {
			if id != 5 {
				t.Fatalf("unexpected id: %d", id)
			}
			deleted = true
			return nil
		},
	}
	svc := newTestCatalogService(suppliers, nil, nil, nil, nil)

	if err := svc.DeleteSupplier(ctx, 5); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !deleted {
		t.Fatalf("expected delete to be called")
	}
}
