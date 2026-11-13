package server

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/netip"
	"os"

	"buf.build/go/protovalidate"
	validatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	Addr     string
	host     string
	port     int
	log      *slog.Logger
	listener net.Listener
	health   *health.Server
}

// Config holds optional construction settings for New.
type Config struct {
	TLS    *tls.Config
	Logger *slog.Logger
}

// Option mutates Config during New.
type Option func(*Config) error

// New provides a new gRPC server.
func New(addr string, opts ...Option) (*Server, error) {
	var (
		conf    Config
		grpcTLS credentials.TransportCredentials
	)
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

	if conf.TLS != nil {
		grpcTLS = credentials.NewTLS(conf.TLS)
	}

	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}

	// Auth interceptor is wired in a follow-up task.
	s := grpc.NewServer(
		grpc.Creds(grpcTLS),
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

	host, port, err := advertiseAddr(l.Addr())
	if err != nil {
		_ = l.Close()

		return nil, err
	}

	return &Server{
		Addr:     addr,
		Server:   s,
		log:      log,
		host:     host,
		port:     port,
		listener: l,
		health:   healthSrv,
	}, nil
}

// Listen serves gRPC on the bound listener until Shutdown.
func (s *Server) Listen() error {
	return s.Serve(s.listener)
}

// Shutdown marks health NOT_SERVING and GracefulStops the server.
func (s *Server) Shutdown() error {
	s.log.Debug("receive shutdown grpc")

	if s.health != nil {
		// Flip health to NOT_SERVING so load balancers drain us before we stop.
		s.health.Shutdown()
	}

	s.GracefulStop()

	return nil
}

// Host returns the advertised gRPC host, honoring PROXY_GRPC_HOST when set.
func (s *Server) Host() string {
	if e, ok := os.LookupEnv("PROXY_GRPC_HOST"); ok {
		return e
	}

	return s.host
}

// Port returns the bound listen port.
func (s *Server) Port() int {
	return s.port
}

// Listener returns the socket the server accepts connections on.
func (s *Server) Listener() net.Listener {
	return s.listener
}

func advertiseAddr(addr net.Addr) (string, int, error) {
	ap, err := netip.ParseAddrPort(addr.String())
	if err != nil {
		return "", 0, err
	}

	adv, err := hostaddr.Advertise(ap)
	if err != nil {
		// Fall back to the bound address when a private IP cannot be chosen
		// (for example, only loopback is up in local/dev).
		return ap.Addr().String(), int(ap.Port()), nil
	}

	return adv.Addr().String(), int(adv.Port()), nil
}
