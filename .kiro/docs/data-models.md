# Data Models

## Overview

The Aphex CLI uses structured Go types for configuration, operations, and data exchange. All data models are designed for type safety and clear interfaces between components.

## Configuration Models

### LoginOptions
Authentication configuration for OIDC setup.

```go
type LoginOptions struct {
    KubeconfigPath string  // Path to kubeconfig file
}
```

**Note**: Verbose output is now controlled by global `--log-level` and `--verbose` flags, not per-command options.

### Pipeline Operation Options

#### CreateOptions
```go
type CreateOptions struct {
    Name        string  // Pipeline name
    FilePath    string  // YAML file path
    AphexOrg    string  // Aphex organization name
    RepoOrg     string  // GitHub organization name
    RepoName    string  // GitHub repository name
    TenantName  string  // Tenant name (set to pipeline name)
    IngressHost string  // Ingress host (hardcoded to webhooks.homelab.local)
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

#### DeleteOptions
```go
type DeleteOptions struct {
    Name      string  // Pipeline name
    Namespace string  // Target namespace
    Force     bool    // Skip confirmation
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

#### ListOptions
```go
type ListOptions struct {
    Namespace     string        // Target namespace
    AllNamespaces bool          // List across all namespaces
    OutputFormat  output.Format // Output format
    Quiet         bool          // Suppress output
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

### Organization Operation Options

#### BootstrapOptions
```go
type BootstrapOptions struct {
    Name          string  // Organization name
    DisplayName   string  // Human-readable organization name
    AdminEmail    string  // Admin email address
    WebhookSecret string  // GitHub webhook secret (auto-generated if empty)
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

#### ListOptions
```go
type ListOptions struct {
    Quiet   bool  // Suppress non-essential output
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

#### DeleteOptions
```go
type DeleteOptions struct {
    Name    string  // Organization name
    Force   bool    // Skip confirmation prompt
}
```

**Note**: The `Verbose bool` field was removed. Logging is now controlled by global `--log-level` and `--verbose` flags.

## Output Models

### PipelineInfo
Structured pipeline information for JSON/YAML output.

```go
type PipelineInfo struct {
    Namespace string `json:"namespace" yaml:"namespace"`
    Name      string `json:"name" yaml:"name"`
    Created   string `json:"created" yaml:"created"`
}
```

## Error Models

### PermissionError
Structured permission error with remediation context.

```go
type PermissionError struct {
    Resource  string  // Kubernetes resource type
    Verb      string  // Kubernetes verb
    Namespace string  // Target namespace
    Reason    string  // API error reason
}
```

## Client Models

### Kubernetes Client
Typed controller-runtime client with registered schemes for all resource types.

```go
// Created via NewAphexClient() in pkg/k8s/aphex_client.go
type Client interface {
    Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
    List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
    Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
    Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
    // ... other controller-runtime client methods
}

// Registered types in scheme:
// - corev1.Namespace
// - tektonv1.Pipeline
// - platformv1alpha1.Organization
// - platformv1alpha1.RepoBinding
```

## Data Flow

1. **User Input** → CLI parsing → Options structs
2. **Options structs** → Business logic → Kubernetes API calls
3. **API responses** → Output models → Formatted display
4. **Errors** → Error models → User-friendly messages

## Validation Rules

### Pipeline Names
- Must be valid Kubernetes resource names
- Alphanumeric characters and hyphens only
- Must start and end with alphanumeric character

### Namespace Names
- Must be valid Kubernetes namespace names
- Follow DNS label standards
- Lowercase alphanumeric and hyphens only

### File Paths
- Must exist and be readable
- Must contain valid YAML
- Must define Tekton Pipeline resource

**Source**
- `pkg/auth/login.go` - Authentication models
- `pkg/pipeline/` - Pipeline operation models
- `pkg/organization/` - Organization operation models
- `pkg/output/formatter.go` - Output models
- `pkg/auth/permissions.go` - Error models
- `pkg/k8s/aphex_client.go` - Typed client setup and scheme registration
