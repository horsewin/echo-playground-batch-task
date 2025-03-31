package config

import (
	"log"
	"os"
	"strconv"

	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
)

type Config struct {
	DB  database.Config
	SFN struct {
		TaskToken string
	}
	EnableTracing bool
}

// LoadConfig は設定を読み込みます
func LoadConfig(taskToken string) (*Config, error) {
	cfg := &Config{
		DB: database.Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
			UserName: getEnvOrDefault("DB_USERNAME", "sbcntrapp"),
			Password: getEnvOrDefault("DB_PASSWORD", "password"),
			DBName:   getEnvOrDefault("DB_NAME", "sbcntrapp"),
		},
		SFN: struct {
			TaskToken string
		}{
			TaskToken: taskToken,
		},
		EnableTracing: getEnvOrDefault("ENABLE_TRACING", "false") == "true",
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Environment variable %s is not set, using default value", key)
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
