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

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	GetByID(ctx context.Context, id int64) (*domain.Supplier, error)
	List(ctx context.Context, filter domain.SupplierListFilter) (*domain.SupplierListResult, error)
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
              VALUES (:name, :code, :logo, :api_url, :is_active, :created_at, :updated_at)
              RETURNING id`
	rows, err := r.db.NamedQueryContext(ctx, query, supplier)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return fmt.Errorf("failed to create supplier: no id returned")
	}
	if err := rows.Scan(&supplier.ID); err != nil {
		return fmt.Errorf("failed to scan supplier id: %w", err)
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

func (r *postgresSupplierRepository) List(
	ctx context.Context,
	filter domain.SupplierListFilter,
) (*domain.SupplierListResult, error) {
	where := make([]string, 0, 1)
	args := make([]interface{}, 0, 3)

	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+escapeLikePattern(search)+"%")
		where = append(where, fmt.Sprintf(
			"(name ILIKE $%d ESCAPE '\\' OR COALESCE(code, '') ILIKE $%d ESCAPE '\\')",
			len(args), len(args),
		))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int32
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM suppliers`+whereSQL, args...); err != nil {
		return nil, fmt.Errorf("failed to count suppliers: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := supplierSelectSQL + whereSQL +
		fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	var suppliers []domain.Supplier
	if err := r.db.SelectContext(ctx, &suppliers, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}
	return &domain.SupplierListResult{Suppliers: suppliers, Total: total}, nil
}

func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
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
