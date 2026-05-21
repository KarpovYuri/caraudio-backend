package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
}

type postgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, login, password, role, created_at, updated_at)
              VALUES (:id, :login, :password, :role, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, login, password, role, created_at, updated_at FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}
	return &user, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, login, password, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `UPDATE users
              SET login = :login, password = :password, role = :role, updated_at = :updated_at
              WHERE id = :id`
	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *postgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
