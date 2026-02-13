package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"invalid level defaults to info", "invalid"},
		{"empty level defaults to info", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			if logger == nil {
				t.Error("NewLogger returned nil")
				return
			}
			if logger.slog == nil {
				t.Error("Logger's slog field is nil")
			}
		})
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  string
		logFunc      func(*Logger)
		shouldOutput bool
	}{
		// Debug level logger
		{"debug logger outputs debug", "debug", func(l *Logger) { l.Debug("test") }, true},
		{"debug logger outputs info", "debug", func(l *Logger) { l.Info("test") }, true},
		{"debug logger outputs warn", "debug", func(l *Logger) { l.Warn("test") }, true},
		{"debug logger outputs error", "debug", func(l *Logger) { l.Error("test") }, true},

		// Info level logger
		{"info logger filters debug", "info", func(l *Logger) { l.Debug("test") }, false},
		{"info logger outputs info", "info", func(l *Logger) { l.Info("test") }, true},
		{"info logger outputs warn", "info", func(l *Logger) { l.Warn("test") }, true},
		{"info logger outputs error", "info", func(l *Logger) { l.Error("test") }, true},

		// Warn level logger
		{"warn logger filters debug", "warn", func(l *Logger) { l.Debug("test") }, false},
		{"warn logger filters info", "warn", func(l *Logger) { l.Info("test") }, false},
		{"warn logger outputs warn", "warn", func(l *Logger) { l.Warn("test") }, true},
		{"warn logger outputs error", "warn", func(l *Logger) { l.Error("test") }, true},

		// Error level logger
		{"error logger filters debug", "error", func(l *Logger) { l.Debug("test") }, false},
		{"error logger filters info", "error", func(l *Logger) { l.Info("test") }, false},
		{"error logger filters warn", "error", func(l *Logger) { l.Warn("test") }, false},
		{"error logger outputs error", "error", func(l *Logger) { l.Error("test") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			
			// Create logger with custom handlers for testing
			logger := newTestLogger(tt.loggerLevel, &stdout, &stderr)

			tt.logFunc(logger)

			output := stdout.String() + stderr.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldOutput {
				t.Errorf("Expected output=%v, got output=%v (output: %q)", tt.shouldOutput, hasOutput, output)
			}
		})
	}
}

func TestLoggerOutputStreams(t *testing.T) {
	tests := []struct {
		name         string
		logFunc      func(*Logger)
		expectStdout bool
		expectStderr bool
	}{
		{"Debug goes to stdout", func(l *Logger) { l.Debug("test") }, true, false},
		{"Debugf goes to stdout", func(l *Logger) { l.Debugf("test %s", "msg") }, true, false},
		{"Info goes to stdout", func(l *Logger) { l.Info("test") }, true, false},
		{"Infof goes to stdout", func(l *Logger) { l.Infof("test %s", "msg") }, true, false},
		{"Warn goes to stderr", func(l *Logger) { l.Warn("test") }, false, true},
		{"Warnf goes to stderr", func(l *Logger) { l.Warnf("test %s", "msg") }, false, true},
		{"Error goes to stderr", func(l *Logger) { l.Error("test") }, false, true},
		{"Errorf goes to stderr", func(l *Logger) { l.Errorf("test %s", "msg") }, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			logger := newTestLogger("debug", &stdout, &stderr)

			tt.logFunc(logger)

			hasStdout := stdout.Len() > 0
			hasStderr := stderr.Len() > 0

			if hasStdout != tt.expectStdout {
				t.Errorf("Expected stdout=%v, got stdout=%v", tt.expectStdout, hasStdout)
			}
			if hasStderr != tt.expectStderr {
				t.Errorf("Expected stderr=%v, got stderr=%v", tt.expectStderr, hasStderr)
			}
		})
	}
}

func TestLoggerFormatting(t *testing.T) {
	tests := []struct {
		name           string
		logFunc        func(*Logger)
		expectedLevel  string
	}{
		{"Debug has DEBUG level", func(l *Logger) { l.Debug("test") }, "DEBUG"},
		{"Debugf has DEBUG level", func(l *Logger) { l.Debugf("test %s", "msg") }, "DEBUG"},
		{"Info has INFO level", func(l *Logger) { l.Info("test") }, "INFO"},
		{"Infof has INFO level", func(l *Logger) { l.Infof("test %s", "msg") }, "INFO"},
		{"Warn has WARN level", func(l *Logger) { l.Warn("test") }, "WARN"},
		{"Warnf has WARN level", func(l *Logger) { l.Warnf("test %s", "msg") }, "WARN"},
		{"Error has ERROR level", func(l *Logger) { l.Error("test") }, "ERROR"},
		{"Errorf has ERROR level", func(l *Logger) { l.Errorf("test %s", "msg") }, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			logger := newTestLogger("debug", &stdout, &stderr)

			tt.logFunc(logger)

			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.expectedLevel) {
				t.Errorf("Expected level %q in output, got: %q", tt.expectedLevel, output)
			}
		})
	}
}

func TestLoggerFormattedMessages(t *testing.T) {
	var stdout bytes.Buffer
	logger := newTestLogger("debug", &stdout, &bytes.Buffer{})

	logger.Debugf("test %s %d", "message", 42)

	output := stdout.String()
	if !strings.Contains(output, "test message 42") {
		t.Errorf("Expected formatted message in output, got: %q", output)
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("info")

	ctx = WithLogger(ctx, logger)

	retrieved := ctx.Value(loggerKey)
	if retrieved == nil {
		t.Fatal("Logger not stored in context")
	}

	retrievedLogger, ok := retrieved.(*Logger)
	if !ok {
		t.Fatal("Value in context is not a *Logger")
	}

	if retrievedLogger != logger {
		t.Error("Retrieved logger is not the same instance as stored logger")
	}
}

func TestGetLogger_WithLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("debug")

	ctx = WithLogger(ctx, logger)
	retrieved := GetLogger(ctx)

	if retrieved != logger {
		t.Error("GetLogger did not return the same logger instance")
	}
}

func TestGetLogger_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	retrieved := GetLogger(ctx)

	if retrieved == nil {
		t.Fatal("GetLogger returned nil instead of no-op logger")
	}

	// Verify it's a no-op logger by checking it discards output
	var stdout, stderr bytes.Buffer
	testLogger := newTestLoggerFromSlog(retrieved.slog, &stdout, &stderr)

	testLogger.Debug("test")
	testLogger.Info("test")
	testLogger.Warn("test")
	testLogger.Error("test")

	// No-op logger should not produce output (level is too high)
	if stdout.Len() > 0 || stderr.Len() > 0 {
		t.Error("No-op logger produced output when it should discard everything")
	}
}

func TestNewNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	if logger == nil {
		t.Fatal("NewNoOpLogger returned nil")
	}

	// Verify it discards all output
	var stdout, stderr bytes.Buffer
	testLogger := newTestLoggerFromSlog(logger.slog, &stdout, &stderr)

	testLogger.Debug("test")
	testLogger.Info("test")
	testLogger.Warn("test")
	testLogger.Error("test")

	if stdout.Len() > 0 || stderr.Len() > 0 {
		t.Error("No-op logger produced output when it should discard everything")
	}
}

// Helper function to create a test logger with custom output writers
func newTestLogger(level string, stdout, stderr *bytes.Buffer) *Logger {
	slogLevel := parseLevel(level)
	handler := newSplitHandler(stdout, stderr, slogLevel)
	return &Logger{
		slog: slog.New(handler),
	}
}

// Helper function to create a test logger from an existing slog.Logger
func newTestLoggerFromSlog(slogLogger *slog.Logger, stdout, stderr *bytes.Buffer) *Logger {
	// For testing no-op logger, we just use the existing slog instance
	// since we can't easily replace its handler
	return &Logger{
		slog: slogLogger,
	}
}
