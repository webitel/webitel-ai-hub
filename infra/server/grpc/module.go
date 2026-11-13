// Package server runs the AI Hub gRPC listener with the standard interceptor
// chain, health, reflection and optional mTLS.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/health"

	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/infra/mtls"
)

// Module wires the gRPC server into the fx graph and registers its health check.
var Module = fx.Module("grpc_server",
	fx.Provide(
		fx.Annotate(
			ProvideServer,
		),
	),
	fx.Invoke(registerHealth),
)

// ProvideServer builds the gRPC server and attaches fx start/stop hooks.
func ProvideServer(conf *config.Config, logger *slog.Logger, mtlsCfg *mtls.Config, lc fx.Lifecycle) (*Server, error) {
	srv, err := New(conf.Service.Addr, func(c *Config) error {
		if mtlsCfg.Server != nil {
			c.TLS = mtlsCfg.Server.Clone()
		}

		c.Logger = logger

		return nil
	})
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				logger.Info(fmt.Sprintf("listen grpc %s:%d", srv.Host(), srv.Port()))

				if err := srv.Listen(); err != nil {
					logger.Error("grpc server error", "err", err)
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			if err := srv.Shutdown(); err != nil {
				logger.Error("error stopping grpc server", "err", err.Error())

				return err
			}

			return nil
		},
	})

	return srv, nil
}

func registerHealth(srv *Server, h *health.Registry) {
	h.Critical("grpc", health.ListenerCheck(srv.Listener()))
}
