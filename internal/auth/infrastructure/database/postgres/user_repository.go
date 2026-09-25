package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	ListUsers(ctx context.Context, filter domain.UserListFilter) (*domain.UserListResult, error)
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

func (r *postgresUserRepository) ListUsers(
	ctx context.Context,
	filter domain.UserListFilter,
) (*domain.UserListResult, error) {
	where := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)

	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+escapeLikePattern(search)+"%")
		where = append(where, fmt.Sprintf("login ILIKE $%d ESCAPE '\\'", len(args)))
	}
	if role := strings.TrimSpace(filter.Role); role != "" {
		args = append(args, role)
		where = append(where, fmt.Sprintf("role = $%d", len(args)))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int32
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`+whereSQL, args...); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := `SELECT id, login, password, role, created_at, updated_at FROM users` + whereSQL +
		fmt.Sprintf(" ORDER BY login ASC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	var users []domain.User
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return &domain.UserListResult{Users: users, Total: total}, nil
}

func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
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
