package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type ProductAttributeRepository interface {
	Create(ctx context.Context, attr *domain.ProductAttribute) error
	GetByID(ctx context.Context, id string) (*domain.ProductAttribute, error)
	ListByProductID(ctx context.Context, productID string) ([]domain.ProductAttribute, error)
	Update(ctx context.Context, attr *domain.ProductAttribute) error
	Delete(ctx context.Context, id string) error
}

type postgresProductAttributeRepository struct {
	db *sqlx.DB
}

func NewPostgresProductAttributeRepository(db *sqlx.DB) ProductAttributeRepository {
	return &postgresProductAttributeRepository{db: db}
}

func (r *postgresProductAttributeRepository) Create(ctx context.Context, attr *domain.ProductAttribute) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO product_attributes (id, product_id, name, value, sort_order, created_at, updated_at)
         VALUES (:id, :product_id, :name, :value, :sort_order, :created_at, :updated_at)`, attr)
	if err != nil {
		return fmt.Errorf("failed to create product attribute: %w", err)
	}
	return nil
}

func (r *postgresProductAttributeRepository) GetByID(ctx context.Context, id string) (*domain.ProductAttribute, error) {
	var attr domain.ProductAttribute
	err := r.db.GetContext(ctx, &attr, productAttributeSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductAttributeNotFound
		}
		return nil, fmt.Errorf("failed to get product attribute: %w", err)
	}
	return &attr, nil
}

func (r *postgresProductAttributeRepository) ListByProductID(
	ctx context.Context,
	productID string,
) ([]domain.ProductAttribute, error) {
	var attrs []domain.ProductAttribute
	err := r.db.SelectContext(ctx, &attrs,
		productAttributeSelectSQL+` WHERE product_id = $1 ORDER BY sort_order ASC, name ASC`, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product attributes: %w", err)
	}
	return attrs, nil
}

func (r *postgresProductAttributeRepository) Update(ctx context.Context, attr *domain.ProductAttribute) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE product_attributes SET name = :name, value = :value, sort_order = :sort_order,
         updated_at = :updated_at WHERE id = :id`, attr)
	if err != nil {
		return fmt.Errorf("failed to update product attribute: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrProductAttributeNotFound
	}
	return nil
}

func (r *postgresProductAttributeRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM product_attributes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product attribute: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrProductAttributeNotFound
	}
	return nil
}

const productAttributeSelectSQL = `SELECT id, product_id, name, value, sort_order, created_at, updated_at FROM product_attributes`
