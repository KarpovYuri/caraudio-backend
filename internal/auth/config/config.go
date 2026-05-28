package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort          string         `mapstructure:"grpc_port"`
	HTTPPort          string         `mapstructure:"http_port"`
	JWTSecret         string         `mapstructure:"jwt_secret"`
	AllowedOrigins    []string       `mapstructure:"allowed_origins"`
	CookieSecure      bool           `mapstructure:"cookie_secure"`
	HTTPReadTimeout   time.Duration  `mapstructure:"http_read_timeout"`
	HTTPWriteTimeout  time.Duration  `mapstructure:"http_write_timeout"`
	HTTPIdleTimeout   time.Duration  `mapstructure:"http_idle_timeout"`
	ShutdownTimeout   time.Duration  `mapstructure:"shutdown_timeout"`
	TokenCleanupEvery time.Duration  `mapstructure:"token_cleanup_every"`
	Database          DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info(".env file not loaded, relying on process environment", "error", err)
	}

	viper.AddConfigPath("./config")
	viper.SetConfigName("auth_service")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var notFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundError) {
			slog.Info("config file not found, using env vars", "error", err)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if dbUser := os.Getenv("AUTH_DATABASE_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPassword := os.Getenv("AUTH_DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if dbName := os.Getenv("AUTH_DATABASE_DBNAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}
	if jwtSecret := os.Getenv("AUTH_JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}
	if grpcPort := os.Getenv("AUTH_GRPC_PORT"); grpcPort != "" {
		cfg.GRPCPort = grpcPort
	}
	if httpPort := os.Getenv("AUTH_HTTP_PORT"); httpPort != "" {
		cfg.HTTPPort = httpPort
	}
	if allowedOrigins := os.Getenv("AUTH_ALLOWED_ORIGINS"); allowedOrigins != "" {
		cfg.AllowedOrigins = parseCommaSeparatedList(allowedOrigins)
	}
	if cookieSecure := os.Getenv("AUTH_COOKIE_SECURE"); cookieSecure != "" {
		parsed, err := strconv.ParseBool(cookieSecure)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_COOKIE_SECURE value: %w", err)
		}
		cfg.CookieSecure = parsed
	}
	if readTimeout := os.Getenv("AUTH_HTTP_READ_TIMEOUT"); readTimeout != "" {
		duration, err := time.ParseDuration(readTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_HTTP_READ_TIMEOUT value: %w", err)
		}
		cfg.HTTPReadTimeout = duration
	}
	if writeTimeout := os.Getenv("AUTH_HTTP_WRITE_TIMEOUT"); writeTimeout != "" {
		duration, err := time.ParseDuration(writeTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_HTTP_WRITE_TIMEOUT value: %w", err)
		}
		cfg.HTTPWriteTimeout = duration
	}
	if idleTimeout := os.Getenv("AUTH_HTTP_IDLE_TIMEOUT"); idleTimeout != "" {
		duration, err := time.ParseDuration(idleTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_HTTP_IDLE_TIMEOUT value: %w", err)
		}
		cfg.HTTPIdleTimeout = duration
	}
	if shutdownTimeout := os.Getenv("AUTH_SHUTDOWN_TIMEOUT"); shutdownTimeout != "" {
		duration, err := time.ParseDuration(shutdownTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_SHUTDOWN_TIMEOUT value: %w", err)
		}
		cfg.ShutdownTimeout = duration
	}
	if cleanupEvery := os.Getenv("AUTH_TOKEN_CLEANUP_EVERY"); cleanupEvery != "" {
		duration, err := time.ParseDuration(cleanupEvery)
		if err != nil {
			return nil, fmt.Errorf("invalid AUTH_TOKEN_CLEANUP_EVERY value: %w", err)
		}
		cfg.TokenCleanupEvery = duration
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("AUTH_JWT_SECRET is required")
	}
	if cfg.Database.User == "" {
		return nil, errors.New("AUTH_DATABASE_USER is required")
	}
	if cfg.Database.Password == "" {
		return nil, errors.New("AUTH_DATABASE_PASSWORD is required")
	}
	if cfg.Database.DBName == "" {
		return nil, errors.New("AUTH_DATABASE_DBNAME is required")
	}
	if cfg.GRPCPort == "" {
		return nil, errors.New("AUTH_GRPC_PORT is required")
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = ":8080"
	}
	if len(cfg.AllowedOrigins) == 0 {
		return nil, errors.New("AUTH_ALLOWED_ORIGINS (or AUTH_ALLOWED_ORIGIN) is required")
	}
	if cfg.HTTPReadTimeout <= 0 {
		cfg.HTTPReadTimeout = 15 * time.Second
	}
	if cfg.HTTPWriteTimeout <= 0 {
		cfg.HTTPWriteTimeout = 15 * time.Second
	}
	if cfg.HTTPIdleTimeout <= 0 {
		cfg.HTTPIdleTimeout = 60 * time.Second
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 15 * time.Second
	}
	if cfg.TokenCleanupEvery <= 0 {
		cfg.TokenCleanupEvery = 10 * time.Minute
	}

	slog.Info("auth service configuration loaded")
	return &cfg, nil
}

func parseCommaSeparatedList(value string) []string {
	raw := strings.Split(value, ",")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
