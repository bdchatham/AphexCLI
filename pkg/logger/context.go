package logger

import (
	"context"
	"io"
	"log/slog"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves the logger from context
// Returns a no-op logger if none exists (fail-safe)
func GetLogger(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	// Return no-op logger as fallback
	return NewNoOpLogger()
}

// NewNoOpLogger creates a logger that discards all output
func NewNoOpLogger() *Logger {
	// Create a handler that discards everything
	opts := &slog.HandlerOptions{
		Level: slog.LevelError + 1, // Higher than any real level
	}
	handler := slog.NewTextHandler(io.Discard, opts)
	
	return &Logger{
		slog: slog.New(handler),
	}
}
