package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	Environment string

	// TLS configuration
	TLS struct {
		CertFile string
		KeyFile  string
	}

	// CORS configuration
	AllowedOrigins []string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnvInt("PORT", 8080),
		DatabaseURL: getEnvString("DATABASE_URL", "postgres://localhost/gamedb?sslmode=disable"),
		JWTSecret:   getEnvString("JWT_SECRET", "your-super-secret-key-change-in-production"),
		Environment: getEnvString("ENVIRONMENT", "development"),
		TLS: struct {
			CertFile string
			KeyFile  string
		}{

			CertFile: getEnvString("TLS_CERT_FILE", "server.crt"),
			KeyFile:  getEnvString("TLS_KEY_FILE", "server.key"),
		},
		AllowedOrigins: strings.Split(getEnvString("ALLOWED_ORIGINS", "*"), ","),
	}

	cfg.Port = *flag.Int("port", cfg.Port, "Port to run the server on")
	cfg.DatabaseURL = *flag.String("db_url", cfg.DatabaseURL, "Database connection URL")

	flag.Parse()

	// Validate TLS configuration
	if cfg.TLS.CertFile == "" || cfg.TLS.KeyFile == "" {
		return cfg, os.ErrInvalid
	}

	if _, err := os.Stat(cfg.TLS.CertFile); os.IsNotExist(err) {
		return cfg, fmt.Errorf("TLS certificate file does not exist: %w", err)
	}
	if _, err := os.Stat(cfg.TLS.KeyFile); os.IsNotExist(err) {
		return cfg, fmt.Errorf("TLS key file does not exist: %w", err)
	}

	return cfg, nil
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
