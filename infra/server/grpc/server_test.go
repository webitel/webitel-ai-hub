package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

func TestNewAdvertiseAddr(t *testing.T) {
	t.Parallel()

	srv, err := New("127.0.0.1:0", withTestLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	t.Cleanup(func() {
		srv.Stop()
		_ = srv.Listener().Close()
	})

	adv, err := srv.AdvertiseAddr()
	if err != nil {
		t.Fatalf("AdvertiseAddr() error = %v", err)
	}

	if !adv.Addr().IsLoopback() {
		t.Fatalf("AdvertiseAddr() = %v, want loopback", adv)
	}

	if adv.Port() == 0 {
		t.Fatal("AdvertiseAddr() port = 0")
	}
}

func TestShutdownHonorsContext(t *testing.T) {
	t.Parallel()

	srv, err := New("127.0.0.1:0", withTestLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	go func() {
		_ = srv.Listen()
	}()

	conn, err := grpc.NewClient(
		srv.Listener().Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		client := healthgrpc.NewHealthClient(conn)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		stream, err := client.Watch(ctx, &healthgrpc.HealthCheckRequest{})
		if err != nil {
			done <- err

			return
		}

		close(started)
		<-release

		_, _ = stream.Recv()

		done <- nil
	}()

	select {
	case <-started:
	case err := <-done:
		t.Fatalf("watch setup failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("watch did not start")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	finished := make(chan struct{})

	go func() {
		srv.Shutdown(ctx)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown() did not return after context deadline")
	}

	close(release)
}

func TestListenAddrFallsBackToListener(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	t.Cleanup(func() { _ = ln.Close() })

	srv := &Server{listener: ln}

	got := listenAddr(srv)
	if got == "" {
		t.Fatal("listenAddr() = empty")
	}
}

func withTestLogger() Option {
	return func(c *Config) error {
		c.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

		return nil
	}
}
