package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort           string         `mapstructure:"grpc_port"`
	JWTSecret          string         `mapstructure:"jwt_secret"`
	JWTExpirationHours int            `mapstructure:"jwt_expiration_hours"`
	AllowedOrigins     []string       `mapstructure:"allowed_origins"`
	CookieSecure       bool           `mapstructure:"cookie_secure"`
	Database           DatabaseConfig `mapstructure:"database"`
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
		log.Printf(".env file not loaded, relying on process environment: %v", err)
	}

	viper.AddConfigPath("./config")
	viper.SetConfigName("auth_service")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var notFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundError) {
			log.Printf("Config file not found: %v, using env vars", err)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if dbPassword := os.Getenv("AUTH_DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if jwtSecret := os.Getenv("AUTH_JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}
	if grpcPort := os.Getenv("AUTH_GRPC_PORT"); grpcPort != "" {
		cfg.GRPCPort = grpcPort
	}
	if allowedOrigin := os.Getenv("AUTH_ALLOWED_ORIGIN"); allowedOrigin != "" {
		cfg.AllowedOrigins = []string{allowedOrigin}
	}
	if allowedOrigins := os.Getenv("AUTH_ALLOWED_ORIGINS"); allowedOrigins != "" {
		cfg.AllowedOrigins = parseCommaSeparatedList(allowedOrigins)
	}
	if cookieSecure := os.Getenv("AUTH_COOKIE_SECURE"); cookieSecure != "" {
		cfg.CookieSecure = cookieSecure == "true"
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("AUTH_JWT_SECRET is required")
	}
	if cfg.Database.Password == "" {
		return nil, errors.New("AUTH_DATABASE_PASSWORD is required")
	}
	if cfg.GRPCPort == "" {
		return nil, errors.New("AUTH_GRPC_PORT is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return nil, errors.New("AUTH_ALLOWED_ORIGINS (or AUTH_ALLOWED_ORIGIN) is required")
	}

	log.Println("Auth Service Configuration loaded successfully")
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
