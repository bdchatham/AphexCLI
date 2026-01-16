# Centralized Logging System - Requirements

## Overview

Create a centralized logging system for the Aphex CLI that encapsulates verbosity complexity and provides consistent logging across all commands and modules.

## Problem Statement

Currently, the Aphex CLI handles logging in an inconsistent manner:
- Each command defines its own `--verbose` flag
- Commands pass `verbose` boolean to package functions
- Package functions must check the verbose flag and decide what to log
- No centralized control over log levels or output formatting
- Verbosity logic is scattered across command and package layers

This creates maintenance burden and inconsistent logging behavior across the CLI.

## Goals

1. **Centralized Configuration**: Configure logging once at the root command level
2. **Context-Based Access**: Make logger available via context throughout the application
3. **Encapsulated Complexity**: Hide verbosity and log level logic inside the logger
4. **Simple API**: Commands and packages just call logger methods without checking flags
5. **Consistent Behavior**: All logging follows the same patterns and formatting

## Non-Goals

- Structured logging (JSON output) - keep simple text-based logging
- Log file persistence - CLI logs to stdout/stderr only
- Log rotation or archiving - not applicable for CLI tool
- Remote logging or telemetry - local output only

## User Stories

### US-1: Platform Engineer Configures Log Level
**As a** platform engineer  
**I want to** set a global log level when running any Aphex command  
**So that** I can control verbosity without modifying individual command flags

**Acceptance Criteria:**
- Root command accepts `--log-level` flag with values: `debug`, `info`, `warn`, `error`
- Root command accepts `--verbose` flag as shorthand for `--log-level=debug`
- Log level is configured once and applies to all subcommands
- Default log level is `info` when no flags are provided

### US-2: Command Uses Logger Without Checking Flags
**As a** CLI developer  
**I want to** use a logger from context without checking verbosity flags  
**So that** I can write cleaner code focused on business logic

**Acceptance Criteria:**
- Commands can call `logger := GetLogger(ctx)` to retrieve logger
- Logger provides methods: `Debug()`, `Info()`, `Warn()`, `Error()`
- Logger automatically filters messages based on configured log level
- Commands don't need to check `--verbose` or `--log-level` flags

### US-3: Package Functions Use Logger From Context
**As a** CLI developer  
**I want to** access the logger in package functions via context  
**So that** package code doesn't need to know about CLI flags

**Acceptance Criteria:**
- Package functions receive `context.Context` as first parameter
- Package functions call `logger := GetLogger(ctx)` to retrieve logger
- Logger works the same in packages as in commands
- No need to pass `verbose bool` parameters to package functions

### US-4: Logger Provides Consistent Output Formatting
**As a** platform engineer  
**I want** all log messages to follow consistent formatting  
**So that** I can easily parse and understand CLI output

**Acceptance Criteria:**
- Log messages include level prefix: `[DEBUG]`, `[INFO]`, `[WARN]`, `[ERROR]`
- Debug and Info messages go to stdout
- Warn and Error messages go to stderr
- Messages are formatted consistently across all commands

## Technical Requirements

### TR-1: Logger Struct
Create a `Logger` struct in `pkg/logger/logger.go` with:
- Log level field (debug, info, warn, error)
- Methods: `Debug(msg string)`, `Info(msg string)`, `Warn(msg string)`, `Error(msg string)`
- Methods: `Debugf(format string, args ...interface{})`, `Infof(format string, args ...interface{})`, etc.
- Internal logic to filter messages based on log level

### TR-2: Context Integration
Implement context-based logger storage and retrieval:
- `NewLogger(level string) *Logger` - creates logger with specified level
- `WithLogger(ctx context.Context, logger *Logger) context.Context` - adds logger to context
- `GetLogger(ctx context.Context) *Logger` - retrieves logger from context
- `GetLogger` returns a no-op logger if none exists in context (fail-safe)

### TR-3: Root Command Integration
Modify `cmd/aphex/main.go` to:
- Add global `--log-level` flag (values: debug, info, warn, error)
- Add global `--verbose` flag (sets log-level to debug)
- Add `Before` hook to root command that:
  - Reads log level from flags
  - Creates logger with configured level
  - Adds logger to context
  - Passes context to subcommands

### TR-4: Remove Individual Verbose Flags
Update all command files to:
- Remove `--verbose` flags from individual commands
- Remove `verbose bool` parameters from package function calls
- Replace `if verbose { fmt.Println(...) }` with `logger.Debug(...)`
- Use `GetLogger(ctx)` to retrieve logger

### TR-5: Package Function Updates
Update all package functions to:
- Accept `context.Context` as first parameter (if not already)
- Use `logger := GetLogger(ctx)` to retrieve logger
- Replace `if opts.Verbose { fmt.Println(...) }` with `logger.Debug(...)`
- Remove `Verbose bool` fields from option structs

## Log Level Behavior

| Level | Debug | Info | Warn | Error |
|-------|-------|------|------|-------|
| debug | ✓     | ✓    | ✓    | ✓     |
| info  | ✗     | ✓    | ✓    | ✓     |
| warn  | ✗     | ✗    | ✓    | ✓     |
| error | ✗     | ✗    | ✗    | ✓     |

## Implementation Phases

### Phase 1: Create Logger Package
- Create `pkg/logger/logger.go`
- Implement `Logger` struct with methods
- Implement context integration functions
- Write unit tests for logger behavior

### Phase 2: Integrate with Root Command
- Add global flags to root command
- Add `Before` hook to create and inject logger
- Test that logger is available in subcommands

### Phase 3: Update Commands
- Update `internal/commands/pipeline.go`
- Update `internal/commands/organization.go`
- Update `internal/commands/auth.go` (if exists)
- Remove individual `--verbose` flags

### Phase 4: Update Packages
- Update `pkg/pipeline/create.go`
- Update `pkg/pipeline/delete.go`
- Update `pkg/pipeline/list.go`
- Update `pkg/organization/bootstrap.go`
- Update `pkg/organization/delete.go`
- Update `pkg/organization/list.go`
- Remove `Verbose` fields from option structs

### Phase 5: Testing and Validation
- Test all commands with different log levels
- Verify output goes to correct streams (stdout/stderr)
- Verify backward compatibility (commands still work)
- Update documentation

## Design Principles Alignment

This design aligns with Aphex CLI design principles:

**Separation of Concerns:**
- Logging logic is centralized in `pkg/logger`
- Commands focus on business logic, not logging mechanics
- Packages don't need to know about CLI flags

**Clean Code:**
- Expressive method names: `logger.Debug()`, `logger.Info()`
- Single Responsibility: Logger handles all logging concerns
- DRY: No repeated verbosity checks across codebase

**Context-Based Architecture:**
- Follows Go idioms for passing cross-cutting concerns via context
- Logger is available anywhere context is available
- No global state or singleton patterns

## Success Criteria

1. All commands use centralized logger via context
2. No individual `--verbose` flags remain on subcommands
3. Global `--log-level` and `--verbose` flags control all logging
4. Package functions use `GetLogger(ctx)` instead of checking verbose flags
5. All log output is consistently formatted
6. Backward compatibility maintained (commands still work as before)

## Source References

- `AphexCLI/cmd/aphex/main.go` - Root command structure
- `AphexCLI/internal/commands/pipeline.go` - Current verbose flag usage
- `AphexCLI/internal/commands/organization.go` - Current verbose flag usage
- `AphexCLI/pkg/pipeline/create.go` - Current verbose parameter passing
- `AphexCLI/CLAUDE.md` - Design principles and code quality standards
