# Ice Cream Store Gateway Service Makefile
# This Makefile provides commands for managing the gateway service

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := gateway-service
BUILD_DIR := bin

## 🍦 Ice Cream Store Gateway Service Management

help: ## Show this help message
	@echo "$(CYAN)🍦🌐 Ice Cream Store Gateway Service$(RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(YELLOW)📋 Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🚀 Common Workflows:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start"
	@echo "  $(GREEN)make dev$(RESET)              # Run in development mode"
	@echo ""
	@echo "$(YELLOW)📖 Prerequisites:$(RESET)"
	@echo "  • Go 1.21+ must be installed"
	@echo "  • Data service should be running for full functionality"
	@echo "  • Auth service should be running for full functionality"
	@echo ""

## 🚀 Service Management Commands

start: build ## Start the gateway service
	@echo "$(CYAN)🚀 Starting Gateway Service...$(RESET)"
	@echo "$(YELLOW)📡 Gateway will be available at http://localhost:8080$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

stop: ## Stop the gateway service (if running in background)
	@echo "$(YELLOW)🛑 Stopping Gateway Service...$(RESET)"
	@pkill -f "$(BINARY_NAME)" 2>/dev/null || echo "$(YELLOW)No running gateway service found$(RESET)"

restart: stop start ## Restart the gateway service
	@echo "$(GREEN)🔄 Gateway Service restarted successfully!$(RESET)"

## 📊 Monitoring & Debugging Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking gateway service health...$(RESET)"
	@if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Gateway service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Gateway service is not responding$(RESET)"; \
		exit 1; \
	fi

test: ## Run basic connectivity test
	@echo "$(CYAN)🧪 Running Gateway Service tests...$(RESET)"
	@curl -f http://localhost:8080/api/health > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Health check test passed!$(RESET)" || \
		(echo "$(RED)❌ Health check test failed!$(RESET)" && exit 1)
	@curl -f http://localhost:8080/api/hello > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Hello endpoint test passed!$(RESET)" || \
		(echo "$(RED)❌ Hello endpoint test failed!$(RESET)" && exit 1)

## 🔧 Development Commands

build: deps ## Build the gateway service binary
	@echo "$(CYAN)🔨 Building Gateway Service...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

dev: ## Run the service in development mode
	@echo "$(CYAN)🛠️  Running Gateway Service in development mode...$(RESET)"
	@echo "$(YELLOW)📡 Service will be available at http://localhost:8080$(RESET)"
	@go run main.go

deps: ## Install Go dependencies
	@echo "$(CYAN)📦 Installing Go dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(RESET)"

lint: ## Run linter
	@echo "$(CYAN)🔍 Running linter...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✅ Linting completed$(RESET)"

format: ## Format Go code
	@echo "$(CYAN)✨ Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(RESET)"

## 📦 Setup & Installation Commands

install: deps build ## Complete setup: install dependencies and build service
	@echo "$(GREEN)🎉 Gateway Service installation completed!$(RESET)"
	@echo "$(CYAN)📝 Service Details:$(RESET)"
	@echo "   Gateway API: http://localhost:8080"
	@echo "   Health:      http://localhost:8080/api/health"
	@echo "   Hello API:   http://localhost:8080/api/hello"
	@echo ""
	@echo "$(CYAN)🧪 Test Commands:$(RESET)"
	@echo "   make test                      # Run all tests"
	@echo "   curl http://localhost:8080/api/health"

fresh: clean deps build test info ## Fresh start: clean, build, test, and show info
	@echo "$(GREEN)🎉 Fresh Gateway Service setup completed!$(RESET)"
	@echo "$(YELLOW)💡 Service binary is ready. Run 'make start' in a separate terminal to launch the service$(RESET)"
	@echo "$(CYAN)🧪 Running tests to verify the binary works...$(RESET)"
	@echo "$(YELLOW)Note: Gateway tests require the service to be started manually$(RESET)"

## 🧹 Cleanup Commands

clean: ## Clean up build artifacts
	@echo "$(YELLOW)🧹 Cleaning up Gateway Service...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

clean-all: clean ## Clean up everything including Go build cache
	@echo "$(YELLOW)🧹 Deep cleaning...$(RESET)"
	@go clean -cache -modcache -testcache 2>/dev/null || true
	@echo "$(GREEN)✅ Deep cleanup completed$(RESET)"

## 📋 Information Commands

info: ## Show service information and URLs
	@echo "$(CYAN)📋 Gateway Service Information$(RESET)"
	@echo "================================="
	@echo ""
	@echo "$(YELLOW)🔗 Service URLs:$(RESET)"
	@echo "   Base URL:     http://localhost:8080"
	@echo "   Health Check: http://localhost:8080/api/health"
	@echo "   API Docs:     http://localhost:8080/api/"
	@echo ""
	@echo "$(YELLOW)🔑 API Endpoints:$(RESET)"
	@echo "   GET  /api/health             # Health check"
	@echo "   GET  /api/hello              # Hello endpoint"
	@echo "   POST /api/hello              # Hello with name"
	@echo ""
	@echo "$(YELLOW)🧪 Test Commands:$(RESET)"
	@echo "   # Health check"
	@echo "   curl http://localhost:8080/api/health"
	@echo ""
	@echo "   # Hello endpoint"
	@echo "   curl http://localhost:8080/api/hello"
	@echo ""
	@echo "   # Hello with name"
	@echo "   curl -X POST http://localhost:8080/api/hello \\"
	@echo "        -H 'Content-Type: application/json' \\"
	@echo "        -d '{\"name\":\"World\"}'"
	@echo ""

version: ## Show version information
	@echo "$(CYAN)📋 Version Information$(RESET)"
	@echo "========================="
	@echo "Gateway Service: 1.0.0"
	@echo "Go Version: $$(go version)"

# List all targets for tab completion
.PHONY: help start stop restart health test build dev deps lint format install fresh clean clean-all info version 