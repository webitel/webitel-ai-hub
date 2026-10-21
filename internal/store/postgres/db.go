package postgres

import (
	"context"
	"sync"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/infra/pgw"

	"github.com/webitel/webitel-ai-hub/internal/model"
)

// DB owns the pgw pool manager.
//
// Ping fails with pgw.ErrUnreachable until Open connects the primary.
// A later outage makes Primary return the same error while pgw reconnects,
// which the readiness check reports.
type DB struct {
	mu      sync.RWMutex
	manager *pgw.PoolManager
}

// NewDB returns a database handle that is not connected.
func NewDB() *DB {
	return &DB{}
}

// Open dials the primary pool. Standbys are not configured.
//
// MaxOpenConns maps to pgw MaxConns and MaxIdleConns maps to MinConns.
// pgw has no connection lifetime settings, so those appconfig fields are ignored.
func (d *DB) Open(ctx context.Context, pg appconfig.Postgres) error {
	manager, err := pgw.NewPoolManager(ctx, managerOptions(pg)...)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.manager = manager
	d.mu.Unlock()

	return nil
}

// Close shuts the pool down. It is safe to call before Open and more than once.
func (d *DB) Close() {
	d.mu.Lock()
	manager := d.manager
	d.manager = nil
	d.mu.Unlock()

	if manager == nil {
		return
	}

	manager.Close()
}

// Primary returns the read-write pool, or pgw.ErrUnreachable when it is down.
func (d *DB) Primary() (*pgw.Pool, error) {
	d.mu.RLock()
	manager := d.manager
	d.mu.RUnlock()

	if manager == nil {
		return nil, pgw.ErrUnreachable
	}

	return manager.Primary()
}

// Ping reports whether the primary can answer SELECT 1.
func (d *DB) Ping(ctx context.Context) error {
	pool, err := d.Primary()
	if err != nil {
		return err
	}

	var one int

	return pool.QueryRow(ctx, "SELECT 1").Scan(&one)
}

func managerOptions(pg appconfig.Postgres) []pgw.ConfigOption {
	return []pgw.ConfigOption{
		pgw.WithApplicationName(model.ServiceName),
		pgw.WithPrimaryConfig(primaryConfig(pg)),
	}
}

func primaryConfig(pg appconfig.Postgres) pgw.PrimaryConfig {
	cfg := pgw.PrimaryConfig{DSN: pg.DSN}
	if pg.MaxOpenConns > 0 {
		cfg.MaxConns = pg.MaxOpenConns
	}

	if pg.MaxIdleConns > 0 {
		cfg.MinConns = pg.MaxIdleConns
	}

	return cfg
}
