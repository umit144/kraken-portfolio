package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey    string
	ApiSecret string
}

var (
	ErrNoAPIKey    = fmt.Errorf("KRAKEN_API_KEY is not set")
	ErrNoAPISecret = fmt.Errorf("KRAKEN_API_SECRET is not set")
)

func LoadEnv(filePath string) error {
	if filePath != "" {
		if err := godotenv.Load(filePath); err != nil {
			return fmt.Errorf("error loading env file: %w", err)
		}
	}
	return nil
}

func New(apiKey, apiSecret string) (*Config, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}
	if apiSecret == "" {
		return nil, ErrNoAPISecret
	}

	return &Config{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
	}, nil
}

func LoadConfig(envPath string) (*Config, error) {
	if err := LoadEnv(envPath); err != nil {
		return nil, err
	}

	apiKey := os.Getenv("KRAKEN_API_KEY")
	apiSecret := os.Getenv("KRAKEN_API_SECRET")

	return New(apiKey, apiSecret)
}

func (c *Config) Validate() error {
	if c.ApiKey == "" {
		return ErrNoAPIKey
	}
	if c.ApiSecret == "" {
		return ErrNoAPISecret
	}
	return nil
}
