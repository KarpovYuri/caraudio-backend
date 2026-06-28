package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type SupplierCategoryMappingRepository interface {
	Create(ctx context.Context, m *domain.SupplierCategoryMapping) error
	GetByID(ctx context.Context, id string) (*domain.SupplierCategoryMapping, error)
	List(ctx context.Context, filter domain.SupplierCategoryMappingFilter) ([]domain.SupplierCategoryMapping, error)
	Update(ctx context.Context, m *domain.SupplierCategoryMapping) error
	Delete(ctx context.Context, id string) error
}

type SupplierProductMappingRepository interface {
	Create(ctx context.Context, m *domain.SupplierProductMapping) error
	GetByID(ctx context.Context, id string) (*domain.SupplierProductMapping, error)
	List(ctx context.Context, filter domain.SupplierProductMappingFilter) ([]domain.SupplierProductMapping, error)
	Update(ctx context.Context, m *domain.SupplierProductMapping) error
	Delete(ctx context.Context, id string) error
}

type postgresSupplierCategoryMappingRepository struct {
	db *sqlx.DB
}

func NewPostgresSupplierCategoryMappingRepository(db *sqlx.DB) SupplierCategoryMappingRepository {
	return &postgresSupplierCategoryMappingRepository{db: db}
}

func (r *postgresSupplierCategoryMappingRepository) Create(
	ctx context.Context,
	m *domain.SupplierCategoryMapping,
) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO supplier_category_mappings (
           id, category_id, supplier_id, external_id, external_name, notes, created_at, updated_at
         ) VALUES (
           :id, :category_id, :supplier_id, :external_id, :external_name, :notes, :created_at, :updated_at
         )`, m)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create supplier category mapping: %w", err)
	}
	return nil
}

func (r *postgresSupplierCategoryMappingRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.SupplierCategoryMapping, error) {
	var m domain.SupplierCategoryMapping
	err := r.db.GetContext(ctx, &m, supplierCategoryMappingSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSupplierMappingNotFound
		}
		return nil, fmt.Errorf("failed to get supplier category mapping: %w", err)
	}
	return &m, nil
}

func (r *postgresSupplierCategoryMappingRepository) List(
	ctx context.Context,
	filter domain.SupplierCategoryMappingFilter,
) ([]domain.SupplierCategoryMapping, error) {
	where, args := buildMappingWhere(filter.SupplierID, filter.CategoryID, "supplier_id", "category_id")
	query := supplierCategoryMappingSelectSQL + where + ` ORDER BY created_at DESC`

	var list []domain.SupplierCategoryMapping
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list supplier category mappings: %w", err)
	}
	return list, nil
}

func (r *postgresSupplierCategoryMappingRepository) Update(
	ctx context.Context,
	m *domain.SupplierCategoryMapping,
) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE supplier_category_mappings SET
           category_id = :category_id, supplier_id = :supplier_id,
           external_id = :external_id, external_name = :external_name,
           notes = :notes, updated_at = :updated_at
         WHERE id = :id`, m)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update supplier category mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSupplierMappingNotFound
	}
	return nil
}

func (r *postgresSupplierCategoryMappingRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM supplier_category_mappings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier category mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSupplierMappingNotFound
	}
	return nil
}

type postgresSupplierProductMappingRepository struct {
	db *sqlx.DB
}

func NewPostgresSupplierProductMappingRepository(db *sqlx.DB) SupplierProductMappingRepository {
	return &postgresSupplierProductMappingRepository{db: db}
}

func (r *postgresSupplierProductMappingRepository) Create(
	ctx context.Context,
	m *domain.SupplierProductMapping,
) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO supplier_product_mappings (
           id, product_id, supplier_id, external_id, external_sku, external_name, notes, created_at, updated_at
         ) VALUES (
           :id, :product_id, :supplier_id, :external_id, :external_sku, :external_name, :notes, :created_at, :updated_at
         )`, m)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create supplier product mapping: %w", err)
	}
	return nil
}

func (r *postgresSupplierProductMappingRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.SupplierProductMapping, error) {
	var m domain.SupplierProductMapping
	err := r.db.GetContext(ctx, &m, supplierProductMappingSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSupplierMappingNotFound
		}
		return nil, fmt.Errorf("failed to get supplier product mapping: %w", err)
	}
	return &m, nil
}

func (r *postgresSupplierProductMappingRepository) List(
	ctx context.Context,
	filter domain.SupplierProductMappingFilter,
) ([]domain.SupplierProductMapping, error) {
	where, args := buildMappingWhere(filter.SupplierID, filter.ProductID, "supplier_id", "product_id")
	query := supplierProductMappingSelectSQL + where + ` ORDER BY created_at DESC`

	var list []domain.SupplierProductMapping
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list supplier product mappings: %w", err)
	}
	return list, nil
}

func (r *postgresSupplierProductMappingRepository) Update(
	ctx context.Context,
	m *domain.SupplierProductMapping,
) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE supplier_product_mappings SET
           product_id = :product_id, supplier_id = :supplier_id,
           external_id = :external_id, external_sku = :external_sku, external_name = :external_name,
           notes = :notes, updated_at = :updated_at
         WHERE id = :id`, m)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update supplier product mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSupplierMappingNotFound
	}
	return nil
}

func (r *postgresSupplierProductMappingRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM supplier_product_mappings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier product mapping: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSupplierMappingNotFound
	}
	return nil
}

func buildMappingWhere(supplierID int64, entityID, supplierCol, entityCol string) (string, []interface{}) {
	where := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if supplierID != 0 {
		args = append(args, supplierID)
		where = append(where, fmt.Sprintf("%s = $%d", supplierCol, len(args)))
	}
	if entityID != "" {
		args = append(args, entityID)
		where = append(where, fmt.Sprintf("%s = $%d", entityCol, len(args)))
	}
	if len(where) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(where, " AND "), args
}

const supplierCategoryMappingSelectSQL = `SELECT id, category_id, supplier_id, external_id, external_name, notes, created_at, updated_at FROM supplier_category_mappings`

const supplierProductMappingSelectSQL = `SELECT id, product_id, supplier_id, external_id, external_sku, external_name, notes, created_at, updated_at FROM supplier_product_mappings`
