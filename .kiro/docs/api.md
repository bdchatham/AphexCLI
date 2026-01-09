# API

## Overview

The Aphex CLI provides a command-line interface for managing Tekton pipelines on the Arbiter platform. This document describes the CLI commands, options, and integration patterns.

## Commands

### Authentication Commands

#### `aphex auth login`
Authenticate with the platform using OIDC.

```bash
aphex auth login [options]

Options:
  --kubeconfig string   Path to kubeconfig file
  --verbose, -v         Verbose output
```

### Pipeline Commands

#### `aphex pipeline create`
Create a new Tekton Pipeline resource.

```bash
aphex pipeline create [name] [options]

Options:
  --file, -f string       Pipeline definition file path (required)
  --namespace, -n string  Kubernetes namespace
  --kubeconfig string     Path to kubeconfig file
  --output, -o string     Output format (table, json, yaml)
  --verbose, -v           Verbose output
```

#### `aphex pipeline delete`
Delete a Tekton Pipeline resource.

```bash
aphex pipeline delete [name] [options]

Options:
  --namespace, -n string  Kubernetes namespace
  --kubeconfig string     Path to kubeconfig file
  --force                 Skip confirmation prompt
  --output, -o string     Output format (table, json, yaml)
  --verbose, -v           Verbose output
```

#### `aphex pipeline list`
List Tekton Pipeline resources.

```bash
aphex pipeline list [options]

Options:
  --namespace, -n string  Kubernetes namespace
  --kubeconfig string     Path to kubeconfig file
  --all-namespaces        List pipelines across all namespaces
  --output, -o string     Output format (table, json, yaml)
  --quiet                 Suppress non-essential output
  --verbose, -v           Verbose output
```

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

### Scripting Usage
```bash
#!/bin/bash
# Create pipeline and check status
aphex pipeline create my-pipeline --file pipeline.yaml --output json > result.json
if [ $? -eq 0 ]; then
  echo "Pipeline created successfully"
  aphex pipeline list --namespace tenant-1 --output table
fi
```

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Deploy Pipeline
  run: |
    aphex auth login
    aphex pipeline create ${{ github.event.repository.name }} \
      --file .tekton/pipeline.yaml \
      --namespace ${{ env.TENANT_NAMESPACE }}
```

**Source**
- `internal/commands/auth.go` - Authentication commands
- `internal/commands/pipeline.go` - Pipeline commands  
- `pkg/output/formatter.go` - Output formatting
- `pkg/auth/errors.go` - Error message formatting
