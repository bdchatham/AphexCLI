# API

## Overview

The Aphex CLI provides a command-line interface for managing Tekton pipelines on the Aphex platform. This document describes the CLI commands, options, integration patterns, and the logger API for developers.

## Global Flags

All commands support the following global flags for logging control:

### `--log-level <level>` or `-l <level>`
Set the log level for all commands.

**Values**: `debug`, `info`, `warn`, `error`  
**Default**: `info`

Controls the verbosity of output across all commands and packages. Messages at or above the specified level are displayed.

**Examples**:
```bash
aphex pipeline list --log-level=debug
aphex organization bootstrap my-org --admin-email admin@example.com -l warn
```

### `--verbose` or `-v`
Enable verbose output (equivalent to `--log-level=debug`).

Shorthand flag for maximum verbosity. When specified, overrides any `--log-level` setting.

**Examples**:
```bash
aphex pipeline create my-pipeline --file pipeline.yaml --verbose
aphex organization list -v
```

## Log Levels

The CLI supports four hierarchical log levels:

- **debug**: Most verbose - shows all messages including detailed debugging information
- **info**: Default level - shows informational messages, warnings, and errors
- **warn**: Shows only warnings and errors
- **error**: Least verbose - shows only error messages

### Output Streams

- **Debug and Info** messages are written to **stdout**
- **Warn and Error** messages are written to **stderr**

This separation allows for flexible output handling in scripts and automation.

## Commands

### Authentication Commands

#### `aphex auth login`
Authenticate with the platform using OIDC.

```bash
aphex auth login [options]

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

### Organization Commands

#### `aphex organization bootstrap`
Bootstrap a new organization with namespace, webhook infrastructure, and RBAC.

```bash
aphex organization bootstrap --admin-email <email> <organization-name> [options]

Options:
  --admin-email string      Admin email address for the organization (required)
  --display-name string     Human-readable organization name (defaults to organization name)
  --webhook-secret string   GitHub webhook secret (auto-generated if not provided)
  --output, -o string       Output format (table, json, yaml)

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Creates Organization CRD in platform-system namespace
- Provisions organization namespace (org-{name})
- Creates GitHub webhook secret and Cloudflared credentials
- Sets up RBAC for organization admins
- Configures per-organization webhook URL (webhooks-{org}.homelab.local)
- Updates Cloudflared credentials if CLOUDFLARE_API_TOKEN environment variable is set
- Requires platform-admin permissions

**Environment Variables:**
- `CLOUDFLARE_API_TOKEN`: API token for Cloudflare tunnel management

#### `aphex organization list`
List all organizations and their status.

```bash
aphex organization list [options]

Options:
  --output, -o string       Output format (table, json, yaml)
  --quiet                   Suppress non-essential output

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

#### `aphex organization delete`
Delete an organization and all its resources.

```bash
aphex organization delete <organization-name> [options]

Options:
  --force                   Skip confirmation prompt

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Deletes Organization CRD from platform-system namespace
- Removes organization namespace and all resources
- Requires platform-admin permissions

### Pipeline Commands

#### `aphex pipeline create`
Create a new Tekton Pipeline resource and associated RepoBinding for webhook integration.

```bash
aphex pipeline create <name> --file <pipeline-file> --aphex-org <org> --repo-org <org> --repo-name <repo> [options]

Options:
  --file, -f string       Pipeline definition file path (required)
  --aphex-org string      Aphex organization name (required)
  --repo-org string       GitHub organization name (required)
  --repo-name string      GitHub repository name (required)
  --output, -o string     Output format (table, json, yaml)

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Creates pipeline in namespace matching pipeline name
- Automatically creates namespace if it doesn't exist
- Creates RepoBinding in platform-system namespace for webhook integration
- Sets tenantName to pipeline name automatically
- Hardcodes ingressHost to "webhooks.homelab.local"

#### `aphex pipeline delete`
Delete a Tekton Pipeline resource by discovering its namespace automatically.

```bash
aphex pipeline delete [name] [options]

Options:
  --force                 Skip confirmation prompt
  --output, -o string     Output format (table, json, yaml)

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Searches for pipeline across all accessible namespaces
- Uses pipeline name as namespace (pipeline-name = namespace-name convention)
- No namespace specification required

#### `aphex pipeline list`
List all Tekton Pipeline resources across accessible namespaces.

```bash
aphex pipeline list [options]

Options:
  --output, -o string     Output format (table, json, yaml)
  --quiet                 Suppress non-essential output

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Lists pipelines from all accessible namespaces automatically
- No namespace specification required or supported

### Knowledge Base Commands

#### `aphex knowledgebase create`
Create a new KnowledgeBase resource for tracking documentation repositories.

```bash
aphex knowledgebase create <name> [options]

Options:
  --namespace string      Namespace for the knowledge base [default: default]
  --repo-url string       Repository URL (https://github.com/org/repo)
  --branch string         Git branch to track [default: main]
  --docs-path string      Documentation path within repository [default: .kiro/docs]

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Creates KnowledgeBase CRD in specified namespace
- Validates repository URL starts with https://github.com/
- Validates documentation path starts with .kiro/docs
- Supports interactive mode when --repo-url is not provided
- Platform controller reconciles and validates the resource

**Interactive Mode:**
When --repo-url is not provided, prompts for:
- Repository URL
- Git branch (default: main)
- Documentation path (default: .kiro/docs)

#### `aphex knowledgebase list`
List all KnowledgeBase resources across all namespaces.

```bash
aphex knowledgebase list [options]

Options:
  --output, -o string     Output format (table, json, yaml)
  --quiet                 Suppress non-essential output

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Lists knowledge bases from all accessible namespaces
- Displays name, namespace, repository count, and phase
- Supports table, JSON, and YAML output formats

#### `aphex knowledgebase delete`
Delete a KnowledgeBase resource.

```bash
aphex knowledgebase delete <name> [options]

Options:
  --namespace string      Namespace of the knowledge base [default: default]
  --force                 Skip confirmation prompt

Global Options:
  --log-level, -l string   Set log level (debug, info, warn, error) [default: info]
  --verbose, -v            Enable verbose output (equivalent to --log-level=debug)
  --kubeconfig string      Path to kubeconfig file
```

**Behavior:**
- Fetches KnowledgeBase resource to display details
- Shows confirmation prompt with tracked repositories
- Deletes KnowledgeBase CRD from specified namespace
- Confirmation can be skipped with --force flag

## Authentication

Uses OIDC authentication via kubelogin exec plugin:
- Browser-based authentication flow
- Integrates with platform Dex endpoint
- Automatic token refresh handled by exec plugin

## Request/Response Formats

### Output Formats

**Table Format (default)**:
```
NAMESPACE    NAME         CREATED
tenant-1     my-pipeline  2024-01-09T10:30:00Z
```

**JSON Format**:
```json
[
  {
    "namespace": "tenant-1",
    "name": "my-pipeline", 
    "created": "2024-01-09T10:30:00Z"
  }
]
```

**YAML Format**:
```yaml
- namespace: tenant-1
  name: my-pipeline
  created: "2024-01-09T10:30:00Z"
```

## Error Handling

### Authentication Errors
- Exit code: 1
- Message includes remediation steps
- Suggests running `aphex auth login`

### Permission Errors  
- Exit code: 1
- Shows required group membership
- Includes current user groups
- Provides access request instructions

### Network Errors
- Exit code: 1
- Provides troubleshooting guidance
- Suggests checking cluster connectivity

## Integration Examples

### Logger API for Developers

The Aphex CLI provides a centralized logger that can be accessed from any command or package function via context. This section describes how to use the logger API in your code.

#### Accessing the Logger

**In Commands** (`internal/commands/*.go`):
```go
func myCommandAction(ctx context.Context, cmd *cli.Command) error {
    // Get logger from context
    logger := logger.GetLogger(ctx)
    
    // Use logger methods
    logger.Debug("Starting command execution")
    logger.Info("Processing request")
    logger.Warn("Deprecated flag used")
    logger.Error("Operation failed")
    
    return nil
}
```

**In Package Functions** (`pkg/*/*.go`):
```go
func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
    // Get logger from context
    logger := logger.GetLogger(ctx)
    
    // Use logger for debugging and information
    logger.Debugf("Creating resource with name: %s", opts.Name)
    logger.Infof("Resource created successfully: %s", opts.Name)
    
    return nil
}
```

#### Logger Methods

The logger provides both simple and formatted logging methods:

**Simple Methods**:
- `logger.Debug(msg string)` - Log debug message
- `logger.Info(msg string)` - Log informational message
- `logger.Warn(msg string)` - Log warning message
- `logger.Error(msg string)` - Log error message

**Formatted Methods**:
- `logger.Debugf(format string, args ...interface{})` - Log formatted debug message
- `logger.Infof(format string, args ...interface{})` - Log formatted info message
- `logger.Warnf(format string, args ...interface{})` - Log formatted warning message
- `logger.Errorf(format string, args ...interface{})` - Log formatted error message

#### Usage Guidelines

**When to use each level**:

- **Debug**: Detailed information for troubleshooting
  - Variable values and state
  - Function entry/exit points
  - Intermediate calculation results
  - API request/response details

- **Info**: General informational messages
  - Operation start/completion
  - Resource creation/deletion
  - Configuration values
  - Progress updates

- **Warn**: Warning conditions that don't prevent operation
  - Deprecated features being used
  - Fallback behavior triggered
  - Non-critical errors recovered from
  - Performance concerns

- **Error**: Error conditions that prevent operation
  - API failures
  - Invalid input
  - Resource not found
  - Permission denied

#### Example: Migrating from Verbose Flags

**Before** (old pattern with verbose flags):
```go
type CreateOptions struct {
    Name    string
    Verbose bool
}

func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
    if opts.Verbose {
        fmt.Printf("Creating resource: %s\n", opts.Name)
    }
    
    // ... create resource ...
    
    if opts.Verbose {
        fmt.Printf("Resource created successfully\n")
    }
    
    return nil
}
```

**After** (new pattern with centralized logger):
```go
type CreateOptions struct {
    Name string
    // No Verbose field needed
}

func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
    logger := logger.GetLogger(ctx)
    
    logger.Debugf("Creating resource: %s", opts.Name)
    
    // ... create resource ...
    
    logger.Info("Resource created successfully")
    
    return nil
}
```

#### No-Op Logger Fallback

The logger is designed to be fail-safe. If `GetLogger(ctx)` is called on a context that doesn't contain a logger, it returns a no-op logger that discards all output without error:

```go
// This is safe even if context doesn't have a logger
logger := logger.GetLogger(ctx)
logger.Debug("This won't panic even if logger wasn't injected")
```

This allows package functions to work correctly even if called outside the normal CLI flow (e.g., in tests or as a library).

#### Testing with Logger

When writing tests, you can inject a logger into the test context:

```go
func TestMyFunction(t *testing.T) {
    // Create a logger for testing
    log := logger.NewLogger("debug")
    
    // Inject into context
    ctx := logger.WithLogger(context.Background(), log)
    
    // Call function with context
    err := MyFunction(ctx, options)
    
    // Assert results
    assert.NoError(t, err)
}
```

Or use the no-op logger for tests that don't need logging:

```go
func TestMyFunction(t *testing.T) {
    // Context without logger will use no-op logger
    ctx := context.Background()
    
    // Function will work fine with no-op logger
    err := MyFunction(ctx, options)
    
    assert.NoError(t, err)
}
```

**Source**
- `pkg/logger/logger.go` - Logger methods and interface implementation
- `pkg/logger/context.go` - Context integration functions (WithLogger, GetLogger, NewNoOpLogger)
- `internal/commands/pipeline.go` - Example command usage pattern
- `pkg/pipeline/create.go` - Example package function usage pattern

### Scripting Usage
```bash
#!/bin/bash
# Create pipeline and check status with debug logging
aphex pipeline create my-pipeline \
  --file pipeline.yaml \
  --log-level=debug \
  --output json > result.json

if [ $? -eq 0 ]; then
  echo "Pipeline created successfully"
  aphex pipeline list --output table
fi

# Separate normal output from errors
aphex pipeline list --log-level=info > pipelines.txt 2> errors.log
```

### CI/CD Integration
```yaml
# GitHub Actions example - Organization Bootstrap (Admin only)
- name: Bootstrap Organization
  run: |
    aphex auth login
    aphex organization bootstrap ${{ github.repository_owner }} \
      --admin-email admin@company.com \
      --log-level=info

# GitHub Actions example - Pipeline Creation with verbose logging
- name: Deploy Pipeline with Webhook Integration
  run: |
    aphex auth login
    aphex pipeline create ${{ github.event.repository.name }} \
      --file .tekton/pipeline.yaml \
      --aphex-org my-organization \
      --repo-org ${{ github.repository_owner }} \
      --repo-name ${{ github.event.repository.name }} \
      --verbose
```

## Organization Management

Organization bootstrap creates a complete multi-tenant setup:

**Created Resources:**
- Organization CRD in `platform-system` namespace
- Organization namespace: `org-{organization-name}`
- GitHub webhook secret: `github-webhook-{organization-name}`
- RBAC roles and bindings for organization admins

**Organization Specification:**
```yaml
apiVersion: aphex/v1alpha1
kind: Organization
metadata:
  name: {organization-name}
  namespace: platform-system
spec:
  displayName: {display-name}
  adminUsers: [{admin-email}]
  webhookSecret: {auto-generated-or-provided}
status:
  namespace: org-{organization-name}
  webhookURL: https://webhooks-{organization-name}.homelab.local
  phase: Active
```

**Per-Organization Infrastructure:**
Each organization gets its own isolated webhook infrastructure:
- Dedicated Cloudflared tunnel with unique subdomain
- EventListener in organization namespace
- Webhook secret scoped to organization
- Complete isolation between organizations

## RepoBinding Integration

Pipeline creation automatically provisions webhook integration:

**Created Resources:**
- Pipeline in `{pipeline-name}` namespace
- RepoBinding in `platform-system` namespace
- Webhook configuration for GitHub integration

**RepoBinding Specification:**
```yaml
apiVersion: aphex/v1alpha1
kind: RepoBinding
metadata:
  name: {pipeline-name}-binding
  namespace: platform-system
spec:
  aphexOrg: {provided-aphex-org}
  repoOrg: {provided-repo-org}
  repoName: {provided-repo-name}
  tenantName: {pipeline-name}
  pipelineName: {pipeline-name}
  ingressHost: webhooks.homelab.local
```

**Source**
- `internal/commands/auth.go` - Authentication commands
- `internal/commands/organization.go` - Organization commands
- `internal/commands/pipeline.go` - Pipeline commands  
- `pkg/output/formatter.go` - Output formatting
- `pkg/auth/errors.go` - Error message formatting


## Knowledge Base Management

Knowledge bases track documentation repositories for the Archon RAG system. The CLI provides commands to manage KnowledgeBase custom resources that specify which repositories to monitor for documentation changes.

**Created Resources:**
- KnowledgeBase CRD in specified namespace
- Platform controller validates and reconciles the resource
- Archon agent monitors specified repositories

**KnowledgeBase Specification:**
```yaml
apiVersion: aphex/v1alpha1
kind: KnowledgeBase
metadata:
  name: {knowledge-base-name}
  namespace: {namespace}
spec:
  displayName: {knowledge-base-name}
  repositories:
    - url: https://github.com/org/repo
      branch: main
      paths:
        - .kiro/docs
status:
  phase: Ready
  message: "Tracking 1 repositories"
  lastReconcileTime: "2026-01-16T10:30:00Z"
```

**Validation Rules:**
- Repository URLs must start with `https://github.com/`
- Documentation paths must start with `.kiro/docs`
- At least one repository must be specified

**Integration with Archon:**
- Archon agent monitors repositories specified in KnowledgeBase resources
- Documentation changes trigger re-ingestion into vector store
- RAG queries use ingested documentation for context

**Source**
- `internal/commands/knowledgebase.go` - Knowledge base commands
- `pkg/knowledgebase/knowledgebase.go` - Knowledge base business logic
- `pkg/knowledgebase/formatter.go` - Knowledge base output formatting
- `AphexPlatformInfrastructure/platform/platform-controller/controller/api/v1alpha1/knowledgebase_types.go` - KnowledgeBase CRD definition
- `AphexPlatformInfrastructure/platform/platform-controller/controller/controllers/knowledgebase_controller.go` - KnowledgeBase controller
