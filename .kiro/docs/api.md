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

### Organization Commands

#### `aphex organization bootstrap`
Bootstrap a new organization with namespace, webhook secrets, and RBAC.

```bash
aphex organization bootstrap [organization-name] [options]

Options:
  --admin-email string      Admin email address for the organization (required)
  --display-name string     Human-readable organization name (defaults to organization name)
  --webhook-secret string   GitHub webhook secret (auto-generated if not provided)
  --kubeconfig string       Path to kubeconfig file
  --output, -o string       Output format (table, json, yaml)
  --verbose, -v             Verbose output
```

**Behavior:**
- Creates Organization CRD in platform-system namespace
- Provisions organization namespace (org-{name})
- Creates GitHub webhook secret for the organization
- Sets up RBAC for organization admins
- Requires platform-admin permissions

#### `aphex organization list`
List all organizations and their status.

```bash
aphex organization list [options]

Options:
  --kubeconfig string       Path to kubeconfig file
  --output, -o string       Output format (table, json, yaml)
  --quiet                   Suppress non-essential output
  --verbose, -v             Verbose output
```

### Pipeline Commands

#### `aphex pipeline create`
Create a new Tekton Pipeline resource and associated RepoBinding for webhook integration.

```bash
aphex pipeline create [name] [options]

Options:
  --file, -f string       Pipeline definition file path (required)
  --repo-org string       GitHub organization name (required)
  --repo-name string      GitHub repository name (required)
  --kubeconfig string     Path to kubeconfig file
  --output, -o string     Output format (table, json, yaml)
  --verbose, -v           Verbose output
```

**Behavior:**
- Creates pipeline in namespace matching pipeline name
- Automatically creates namespace if it doesn't exist
- Creates RepoBinding in platform-system namespace for webhook integration
- Sets tenantName to pipeline name for resource organization

#### `aphex pipeline delete`
Delete a Tekton Pipeline resource by discovering its namespace automatically.

```bash
aphex pipeline delete [name] [options]

Options:
  --kubeconfig string     Path to kubeconfig file
  --force                 Skip confirmation prompt
  --output, -o string     Output format (table, json, yaml)
  --verbose, -v           Verbose output
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
  --kubeconfig string     Path to kubeconfig file
  --output, -o string     Output format (table, json, yaml)
  --quiet                 Suppress non-essential output
  --verbose, -v           Verbose output
```

**Behavior:**
- Lists pipelines from all accessible namespaces automatically
- No namespace specification required or supported

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
# GitHub Actions example - Organization Bootstrap (Admin only)
- name: Bootstrap Organization
  run: |
    aphex auth login
    aphex organization bootstrap ${{ github.repository_owner }} \
      --admin-email admin@company.com

# GitHub Actions example - Pipeline Creation
- name: Deploy Pipeline with Webhook Integration
  run: |
    aphex auth login
    aphex pipeline create ${{ github.event.repository.name }} \
      --file .tekton/pipeline.yaml \
      --repo-org ${{ github.repository_owner }} \
      --repo-name ${{ github.event.repository.name }}
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
apiVersion: arbiter.io/v1alpha1
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
apiVersion: arbiter.io/v1alpha1
kind: RepoBinding
metadata:
  name: {pipeline-name}-binding
  namespace: platform-system
spec:
  repoOrg: {provided-org}
  repoName: {provided-repo}
  tenantName: {pipeline-name}
  pipelineName: {pipeline-name}
  ingressHost: webhooks.homelab.local
```

**Source**
- `internal/commands/auth.go` - Authentication commands
- `internal/commands/pipeline.go` - Pipeline commands  
- `pkg/output/formatter.go` - Output formatting
- `pkg/auth/errors.go` - Error message formatting
