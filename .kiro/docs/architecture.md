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
