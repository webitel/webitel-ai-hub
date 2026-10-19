package server

import (
	"context"
	"log/slog"
	"os"

	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/appconfig"

	"github.com/webitel/webitel-ai-hub/config"
)

// ProvideLogger builds a slog logger from the log section.
// OpenTelemetry is not attached; a console handler is used when no sink is configured.
func ProvideLogger(cfg *config.Config, lc fx.Lifecycle) (*slog.Logger, error) {
	settings := cfg.Log
	if !settings.Console && settings.File == "" {
		settings.Console = true
	}

	handler, err := buildHandler(settings, lc)
	if err != nil {
		return nil, err
	}

	return slog.New(handler), nil
}

func buildHandler(settings appconfig.Log, lc fx.Lifecycle) (slog.Handler, error) {
	opts := &slog.HandlerOptions{Level: parseLevel(settings.Level)}

	var handlers []slog.Handler
	if settings.Console {
		handlers = append(handlers, consoleHandler(settings.JSON, opts))
	}

	if settings.File != "" {
		handler, err := fileHandler(settings.File, opts, lc)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, handler)
	}

	return combineHandlers(handlers), nil
}

func consoleHandler(json bool, opts *slog.HandlerOptions) slog.Handler {
	if json {
		return slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.NewTextHandler(os.Stdout, opts)
}

func fileHandler(path string, opts *slog.HandlerOptions, lc fx.Lifecycle) (slog.Handler, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return file.Close()
		},
	})

	return slog.NewTextHandler(file, opts), nil
}

func combineHandlers(handlers []slog.Handler) slog.Handler {
	if len(handlers) == 1 {
		return handlers[0]
	}

	return &multiHandler{handlers: handlers}
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
