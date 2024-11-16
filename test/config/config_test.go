package config_test

import (
	"os"
	"testing"

	"github.com/umit144/kraken-portfolio/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		wantErr   error
	}{
		{
			name:      "valid config",
			apiKey:    "test-key",
			apiSecret: "test-secret",
			wantErr:   nil,
		},
		{
			name:      "missing api key",
			apiKey:    "",
			apiSecret: "test-secret",
			wantErr:   config.ErrNoAPIKey,
		},
		{
			name:      "missing api secret",
			apiKey:    "test-key",
			apiSecret: "",
			wantErr:   config.ErrNoAPISecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.New(tt.apiKey, tt.apiSecret)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.apiKey, cfg.ApiKey)
				assert.Equal(t, tt.apiSecret, cfg.ApiSecret)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name    string
		envPath string
		setup   func() error
		cleanup func()
		wantErr bool
	}{
		{
			name:    "valid env file",
			envPath: "test.env",
			setup: func() error {
				return os.WriteFile("test.env", []byte(
					"KRAKEN_API_KEY=test-key\nKRAKEN_API_SECRET=test-secret",
				), 0644)
			},
			cleanup: func() {
				os.Remove("test.env")
			},
			wantErr: false,
		},
		{
			name:    "non-existent file",
			envPath: "nonexistent.env",
			setup:   func() error { return nil },
			cleanup: func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatal(err)
			}
			defer tt.cleanup()

			err := config.LoadEnv(tt.envPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		envPath string
		envVars map[string]string
		setup   func() error
		cleanup func()
		wantErr bool
	}{
		{
			name:    "load from env file",
			envPath: "test.env",
			setup: func() error {
				return os.WriteFile("test.env", []byte(
					"KRAKEN_API_KEY=test-key\nKRAKEN_API_SECRET=test-secret",
				), 0644)
			},
			cleanup: func() {
				os.Remove("test.env")
			},
			wantErr: false,
		},
		{
			name:    "load from environment",
			envPath: "",
			envVars: map[string]string{
				"KRAKEN_API_KEY":    "test-key",
				"KRAKEN_API_SECRET": "test-secret",
			},
			setup: func() error {
				for k, v := range map[string]string{
					"KRAKEN_API_KEY":    "test-key",
					"KRAKEN_API_SECRET": "test-secret",
				} {
					if err := os.Setenv(k, v); err != nil {
						return err
					}
				}
				return nil
			},
			cleanup: func() {
				os.Unsetenv("KRAKEN_API_KEY")
				os.Unsetenv("KRAKEN_API_SECRET")
			},
			wantErr: false,
		},
		{
			name:    "missing env vars",
			envPath: "",
			setup:   func() error { return nil },
			cleanup: func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatal(err)
			}
			defer tt.cleanup()

			cfg, err := config.LoadConfig(tt.envPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.NotEmpty(t, cfg.ApiKey)
				assert.NotEmpty(t, cfg.ApiSecret)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr error
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				ApiKey:    "test-key",
				ApiSecret: "test-secret",
			},
			wantErr: nil,
		},
		{
			name: "missing api key",
			cfg: &config.Config{
				ApiKey:    "",
				ApiSecret: "test-secret",
			},
			wantErr: config.ErrNoAPIKey,
		},
		{
			name: "missing api secret",
			cfg: &config.Config{
				ApiKey:    "test-key",
				ApiSecret: "",
			},
			wantErr: config.ErrNoAPISecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
