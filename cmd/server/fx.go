package server

import (
	"time"

	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/health"
	healthfx "github.com/webitel/webitel-go-kit/infra/health/fx"
	otelhealth "github.com/webitel/webitel-go-kit/infra/otel/instrumentation/health"
	pgwfx "github.com/webitel/webitel-go-kit/infra/pgw/fx"

	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/internal/model"
)

// pgwStartTimeout bounds the initial primary connect. pgw retries on its own
// background context, so this must outlive the library's default retry budget.
const pgwStartTimeout = 60 * time.Second

// NewApp builds the server graph. Migrations are not part of it.
func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.StartTimeout(pgwStartTimeout),
		fx.Provide(
			func() *config.Config { return cfg },
			ProvideLogger,
		),
		healthfx.Module(healthfx.Config{
			HTTPAddr:     cfg.Health.Address,
			StartTimeout: cfg.Health.StartTimeoutDuration(),
			StopTimeout:  cfg.Health.StopTimeoutDuration(),
		}, health.WithTransport(otelhealth.NewTransport())),
		pgwfx.Module(pgwfx.Config{
			ApplicationName: model.ServiceName,
			DSN:             cfg.Postgres.DSN,
			MaxConns:        cfg.Postgres.MaxOpenConns,
			MinConns:        cfg.Postgres.MaxIdleConns,
			RegisterHealth:  true,
		}),
		healthfx.Shutdown(),
	)
}
