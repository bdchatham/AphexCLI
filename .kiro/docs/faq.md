# FAQ

## General Questions

### What is this repository for?
The Aphex CLI is a command-line tool for managing Tekton pipelines on the Arbiter platform. It provides a simplified interface for creating, deleting, and listing pipeline instances without requiring deep Kubernetes knowledge.

### How does this fit into the larger system?
The CLI integrates with the Arbiter platform's Kubernetes cluster and authentication system. It uses OIDC authentication via Dex and manages Tekton Pipeline resources in tenant namespaces.

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
