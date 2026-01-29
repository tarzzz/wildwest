.PHONY: help run dev build clean test test-unit test-integration test-coverage lint fmt vet migrate-up migrate-down migrate-create migrate-force migrate-version db-reset docker-build docker-up docker-down docker-logs docker-restart docker-clean install-tools

# Variables
BINARY_NAME=api
MAIN_PATH=cmd/api/main.go
MIGRATIONS_PATH=migrations
DB_URL=postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

##@ General

help: ## Display this help message
	@echo "$(BLUE)User Management API - Makefile Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(YELLOW)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(BLUE)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(GREEN)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

run: ## Run the application locally
	@echo "$(GREEN)Running application...$(NC)"
	@go run $(MAIN_PATH)

dev: ## Run with hot reload using Air
	@echo "$(GREEN)Starting development server with hot reload...$(NC)"
	@if ! command -v air > /dev/null; then \
		echo "$(YELLOW)Air not found. Installing...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@air

build: ## Build the binary
	@echo "$(GREEN)Building binary...$(NC)"
	@go build -ldflags="-w -s" -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Binary built: bin/$(BINARY_NAME)$(NC)"

clean: ## Remove build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/ tmp/
	@echo "$(GREEN)Clean complete$(NC)"

install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)Tools installed$(NC)"

##@ Testing

test: ## Run all tests
	@echo "$(GREEN)Running all tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v -short -race ./...

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	@go test -v -run Integration ./...

test-coverage: ## Generate test coverage report
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

##@ Code Quality

lint: ## Run golangci-lint
	@echo "$(GREEN)Running linter...$(NC)"
	@if ! command -v golangci-lint > /dev/null; then \
		echo "$(YELLOW)golangci-lint not found. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

##@ Database

migrate-up: ## Run all migrations
	@echo "$(GREEN)Running migrations up...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(YELLOW)migrate not found. Installing...$(NC)"; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations applied$(NC)"

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(YELLOW)migrate not found. Installing...$(NC)"; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "$(GREEN)Migration rolled back$(NC)"

migrate-create: ## Create new migration (usage: make migrate-create name=add_users)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: name parameter required$(NC)"; \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating migration: $(name)$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(YELLOW)migrate not found. Installing...$(NC)"; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo "$(GREEN)Migration files created$(NC)"

migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "$(RED)Error: version parameter required$(NC)"; \
		echo "Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration to version $(version)...$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)

migrate-version: ## Show current migration version
	@echo "$(GREEN)Current migration version:$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

db-reset: ## Drop and recreate database
	@echo "$(RED)WARNING: This will delete all data!$(NC)"
	@echo "Press Ctrl+C to cancel or Enter to continue..."
	@read confirm
	@echo "$(YELLOW)Resetting database...$(NC)"
	@psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS userdb;"
	@psql -h localhost -U postgres -c "CREATE DATABASE userdb;"
	@echo "$(GREEN)Database reset complete. Run 'make migrate-up' to apply migrations$(NC)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t user-management-api:latest .
	@echo "$(GREEN)Docker image built$(NC)"

docker-up: ## Start services with docker-compose
	@echo "$(GREEN)Starting Docker services...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Services started$(NC)"
	@echo "$(BLUE)API: http://localhost:8080$(NC)"
	@echo "$(BLUE)PostgreSQL: localhost:5432$(NC)"

docker-down: ## Stop services
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)Services stopped$(NC)"

docker-logs: ## View container logs
	@docker-compose logs -f

docker-restart: ## Restart services
	@echo "$(YELLOW)Restarting Docker services...$(NC)"
	@docker-compose restart
	@echo "$(GREEN)Services restarted$(NC)"

docker-clean: ## Remove containers, volumes, and images
	@echo "$(RED)WARNING: This will remove all containers, volumes, and images!$(NC)"
	@echo "Press Ctrl+C to cancel or Enter to continue..."
	@read confirm
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	@docker-compose down -v --rmi all
	@echo "$(GREEN)Docker cleanup complete$(NC)"

##@ Utilities

.DEFAULT_GOAL := help
