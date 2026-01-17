# Centralized Logging System - Design

## Overview

This design implements a centralized logging system for the Aphex CLI that provides consistent, context-based logging across all commands and packages. The logger encapsulates verbosity complexity and provides a clean API for logging at different levels without requiring commands or packages to check flags.

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Root Command                             │
│  (cmd/aphex/main.go)                                        │
│                                                              │
│  Flags: --log-level, --verbose                             │
│  Before Hook: Creates Logger → Injects into Context        │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   │ Context with Logger
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    Subcommands                               │
│  (internal/commands/*.go)                                   │
│                                                              │
│  logger := GetLogger(ctx)                                   │
│  logger.Debug("message")                                    │
│  logger.Info("message")                                     │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   │ Context with Logger
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    Package Functions                         │
│  (pkg/pipeline/*.go, pkg/organization/*.go)                │
│                                                              │
│  func Create(ctx context.Context, ...) error {             │
│      logger := GetLogger(ctx)                               │
│      logger.Debug("creating resource...")                   │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
User runs command with flags
         │
         ▼
Root Command Before Hook
         │
         ├─ Parse --log-level or --verbose flag
         ├─ Create Logger with level
         └─ Inject Logger into Context
         │
         ▼
Subcommand Action
         │
         ├─ GetLogger(ctx)
         └─ Call logger.Debug/Info/Warn/Error
         │
         ▼
Package Function
         │
         ├─ GetLogger(ctx)
         └─ Call logger.Debug/Info/Warn/Error
         │
         ▼
Logger filters and outputs based on level
```

## Components and Interfaces

### Logger Struct

**Location:** `pkg/logger/logger.go`

```go
package logger

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Level represents the logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging with level filtering
type Logger struct {
	level  Level
	stdout io.Writer
	stderr io.Writer
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	return &Logger{
		level:  parseLevel(level),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Debug logs a debug message (only if level is Debug)
func (l *Logger) Debug(msg string) {
	if l.level <= LevelDebug {
		fmt.Fprintf(l.stdout, "[DEBUG] %s\n", msg)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		fmt.Fprintf(l.stdout, "[DEBUG] "+format+"\n", args...)
	}
}

// Info logs an info message (if level is Debug or Info)
func (l *Logger) Info(msg string) {
	if l.level <= LevelInfo {
		fmt.Fprintf(l.stdout, "[INFO] %s\n", msg)
	}
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		fmt.Fprintf(l.stdout, "[INFO] "+format+"\n", args...)
	}
}

// Warn logs a warning message (if level is Debug, Info, or Warn)
func (l *Logger) Warn(msg string) {
	if l.level <= LevelWarn {
		fmt.Fprintf(l.stderr, "[WARN] %s\n", msg)
	}
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		fmt.Fprintf(l.stderr, "[WARN] "+format+"\n", args...)
	}
}

// Error logs an error message (always logged)
func (l *Logger) Error(msg string) {
	fmt.Fprintf(l.stderr, "[ERROR] %s\n", msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(l.stderr, "[ERROR] "+format+"\n", args...)
}

// parseLevel converts string level to Level constant
func parseLevel(level string) Level {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo // default to info
	}
}
```

### Context Integration

**Location:** `pkg/logger/context.go`

```go
package logger

import "context"

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
	return &Logger{
		level:  LevelError + 1, // Higher than any real level
		stdout: io.Discard,
		stderr: io.Discard,
	}
}
```

### Root Command Integration

**Location:** `cmd/aphex/main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bdchatham/AphexCLI/internal/commands"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/urfave/cli/v3"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
)

func main() {
	app := &cli.Command{
		Name:    "aphex",
		Usage:   "Manage Tekton pipelines on the Aphex platform",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Set log level (debug, info, warn, error)",
				Value:   "info",
				Aliases: []string{"l"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose output (equivalent to --log-level=debug)",
				Aliases: []string{"v"},
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Determine log level
			logLevel := cmd.String("log-level")
			if cmd.Bool("verbose") {
				logLevel = "debug"
			}

			// Create logger and inject into context
			log := logger.NewLogger(logLevel)
			ctx = logger.WithLogger(ctx, log)

			return ctx, nil
		},
		Commands: []*cli.Command{
			commands.AuthCommand(),
			commands.OrganizationCommand(),
			commands.PipelineCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Default action when no subcommands are provided
			return cli.ShowAppHelp(cmd)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

### Command Updates

**Example: `internal/commands/pipeline.go`**

Changes:
1. Remove `--verbose` flag from each subcommand
2. Use `GetLogger(ctx)` instead of checking `cmd.Bool("verbose")`
3. Remove `Verbose: cmd.Bool("verbose")` from option structs

```go
// Before
&cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Verbose output",
}

// After - flag removed entirely

// Before
if cmd.Bool("verbose") {
	fmt.Printf("Creating pipeline %q\n", pipelineName)
}

// After
logger := logger.GetLogger(ctx)
logger.Debugf("Creating pipeline %q", pipelineName)

// Before
opts := pipeline.CreateOptions{
	Name:     args.First(),
	FilePath: cmd.String("file"),
	Verbose:  cmd.Bool("verbose"),
}

// After
opts := pipeline.CreateOptions{
	Name:     args.First(),
	FilePath: cmd.String("file"),
	// Verbose field removed
}
```

### Package Function Updates

**Example: `pkg/pipeline/create.go`**

Changes:
1. Remove `Verbose bool` field from option structs
2. Use `logger := logger.GetLogger(ctx)` at function start
3. Replace `if opts.Verbose { fmt.Printf(...) }` with `logger.Debug(...)`

```go
// Before
type CreateOptions struct {
	Name        string
	FilePath    string
	AphexOrg    string
	RepoOrg     string
	RepoName    string
	TenantName  string
	IngressHost string
	Verbose     bool
}

// After
type CreateOptions struct {
	Name        string
	FilePath    string
	AphexOrg    string
	RepoOrg     string
	RepoName    string
	TenantName  string
	IngressHost string
	// Verbose field removed
}

// Before
func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
	if opts.Verbose {
		fmt.Printf("Creating RepoBinding for pipeline %q\n", opts.Name)
	}
	// ...
}

// After
func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
	logger := logger.GetLogger(ctx)
	logger.Debugf("Creating RepoBinding for pipeline %q", opts.Name)
	// ...
}
```

## Data Models

### Logger

```go
type Logger struct {
	level  Level      // Current log level filter
	stdout io.Writer  // Output for Debug and Info
	stderr io.Writer  // Output for Warn and Error
}
```

### Level

```go
type Level int

const (
	LevelDebug Level = iota  // 0 - Most verbose
	LevelInfo                // 1 - Default
	LevelWarn                // 2 - Warnings only
	LevelError               // 3 - Errors only
)
```

### Context Key

```go
type contextKey string

const loggerKey contextKey = "logger"
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Logger Level Filtering

*For any* logger with level L and any message at level M, the message should be output if and only if M >= L.

**Validates: Requirements US-1.1, US-1.4**

### Property 2: Context Propagation

*For any* context with an injected logger, calling `GetLogger(ctx)` should return the same logger instance that was injected.

**Validates: Requirements US-2.1, US-3.1**

### Property 3: No-Op Logger Fallback

*For any* context without a logger, calling `GetLogger(ctx)` should return a no-op logger that discards all output without error.

**Validates: Requirements US-2.2, US-3.2**

### Property 4: Output Stream Routing

*For any* logger, Debug and Info messages should go to stdout, and Warn and Error messages should go to stderr.

**Validates: Requirements US-4.2**

### Property 5: Verbose Flag Equivalence

*For any* command invocation with `--verbose`, the log level should be set to debug, equivalent to `--log-level=debug`.

**Validates: Requirements US-1.2**

### Property 6: Level Prefix Consistency

*For any* logged message at level L, the output should include the prefix `[L]` where L is DEBUG, INFO, WARN, or ERROR.

**Validates: Requirements US-4.1**

## Error Handling

### Invalid Log Level

When an invalid log level string is provided:
- Default to `info` level
- No error is raised (fail-safe behavior)
- Logger continues to function normally

```go
func parseLevel(level string) Level {
	switch level {
	case "debug", "info", "warn", "error":
		return /* appropriate level */
	default:
		return LevelInfo // Safe default
	}
}
```

### Missing Logger in Context

When `GetLogger(ctx)` is called on a context without a logger:
- Return a no-op logger that discards all output
- No panic or error
- Allows code to function even if logger wasn't injected

```go
func GetLogger(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return NewNoOpLogger() // Fail-safe
}
```

### Context Cancellation

Logger operations are not affected by context cancellation:
- Logger doesn't check `ctx.Done()`
- Logging is synchronous and immediate
- Context is only used for logger storage/retrieval

## Testing Strategy

### Unit Tests

**Logger Level Filtering (`pkg/logger/logger_test.go`):**
- Test each level (debug, info, warn, error) filters correctly
- Test that Debug messages don't appear at Info level
- Test that Info messages don't appear at Warn level
- Test that Warn messages don't appear at Error level
- Test that Error messages always appear

**Context Integration (`pkg/logger/context_test.go`):**
- Test `WithLogger` stores logger in context
- Test `GetLogger` retrieves stored logger
- Test `GetLogger` returns no-op logger when none exists
- Test no-op logger discards all output

**Output Stream Routing (`pkg/logger/logger_test.go`):**
- Test Debug and Info go to stdout
- Test Warn and Error go to stderr
- Use `bytes.Buffer` to capture and verify output

**Level Parsing (`pkg/logger/logger_test.go`):**
- Test valid level strings ("debug", "info", "warn", "error")
- Test invalid level strings default to "info"
- Test case sensitivity

### Property-Based Tests

**Property 1: Level Filtering**
```go
// For any level L and message level M, output occurs iff M >= L
func TestProperty_LevelFiltering(t *testing.T) {
	// Generate random combinations of logger level and message level
	// Verify filtering behavior matches expected rules
}
```
**Feature: centralized-logging, Property 1: Logger Level Filtering**

**Property 2: Context Propagation**
```go
// For any logger injected into context, GetLogger returns same instance
func TestProperty_ContextPropagation(t *testing.T) {
	// Generate random log levels
	// Create logger, inject into context, retrieve
	// Verify same instance returned
}
```
**Feature: centralized-logging, Property 2: Context Propagation**

**Property 3: No-Op Logger Fallback**
```go
// For any context without logger, GetLogger returns no-op logger
func TestProperty_NoOpFallback(t *testing.T) {
	// Generate random contexts without logger
	// Verify GetLogger returns no-op logger
	// Verify no-op logger discards all output
}
```
**Feature: centralized-logging, Property 3: No-Op Logger Fallback**

**Property 4: Output Stream Routing**
```go
// For any message, Debug/Info go to stdout, Warn/Error go to stderr
func TestProperty_OutputStreamRouting(t *testing.T) {
	// Generate random messages at each level
	// Capture stdout and stderr
	// Verify correct routing
}
```
**Feature: centralized-logging, Property 4: Output Stream Routing**

### Integration Tests

**End-to-End Command Test:**
- Run `aphex pipeline list --log-level=debug`
- Verify debug messages appear
- Run `aphex pipeline list --log-level=error`
- Verify debug messages don't appear

**Verbose Flag Test:**
- Run `aphex pipeline list --verbose`
- Verify equivalent to `--log-level=debug`

**Package Function Test:**
- Call package function with context containing logger
- Verify logger is accessible and works correctly

## Implementation Notes

### urfave/cli v3 Context Handling

The urfave/cli v3 library uses `context.Context` for command actions. The `Before` hook can modify the context and return it:

```go
Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	// Modify context
	ctx = logger.WithLogger(ctx, log)
	return ctx, nil
}
```

This modified context is passed to all subcommands and their actions.

### Backward Compatibility

This change maintains backward compatibility:
- Commands still work the same way from user perspective
- Output format remains consistent (just adds level prefixes)
- No breaking changes to command-line interface
- Existing scripts using `--verbose` continue to work

### Migration Strategy

1. Create logger package with full implementation
2. Update root command to inject logger
3. Update commands one at a time (can coexist during migration)
4. Update packages one at a time (can coexist during migration)
5. Remove `Verbose` fields from option structs last

### Performance Considerations

- Logger level checks are simple integer comparisons (fast)
- No reflection or complex logic in hot path
- Output is synchronous (no goroutines or channels)
- Minimal memory overhead (logger is small struct)

## Source References

**Current Implementation:**
- `AphexCLI/cmd/aphex/main.go` - Root command structure
- `AphexCLI/internal/commands/pipeline.go` - Current verbose flag usage
- `AphexCLI/internal/commands/organization.go` - Current verbose flag usage
- `AphexCLI/pkg/pipeline/create.go` - Current verbose parameter passing
- `AphexCLI/pkg/organization/organization.go` - Current verbose parameter passing

**Design Principles:**
- `AphexCLI/CLAUDE.md` - Code quality standards and design principles
- `AphexPipelineInfrastructure/CLAUDE.md` - Clean code principles

**Dependencies:**
- `github.com/urfave/cli/v3` - CLI framework with context support
- Standard library: `context`, `fmt`, `io`, `os`
