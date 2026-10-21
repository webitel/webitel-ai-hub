// Package postgres connects the service to PostgreSQL through pgw and maps
// driver errors onto go-kit errors.
//
// Migrations are not applied here. The migrate command owns schema ai_hub.
package postgres

import (
	"context"

	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/health"

	"github.com/webitel/webitel-ai-hub/config"
)

const healthCheckName = "postgres"

// Module connects pgw when the application starts and registers the database
// as a critical readiness check.
var Module = fx.Module("postgres",
	fx.Provide(ProvideDB),
	fx.Invoke(registerHealth),
)

// ProvideDB dials the primary during OnStart and closes it during OnStop.
func ProvideDB(cfg *config.Config, lc fx.Lifecycle) *DB {
	db := NewDB()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.Open(ctx, cfg.Postgres)
		},
		OnStop: func(context.Context) error {
			db.Close()

			return nil
		},
	})

	return db
}

func registerHealth(db *DB, registry *health.Registry) {
	registry.Informational(healthCheckName, db.Ping)
}
