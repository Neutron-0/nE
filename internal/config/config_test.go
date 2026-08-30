package config_test

import (
	"testing"
	"time"

	"ne/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Server.Port != 4533 {
		t.Errorf("expected default port 4533, got %d", cfg.Server.Port)
	}
	if cfg.Scanner.WorkerCount != 4 {
		t.Errorf("expected default worker count 4, got %d", cfg.Scanner.WorkerCount)
	}
	if cfg.Auth.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("expected default access token expiry 15m, got %v", cfg.Auth.AccessTokenExpiry)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("NE_PORT", "8080")
	t.Setenv("NE_LOG_LEVEL", "debug")
	t.Setenv("NE_JWT_SECRET", "super-secret-key-1234567890123456")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Server.LogLevel)
	}
	if cfg.Auth.JWTSecretKey != "super-secret-key-1234567890123456" {
		t.Errorf("expected custom jwt secret, got %s", cfg.Auth.JWTSecretKey)
	}
}
