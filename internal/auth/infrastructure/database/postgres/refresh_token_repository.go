package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	DeleteByHash(ctx context.Context, hash string) error
}

type PostgresRefreshTokenRepository struct {
	db *sqlx.DB
}

func NewPostgresRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Create(
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

func (r *PostgresRefreshTokenRepository) GetByHash(
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
		return nil, err
	}

	return &token, nil
}

func (r *PostgresRefreshTokenRepository) DeleteByHash(
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
