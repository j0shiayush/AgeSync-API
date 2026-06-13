package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App AppConfig
	DB  DBConfig
}

type AppConfig struct {
	Env  string
	Port int
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}


func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	appPort, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid APP_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid DB_PORT: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: appPort,
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "secret"),
			Name:     getEnv("DB_NAME", "userapi"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}