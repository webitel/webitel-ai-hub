package config

import (
	"testing"

	"github.com/webitel/webitel-go-kit/appconfig"
)

func TestValidateServerRequiresDSN(t *testing.T) {
	cfg := &Config{}

	if err := cfg.validateServer(); err == nil {
		t.Fatal("validateServer() error = nil, want missing dsn")
	}
}

func TestValidateServerDefaultsLogLevel(t *testing.T) {
	cfg := &Config{Postgres: appconfig.Postgres{DSN: "postgres://localhost/webitel"}}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want info", cfg.Log.Level)
	}
}

func TestRequireDSNAcceptsValue(t *testing.T) {
	cfg := &Config{Postgres: appconfig.Postgres{DSN: "postgres://localhost/webitel"}}

	if err := cfg.requireDSN(); err != nil {
		t.Fatalf("requireDSN() error = %v", err)
	}
}
