# Data Models

## Overview

The Aphex CLI uses structured Go types for configuration, operations, and data exchange. All data models are designed for type safety and clear interfaces between components.

## Configuration Models

### LoginOptions
Authentication configuration for OIDC setup.

```go
type LoginOptions struct {
    KubeconfigPath string  // Path to kubeconfig file
    Verbose        bool    // Enable verbose output
}
```

### Pipeline Operation Options

#### CreateOptions
```go
type CreateOptions struct {
    Name      string  // Pipeline name
    Namespace string  // Target namespace  
    FilePath  string  // YAML file path
    Verbose   bool    // Enable verbose output
}
```

#### DeleteOptions
```go
type DeleteOptions struct {
    Name      string  // Pipeline name
    Namespace string  // Target namespace
    Force     bool    // Skip confirmation
    Verbose   bool    // Enable verbose output
}
```

#### ListOptions
```go
type ListOptions struct {
    Namespace     string        // Target namespace
    AllNamespaces bool          // List across all namespaces
    OutputFormat  output.Format // Output format
    Quiet         bool          // Suppress output
    Verbose       bool          // Enable verbose output
}
```

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
Wrapper around client-go with additional context.

```go
type Client struct {
    Clientset *kubernetes.Clientset  // K8s client
    Config    *rest.Config           // REST config
    Namespace string                 // Default namespace
}
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
- `pkg/output/formatter.go` - Output models
- `pkg/auth/permissions.go` - Error models
- `pkg/k8s/client.go` - Client models
