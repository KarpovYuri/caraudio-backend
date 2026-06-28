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

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	CountBySupplier(ctx context.Context, supplierID int64) (int64, error)
	CountByBrand(ctx context.Context, brandID string) (int64, error)
}

type postgresProductRepository struct {
	db *sqlx.DB
}

func NewPostgresProductRepository(db *sqlx.DB) ProductRepository {
	return &postgresProductRepository{db: db}
}

func (r *postgresProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `INSERT INTO products (
	            id, category_id, brand_id, supplier_id, name, description, price_cents, sku, stock, is_active, created_at, updated_at
	          ) VALUES (
	            :id, :category_id, :brand_id, :supplier_id, :name, :description, :price_cents, :sku, :stock, :is_active, :created_at, :updated_at
	          )`
	_, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *postgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, productSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

func (r *postgresProductRepository) List(
	ctx context.Context,
	filter domain.ProductListFilter,
) (*domain.ProductListResult, error) {
	where := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)

	if filter.CategoryID != "" {
		args = append(args, filter.CategoryID)
		where = append(where, fmt.Sprintf("category_id = $%d", len(args)))
	}
	if filter.BrandID != "" {
		args = append(args, filter.BrandID)
		where = append(where, fmt.Sprintf("brand_id = $%d", len(args)))
	}
	if filter.SupplierID != 0 {
		args = append(args, filter.SupplierID)
		where = append(where, fmt.Sprintf("supplier_id = $%d", len(args)))
	}
	if filter.ActiveOnly {
		where = append(where, "is_active = TRUE")
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM products` + whereSQL
	var total int32
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	listQuery := productSelectSQL + whereSQL +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	var products []domain.Product
	if err := r.db.SelectContext(ctx, &products, listQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return &domain.ProductListResult{Products: products, Total: total}, nil
}

func (r *postgresProductRepository) Update(ctx context.Context, product *domain.Product) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE products SET
           category_id = :category_id,
           brand_id = :brand_id,
           supplier_id = :supplier_id,
           name = :name,
           description = :description,
           price_cents = :price_cents,
           sku = :sku,
           stock = :stock,
           is_active = :is_active,
           updated_at = :updated_at
         WHERE id = :id`, product)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update product: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *postgresProductRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *postgresProductRepository) CountByBrand(ctx context.Context, brandID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM products WHERE brand_id = $1`, brandID)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by brand: %w", err)
	}
	return count, nil
}

func (r *postgresProductRepository) CountBySupplier(ctx context.Context, supplierID int64) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM products WHERE supplier_id = $1`, supplierID)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by supplier: %w", err)
	}
	return count, nil
}

const productSelectSQL = `SELECT id, category_id, brand_id, supplier_id, name, description, price_cents, sku, stock, is_active, created_at, updated_at FROM products`
