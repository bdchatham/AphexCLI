# Aphex CLI

A command-line tool for managing Tekton pipelines and organizations on the Arbiter platform.

## Features

- **OIDC Authentication**: Browser-based authentication via platform Dex
- **Organization Management**: Bootstrap and manage multi-tenant organizations
- **Pipeline Management**: Create, delete, and list Tekton pipelines
- **Webhook Integration**: Automatic GitHub webhook provisioning via RepoBindings
- **Interactive Mode**: Guided prompts for missing parameters
- **Multiple Output Formats**: Table, JSON, and YAML output support
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

# List pipelines
aphex pipeline list

# Delete a pipeline
aphex pipeline delete my-pipeline
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
