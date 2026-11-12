package config

import (
	"testing"

	"github.com/webitel/webitel-go-kit/appconfig"
)

func testPostgres(dsn string) appconfig.Postgres {
	return appconfig.Postgres{DSN: dsn}
}

func TestValidateServerRequiresDSN(t *testing.T) {
	cfg := &Config{}

	if err := cfg.validateServer(); err == nil {
		t.Fatal("validateServer() error = nil, want missing dsn")
	}
}

func TestValidateServerDefaultsLogLevel(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want info", cfg.Log.Level)
	}
}

func TestValidateServerDefaultsHealthTimeouts(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Health.StartTimeout != defaultHealthStartTimeoutSec {
		t.Fatalf("StartTimeout = %d, want %d", cfg.Health.StartTimeout, defaultHealthStartTimeoutSec)
	}

	if cfg.Health.StopTimeout != 0 {
		t.Fatalf("StopTimeout = %d, want 0", cfg.Health.StopTimeout)
	}
}

func TestValidateServerKeepsStopTimeout(t *testing.T) {
	cfg := &Config{
		Postgres: testPostgres("postgres://localhost/webitel"),
		Health:   HealthConfig{StopTimeout: 20},
	}

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Health.StopTimeout != 20 {
		t.Fatalf("StopTimeout = %d, want 20", cfg.Health.StopTimeout)
	}
}

func TestRequireDSNAcceptsValue(t *testing.T) {
	cfg := &Config{Postgres: testPostgres("postgres://localhost/webitel")}

	if err := cfg.requireDSN(); err != nil {
		t.Fatalf("requireDSN() error = %v", err)
	}
}
