package server

import (
	"go.uber.org/fx"

	discoveryfx "github.com/webitel/webitel-go-kit/infra/discovery/fx"
	"github.com/webitel/webitel-go-kit/infra/health"
	healthfx "github.com/webitel/webitel-go-kit/infra/health/fx"
	otelhealth "github.com/webitel/webitel-go-kit/infra/otel/instrumentation/health"
	pgwfx "github.com/webitel/webitel-go-kit/infra/pgw/fx"

	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/infra/mtls"
	grpcsrv "github.com/webitel/webitel-ai-hub/infra/server/grpc"
	"github.com/webitel/webitel-ai-hub/internal/model"
)

// NewApp builds the server graph. Migrations are not part of it.
func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.StartTimeout(cfg.Health.StartTimeoutDuration()),
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
			StandbyDSN:      cfg.Postgres.StandbyDSN,
			RegisterHealth:  true,
		}),
		discoveryfx.Module(discoveryfx.Config{
			Addr:      cfg.Consul.Addr,
			Name:      model.ServiceName,
			Version:   model.Version,
			Endpoints: []string{"grpc://" + cfg.Service.Addr},
			Metadata: map[string]string{
				"commit":         model.Commit,
				"commitDate":     model.CommitDate,
				"branch":         model.Branch,
				"buildTimestamp": model.BuildTimestamp,
			},
		}),
		mtls.Module,
		grpcsrv.Module,
		healthfx.Shutdown(),
	)
}
