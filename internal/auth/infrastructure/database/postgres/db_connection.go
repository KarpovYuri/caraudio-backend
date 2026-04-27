package postgres

import (
	"fmt"
	"log/slog"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db connection after failed ping", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connection established")
	return db, nil
}
