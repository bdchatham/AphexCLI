# Operations

## Deployment

### Installation for End Users

**Recommended: Homebrew (macOS)**

```bash
# Add the tap
brew tap bdchatham/aphex

# Install aphex (automatically installs kubelogin dependency)
brew install aphex

# Verify installation
aphex --version
```

**Manual Installation (Linux or without Homebrew)**

Download the latest release from GitHub:

```bash
# Linux (x86_64)
curl -L https://github.com/bdchatham/AphexCLI/releases/latest/download/aphex-linux-amd64 -o aphex
chmod +x aphex
sudo mv aphex /usr/local/bin/

# Verify installation
aphex --version
```

**Prerequisites for Manual Installation:**
- `kubelogin` must be installed for OIDC authentication:
  ```bash
  # macOS
  brew install Azure/kubelogin/kubelogin
  
  # Linux - download from https://github.com/Azure/kubelogin/releases
  ```

### Building the CLI
```bash
# Build for current platform
make build

# Build for all platforms
make build-all
```

### Release Process
1. Update `VERSION` file with semantic version
2. Run `make release` to create GitHub release

The release process automatically:
- Runs tests
- Builds cross-platform binaries
- Generates checksums
- Creates git tag
- Publishes GitHub release

## Monitoring

### Build Status
- GitHub Actions (if configured) for automated builds
- Manual testing on target platforms

### User Adoption
- GitHub release download metrics
- Issue reports and feature requests

## Alerting

No automated alerting configured. Issues reported via:
- GitHub Issues
- Platform team communication channels

## Runbooks

### Common Issues

#### Authentication Failures
**Symptom**: `aphex auth login` fails with OIDC errors

**Resolution**:
1. Verify kubelogin installed: `kubelogin --version`
2. Check Dex endpoint: `curl https://dex.home.local`
3. Clear existing kubeconfig context: `kubectl config delete-context aphex`
4. Re-run: `aphex auth login`

#### Permission Denied Errors
**Symptom**: Commands fail with 403 Forbidden

**Resolution**:
1. Check current namespace: `kubectl config view --minify`
2. Verify RBAC permissions: `kubectl auth can-i create pipelines`
3. Contact platform administrator for group membership

#### Pipeline Creation Failures
**Symptom**: `aphex pipeline create` fails

**Resolution**:
1. Validate YAML file: `kubectl apply --dry-run=client -f pipeline.yaml`
2. Verify required parameters: `--repo-org` and `--repo-name` flags provided
3. Check RepoBinding CRD exists: `kubectl get crd repobindings.aphex`
4. Verify Tekton is installed: `kubectl get crd pipelines.tekton.dev`
5. Check platform-system namespace exists: `kubectl get namespace platform-system`

#### Pipeline Not Found Errors
**Symptom**: `aphex pipeline delete` reports pipeline not found

**Resolution**:
1. List all pipelines: `aphex pipeline list`
2. Check if pipeline exists in expected namespace: `kubectl get pipeline <name> -n <name>`
3. Verify RBAC permissions across namespaces: `kubectl auth can-i list pipelines --all-namespaces`

#### RepoBinding Creation Failures
**Symptom**: Pipeline created but RepoBinding creation fails

**Resolution**:
1. Check RepoBinding CRD: `kubectl get crd repobindings.aphex`
2. Verify platform-system namespace: `kubectl get namespace platform-system`
3. Check existing RepoBinding: `kubectl get repobinding <pipeline-name>-binding -n platform-system`
4. Validate GitHub org/repo names match repository structure

### Troubleshooting

#### Debug Mode
Use `--log-level=debug` or `--verbose` flag on any command for detailed output:

```bash
# Maximum verbosity with --verbose
aphex pipeline create my-pipeline --file pipeline.yaml --verbose

# Or specify debug level explicitly
aphex pipeline list --log-level=debug
```

#### Log Analysis
Check kubectl logs for related errors:
```bash
kubectl logs -n tekton-pipelines -l app=tekton-pipelines-controller
```

## Maintenance

### Version Updates
- Update `VERSION` file for new releases
- Test on all supported platforms before release
- Update installation instructions if needed

### Dependency Updates
- Regular `go mod tidy` and dependency updates
- Test with latest kubelogin versions
- Verify compatibility with platform Kubernetes versions

**Source**
- `Makefile` - Build and release automation
- `scripts/release.sh` - Release process
- `pkg/auth/login.go` - Authentication logic
- `pkg/k8s/errors.go` - Error handling and troubleshooting
