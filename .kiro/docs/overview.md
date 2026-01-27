# Overview

## Purpose

The Aphex CLI is a command-line tool for managing Tekton pipelines on the Aphex platform. It provides a simplified interface for creating, deleting, and listing pipeline instances using native Tekton concepts, without requiring deep Kubernetes knowledge.

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

# Bootstrap a new organization
aphex organization bootstrap my-org --admin-email admin@example.com

# List pipelines in current namespace
aphex pipeline list

# Create a pipeline
aphex pipeline create my-pipeline --file pipeline.yaml \
  --aphex-org my-org --repo-org github-org --repo-name my-repo

# Set secrets for an organization
aphex secret set --org my-org github-token=ghp_xxx

# Create a knowledge base
aphex knowledgebase create my-kb --org my-org

# Create an agent (model server)
aphex agent create llama-70b \
  --model meta-llama/Llama-3.1-70B-Instruct \
  --gpu-count 4

# Create an agent with RAG capabilities
aphex agent create llama-rag \
  --model meta-llama/Llama-3.1-70B-Instruct \
  --kb-name my-kb \
  --orchestration

# Delete a pipeline
aphex pipeline delete my-pipeline
```

## Key Concepts

- **Organization**: Multi-tenant isolation boundary with dedicated namespace and webhook infrastructure
- **Pipeline**: Tekton Pipeline resource defining a workflow, managed via RepoBinding CRD
- **RepoBinding**: Platform CRD that connects GitHub repositories to pipelines
- **Secret**: Organization-scoped secrets for GitHub tokens, credentials, etc.
- **Knowledge Base**: RAG knowledge base instance for document ingestion and retrieval
- **Agent**: Model server (vLLM) with optional KnowledgeBase integration and orchestration
- **Namespace**: Kubernetes namespace for tenant isolation (auto-managed by controller)
- **OIDC Authentication**: Browser-based authentication via Dex
- **Interactive Mode**: Automatic prompting when arguments are missing

**Source**
- `cmd/aphex/main.go` - CLI entry point
- `internal/commands/` - Command implementations
- `pkg/` - Core functionality packages
