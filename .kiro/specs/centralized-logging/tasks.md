# Implementation Plan: Centralized Logging System

## Overview

Implement a centralized logging system for the Aphex CLI that provides context-based logging with level filtering. The implementation follows a phased approach: create the logger package, integrate with the root command, then update commands and packages to use the centralized logger.

## Tasks

- [x] 1. Create logger package with core functionality
  - [x] 1.1 Create `pkg/logger/logger.go` with Logger struct and methods
    - Implement `Level` type and constants (Debug, Info, Warn, Error)
    - Implement `Logger` struct with level, stdout, stderr fields
    - Implement `NewLogger(level string) *Logger` constructor
    - Implement logging methods: `Debug()`, `Info()`, `Warn()`, `Error()`
    - Implement formatted logging methods: `Debugf()`, `Infof()`, `Warnf()`, `Errorf()`
    - Implement `parseLevel(level string) Level` helper
    - Implement level filtering logic in each method
    - Route Debug/Info to stdout, Warn/Error to stderr
    - _Requirements: TR-1, US-4.1, US-4.2_

  - [ ]* 1.2 Write unit tests for Logger struct
    - Test level filtering for each level (debug, info, warn, error)
    - Test that messages below current level are filtered
    - Test output stream routing (stdout vs stderr)
    - Test formatted logging methods
    - Test parseLevel with valid and invalid inputs
    - _Requirements: TR-1, US-4.1, US-4.2_

  - [x] 1.3 Create `pkg/logger/context.go` with context integration
    - Implement `contextKey` type for type-safe context keys
    - Implement `WithLogger(ctx context.Context, logger *Logger) context.Context`
    - Implement `GetLogger(ctx context.Context) *Logger`
    - Implement `NewNoOpLogger() *Logger` for fallback
    - _Requirements: TR-2, US-2.1, US-3.1_

  - [ ]* 1.4 Write unit tests for context integration
    - Test `WithLogger` stores logger in context
    - Test `GetLogger` retrieves stored logger
    - Test `GetLogger` returns no-op logger when none exists
    - Test no-op logger discards all output
    - _Requirements: TR-2, US-2.1, US-3.1_

  - [ ]* 1.5 Write property test for level filtering
    - **Property 1: Logger Level Filtering**
    - **Validates: Requirements US-1.1, US-1.4**
    - Generate random combinations of logger level and message level
    - Verify message is output if and only if message level >= logger level
    - _Requirements: US-1.1, US-1.4_

  - [ ]* 1.6 Write property test for context propagation
    - **Property 2: Context Propagation**
    - **Validates: Requirements US-2.1, US-3.1**
    - Generate random log levels
    - Create logger, inject into context, retrieve
    - Verify same instance is returned
    - _Requirements: US-2.1, US-3.1_

  - [ ]* 1.7 Write property test for no-op fallback
    - **Property 3: No-Op Logger Fallback**
    - **Validates: Requirements US-2.2, US-3.2**
    - Generate random contexts without logger
    - Verify GetLogger returns no-op logger
    - Verify no-op logger discards all output
    - _Requirements: US-2.2, US-3.2_

  - [ ]* 1.8 Write property test for output stream routing
    - **Property 4: Output Stream Routing**
    - **Validates: Requirements US-4.2**
    - Generate random messages at each level
    - Capture stdout and stderr
    - Verify Debug/Info go to stdout, Warn/Error go to stderr
    - _Requirements: US-4.2_

- [x] 2. Integrate logger with root command
  - [x] 2.1 Update `cmd/aphex/main.go` to add global flags and Before hook
    - Add `--log-level` flag with values: debug, info, warn, error (default: info)
    - Add `--verbose` flag as shorthand for `--log-level=debug`
    - Add `Before` hook to root command
    - In Before hook: parse log level from flags (verbose takes precedence)
    - In Before hook: create logger with `logger.NewLogger(logLevel)`
    - In Before hook: inject logger into context with `logger.WithLogger(ctx, log)`
    - Return modified context from Before hook
    - _Requirements: TR-3, US-1.1, US-1.2, US-1.3_

  - [ ]* 2.2 Write integration test for root command logger injection
    - Test that logger is available in subcommand context
    - Test `--log-level` flag sets correct level
    - Test `--verbose` flag sets debug level
    - Test default level is info
    - _Requirements: TR-3, US-1.1, US-1.2, US-1.3_

  - [ ]* 2.3 Write property test for verbose flag equivalence
    - **Property 5: Verbose Flag Equivalence**
    - **Validates: Requirements US-1.2**
    - For any command with `--verbose`, verify log level is debug
    - Verify equivalent to `--log-level=debug`
    - _Requirements: US-1.2_

- [x] 3. Checkpoint - Verify logger package works end-to-end
  - Ensure all tests pass
  - Manually test root command with different log levels
  - Verify logger is accessible in subcommands
  - Ask the user if questions arise

- [x] 4. Update pipeline commands to use centralized logger
  - [x] 4.1 Update `internal/commands/pipeline.go`
    - Remove `--verbose` flag from `create` subcommand
    - Remove `--verbose` flag from `delete` subcommand
    - Remove `--verbose` flag from `list` subcommand
    - In `pipelineCreateAction`: use `logger := logger.GetLogger(ctx)`
    - In `pipelineDeleteAction`: use `logger := logger.GetLogger(ctx)`
    - In `pipelineListAction`: use `logger := logger.GetLogger(ctx)`
    - Remove `Verbose: cmd.Bool("verbose")` from CreateOptions
    - Remove `Verbose: cmd.Bool("verbose")` from DeleteOptions
    - Remove `Verbose: cmd.Bool("verbose")` from ListOptions
    - _Requirements: TR-4, US-2.1_

  - [x] 4.2 Update `pkg/pipeline/create.go`
    - Remove `Verbose bool` field from `CreateOptions` struct
    - Add `logger := logger.GetLogger(ctx)` at start of `Create()` function
    - Replace `if opts.Verbose { fmt.Printf(...) }` with `logger.Debugf(...)`
    - Replace any other verbose checks with appropriate logger calls
    - _Requirements: TR-5, US-3.1_

  - [x] 4.3 Update `pkg/pipeline/delete.go`
    - Remove `Verbose bool` field from `DeleteOptions` struct
    - Add `logger := logger.GetLogger(ctx)` at start of `Delete()` function
    - Replace verbose checks with logger calls
    - _Requirements: TR-5, US-3.1_

  - [x] 4.4 Update `pkg/pipeline/list.go`
    - Remove `Verbose bool` field from `ListOptions` struct
    - Add `logger := logger.GetLogger(ctx)` at start of `List()` function
    - Replace verbose checks with logger calls
    - _Requirements: TR-5, US-3.1_

- [x] 5. Update organization commands to use centralized logger
  - [x] 5.1 Update `internal/commands/organization.go`
    - Remove `--verbose` flag from `bootstrap` subcommand
    - Remove `--verbose` flag from `list` subcommand
    - Remove `--verbose` flag from `delete` subcommand
    - In `organizationBootstrapAction`: use `logger := logger.GetLogger(ctx)`
    - In `organizationListAction`: use `logger := logger.GetLogger(ctx)`
    - In `organizationDeleteAction`: use `logger := logger.GetLogger(ctx)`
    - Remove `Verbose: cmd.Bool("verbose")` from BootstrapOptions
    - Remove `Verbose: cmd.Bool("verbose")` from ListOptions
    - Remove `Verbose: cmd.Bool("verbose")` from DeleteOptions
    - _Requirements: TR-4, US-2.1_

  - [x] 5.2 Update `pkg/organization/organization.go`
    - Remove `Verbose bool` field from `BootstrapOptions` struct
    - Remove `Verbose bool` field from `ListOptions` struct
    - Remove `Verbose bool` field from `DeleteOptions` struct
    - Add `logger := logger.GetLogger(ctx)` at start of `Bootstrap()` function
    - Add `logger := logger.GetLogger(ctx)` at start of `List()` function
    - Add `logger := logger.GetLogger(ctx)` at start of `Delete()` function
    - Replace `if opts.Verbose { fmt.Printf(...) }` with `logger.Debugf(...)`
    - Replace any other verbose checks with appropriate logger calls
    - _Requirements: TR-5, US-3.1_

- [x] 6. Update auth commands to use centralized logger (if exists)
  - [x] 6.1 Check if `internal/commands/auth.go` exists
    - If exists, remove `--verbose` flags
    - If exists, use `logger := logger.GetLogger(ctx)`
    - If exists, update corresponding package functions
    - _Requirements: TR-4, TR-5_

- [x] 7. Checkpoint - Verify all commands work with centralized logger
  - Ensure all tests pass
  - Manually test each command with different log levels
  - Verify no `--verbose` flags remain on subcommands
  - Verify global `--log-level` and `--verbose` flags work
  - Ask the user if questions arise

- [x] 8. Update documentation
  - [x] 8.1 Update `AphexCLI/README.md`
    - Document global `--log-level` flag
    - Document global `--verbose` flag
    - Update examples to show log level usage
    - Remove references to individual `--verbose` flags
    - _Requirements: US-1.1, US-1.2_

  - [x] 8.2 Update `AphexCLI/.kiro/docs/architecture.md`
    - Add section describing centralized logging architecture
    - Document logger package and context integration
    - Document log levels and filtering behavior
    - _Requirements: TR-1, TR-2_

  - [x] 8.3 Update `AphexCLI/.kiro/docs/api.md`
    - Document logger API for developers
    - Document `GetLogger(ctx)` usage pattern
    - Document logging methods and levels
    - _Requirements: US-2.1, US-3.1_

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Implementation can proceed incrementally (logger package → root command → commands → packages)
- Commands and packages can be updated one at a time (migration-friendly)
