package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/infra/pgw"
)

func TestPingBeforeOpen(t *testing.T) {
	db := NewDB()

	err := db.Ping(context.Background())
	if !errors.Is(err, pgw.ErrUnreachable) {
		t.Fatalf("Ping() error = %v, want ErrUnreachable", err)
	}
}

func TestCloseBeforeOpen(t *testing.T) {
	NewDB().Close()
}

func TestPrimaryConfigMapsPoolLimits(t *testing.T) {
	cfg := primaryConfig(appconfig.Postgres{
		DSN:          "postgres://localhost/webitel",
		MaxOpenConns: 8,
		MaxIdleConns: 2,
	})

	if cfg.DSN != "postgres://localhost/webitel" {
		t.Fatalf("DSN = %q", cfg.DSN)
	}

	if cfg.MaxConns != 8 || cfg.MinConns != 2 {
		t.Fatalf("pool = max %d min %d, want 8 and 2", cfg.MaxConns, cfg.MinConns)
	}
}

func TestPrimaryConfigLeavesZeroLimits(t *testing.T) {
	cfg := primaryConfig(appconfig.Postgres{DSN: "postgres://localhost/webitel"})

	if cfg.MaxConns != 0 || cfg.MinConns != 0 {
		t.Fatalf("pool = max %d min %d, want zeros so pgw defaults apply", cfg.MaxConns, cfg.MinConns)
	}
}
