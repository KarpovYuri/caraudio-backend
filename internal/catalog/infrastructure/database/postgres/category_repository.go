package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
	CountProducts(ctx context.Context, categoryID string) (int64, error)
}

type postgresCategoryRepository struct {
	db *sqlx.DB
}

func NewPostgresCategoryRepository(db *sqlx.DB) CategoryRepository {
	return &postgresCategoryRepository{db: db}
}

func (r *postgresCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `INSERT INTO categories (id, name, slug, parent_id, created_at, updated_at)
              VALUES (:id, :name, :slug, :parent_id, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *postgresCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.GetContext(ctx, &category,
		`SELECT id, name, slug, parent_id, created_at, updated_at FROM categories WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

func (r *postgresCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.SelectContext(ctx, &categories,
		`SELECT id, name, slug, parent_id, created_at, updated_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

func (r *postgresCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE categories
         SET name = :name, slug = :slug, parent_id = :parent_id, updated_at = :updated_at
         WHERE id = :id`, category)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *postgresCategoryRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *postgresCategoryRepository) CountProducts(ctx context.Context, categoryID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM products WHERE category_id = $1`, categoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to count products in category: %w", err)
	}
	return count, nil
}
