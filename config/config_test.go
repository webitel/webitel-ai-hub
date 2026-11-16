package config

import (
	"strings"
	"testing"

	"github.com/webitel/webitel-go-kit/appconfig"
)

func testPostgres(dsn string) appconfig.Postgres {
	return appconfig.Postgres{DSN: dsn}
}

func validServerConfig() *Config {
	return &Config{
		Service:  ServiceConfig{Addr: "127.0.0.1:8080"},
		Postgres: testPostgres("postgres://localhost/webitel"),
		Consul:   appconfig.Consul{Addr: "127.0.0.1:8500"},
	}
}

func TestValidateServerRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:    "requires dsn",
			mutate:  func(c *Config) { c.Postgres.DSN = "" },
			wantErr: "postgres.dsn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validServerConfig()
			tt.mutate(cfg)

			err := cfg.validateServer()
			if err == nil {
				t.Fatal("validateServer() error = nil, want error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateServer() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServerDefaultsLogLevel(t *testing.T) {
	cfg := validServerConfig()

	if err := cfg.validateServer(); err != nil {
		t.Fatalf("validateServer() error = %v", err)
	}

	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want info", cfg.Log.Level)
	}
}

func TestValidateServerDefaultsHealthTimeouts(t *testing.T) {
	cfg := validServerConfig()

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
	cfg := validServerConfig()
	cfg.Health.StopTimeout = 20

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
