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
	@echo "$(CYAN)Gateway Service - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start, test"
	@echo "  $(GREEN)make test$(RESET)             # Run unit tests"
	@echo "  $(GREEN)make dev$(RESET)              # Run in development mode"
	@echo ""
	@echo "$(GREEN)Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

## 🧪 Testing Commands

test: ## Run unit tests
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: ## Run integration tests (requires services to be running)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
	@if curl -f http://localhost:8080/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Gateway is running, running integration tests...$(RESET)"; \
		./scripts/test_all_services.sh; \
	else \
		echo "$(YELLOW)⚠️  Gateway service is not running, skipping integration tests$(RESET)"; \
		echo "$(YELLOW)   Start the service with 'make start' to run integration tests$(RESET)"; \
	fi

test-coverage: ## Run tests with coverage
	@echo "$(CYAN)🧪 Running tests with coverage...$(RESET)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

## 🔧 Development Commands

build: ## Build the gateway service binary
	@echo "$(CYAN)🔨 Building Gateway Service...$(RESET)"
	@go build -o bin/gateway-service .
	@echo "$(GREEN)✅ Build completed: bin/gateway-service$(RESET)"

run: build ## Run the gateway service locally
	@echo "$(CYAN)🚀 Running Gateway Service locally...$(RESET)"
	@./bin/gateway-service

dev: ## Run the service in development mode
	@echo "$(CYAN)🛠️  Running Gateway Service in development mode...$(RESET)"
	@echo "$(YELLOW)⚠️  Make sure all backend services are running first!$(RESET)"
	@go run *.go

start: build ## Start the gateway service (production mode)
	@echo "$(CYAN)🚀 Starting Gateway Service...$(RESET)"
	@./bin/gateway-service

stop: ## Stop the gateway service
	@echo "$(CYAN)🛑 Stopping Gateway Service...$(RESET)"
	@pkill -f gateway-service || echo "$(YELLOW)Service was not running$(RESET)"

restart: stop start ## Restart the gateway service
	@echo "$(GREEN)✅ Gateway Service restarted!$(RESET)"

clean: ## Clean build artifacts and cache
	@echo "$(CYAN)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f gateway-service local-server
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
	@echo "$(GREEN)🎉 Gateway Service is ready!$(RESET)"
	@echo "$(CYAN)Run 'make dev' to start in development mode$(RESET)"

fresh: clean deps build test ## Fresh start: clean, build, test
	@echo "$(GREEN)🎉 Fresh gateway service build completed!$(RESET)"

## 🧪 Service Testing Commands

test-session: ## Test session flow
	@echo "$(CYAN)🧪 Testing session flow...$(RESET)"
	@if [ -f ./test_session_flow.sh ]; then \
		./test_session_flow.sh; \
	else \
		echo "$(YELLOW)Session flow test script not found$(RESET)"; \
	fi

test-services: ## Test all service connections
	@echo "$(CYAN)🧪 Testing service connections...$(RESET)"
	@echo "Testing Session Service..." && curl -f http://localhost:8081/health || echo "$(RED)Session service not responding$(RESET)"
	@echo "Testing Orders Service..." && curl -f http://localhost:8083/health || echo "$(RED)Orders service not responding$(RESET)"
	@echo "Testing Inventory Service..." && curl -f http://localhost:8084/health || echo "$(RED)Inventory service not responding$(RESET)"
	@echo "Testing Data Service..." && curl -f http://localhost:8082/health || echo "$(RED)Data service not responding$(RESET)"
	@echo "Testing UI Service..." && curl -f http://localhost:3000 || echo "$(RED)UI service not responding$(RESET)"

## 📊 Information Commands

info: ## Show service information
	@echo "$(CYAN)📊 Gateway Service Information:$(RESET)"
	@echo "  Service: gateway-service"
	@echo "  Port: 8080"
	@echo "  Go Version: $(shell go version)"
	@echo "  Module: $(shell go list -m)"
	@echo ""

env: ## Show environment variables
	@echo "$(CYAN)🌍 Environment Variables:$(RESET)"
	@echo "  SESSION_SERVICE_URL: $(or $(SESSION_SERVICE_URL),not set (default: http://localhost:8081))"
	@echo "  ORDERS_SERVICE_URL: $(or $(ORDERS_SERVICE_URL),not set (default: http://localhost:8083))"
	@echo "  INVENTORY_SERVICE_URL: $(or $(INVENTORY_SERVICE_URL),not set (default: http://localhost:8084))"
	@echo "  DATA_SERVICE_URL: $(or $(DATA_SERVICE_URL),not set (default: http://localhost:8082))"
	@echo "  UI_SERVICE_URL: $(or $(UI_SERVICE_URL),not set (default: http://localhost:3000))"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information:$(RESET)"
	@echo "  Gateway Service: 1.0.0"
	@echo "  Go Version: $(shell go version)"
	@echo "  Build Date: $(shell date)"

## 🔧 Utility Commands

check-deps: ## Check if required dependencies are installed
	@echo "$(CYAN)🔍 Checking dependencies...$(RESET)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go is not installed$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Go is installed$(RESET)"

health: ## Check gateway health
	@echo "$(CYAN)🔍 Checking gateway health...$(RESET)"
	@curl -f http://localhost:8080/health || echo "$(RED)❌ Gateway is not responding$(RESET)"

## 📝 Documentation

docs: ## Generate documentation
	@echo "$(CYAN)📝 Generating documentation...$(RESET)"
	@go doc -all > docs.txt
	@echo "$(GREEN)✅ Documentation generated: docs.txt$(RESET)"

# List all targets for tab completion
.PHONY: help test test-integration test-coverage build run dev start stop restart clean deps deps-update lint format install fresh test-session test-services info env version check-deps health docs 