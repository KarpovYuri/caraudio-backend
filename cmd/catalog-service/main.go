package main

import (
	"log/slog"
	"os"

	catalogconfig "github.com/KarpovYuri/caraudio-backend/internal/catalog/config"
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

	logger.Info("config loaded", cfg)

}
