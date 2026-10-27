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

func TestValidateServerDefaultsHealthTimeouts(t *testing.T) {
	cfg := &Config{Postgres: appconfig.Postgres{DSN: "postgres://localhost/webitel"}}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Health.StartTimeout != defaultHealthStartTimeoutSec {
		t.Fatalf("StartTimeout = %d, want %d", cfg.Health.StartTimeout, defaultHealthStartTimeoutSec)
	}

	if cfg.Health.StopTimeout != defaultHealthStopTimeoutSec {
		t.Fatalf("StopTimeout = %d, want %d", cfg.Health.StopTimeout, defaultHealthStopTimeoutSec)
	}
}

func TestRequireDSNAcceptsValue(t *testing.T) {
	cfg := &Config{Postgres: appconfig.Postgres{DSN: "postgres://localhost/webitel"}}

	if err := cfg.requireDSN(); err != nil {
		t.Fatalf("requireDSN() error = %v", err)
	}
}
