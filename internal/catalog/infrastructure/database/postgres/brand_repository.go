package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type BrandRepository interface {
	Create(ctx context.Context, brand *domain.Brand) error
	GetByID(ctx context.Context, id string) (*domain.Brand, error)
	List(ctx context.Context, activeOnly bool) ([]domain.Brand, error)
	Update(ctx context.Context, brand *domain.Brand) error
	Delete(ctx context.Context, id string) error
}

type postgresBrandRepository struct {
	db *sqlx.DB
}

func NewPostgresBrandRepository(db *sqlx.DB) BrandRepository {
	return &postgresBrandRepository{db: db}
}

func (r *postgresBrandRepository) Create(ctx context.Context, brand *domain.Brand) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO brands (id, name, slug, description, is_active, created_at, updated_at)
         VALUES (:id, :name, :slug, :description, :is_active, :created_at, :updated_at)`, brand)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create brand: %w", err)
	}
	return nil
}

func (r *postgresBrandRepository) GetByID(ctx context.Context, id string) (*domain.Brand, error) {
	var brand domain.Brand
	err := r.db.GetContext(ctx, &brand, brandSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBrandNotFound
		}
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return &brand, nil
}

func (r *postgresBrandRepository) List(ctx context.Context, activeOnly bool) ([]domain.Brand, error) {
	query := brandSelectSQL
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY name ASC`

	var brands []domain.Brand
	if err := r.db.SelectContext(ctx, &brands, query); err != nil {
		return nil, fmt.Errorf("failed to list brands: %w", err)
	}
	return brands, nil
}

func (r *postgresBrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	result, err := r.db.NamedExecContext(ctx,
		`UPDATE brands SET name = :name, slug = :slug, description = :description,
         is_active = :is_active, updated_at = :updated_at WHERE id = :id`, brand)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update brand: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrBrandNotFound
	}
	return nil
}

func (r *postgresBrandRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM brands WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrBrandNotFound
	}
	return nil
}

const brandSelectSQL = `SELECT id, name, slug, description, is_active, created_at, updated_at FROM brands`
