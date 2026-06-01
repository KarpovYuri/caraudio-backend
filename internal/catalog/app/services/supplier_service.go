package services

import (
	"context"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

func (s *catalogService) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	return s.suppliers.List(ctx)
}

func (s *catalogService) GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error) {
	return s.suppliers.GetByID(ctx, id)
}

func (s *catalogService) CreateSupplier(
	ctx context.Context,
	input domain.SupplierInput,
) (*domain.Supplier, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidArgument
	}

	now := time.Now()
	supplier := &domain.Supplier{
		Name:      name,
		Code:      stringPtrOrNil(strings.TrimSpace(input.Code)),
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
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidArgument
	}

	if _, err := s.suppliers.GetByID(ctx, id); err != nil {
		return nil, err
	}

	supplier := &domain.Supplier{
		ID:        id,
		Name:      name,
		Code:      stringPtrOrNil(strings.TrimSpace(input.Code)),
		IsActive:  input.IsActive,
		UpdatedAt: time.Now(),
	}

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
