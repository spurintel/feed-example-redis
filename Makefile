# Build variables
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%d 2>/dev/null || echo "unknown")
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Build flags
LDFLAGS := -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.Date=$(DATE)'
BUILD_FLAGS := -trimpath -a -ldflags "$(LDFLAGS)"

# Docker image names
LOCAL_IMAGE := spurintelligence/spurredis:local
DEV_IMAGE := spurintelligence/spurredis:dev
PROD_IMAGE := spurintelligence/spurredis:latest

# Platform detection for Docker builds
DOCKER_PLATFORM := $(shell uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')

# Declare phony targets
.PHONY: all help deps format lint test clean
.PHONY: bin bin-linux build-local
.PHONY: run run-local run-prod
.PHONY: publish-docker publish-docker-dev

# Default target
all: format lint test bin

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

format: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@go vet ./...

test: ## Run tests
	@echo "Running tests..."
	@go clean -testcache
	@go test ./internal/...
	@go test ./cmd/...

##@ Build

bin: deps ## Build binary for current platform
	@echo "Building backend version: $(VERSION), commit: $(COMMIT), date: $(DATE)"
	@mkdir -p target
	@go build $(BUILD_FLAGS) -o target/spurredis_$(GOOS)_$(GOARCH) ./cmd/spurredis/

bin-linux: deps ## Build Linux binary
	@echo "Building Linux binary version: $(VERSION), commit: $(COMMIT), date: $(DATE)"
	@mkdir -p target
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o target/spurredis_linux_amd64 ./cmd/spurredis/

build-local: ## Build local Docker image for development
	@echo "Building local Docker image for platform: linux/$(DOCKER_PLATFORM)..."
	@docker build \
		--platform linux/$(DOCKER_PLATFORM) \
		-t $(LOCAL_IMAGE) \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=$(DOCKER_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		.

##@ Run

run: run-local ## Run with local development (default)

run-local: build-local ## Run with local development image
	@echo "Starting local development environment..."
	@docker compose -f compose.yaml -f compose.local.yaml up

run-prod: ## Run with production image
	@echo "Starting production environment..."
	@docker compose up

##@ Clean

clean: ## Clean up build artifacts and local Docker image
	@echo "Cleaning up..."
	@rm -rf target/
	@docker rmi $(LOCAL_IMAGE) 2>/dev/null || true

##@ Publish

publish-docker-dev: ## Publish development Docker image
	@echo "Publishing development Docker image..."
	@docker buildx inspect spurredis-builder >/dev/null 2>&1 || docker buildx create --name spurredis-builder --use
	@docker buildx build --builder spurredis-builder \
		--platform=linux/amd64,linux/arm64 \
		-t $(DEV_IMAGE) \
		--push \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		.

publish-docker: ## Publish production Docker image
	@echo "Publishing production Docker image..."
	@docker buildx inspect spurredis-builder >/dev/null 2>&1 || docker buildx create --name spurredis-builder --use
	@docker buildx build --builder spurredis-builder \
		--platform=linux/amd64,linux/arm64 \
		--push \
		-t $(PROD_IMAGE) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		.