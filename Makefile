# Aphex CLI Build System

.PHONY: build build-all clean test lint

# Build variables
VERSION ?= $(shell cat VERSION)
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Default build for current platform
build:
	go build $(LDFLAGS) -o bin/aphex ./cmd/aphex

# Cross-platform builds
build-all: clean
	# Linux x86_64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/aphex-linux-amd64 ./cmd/aphex
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/aphex-linux-arm64 ./cmd/aphex
	
	# macOS x86_64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/aphex-darwin-amd64 ./cmd/aphex
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/aphex-darwin-arm64 ./cmd/aphex
	
	# Windows x86_64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/aphex-windows-amd64.exe ./cmd/aphex

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Run property tests
test-property:
	go test ./.kiro/aphex-cli/tests/...

# Lint code
lint:
	golangci-lint run

# Install locally
install: build
	cp bin/aphex $(GOPATH)/bin/aphex

# Development build with race detection
dev:
	go build -race $(LDFLAGS) -o bin/aphex-dev ./cmd/aphex

# Release process (local testing)
release-local: test
	@echo "Creating local release for version $(VERSION)"
	@goreleaser release --snapshot --clean

# Release process (GitHub - creates tag and triggers GitHub Actions)
release: test build-all
	@echo "Creating GitHub release for version $(VERSION)"
	@./scripts/release.sh
