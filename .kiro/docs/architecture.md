# Architecture

## System Design

The Aphex CLI is a Go-based command-line tool that integrates with the Arbiter platform's Kubernetes cluster and authentication system. It uses standard kubectl patterns for configuration discovery and OIDC authentication via exec plugins.

## Components

### CLI Framework
- **urfave/cli v3**: Command structure and argument parsing
- **Survey**: Interactive prompts for missing parameters
- **Cross-platform builds**: Support for Linux, macOS, Windows

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
- **Create**: Deploy Tekton Pipeline resources from YAML files with automatic namespace creation
- **Delete**: Remove pipelines using cross-namespace discovery (no namespace required)
- **List**: Display all pipelines across accessible namespaces
- **RepoBinding Integration**: Automatic webhook provisioning for GitHub integration

### Namespace Management
- **Pipeline-Namespace Mapping**: Each pipeline creates its own namespace (pipeline-name = namespace-name)
- **Automatic Namespace Creation**: CLI creates namespaces automatically during pipeline creation
- **Cross-Namespace Discovery**: Delete and list operations search across all accessible namespaces
- **No Manual Namespace Management**: Users never specify namespaces directly

## Technology Stack

- **Go 1.25**: Primary programming language
- **urfave/cli v3**: CLI framework
- **Kubernetes client-go**: Kubernetes API integration
- **Survey v2**: Interactive prompts
- **gopter**: Property-based testing
- **kubelogin**: OIDC exec plugin (external dependency)

## Architectural Patterns

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
- `go.mod` - Go dependencies
