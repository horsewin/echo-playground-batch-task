package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	DB      struct {
		Host     string `mapstructure:"DB_HOST"`
		Port     int    `mapstructure:"DB_PORT"`
		User     string `mapstructure:"DB_USER"`
		Password string `mapstructure:"DB_PASSWORD"`
		DBName   string `mapstructure:"DB_NAME"`
		SSLMode  string `mapstructure:"DB_SSL_MODE"`
	} `mapstructure:"DB"`
	// 必要に応じて設定項目を追加
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// デフォルト値の設定
	viper.SetDefault("APP_NAME", "echo-playground-batch-task")
	viper.SetDefault("DB.HOST", "localhost")
	viper.SetDefault("DB.PORT", 5432)
	viper.SetDefault("DB.USER", "postgres")
	viper.SetDefault("DB.PASSWORD", "postgres")
	viper.SetDefault("DB.DBNAME", "echo_playground")
	viper.SetDefault("DB.SSL_MODE", "disable")

	// 環境変数の読み込み
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
