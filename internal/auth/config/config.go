package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort           string         `mapstructure:"grpc_port"`
	JWTSecret          string         `mapstructure:"jwt_secret"`
	JWTExpirationHours int            `mapstructure:"jwt_expiration_hours"`
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
	viper.AddConfigPath("./config")
	viper.SetConfigName("auth_service")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundError) {
			log.Printf("Config file not found: %v, using env vars if available", err)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if dbPassword := viper.GetString("AUTH_DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if jwtSecret := viper.GetString("AUTH_JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}
	if grpcPort := viper.GetString("AUTH_GRPC_PORT"); grpcPort != "" {
		cfg.GRPCPort = grpcPort
	}

	log.Println("Auth Service Configuration loaded successfully")
	return &cfg, nil
}
