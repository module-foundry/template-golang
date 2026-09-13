// Package logger provides the project-wide structured logger built on log/slog.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type ctxKey int

const loggerKey ctxKey = iota

// Config describes logger construction options.
type Config struct {
	Level   string // debug|info|warn|error
	Format  string // auto|json|text
	Env     string // development|staging|production
	Service string
	Version string
	Output  io.Writer
}

// New builds a slog.Logger with project defaults (text in dev, JSON otherwise).
func New(cfg Config) *slog.Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	format := cfg.Format
	if format == "" || format == "auto" {
		if cfg.Env == "development" {
			format = "text"
		} else {
			format = "json"
		}
	}

	opts := &slog.HandlerOptions{Level: ParseLevel(cfg.Level)}
	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(out, opts)
	} else {
		handler = slog.NewJSONHandler(out, opts)
	}

	log := slog.New(handler).With(
		slog.String("service", cfg.Service),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
	slog.SetDefault(log)
	return log
}

// ParseLevel converts a string level into slog.Level.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext stores a request-scoped logger in the context.
func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromCtx returns the request-scoped logger or a safe default.
func FromCtx(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if log, ok := ctx.Value(loggerKey).(*slog.Logger); ok && log != nil {
		return log
	}
	return slog.Default()
}

// Reporter is the seam for external error reporters (Sentry/OTel). Default noop.
type Reporter interface {
	Report(ctx context.Context, err error)
}

type noopReporter struct{}

func (noopReporter) Report(context.Context, error) {}

// NoopReporter returns a reporter that does nothing.
func NoopReporter() Reporter { return noopReporter{} }
