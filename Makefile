# Project variables
PROJECT_NAME := cdncert
BINARY_NAME := cdncert
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Go build flags
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
GOFLAGS := -trimpath

# Directories
BIN_DIR := bin
DIST_DIR := dist
CERT_DIR := certificates

# Platforms
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64
OS_ARCH := $(shell go env GOOS)-$(shell go env GOARCH)

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: all
all: clean build

.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary for current platform
	@echo "$(BLUE)Building $(BINARY_NAME) for $(OS_ARCH)...$(NC)"
	@mkdir -p $(BIN_DIR)/$(OS_ARCH)
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) \
		go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(OS_ARCH)/$(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BIN_DIR)/$(OS_ARCH)/$(BINARY_NAME)$(NC)"

.PHONY: build-all
build-all: ## Build binaries for all platforms
	@echo "$(BLUE)Building for all platforms...$(NC)"
	@for platform in $(PLATFORMS); do \
		os=$${platform%-*}; \
		arch=$${platform#*-}; \
		output=$(BIN_DIR)/$$platform/$(BINARY_NAME); \
		if [ "$$os" = "windows" ]; then \
			output=$${output}.exe; \
		fi; \
		echo "Building for $$os/$$arch..."; \
		mkdir -p $(BIN_DIR)/$$platform; \
		GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) $(LDFLAGS) -o $$output . || exit 1; \
	done
	@echo "$(GREEN)Multi-platform build complete$(NC)"
	@echo "$(BLUE)Binaries available in:$(NC)"
	@ls -lh $(BIN_DIR)/*/*

.PHONY: install
install: ## Install the binary to $GOBIN
	@echo "$(BLUE)Installing $(BINARY_NAME) to $(GOBIN)...$(NC)"
	go install $(GOFLAGS) $(LDFLAGS) .
	@echo "$(GREEN)Installed: $(GOBIN)/$(BINARY_NAME)$(NC)"

.PHONY: release
release: ## Create a release using goreleaser
	@echo "$(BLUE)Creating release...$(NC)"
	goreleaser release --clean
	@echo "$(GREEN)Release complete$(NC)"

.PHONY: snapshot
snapshot: ## Create a snapshot release
	@echo "$(BLUE)Creating snapshot release...$(NC)"
	goreleaser release --snapshot --clean
	@echo "$(GREEN)Snapshot release complete$(NC)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf $(BIN_DIR) $(DIST_DIR)
	@echo "$(GREEN)Clean complete$(NC)"

.PHONY: clean-all
clean-all: clean ## Clean all artifacts including certificates
	@echo "$(BLUE)Cleaning certificates...$(NC)"
	@rm -rf $(CERT_DIR)
	@echo "$(GREEN)All artifacts cleaned$(NC)"

.PHONY: test
test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete$(NC)"

.PHONY: coverage
coverage: test ## Generate and view test coverage
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go tool cover -html=coverage.out
	@echo "$(GREEN)Coverage report opened$(NC)"

.PHONY: lint
lint: ## Run linters
	@echo "$(BLUE)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)golangci-lint not found. Install from https://golangci-lint.run/$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Linting complete$(NC)"

.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Formatting complete$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)Vetting complete$(NC)"

.PHONY: deps
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)Dependencies downloaded$(NC)"

.PHONY: tidy
tidy: ## Tidy up go.mod and go.sum
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	go mod tidy
	@echo "$(GREEN)Dependencies tidied$(NC)"

.PHONY: verify
verify: fmt vet lint test ## Run all verification steps (fmt, vet, lint, test)
	@echo "$(GREEN)All verifications passed$(NC)"

.PHONY: dev
dev: clean build ## Clean and build for development
	@echo "$(GREEN)Development build complete$(NC)"

.PHONY: run
run: build ## Build and run the application
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	$(BIN_DIR)/$(OS_ARCH)/$(BINARY_NAME)

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version:$(NC) $(VERSION)"
	@echo "$(BLUE)Build Time:$(NC) $(BUILD_TIME)"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(PROJECT_NAME):$(VERSION) .
	@echo "$(GREEN)Docker image built: $(PROJECT_NAME):$(VERSION)$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run --rm -it $(PROJECT_NAME):$(VERSION)

# List built binaries
.PHONY: list
list: ## List all built binaries
	@echo "$(BLUE)Built binaries:$(NC)"
	@if [ -d $(BIN_DIR) ]; then \
		find $(BIN_DIR) -type f -name "$(BINARY_NAME)*" -exec ls -lh {} \; | awk '{print "  " $$9 " (" $$5 ")"}'; \
	else \
		echo "  $(RED)No binaries found. Run 'make build' or 'make build-all'$(NC)"; \
	fi

# Example usage targets
.PHONY: example-dns
example-dns: build ## Example: Obtain certificate using DNS challenge
	@echo "$(BLUE)Example: DNS challenge$(NC)"
	@echo "$(BIN_DIR)/$(OS_ARCH)/$(BINARY_NAME) -domain example.com -email admin@example.com \\"
	@echo "  -dns-access-key YOUR_KEY -dns-secret-key YOUR_SECRET \\"
	@echo "  -cdn-access-key YOUR_KEY -cdn-secret-key YOUR_SECRET \\"
	@echo "  -challenge dns"

.PHONY: example-s3
example-s3: build ## Example: Obtain certificate using S3 challenge
	@echo "$(BLUE)Example: S3 challenge$(NC)"
	@echo "$(BIN_DIR)/$(OS_ARCH)/$(BINARY_NAME) -domain example.com -email admin@example.com \\"
	@echo "  -cdn-access-key YOUR_KEY -cdn-secret-key YOUR_SECRET \\"
	@echo "  -s3-bucket YOUR_BUCKET -s3-endpoint YOUR_ENDPOINT \\"
	@echo "  -s3-region YOUR_REGION -challenge s3"
