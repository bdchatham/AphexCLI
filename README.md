# Aphex CLI

A command-line tool for managing Tekton pipelines and organizations on the Aphex platform.

## How It Works

The Aphex CLI provides a thin, declarative interface to the Aphex Pipeline Infrastructure:

1. **CLI creates RepoBinding CRD**: The CLI reads your pipeline YAML and creates a RepoBinding custom resource with the pipeline definition embedded
2. **Controller reconciles**: The RepoBinding controller in AphexPlatformInfrastructure watches for RepoBinding resources and provisions all necessary infrastructure
3. **Kubernetes-driven**: All resource creation and management happens through Kubernetes reconciliation, not imperative CLI commands

This design follows the Kubernetes operator pattern: the CLI is thin and declarative, while the controller handles all provisioning logic.

## Features

- **OIDC Authentication**: Browser-based authentication via platform Dex
- **Organization Management**: Bootstrap and manage multi-tenant organizations
- **Pipeline Management**: Create, delete, and list Tekton pipelines
- **Knowledge Base Management**: Create, delete, and list knowledge bases for Archon documentation tracking
- **Webhook Integration**: Automatic GitHub webhook provisioning via RepoBindings
- **Interactive Mode**: Guided prompts for missing parameters
- **Multiple Output Formats**: Table, JSON, and YAML output support
- **Centralized Logging**: Global log level control with `--log-level` and `--verbose` flags
- **Cross-Platform**: Builds for Linux, macOS, and Windows

## Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/bdchatham/AphexCLI/releases).

### macOS/Linux
```bash
# Extract and install
tar -xzf aphex_*.tar.gz
sudo mv aphex /usr/local/bin/
```

### Windows
```powershell
# Extract aphex.exe and add to PATH
```

## Quick Start

```bash
# Authenticate with the platform
aphex auth login

# Bootstrap a new organization (admin only)
aphex organization bootstrap my-org --admin-email admin@company.com

# List organizations
aphex organization list

# Create a pipeline with webhook integration
aphex pipeline create my-pipeline \
  --file pipeline.yaml \
  --aphex-org my-org \
  --repo-org github-org \
  --repo-name github-repo

# The controller will provision:
# - Namespace: my-pipeline
# - Pipeline resource in namespace my-pipeline
# - Service account, RBAC, and resource limits
# - Tekton Triggers (TriggerTemplate, Trigger)

# Monitor provisioning progress
kubectl get repobinding my-pipeline-binding -n platform-system -o yaml

# List pipelines
aphex pipeline list

# Delete a pipeline
aphex pipeline delete my-pipeline

# Create a knowledge base for documentation tracking
aphex knowledgebase create platform-docs \
  --namespace platform-system \
  --repo-url https://github.com/bdchatham/AphexPlatformInfrastructure \
  --branch main \
  --docs-path .kiro/docs

# List knowledge bases
aphex knowledgebase list

# Delete a knowledge base
aphex knowledgebase delete platform-docs --namespace platform-system
```

## Knowledge Base Management

Knowledge bases track documentation repositories for the Archon RAG system. The CLI provides commands to manage KnowledgeBase custom resources.

### Create a Knowledge Base

```bash
# Create with all parameters
aphex knowledgebase create my-kb \
  --namespace default \
  --repo-url https://github.com/org/repo \
  --branch main \
  --docs-path .kiro/docs

# Interactive mode (prompts for missing parameters)
aphex knowledgebase create my-kb --namespace default

# Multiple repositories can be added by creating multiple knowledge bases
# or by editing the KnowledgeBase resource directly
```

### List Knowledge Bases

```bash
# List all knowledge bases (table format)
aphex knowledgebase list

# JSON output
aphex knowledgebase list --output json

# YAML output
aphex knowledgebase list --output yaml

# Quiet mode (names only)
aphex knowledgebase list --quiet
```

### Delete a Knowledge Base

```bash
# Delete with confirmation prompt
aphex knowledgebase delete my-kb --namespace default

# Skip confirmation
aphex knowledgebase delete my-kb --namespace default --force
```

### Knowledge Base Details

When you create a knowledge base:
- A KnowledgeBase custom resource is created in the specified namespace
- The platform controller validates the repository URL and documentation path
- The Archon agent monitors the specified repositories for documentation changes
- Documentation is ingested into the vector store for RAG queries

## Logging and Verbosity

The Aphex CLI provides centralized logging control through global flags:

### Global Flags

- **`--log-level <level>`** or **`-l <level>`**: Set the log level (debug, info, warn, error)
  - Default: `info`
  - Controls verbosity for all commands
  
- **`--verbose`** or **`-v`**: Enable verbose output (equivalent to `--log-level=debug`)
  - Shorthand for maximum verbosity

### Log Levels

- **debug**: Show all messages including detailed debugging information
- **info**: Show informational messages and above (default)
- **warn**: Show warnings and errors only
- **error**: Show only error messages

### Examples

```bash
# Run with debug logging to see detailed information
aphex pipeline create my-pipeline --file pipeline.yaml --log-level=debug

# Use verbose flag for maximum detail
aphex pipeline list --verbose

# Suppress informational messages, show only warnings and errors
aphex organization list --log-level=warn

# Show only errors
aphex pipeline delete my-pipeline --log-level=error
```

### Output Streams

- Debug and Info messages go to **stdout**
- Warn and Error messages go to **stderr**

This allows you to separate normal output from error output in scripts:

```bash
# Capture normal output, display errors
aphex pipeline list --log-level=debug > pipelines.txt

# Redirect errors to a file
aphex pipeline create my-pipeline --file pipeline.yaml 2> errors.log
```

## Documentation

Complete documentation is available in `.kiro/docs/`:
- [Overview](.kiro/docs/overview.md) - Purpose and key concepts
- [Architecture](.kiro/docs/architecture.md) - System design and components
- [Operations](.kiro/docs/operations.md) - Deployment and troubleshooting
- [API](.kiro/docs/api.md) - Command reference and examples
- [Data Models](.kiro/docs/data-models.md) - Data structures and schemas
- [FAQ](.kiro/docs/faq.md) - Common questions and answers

## Development

```bash
# Build for current platform
make build

# Run tests
make test

# Build for all platforms
make build-all

# Create release
make release
```

## Archon Integration

This repository participates in the **Archon** RAG system. Documentation under `.kiro/docs/` is ingested to provide context for automated agents and engineers.

See `CLAUDE.md` for the complete documentation contract.

## License

MIT


