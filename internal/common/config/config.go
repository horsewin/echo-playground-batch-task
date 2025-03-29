package config

import (
	"fmt"
	"os"

	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"gopkg.in/yaml.v3"
)

// Config はアプリケーションの設定を表します
type Config struct {
	DB  database.Config `yaml:"db"`
	SFN struct {
		TaskToken string `yaml:"task_token"`
	} `yaml:"sfn"`
}

// LoadConfig は設定ファイルを読み込みます
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
