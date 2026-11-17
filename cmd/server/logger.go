package server

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/appconfig"
	"github.com/webitel/webitel-go-kit/infra/discovery"
	otelsdk "github.com/webitel/webitel-go-kit/infra/otel/sdk"

	"github.com/webitel/webitel-ai-hub/config"
	"github.com/webitel/webitel-ai-hub/internal/model"

	// Register the OTel exporters selected by OTEL_*_EXPORTER.
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/prometheus"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/stdout"
)

// ProvideLogger builds a slog logger from the log section.
// When log.otel is true, the go-kit OTel SDK is configured (traces, metrics,
// logs) from OTEL_* env vars and slog is bridged into the log provider.
func ProvideLogger(cfg *config.Config, lc fx.Lifecycle) (*slog.Logger, error) {
	settings := cfg.Log
	if !settings.Console && !settings.Otel && settings.File == "" {
		settings.Console = true
	}

	opts := &slog.HandlerOptions{Level: parseLevel(settings.Level)}

	handlers, err := collectHandlers(settings, opts, lc)
	if err != nil {
		return nil, err
	}

	logger := slog.New(combineHandlers(handlers, opts))
	slog.SetDefault(logger)

	return logger, nil
}

func collectHandlers(settings appconfig.Log, opts *slog.HandlerOptions, lc fx.Lifecycle) ([]slog.Handler, error) {
	var handlers []slog.Handler

	if settings.Console {
		handlers = append(handlers, consoleHandler(settings.JSON, opts))
	}

	if settings.File != "" {
		handler, err := fileHandler(settings.File, settings.JSON, opts, lc)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, handler)
	}

	if settings.Otel {
		if err := attachOTel(&handlers, lc); err != nil {
			return nil, err
		}
	}

	return handlers, nil
}

// attachOTel configures the global OTel providers and optionally appends a
// slog bridge when a logs exporter is selected.
func attachOTel(handlers *[]slog.Handler, lc fx.Lifecycle) error {
	service := resource.NewSchemaless(
		semconv.ServiceName(model.ServiceName),
		semconv.ServiceVersion(model.Version),
		semconv.ServiceInstanceID(discovery.GenerateInstanceID(model.ServiceName)),
		semconv.ServiceNamespace(model.ServiceNamespace),
	)
	otelHandler := otelslog.NewHandler("slog")

	shutdown, err := otelsdk.Configure(context.Background(), otelsdk.WithResource(service),
		otelsdk.WithLogBridge(
			func() {
				*handlers = append(*handlers, otelHandler)
			},
		),
	)
	if err != nil {
		return err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return shutdown(ctx)
		},
	})

	return nil
}

func consoleHandler(json bool, opts *slog.HandlerOptions) slog.Handler {
	if json {
		return slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.NewTextHandler(os.Stdout, opts)
}

func fileHandler(path string, json bool, opts *slog.HandlerOptions, lc fx.Lifecycle) (slog.Handler, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- path from trusted app config
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return file.Close()
		},
	})

	if json {
		return slog.NewJSONHandler(file, opts), nil
	}

	return slog.NewTextHandler(file, opts), nil
}

func combineHandlers(handlers []slog.Handler, opts *slog.HandlerOptions) slog.Handler {
	switch len(handlers) {
	case 0:
		return slog.NewTextHandler(os.Stdout, opts)
	case 1:
		return handlers[0]
	default:
		return &multiHandler{handlers: handlers}
	}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	last := len(h.handlers) - 1
	for i, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}

		next := record
		if i != last {
			next = record.Clone()
		}

		if err := handler.Handle(ctx, next); err != nil {
			return err
		}
	}

	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		next[i] = handler.WithAttrs(attrs)
	}

	return &multiHandler{handlers: next}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		next[i] = handler.WithGroup(name)
	}

	return &multiHandler{handlers: next}
}
