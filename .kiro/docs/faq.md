# FAQ

## General Questions

### What is this repository for?
The Aphex CLI is a command-line tool for managing Tekton pipelines on the Aphex platform. It provides a simplified interface for creating, deleting, and listing pipeline instances without requiring deep Kubernetes knowledge.

### How does this fit into the larger system?
The CLI integrates with the Aphex platform's Kubernetes cluster and authentication system. It uses OIDC authentication via Dex and manages Tekton Pipeline resources in tenant namespaces.

## Development Questions

### How do I set up my development environment?
1. Install Go 1.25+
2. Install kubelogin: `brew install Azure/kubelogin/kubelogin`
3. Clone repository and run: `make build`
4. Authenticate: `./bin/aphex auth login`

### How do I run tests?
```bash
# Run all tests
make test

# Run property-based tests specifically
go test ./.kiro/{current-spec}/tests/...
```

## Operational Questions

### How do I deploy changes?
1. Update `VERSION` file with new semantic version
2. Run `make release` to create GitHub release with binaries
3. Users install from GitHub releases

### What should I do if authentication fails?
1. Verify kubelogin is installed: `kubelogin --version`
2. Check Dex endpoint: `curl https://dex.home.local`
3. Re-authenticate: `aphex auth login`

### What should I do if I get permission errors?
Check the error message for required group membership and contact your platform administrator to request access to the appropriate group (e.g., platform-engineering).

### How do I troubleshoot pipeline creation failures?
1. Validate YAML: `kubectl apply --dry-run=client -f pipeline.yaml`
2. Check namespace exists: `kubectl get namespace <namespace>`
3. Verify Tekton is installed: `kubectl get crd pipelines.tekton.dev`
4. Use `--log-level=debug` or `--verbose` flag for detailed output

### How do I manage secrets for my organization?
Use the `secret` command to set, list, and delete organization-scoped secrets:

```bash
# Set a secret
aphex secret set --org my-org github-token=ghp_xxx

# List secrets (values are redacted)
aphex secret list --org my-org

# Delete a secret
aphex secret delete --org my-org github-token
```

Secrets are stored in Kubernetes Secrets in the organization namespace and are accessible to pipelines within that organization.

### How do I create a knowledge base?
Use the `knowledgebase create` command to provision a RAG knowledge base:

```bash
# Simple single-repo
aphex knowledgebase create my-kb \
  --repo-url https://github.com/org/repo

# With MCP server (explicit config required)
aphex knowledgebase create my-kb \
  --repo-url https://github.com/org/repo \
  --mcp-image ghcr.io/bdchatham/archon-mcp-server:latest \
  --mcp-port 8090
```

This creates a KnowledgeBase CRD that the platform controller reconciles into a full knowledge base deployment with vector storage, embedding service, and query API.

### How do I create a knowledge base with multiple repositories?
Use the file input pattern for complex configurations:

```bash
# Generate template
aphex knowledgebase generate-spec > kb.yaml

# Edit kb.yaml to add multiple repositories
# Example:
#   repositories:
#   - url: https://github.com/org/repo1
#     branch: main
#     paths: [".kiro/docs"]
#   - url: https://github.com/org/repo2
#     branch: mainline
#     paths: ["docs/**/*.md"]

# Create from file
aphex knowledgebase create --cli-input-yaml kb.yaml
```

The CRD supports multiple repositories, but the CLI flags only support single-repo for simplicity. Use file input for multi-repo configurations.

### How do I create an agent?
Use the `agent create` command to provision a model server with optional RAG capabilities:

```bash
# Simple model server
aphex agent create llama-70b \
  --model meta-llama/Llama-3.1-70B-Instruct \
  --gpu-count 4

# With RAG capabilities
aphex agent create llama-rag \
  --model meta-llama/Llama-3.1-70B-Instruct \
  --kb-name my-kb \
  --orchestration
```

This creates an Agent CRD that the platform controller reconciles into:
- Model server deployment (vLLM)
- Optional orchestrator for unified RAG endpoint
- Services for accessing the model and orchestrator

### What are the Agent deployment patterns?
The Agent CRD supports three patterns:

1. **Model only**: Direct model inference, no RAG
   ```bash
   aphex agent create my-agent --model meta-llama/Llama-3.1-70B-Instruct
   ```

2. **Model + KB**: Manual RAG orchestration by user
   ```bash
   aphex agent create my-agent --model meta-llama/Llama-3.1-70B-Instruct --kb-name my-kb
   ```

3. **Model + KB + Orchestration**: Unified `/v1/chat` endpoint with automatic RAG
   ```bash
   aphex agent create my-agent --model meta-llama/Llama-3.1-70B-Instruct --kb-name my-kb --orchestration
   ```

### How do I use file input for complex Agent configurations?
Use the AWS CLI-style file input pattern:

```bash
# Generate template
aphex agent generate-spec > agent.yaml

# Edit agent.yaml with your configuration

# Create from file
aphex agent create --cli-input-yaml agent.yaml
```

This is useful for complex configurations with custom images, ports, or multiple settings.

## Archon-Specific Questions

### How is this repository ingested by Archon?
Archon reads all Markdown files under `.kiro/docs/` from this public GitHub repository. Documentation follows the contract defined in `CLAUDE.md`.

### How do I update documentation?
Update the relevant files under `.kiro/docs/` and ensure changes are grounded in code. Include "Source" references to relevant files.

**Source**
- `cmd/aphex/main.go` - CLI entry point
- `pkg/auth/login.go` - Authentication logic
- `pkg/auth/errors.go` - Error handling
- `Makefile` - Build and release process
- `CLAUDE.md` - Documentation contract
