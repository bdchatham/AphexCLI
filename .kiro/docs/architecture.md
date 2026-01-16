# Architecture

## System Design

The Aphex CLI is a Go-based command-line tool that integrates with the Arbiter platform's Kubernetes cluster and authentication system. It uses standard kubectl patterns for configuration discovery and OIDC authentication via exec plugins.

## Components

### CLI Framework
- **urfave/cli v3**: Command structure and argument parsing
- **Survey**: Interactive prompts for missing parameters
- **Cross-platform builds**: Support for Linux, macOS, Windows

### Centralized Logging System
- **Context-based logger**: Logger injected into context at root command level
- **Level filtering**: Four log levels (debug, info, warn, error) with automatic filtering
- **Global flags**: `--log-level` and `--verbose` flags control logging for all commands
- **Stream routing**: Debug/Info to stdout, Warn/Error to stderr
- **No-op fallback**: Safe fallback logger when context doesn't contain logger
- **Package integration**: All commands and packages use `logger.GetLogger(ctx)` for consistent logging

### Authentication System
- **OIDC Integration**: Uses kubelogin exec plugin for browser-based authentication
- **Dex Integration**: Points to platform Dex endpoint (https://dex.home.local)
- **Standard kubeconfig**: Follows kubectl discovery patterns

### Kubernetes Integration
- **client-go**: Standard Kubernetes client library
- **Dynamic client**: Proper dynamic client for Tekton Pipeline and RepoBinding resource management
- **RBAC-aware**: Respects platform namespace isolation and permissions
- **Cross-namespace operations**: Searches across namespaces for pipeline discovery
- **Automatic resource provisioning**: Creates namespaces and RepoBindings automatically
- **Timeout handling**: 10-second timeout on authorization checks to prevent hanging
- **Error resilience**: Graceful handling of network issues and API failures

### Pipeline Management
- **Create**: Creates a RepoBinding CRD with embedded pipeline YAML
  - Requires `--aphex-org`, `--repo-org`, and `--repo-name` flags
  - Reads pipeline YAML from file and embeds it in RepoBinding spec
  - Controller provisions all resources (namespace, Pipeline, RBAC, Triggers)
  - CLI is thin and declarative - no direct resource creation
- **Delete**: Remove pipelines using cross-namespace discovery (no namespace required)
- **List**: Display all pipelines across accessible namespaces
- **RepoBinding Integration**: Automatic webhook provisioning for GitHub integration

### Kubernetes Operator Pattern
- **Declarative CLI**: CLI only creates RepoBinding CRDs, doesn't provision resources directly
- **Controller Reconciliation**: RepoBinding controller in ArbiterPipelineInfrastructure handles all provisioning
- **Separation of Concerns**: CLI handles user interaction, controller handles infrastructure
- **Kubernetes-Driven**: All resource management through Kubernetes reconciliation loops

### Namespace Management
- **Pipeline-Namespace Mapping**: Each pipeline has its own namespace (pipeline-name = namespace-name)
- **Controller-Managed**: RepoBinding controller creates namespaces automatically
- **Cross-Namespace Discovery**: Delete and list operations search across all accessible namespaces
- **No Manual Namespace Management**: Users never specify namespaces directly

### Organization Management
- **Bootstrap**: Create new organizations with complete multi-tenant infrastructure
  - Creates Organization CRD in platform-system namespace
  - Provisions organization namespace (org-{name})
  - Sets up webhook infrastructure and RBAC
  - Requires platform-admin permissions
- **List**: Display all organizations and their status
- **Delete**: Remove organizations and all associated resources
  - Deletes Organization CRD
  - Removes organization namespace
  - Requires platform-admin permissions

## Technology Stack

- **Go 1.25**: Primary programming language
- **urfave/cli v3**: CLI framework
- **Kubernetes client-go**: Kubernetes API integration
- **Survey v2**: Interactive prompts
- **gopter**: Property-based testing
- **kubelogin**: OIDC exec plugin (external dependency)

## Architectural Patterns

### Centralized Logging Architecture

The CLI implements a centralized logging system that encapsulates verbosity complexity and provides consistent logging across all commands and packages.

#### Logger Package Structure

**Location**: `pkg/logger/`

The logger package provides two main components:

1. **Logger Struct** (`logger.go`):
   - Encapsulates log level filtering logic
   - Provides methods: `Debug()`, `Info()`, `Warn()`, `Error()` and formatted variants
   - Routes output to appropriate streams (stdout for Debug/Info, stderr for Warn/Error)
   - Performs level filtering automatically (e.g., Debug messages don't appear at Info level)

2. **Context Integration** (`context.go`):
   - `WithLogger(ctx, logger)`: Injects logger into context
   - `GetLogger(ctx)`: Retrieves logger from context
   - `NewNoOpLogger()`: Provides fail-safe fallback logger that discards output

#### Log Levels and Filtering

The logger supports four levels with hierarchical filtering:

| Level | Debug | Info | Warn | Error |
|-------|-------|------|------|-------|
| debug | ✓     | ✓    | ✓    | ✓     |
| info  | ✗     | ✓    | ✓    | ✓     |
| warn  | ✗     | ✗    | ✓    | ✓     |
| error | ✗     | ✗    | ✗    | ✓     |

Messages are automatically filtered based on the configured level. For example, if the logger is set to `info` level, `Debug()` calls produce no output.

#### Context Propagation Flow

```
Root Command (main.go)
  ├─ Parse --log-level or --verbose flag
  ├─ Create Logger with specified level
  └─ Inject Logger into Context via Before hook
      │
      ▼
Subcommands (internal/commands/*.go)
  ├─ Receive context with logger
  ├─ Call logger := logger.GetLogger(ctx)
  └─ Use logger.Debug/Info/Warn/Error methods
      │
      ▼
Package Functions (pkg/*/*.go)
  ├─ Receive context as first parameter
  ├─ Call logger := logger.GetLogger(ctx)
  └─ Use logger methods for all logging
```

#### Design Benefits

1. **Separation of Concerns**: Commands and packages focus on business logic, not logging mechanics
2. **Single Configuration Point**: Log level set once at root command, applies everywhere
3. **Type-Safe Context Keys**: Uses typed context keys to prevent collisions
4. **Fail-Safe Design**: Returns no-op logger if context doesn't contain logger (prevents panics)
5. **Clean API**: Simple method calls (`logger.Debug("msg")`) without flag checking
6. **Consistent Formatting**: All log messages include level prefix ([DEBUG], [INFO], [WARN], [ERROR])

#### Migration from Individual Verbose Flags

The centralized logging system replaces the previous pattern of individual `--verbose` flags on each command:

**Before**:
```go
// Each command had its own --verbose flag
&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}}

// Commands checked the flag
if cmd.Bool("verbose") {
    fmt.Printf("Debug message\n")
}

// Packages received verbose as parameter
type CreateOptions struct {
    Verbose bool
}
```

**After**:
```go
// Global flags on root command only
&cli.StringFlag{Name: "log-level", Value: "info"}
&cli.BoolFlag{Name: "verbose"} // Shorthand for --log-level=debug

// Commands use logger from context
logger := logger.GetLogger(ctx)
logger.Debug("Debug message")

// Packages use logger from context
type CreateOptions struct {
    // No Verbose field
}
```

This eliminates scattered verbosity logic and provides consistent behavior across the entire CLI.

**Source**
- `pkg/logger/logger.go` - Logger implementation with level filtering and stream routing
- `pkg/logger/context.go` - Context integration and no-op logger fallback
- `cmd/aphex/main.go` - Root command logger injection via Before hook

### Standard kubectl Patterns
- Uses standard kubeconfig discovery (--kubeconfig flag → KUBECONFIG env → ~/.kube/config)
- Respects current context namespace from kubeconfig
- Delegates token refresh to exec plugin

### Command Pattern
- Each CLI command implemented as separate action function
- Shared options structures for consistent parameter handling
- BeforeAction hooks for preflight authorization checks

### Error Translation
- Kubernetes API errors translated to user-friendly messages
- Permission errors include required groups and remediation steps
- Network errors provide troubleshooting guidance

## Dependencies

### Upstream Dependencies
- **Arbiter Platform**: Kubernetes cluster with Tekton and Dex
- **kubelogin**: OIDC exec plugin for authentication
- **Dex**: OIDC authentication endpoint (https://dex.home.local)

### Downstream Dependencies
- **Tekton Pipelines**: Creates and manages Pipeline resources
- **Kubernetes RBAC**: Enforces namespace isolation and permissions

**Source**
- `cmd/aphex/main.go` - CLI application structure
- `internal/commands/` - Command implementations
- `pkg/k8s/client.go` - Kubernetes client wrapper
- `pkg/auth/login.go` - OIDC authentication setup
- `pkg/organization/` - Organization management
- `go.mod` - Go dependencies
