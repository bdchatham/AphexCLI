package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger to provide a consistent logging interface
type Logger struct {
	slog *slog.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	slogLevel := parseLevel(level)
	
	// Create a custom handler that routes Debug/Info to stdout and Warn/Error to stderr
	handler := newSplitHandler(os.Stdout, os.Stderr, slogLevel)
	
	return &Logger{
		slog: slog.New(handler),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.slog.Debug(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.slog.Debug(fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.slog.Info(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.slog.Info(fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.slog.Warn(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.slog.Warn(fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.slog.Error(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.slog.Error(fmt.Sprintf(format, args...))
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to info
	}
}

// splitHandler routes Debug/Info to stdout and Warn/Error to stderr
type splitHandler struct {
	stdoutHandler slog.Handler
	stderrHandler slog.Handler
	level         slog.Level
}

func newSplitHandler(stdout, stderr io.Writer, level slog.Level) *splitHandler {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time attribute for cleaner CLI output
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	
	return &splitHandler{
		stdoutHandler: slog.NewTextHandler(stdout, opts),
		stderrHandler: slog.NewTextHandler(stderr, opts),
		level:         level,
	}
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *splitHandler) Handle(ctx context.Context, record slog.Record) error {
	// Route Debug and Info to stdout, Warn and Error to stderr
	if record.Level < slog.LevelWarn {
		return h.stdoutHandler.Handle(ctx, record)
	}
	return h.stderrHandler.Handle(ctx, record)
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{
		stdoutHandler: h.stdoutHandler.WithAttrs(attrs),
		stderrHandler: h.stderrHandler.WithAttrs(attrs),
		level:         h.level,
	}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{
		stdoutHandler: h.stdoutHandler.WithGroup(name),
		stderrHandler: h.stderrHandler.WithGroup(name),
		level:         h.level,
	}
}
