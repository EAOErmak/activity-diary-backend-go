package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost              string
	DBPort              int
	DBUser              string
	DBPassword          string
	DBName              string
	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	AnalyticsServiceURL string
	MigrationsPath      string
}

func Load() Config {
	return Config{
		DBHost:              env("DB_HOST", "postgres"),
		DBPort:              envInt("DB_PORT", 5432),
		DBUser:              env("DB_USER", "postgres"),
		DBPassword:          env("DB_PASSWORD", "postgres"),
		DBName:              env("DB_NAME", "activity_diary"),
		JWTSecret:           env("JWT_SECRET", "change-me"),
		AccessTokenTTL:      time.Duration(envInt("ACCESS_TOKEN_TTL_MINUTES", 60)) * time.Minute,
		RefreshTokenTTL:     time.Duration(envInt("REFRESH_TOKEN_TTL_HOURS", 72)) * time.Hour,
		AnalyticsServiceURL: env("ANALYTICS_SERVICE_URL", "http://analytics:18081"),
		MigrationsPath:      env("MIGRATIONS_PATH", "internal/migrations"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
