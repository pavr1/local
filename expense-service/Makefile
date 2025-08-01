# Expense Service Makefile
# Ice Cream Store Management System

# Variables
SERVICE_NAME=expense-service
BINARY_NAME=expense-service
GO_VERSION=1.23
PORT=8085

# Build variables
BUILD_DIR=bin
MAIN_FILE=main.go
GO_FILES=$(shell find . -name "*.go" -type f)

# Docker variables
DOCKER_IMAGE=icecream-expense-service
DOCKER_TAG=latest

# Database variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=icecream_store

# Test variables
TEST_COVERAGE_DIR=coverage
TEST_COVERAGE_FILE=$(TEST_COVERAGE_DIR)/coverage.out
TEST_COVERAGE_HTML=$(TEST_COVERAGE_DIR)/coverage.html

.PHONY: help build clean test test-coverage test-coverage-html run dev deps deps-update deps-verify lint format check docker-build docker-run docker-stop docker-clean logs reset start stop

## Help target
help: ## Display this help message
	@echo "Available commands for $(SERVICE_NAME):"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Build targets
build: deps ## Build the expense service binary
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## Clean build artifacts and coverage reports
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(TEST_COVERAGE_DIR)
	@echo "✅ Clean completed"

## Test targets
test: ## Run all tests
	@echo "Running tests for $(SERVICE_NAME)..."
	@go test -v ./...
	@echo "✅ Tests completed"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(TEST_COVERAGE_DIR)
	@go test -coverprofile=$(TEST_COVERAGE_FILE) ./...
	@go tool cover -func=$(TEST_COVERAGE_FILE)
	@echo "✅ Coverage analysis completed"

test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@go tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)
	@echo "✅ HTML coverage report generated: $(TEST_COVERAGE_HTML)"

## Development targets
run: build ## Build and run the service locally
	@echo "Starting $(SERVICE_NAME) on port $(PORT)..."
	@EXPENSE_SERVER_PORT=$(PORT) \
	 DB_HOST=$(DB_HOST) \
	 DB_PORT=$(DB_PORT) \
	 DB_USER=$(DB_USER) \
	 DB_PASSWORD=$(DB_PASSWORD) \
	 DB_NAME=$(DB_NAME) \
	 LOG_LEVEL=debug \
	 ./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run the service in development mode (with auto-reload)
	@echo "Starting $(SERVICE_NAME) in development mode..."
	@EXPENSE_SERVER_PORT=$(PORT) \
	 DB_HOST=$(DB_HOST) \
	 DB_PORT=$(DB_PORT) \
	 DB_USER=$(DB_USER) \
	 DB_PASSWORD=$(DB_PASSWORD) \
	 DB_NAME=$(DB_NAME) \
	 LOG_LEVEL=debug \
	 go run $(MAIN_FILE)

## Dependency targets
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "✅ Dependencies ready"

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated"

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify
	@go mod tidy
	@echo "✅ Dependencies verified"

## Code quality targets
lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi
	@echo "✅ Linting completed"

format: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

check: format lint test ## Run all code quality checks

## Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run service in Docker container
	@echo "Running $(SERVICE_NAME) in Docker..."
	@docker run -d \
		--name $(SERVICE_NAME) \
		-p $(PORT):$(PORT) \
		-e EXPENSE_SERVER_PORT=$(PORT) \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=$(DB_PORT) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		-e LOG_LEVEL=info \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ $(SERVICE_NAME) running in Docker on port $(PORT)"

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	@docker stop $(SERVICE_NAME) || true
	@docker rm $(SERVICE_NAME) || true
	@echo "✅ Docker container stopped"

docker-clean: docker-stop ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	@echo "✅ Docker cleanup completed"

## Service management targets (for integration with main project)
logs: ## Show service logs (when running via Docker)
	@echo "Showing logs for $(SERVICE_NAME)..."
	@docker logs -f $(SERVICE_NAME) 2>/dev/null || echo "Service not running in Docker"

reset: clean ## Reset the service (clean and rebuild)
	@echo "Resetting $(SERVICE_NAME)..."
	@$(MAKE) build
	@echo "✅ Reset completed"

start: docker-run ## Start the service (alias for docker-run)

stop: docker-stop ## Stop the service (alias for docker-stop)

## Health check
health: ## Check if the service is running and healthy
	@echo "Checking $(SERVICE_NAME) health..."
	@curl -f http://localhost:$(PORT)/health 2>/dev/null && echo "✅ Service is healthy" || echo "❌ Service is not responding"

## Development workflow
fresh: clean deps build test ## Fresh build (clean, deps, build, test)
	@echo "✅ Fresh build completed for $(SERVICE_NAME)"

quick: format test ## Quick development check (format and test)
	@echo "✅ Quick check completed for $(SERVICE_NAME)"

## Database connection test
db-test: ## Test database connection
	@echo "Testing database connection..."
	@EXPENSE_SERVER_PORT=$(PORT) \
	 DB_HOST=$(DB_HOST) \
	 DB_PORT=$(DB_PORT) \
	 DB_USER=$(DB_USER) \
	 DB_PASSWORD=$(DB_PASSWORD) \
	 DB_NAME=$(DB_NAME) \
	 go run -ldflags="-X main.testMode=true" $(MAIN_FILE) -test-db

## Show configuration
config: ## Show current configuration
	@echo "$(SERVICE_NAME) Configuration:"
	@echo "  Service Name: $(SERVICE_NAME)"
	@echo "  Binary Name:  $(BINARY_NAME)"
	@echo "  Port:         $(PORT)"
	@echo "  Go Version:   $(GO_VERSION)"
	@echo "  DB Host:      $(DB_HOST)"
	@echo "  DB Port:      $(DB_PORT)"
	@echo "  DB Name:      $(DB_NAME)"
	@echo "  DB User:      $(DB_USER)"

## Default target
all: fresh ## Build everything (default target)

.DEFAULT_GOAL := help 