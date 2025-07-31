# Colors for output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
RESET=\033[0m

# Default target
.DEFAULT_GOAL := help

## 🚀 Quick Start Commands

help: ## Show this help message
	@echo "$(CYAN)Inventory Service - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start, test"
	@echo "  $(GREEN)make test$(RESET)             # Run unit tests"
	@echo "  $(GREEN)make logs -f$(RESET)          # Follow live logs"
	@echo ""
	@echo "$(GREEN)Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

## 🧪 Testing Commands

test: ## Run unit tests
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: ## Run integration tests (requires service to be running)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
	@if curl -f http://localhost:8084/api/v1/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Service is running, running integration tests...$(RESET)"; \
		echo "$(YELLOW)ℹ️  Integration tests not yet implemented$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  Inventory service is not running, skipping integration tests$(RESET)"; \
		echo "$(YELLOW)   Start the service with 'make start' to run integration tests$(RESET)"; \
	fi

test-coverage: ## Run tests with coverage
	@echo "$(CYAN)🧪 Running tests with coverage...$(RESET)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

## 🔧 Development Commands

build: ## Build the inventory service binary
	@echo "$(CYAN)🔨 Building Inventory Service...$(RESET)"
	@go build -o bin/inventory-service .
	@echo "$(GREEN)✅ Build completed: bin/inventory-service$(RESET)"

run: build ## Run the inventory service locally
	@echo "$(CYAN)🚀 Running Inventory Service locally...$(RESET)"
	@./bin/inventory-service

dev: ## Run the service in development mode with auto-reload
	@echo "$(CYAN)🛠️  Running Inventory Service in development mode...$(RESET)"
	@echo "$(YELLOW)⚠️  Make sure database is running first!$(RESET)"
	@go run main.go main_http_handler.go

clean: ## Clean build artifacts and cache
	@echo "$(CYAN)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "$(GREEN)✅ Clean completed!$(RESET)"

## 📦 Dependencies

deps: ## Download and tidy dependencies
	@echo "$(CYAN)📦 Managing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

deps-update: ## Update all dependencies
	@echo "$(CYAN)📦 Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

## 🔍 Code Quality

lint: ## Run linters
	@echo "$(CYAN)🔍 Running linters...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed, running go vet instead$(RESET)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✅ Linting completed!$(RESET)"

format: ## Format code
	@echo "$(CYAN)✨ Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted!$(RESET)"

## 🚀 Lifecycle Commands

install: deps build ## Complete setup and start service
	@echo "$(GREEN)🎉 Inventory Service is ready!$(RESET)"
	@echo "$(CYAN)Run 'make dev' to start in development mode$(RESET)"

fresh: clean deps build test ## Fresh start: clean, build, test
	@echo "$(GREEN)🎉 Fresh inventory service build completed!$(RESET)"

## 📊 Information Commands

info: ## Show service information
	@echo "$(CYAN)📊 Inventory Service Information:$(RESET)"
	@echo "  Service: inventory-service"
	@echo "  Port: 8084"
	@echo "  Go Version: $(shell go version)"
	@echo "  Module: $(shell go list -m)"
	@echo ""

env: ## Show environment variables
	@echo "$(CYAN)🌍 Environment Variables:$(RESET)"
	@echo "  INVENTORY_SERVER_HOST: $(or $(INVENTORY_SERVER_HOST),not set (default: 0.0.0.0))"
	@echo "  INVENTORY_SERVER_PORT: $(or $(INVENTORY_SERVER_PORT),not set (default: 8084))"
	@echo "  DB_HOST: $(or $(DB_HOST),not set (default: localhost))"
	@echo "  DB_PORT: $(or $(DB_PORT),not set (default: 5432))"
	@echo "  DB_USER: $(or $(DB_USER),not set (default: postgres))"
	@echo "  DB_NAME: $(or $(DB_NAME),not set (default: icecream_store))"
	@echo "  LOG_LEVEL: $(or $(LOG_LEVEL),not set (default: info))"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information:$(RESET)"
	@echo "  Inventory Service: 1.0.0"
	@echo "  Go Version: $(shell go version)"
	@echo "  Build Date: $(shell date)"

## 🔧 Utility Commands

check-deps: ## Check if required dependencies are installed
	@echo "$(CYAN)🔍 Checking dependencies...$(RESET)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go is not installed$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Go is installed$(RESET)"

## 📝 Documentation

docs: ## Generate documentation
	@echo "$(CYAN)📝 Generating documentation...$(RESET)"
	@go doc -all > docs.txt
	@echo "$(GREEN)✅ Documentation generated: docs.txt$(RESET)"

# List all targets for tab completion
.PHONY: help test test-integration test-coverage build run dev clean deps deps-update lint format install fresh info env version check-deps docs 