.PHONY: build install test clean run help

# Binary name
BINARY_NAME=wildwest

# Build variables
BUILD_DIR=./bin
MAIN_FILE=./main.go

# Version information
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
VERSION=dev
LDFLAGS=-ldflags "-X github.com/tarzzz/wildwest/cmd.Version=$(VERSION) -X github.com/tarzzz/wildwest/cmd.GitCommit=$(GIT_COMMIT)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@codesign --force --deep --sign - --options=runtime $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || true
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@go install $(LDFLAGS)
	@codesign --force --deep --sign - --options=runtime $(shell go env GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || true
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

run: build ## Build and run the application
	@$(BUILD_DIR)/$(BINARY_NAME)

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run formatters and linters

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

all: clean deps lint test build ## Run all checks and build
