# Overview

## Purpose

The Aphex CLI is a command-line tool for managing Tekton pipelines on the Arbiter platform. It provides a simplified interface for creating, deleting, and listing pipeline instances using native Tekton concepts, without requiring deep Kubernetes knowledge.

## Key Features

- **OIDC Authentication**: Integrates with platform Dex/Authentik authentication system
- **Pipeline Management**: Create, delete, and list Tekton Pipeline resources
- **Interactive Mode**: Guided prompts for missing parameters
- **Multiple Output Formats**: Table, JSON, and YAML output support
- **Cross-Platform**: Builds for Linux, macOS, and Windows

## Archon Integration

This repository is ingested by the **Archon** RAG system, which reads all Markdown files under `.kiro/docs/` to build mental models for sourcing code and architectural information.

Documentation in this repository follows the Archon documentation contract defined in `CLAUDE.md` at the repo root.

## Quick Start

```bash
# Authenticate with the platform
aphex auth login

# List pipelines in current namespace
aphex pipeline list

# Create a pipeline
aphex pipeline create my-pipeline --file pipeline.yaml

# Delete a pipeline
aphex pipeline delete my-pipeline
```

## Key Concepts

- **Pipeline**: Tekton Pipeline resource defining a workflow
- **Namespace**: Kubernetes namespace for tenant isolation
- **OIDC Authentication**: Browser-based authentication via Dex
- **Interactive Mode**: Automatic prompting when arguments are missing

**Source**
- `cmd/aphex/main.go` - CLI entry point
- `internal/commands/` - Command implementations
- `pkg/` - Core functionality packages
