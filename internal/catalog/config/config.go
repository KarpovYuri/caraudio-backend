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
	GRPCPort         string         `mapstructure:"grpc_port"`
	HTTPPort         string         `mapstructure:"http_port"`
	JWTSecret        string         `mapstructure:"jwt_secret"`
	AllowedOrigins   []string       `mapstructure:"allowed_origins"`
	HTTPReadTimeout  time.Duration  `mapstructure:"http_read_timeout"`
	HTTPWriteTimeout time.Duration  `mapstructure:"http_write_timeout"`
	HTTPIdleTimeout  time.Duration  `mapstructure:"http_idle_timeout"`
	ShutdownTimeout  time.Duration  `mapstructure:"shutdown_timeout"`
	Uploads          UploadsConfig  `mapstructure:"uploads"`
	Database         DatabaseConfig `mapstructure:"database"`
}

type UploadsConfig struct {
	Dir           string `mapstructure:"dir"`
	PublicPath    string `mapstructure:"public_path"`
	PublicBaseURL string `mapstructure:"public_base_url"`
	MaxLogoBytes  int64  `mapstructure:"max_logo_bytes"`
	MaxLogoSide   int    `mapstructure:"max_logo_side"`
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
	viper.SetConfigName("catalog_service")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var notFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		slog.Info("config file not found, using env vars", "error", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyEnvOverrides(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	slog.Info("catalog service configuration loaded")
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("CATALOG_DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("CATALOG_DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("CATALOG_DATABASE_DBNAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("CATALOG_DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("CATALOG_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("AUTH_JWT_SECRET"); v != "" && cfg.JWTSecret == "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("CATALOG_GRPC_PORT"); v != "" {
		cfg.GRPCPort = v
	}
	if v := os.Getenv("CATALOG_HTTP_PORT"); v != "" {
		cfg.HTTPPort = v
	}
	if v := os.Getenv("CATALOG_ALLOWED_ORIGINS"); v != "" {
		cfg.AllowedOrigins = parseCommaSeparatedList(v)
	}
	if v := os.Getenv("CATALOG_HTTP_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTPReadTimeout = d
		}
	}
	if v := os.Getenv("CATALOG_HTTP_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTPWriteTimeout = d
		}
	}
	if v := os.Getenv("CATALOG_HTTP_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTPIdleTimeout = d
		}
	}
	if v := os.Getenv("CATALOG_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ShutdownTimeout = d
		}
	}
	if v := os.Getenv("CATALOG_UPLOADS_DIR"); v != "" {
		cfg.Uploads.Dir = v
	}
	if v := os.Getenv("CATALOG_UPLOADS_PUBLIC_PATH"); v != "" {
		cfg.Uploads.PublicPath = v
	}
	if v := os.Getenv("CATALOG_UPLOADS_PUBLIC_BASE_URL"); v != "" {
		cfg.Uploads.PublicBaseURL = v
	}
	if v := os.Getenv("CATALOG_UPLOADS_MAX_LOGO_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Uploads.MaxLogoBytes = n
		}
	}
	if v := os.Getenv("CATALOG_UPLOADS_MAX_LOGO_SIDE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Uploads.MaxLogoSide = n
		}
	}
}

func validate(cfg *Config) error {
	if cfg.JWTSecret == "" {
		return errors.New("CATALOG_JWT_SECRET or AUTH_JWT_SECRET is required")
	}
	if cfg.Database.User == "" {
		return errors.New("CATALOG_DATABASE_USER is required")
	}
	if cfg.Database.Password == "" {
		return errors.New("CATALOG_DATABASE_PASSWORD is required")
	}
	if cfg.Database.DBName == "" {
		return errors.New("CATALOG_DATABASE_DBNAME is required")
	}
	if cfg.Database.Host == "" {
		return errors.New("CATALOG_DATABASE_HOST is required")
	}
	if cfg.GRPCPort == "" {
		return errors.New("CATALOG_GRPC_PORT is required")
	}
	if cfg.HTTPPort == "" {
		return errors.New("CATALOG_HTTP_PORT is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return errors.New("CATALOG_ALLOWED_ORIGINS is required")
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
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Uploads.Dir == "" {
		cfg.Uploads.Dir = "./uploads"
	}
	if cfg.Uploads.PublicPath == "" {
		cfg.Uploads.PublicPath = "/static"
	}
	if cfg.Uploads.MaxLogoBytes <= 0 {
		cfg.Uploads.MaxLogoBytes = 10 << 20 // 10 MiB
	}
	if cfg.Uploads.MaxLogoSide <= 0 {
		cfg.Uploads.MaxLogoSide = 512
	}
	return nil
}

func parseCommaSeparatedList(value string) []string {
	raw := strings.Split(value, ",")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
