// Package server runs the AI Hub gRPC listener with the standard interceptor
// chain, health, reflection and optional mTLS.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/health"
	wgrpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"

	"github.com/webitel/webitel-ai-hub/config"
)

// Module wires the gRPC server into the fx graph and registers its health check.
var Module = fx.Module("grpc_server",
	fx.Provide(
		ProvideServer,
	),
	fx.Invoke(registerHealth),
)

// ProvideServer builds the gRPC server and attaches fx start/stop hooks.
func ProvideServer(conf *config.Config, logger *slog.Logger, lc fx.Lifecycle) (*Server, error) {
	creds, err := wgrpc.ServerCredentials(conf.Service.Connection)
	if err != nil {
		return nil, err
	}

	srv, err := New(conf.Service.Addr, func(c *Config) error {
		c.Creds = creds
		c.Logger = logger

		return nil
	})
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				logger.Info(fmt.Sprintf("listen grpc %s", listenAddr(srv)))

				if err := srv.Listen(); err != nil {
					logger.Error("grpc server error", "err", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.Shutdown(ctx)

			return nil
		},
	})

	return srv, nil
}

func registerHealth(srv *Server, h *health.Registry) {
	h.Critical("grpc", health.ListenerCheck(srv.Listener()))
}

func listenAddr(srv *Server) string {
	if adv, err := srv.AdvertiseAddr(); err == nil {
		return adv.String()
	}

	return srv.Listener().Addr().String()
}
