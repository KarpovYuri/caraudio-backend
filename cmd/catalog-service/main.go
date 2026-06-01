package main

import (
	"log/slog"
	"os"

	catalogconfig "github.com/KarpovYuri/caraudio-backend/internal/catalog/config"
	catalogdb "github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/database/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := catalogconfig.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := catalogdb.InitDB(&cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "error", err)
		}
	}()

	supplierRepo := catalogdb.NewPostgresSupplierRepository(db)

	slog.Info("supplier repository initialized",
		slog.Any("supplierRepo", supplierRepo),
	)
}
