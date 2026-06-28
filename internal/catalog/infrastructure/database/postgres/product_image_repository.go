package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type ProductImageRepository interface {
	Create(ctx context.Context, image *domain.ProductImage) error
	GetByID(ctx context.Context, id string) (*domain.ProductImage, error)
	ListByProductID(ctx context.Context, productID string) ([]domain.ProductImage, error)
	Update(ctx context.Context, image *domain.ProductImage) error
	Delete(ctx context.Context, id string) error
	UnsetPrimaryForProduct(ctx context.Context, productID string, exceptID string) error
}

type postgresProductImageRepository struct {
	db *sqlx.DB
}

func NewPostgresProductImageRepository(db *sqlx.DB) ProductImageRepository {
	return &postgresProductImageRepository{db: db}
}

func (r *postgresProductImageRepository) Create(ctx context.Context, image *domain.ProductImage) error {
	query := `INSERT INTO product_images (
	            id, product_id, url, alt_text, sort_order, is_primary, created_at, updated_at
	          ) VALUES (
	            :id, :product_id, :url, :alt_text, :sort_order, :is_primary, :created_at, :updated_at
	          )`
	_, err := r.db.NamedExecContext(ctx, query, image)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create product image: %w", err)
	}
	return nil
}

func (r *postgresProductImageRepository) GetByID(ctx context.Context, id string) (*domain.ProductImage, error) {
	var image domain.ProductImage
	err := r.db.GetContext(ctx, &image, productImageSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductImageNotFound
		}
		return nil, fmt.Errorf("failed to get product image: %w", err)
	}
	return &image, nil
}

func (r *postgresProductImageRepository) ListByProductID(
	ctx context.Context,
	productID string,
) ([]domain.ProductImage, error) {
	var images []domain.ProductImage
	err := r.db.SelectContext(ctx, &images,
		productImageSelectSQL+` WHERE product_id = $1 ORDER BY sort_order ASC, created_at ASC`, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product images: %w", err)
	}
	return images, nil
}

func (r *postgresProductImageRepository) Update(ctx context.Context, image *domain.ProductImage) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE product_images SET
           url = :url,
           alt_text = :alt_text,
           sort_order = :sort_order,
           is_primary = :is_primary,
           updated_at = :updated_at
         WHERE id = :id`, image)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update product image: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductImageNotFound
	}
	return nil
}

func (r *postgresProductImageRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM product_images WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product image: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductImageNotFound
	}
	return nil
}

func (r *postgresProductImageRepository) UnsetPrimaryForProduct(
	ctx context.Context,
	productID string,
	exceptID string,
) error {
	if exceptID == "" {
		_, err := r.db.ExecContext(ctx,
			`UPDATE product_images SET is_primary = FALSE, updated_at = NOW() WHERE product_id = $1 AND is_primary = TRUE`,
			productID)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE product_images SET is_primary = FALSE, updated_at = NOW()
         WHERE product_id = $1 AND is_primary = TRUE AND id <> $2`,
		productID, exceptID)
	return err
}

const productImageSelectSQL = `SELECT id, product_id, url, alt_text, sort_order, is_primary, created_at, updated_at FROM product_images`
