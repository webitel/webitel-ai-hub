package server

import (
	"context"
	"log/slog"
	"net"
	"net/netip"

	"buf.build/go/protovalidate"
	validatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/webitel/webitel-go-kit/pkg/hostaddr"
	kitinterceptors "github.com/webitel/webitel-go-kit/pkg/interceptors"

	"github.com/webitel/webitel-ai-hub/infra/server/grpc/interceptors"
	"github.com/webitel/webitel-ai-hub/internal/model"
)

// Server is the gRPC listener with health and discovery metadata helpers.
type Server struct {
	*grpc.Server

	log      *slog.Logger
	listener net.Listener
	health   *health.Server
}

// Config holds optional construction settings for New.
type Config struct {
	Creds  credentials.TransportCredentials
	Logger *slog.Logger
}

// Option mutates Config during New.
type Option func(*Config) error

// New provides a new gRPC server.
func New(addr string, opts ...Option) (*Server, error) {
	var conf Config
	for _, opt := range opts {
		if err := opt(&conf); err != nil {
			return nil, err
		}
	}

	log := conf.Logger
	if log == nil {
		log = slog.Default()
	}

	if addr == "" {
		addr = ":0"
	}

	creds := conf.Creds
	if creds == nil {
		creds = insecure.NewCredentials()
	}

	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.NewUnaryErrorInterceptor(log),
			kitinterceptors.RecoveryUnaryServerInterceptor(log),
			validatemiddleware.UnaryServerInterceptor(validator),
		),
	)

	healthSrv := health.NewServer()
	healthgrpc.RegisterHealthServer(s, healthSrv)
	healthSrv.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus(model.ServiceName, healthgrpc.HealthCheckResponse_SERVING)

	reflection.Register(s)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Server:   s,
		log:      log,
		listener: l,
		health:   healthSrv,
	}, nil
}

// Listen serves gRPC on the bound listener until Shutdown.
func (s *Server) Listen() error {
	return s.Serve(s.listener)
}

// Shutdown marks health NOT_SERVING, then GracefulStops; on ctx done it Stop()s
// so later fx hooks (Consul Deregister) can still run.
func (s *Server) Shutdown(ctx context.Context) {
	s.log.Debug("receive shutdown grpc")

	if s.health != nil {
		s.health.Shutdown()
	}

	done := make(chan struct{})

	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.Stop()
	}
}

// AdvertiseAddr returns the address to publish to Consul, preferring a
// private interface IP when the listener is bound to an unspecified address.
func (s *Server) AdvertiseAddr() (netip.AddrPort, error) {
	listen, err := netip.ParseAddrPort(s.listener.Addr().String())
	if err != nil {
		return netip.AddrPort{}, err
	}

	return hostaddr.Advertise(listen)
}

// Listener returns the socket the server accepts connections on.
func (s *Server) Listener() net.Listener {
	return s.listener
}
