package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/jmoiron/sqlx"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	GetByID(ctx context.Context, id int64) (*domain.Supplier, error)
	List(ctx context.Context) ([]domain.Supplier, error)
	Update(ctx context.Context, supplier *domain.Supplier) error
	Delete(ctx context.Context, id int64) error
}

type postgresSupplierRepository struct {
	db *sqlx.DB
}

func NewPostgresSupplierRepository(db *sqlx.DB) SupplierRepository {
	return &postgresSupplierRepository{db: db}
}

func (r *postgresSupplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	query := `INSERT INTO suppliers (name, code, logo, api_url, is_active, created_at, updated_at) 
              VALUES (:name, :code, :logo, :api_url, :is_active, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, supplier)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	return nil
}

func (r *postgresSupplierRepository) GetByID(ctx context.Context, id int64) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.GetContext(ctx, &supplier, supplierSelectSQL+` WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	return &supplier, nil
}

func (r *postgresSupplierRepository) List(ctx context.Context) ([]domain.Supplier, error) {
	query := supplierSelectSQL + " ORDER BY name ASC"

	var suppliers []domain.Supplier
	if err := r.db.SelectContext(ctx, &suppliers, query); err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}
	return suppliers, nil
}

func (r *postgresSupplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	query := `UPDATE suppliers SET 
                name = :name, code = :code, logo = :logo, api_url = :api_url, 
                is_active = :is_active, updated_at = :updated_at 
              WHERE id = :id`
	result, err := r.db.NamedExecContext(ctx, query, supplier)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to update supplier: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}

func (r *postgresSupplierRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM suppliers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}

const supplierSelectSQL = `SELECT id, name, code, logo, api_url, is_active, created_at, updated_at FROM suppliers`
