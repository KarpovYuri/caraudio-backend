package services

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListSuppliers(
	ctx context.Context,
	filter domain.SupplierListFilter,
) (*domain.SupplierListResult, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	return s.suppliers.List(ctx, filter)
}

func (s *catalogService) GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error) {
	return s.suppliers.GetByID(ctx, id)
}

func (s *catalogService) CreateSupplier(
	ctx context.Context,
	input domain.SupplierInput,
) (*domain.Supplier, error) {
	name, apiURL, err := normalizeSupplierFields(input.Name, input.ApiUrl)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	supplier := &domain.Supplier{
		Name:      name,
		Code:      stringPtrOrNil(strings.TrimSpace(input.Code)),
		Logo:      strings.TrimSpace(input.Logo),
		ApiUrl:    apiURL,
		IsActive:  input.IsActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.suppliers.Create(ctx, supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (s *catalogService) UpdateSupplier(
	ctx context.Context,
	id int64,
	input domain.SupplierInput,
) (*domain.Supplier, error) {
	name, apiURL, err := normalizeSupplierFields(input.Name, input.ApiUrl)
	if err != nil {
		return nil, err
	}

	if _, err := s.suppliers.GetByID(ctx, id); err != nil {
		return nil, err
	}

	supplier := &domain.Supplier{
		ID:        id,
		Name:      name,
		Code:      stringPtrOrNil(strings.TrimSpace(input.Code)),
		Logo:      strings.TrimSpace(input.Logo),
		ApiUrl:    apiURL,
		IsActive:  input.IsActive,
		UpdatedAt: time.Now(),
	}

	if err := s.suppliers.Update(ctx, supplier); err != nil {
		return nil, err
	}
	return s.suppliers.GetByID(ctx, id)
}

func (s *catalogService) UpdateSupplierLogo(
	ctx context.Context,
	id int64,
	logoURL string,
) (*domain.Supplier, error) {
	logoURL = strings.TrimSpace(logoURL)
	if logoURL == "" {
		return nil, domain.ErrInvalidArgument
	}

	supplier, err := s.suppliers.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	supplier.Logo = logoURL
	supplier.UpdatedAt = time.Now()
	if err := s.suppliers.Update(ctx, supplier); err != nil {
		return nil, err
	}
	return s.suppliers.GetByID(ctx, id)
}

func (s *catalogService) DeleteSupplier(ctx context.Context, id int64) error {
	/* // Проверка на наличие продуктов у поставщика временно отключена
	   count, err := s.products.CountBySupplier(ctx, id)
	   if err != nil {
	      return err
	   }
	   if count > 0 {
	      return domain.ErrSupplierHasProducts
	   }
	*/

	return s.suppliers.Delete(ctx, id)
}

func (s *catalogService) validateSupplierID(ctx context.Context, supplierID int64) error {
	if supplierID == 0 {
		return nil
	}
	_, err := s.suppliers.GetByID(ctx, supplierID)
	return err
}

func normalizeSupplierFields(name, apiURL string) (string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", domain.ErrInvalidArgument
	}

	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		return "", "", domain.ErrInvalidArgument
	}

	parsed, err := url.ParseRequestURI(apiURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", "", domain.ErrInvalidArgument
	}

	return name, apiURL, nil
}
