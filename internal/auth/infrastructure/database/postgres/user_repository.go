package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
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
		if err.Error() == "pq: duplicate key value violates unique constraint \"idx_users_login\"" ||
			err.Error() == "pq: duplicate key value violates unique constraint \"users_login_key\"" {
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
