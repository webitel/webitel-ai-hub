// Package mtls loads mutual-TLS material for the gRPC server and outbound clients.
package mtls

import (
	"crypto/tls"

	"go.uber.org/fx"

	wgrpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"

	"github.com/webitel/webitel-ai-hub/config"
)

// Module provides the process mTLS config from service.conn.
var Module = fx.Module("mtls",
	fx.Provide(
		ProvideConfig,
	))

// Config holds the server and client TLS settings. Both may be nil when
// verify_certs is false.
type Config struct {
	Client *tls.Config
	Server *tls.Config
}

// ProvideConfig builds server/client TLS from appconfig.GRPCConn via go-kit
// transport helpers. When verify_certs is false, both sides stay nil (plain gRPC).
func ProvideConfig(cfg *config.Config) (*Config, error) {
	conn := cfg.Service.Connection
	out := &Config{}

	server, err := wgrpc.ServerTLSConfig(conn)
	if err != nil {
		return nil, err
	}

	client, err := wgrpc.ClientTLSConfig(conn)
	if err != nil {
		return nil, err
	}

	out.Server = server
	out.Client = client

	return out, nil
}
