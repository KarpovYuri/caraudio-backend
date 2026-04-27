package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	ReplaceForUser(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	DeleteByHash(ctx context.Context, hash string) error
	DeleteByUserId(ctx context.Context, userId string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

type PgRefreshTokenRepository struct {
	db *sqlx.DB
}

func NewPgRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &PgRefreshTokenRepository{db: db}
}

func (r *PgRefreshTokenRepository) Create(
	ctx context.Context,
	token *domain.RefreshToken,
) error {

	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

func (r *PgRefreshTokenRepository) GetByHash(
	ctx context.Context,
	hash string,
) (*domain.RefreshToken, error) {

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token domain.RefreshToken

	err := r.db.GetContext(ctx, &token, query, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	return &token, nil
}

func (r *PgRefreshTokenRepository) DeleteByHash(
	ctx context.Context,
	hash string,
) error {

	query := `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(ctx, query, hash)
	return err
}

func (r *PgRefreshTokenRepository) DeleteByUserId(
	ctx context.Context,
	userId string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
`
	_, err := r.db.ExecContext(ctx, query, userId)
	return err
}

func (r *PgRefreshTokenRepository) ReplaceForUser(
	ctx context.Context,
	token *domain.RefreshToken,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(
		ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		token.UserID,
	); err != nil {
		return fmt.Errorf("failed to delete old refresh tokens: %w", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	); err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit refresh token transaction: %w", err)
	}
	tx = nil
	return nil
}

func (r *PgRefreshTokenRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM refresh_tokens WHERE expires_at <= $1`,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows for refresh token cleanup: %w", err)
	}
	return rowsAffected, nil
}
