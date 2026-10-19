package server

import (
	"time"

	"go.uber.org/fx"

	healthfx "github.com/webitel/webitel-go-kit/infra/health/fx"

	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/internal/store/postgres"
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
		healthfx.Module(healthfx.Config{HTTPAddr: cfg.Health.Addr}),
		postgres.Module,
		healthfx.Shutdown(),
	)
}
