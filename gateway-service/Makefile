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
	@echo "  $(GREEN)make start$(RESET)            # Build latest code and start service"
	@echo "  $(GREEN)make dev$(RESET)              # Run in development mode"
	@echo ""
	@echo "$(YELLOW)📖 Prerequisites:$(RESET)"
	@echo "  • Go 1.21+ must be installed"
	@echo "  • Data service should be running for full functionality"
	@echo "  • Session service should be running for full functionality"
	@echo ""

## 🚀 Service Management Commands

start: ## Build and start the gateway service using Docker (always uses latest code)
	@echo "$(CYAN)🚀 Starting Gateway Service...$(RESET)"
	@chmod +x scripts/start.sh
	@./scripts/start.sh

stop: ## Stop the gateway service containers
	@echo "$(YELLOW)🛑 Stopping Gateway Service...$(RESET)"
	@chmod +x scripts/stop.sh
	@./scripts/stop.sh

restart: stop start ## Restart the gateway service (with fresh build)
	@echo "$(GREEN)🔄 Gateway Service restarted successfully!$(RESET)"

reset: ## Reset the gateway service (clean restart)
	@echo "$(CYAN)🔄 Resetting Gateway Service...$(RESET)"
	@chmod +x scripts/reset.sh
	@./scripts/reset.sh

logs: ## Show gateway service logs
	@chmod +x scripts/logs.sh
	@./scripts/logs.sh

logs-follow: ## Follow gateway service logs
	@chmod +x scripts/logs.sh
	@./scripts/logs.sh -f

## 📊 Monitoring & Debugging Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking gateway service health...$(RESET)"
	@if curl -f http://localhost:8082/api/hello > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Gateway service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Gateway service is not responding$(RESET)"; \
		exit 1; \
	fi

test: ## Run comprehensive tests
	@echo "$(CYAN)🧪 Running Gateway Service tests...$(RESET)"
	@chmod +x scripts/test.sh
	@./scripts/test.sh

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

install: deps build-docker ## Complete setup: install dependencies and build Docker image
	@echo "$(GREEN)🎉 Gateway Service installation completed!$(RESET)"
	@echo "$(CYAN)📝 Service Details:$(RESET)"
	@echo "   Gateway API: http://localhost:8082"
	@echo "   Health:      http://localhost:8082/api/hello"
	@echo "   Hello API:   http://localhost:8082/api/hello"
	@echo ""
	@echo "$(CYAN)🧪 Test Commands:$(RESET)"
	@echo "   make test                      # Run all tests"
	@echo "   curl http://localhost:8082/api/hello"

fresh: clean deps build-docker-fresh start health test info ## Fresh start: clean, force rebuild Docker image, start, and test
	@echo "$(GREEN)🎉 Fresh Gateway Service setup completed!$(RESET)"
	@echo ""
	@echo "$(CYAN)📊 Gateway Service Status:$(RESET)"
	@echo "   🌐 Gateway API: $(GREEN)http://localhost:8082$(RESET)"
	@echo "   🔌 Health Check: $(GREEN)http://localhost:8082/api/hello$(RESET)"
	@echo "   🔐 Session Endpoints: $(GREEN)http://localhost:8082/api/v1/auth/*$(RESET)"
	@echo "   🛒 Orders Proxy: $(GREEN)http://localhost:8082/api/v1/orders/*$(RESET)"

build-docker: ## Build the Docker image
	@echo "$(CYAN)🔨 Building Gateway Service Docker image...$(RESET)"
	@cd docker && docker-compose build

build-docker-fresh: ## Force rebuild the Docker image (no cache)
	@echo "$(CYAN)🔨 Force rebuilding Gateway Service Docker image (no cache)...$(RESET)"
	@cd docker && docker-compose build --no-cache --force-rm
	@echo "$(GREEN)✅ Fresh build completed!$(RESET)"

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
	@echo "   Base URL:     http://localhost:8082"
	@echo "   Health Check: http://localhost:8082/api/health"
	@echo "   API Docs:     http://localhost:8082/api/"
	@echo ""
	@echo "$(YELLOW)🔑 API Endpoints:$(RESET)"
	@echo "   GET  /api/health             # Gateway health (checks all services)"
	@echo "   GET  /api/hello              # Hello endpoint"
	@echo "   POST /api/hello              # Hello with name"
	@echo "   POST /api/v1/auth/login      # Login (creates session)"
	@echo "   POST /api/v1/auth/logout     # Logout (revokes session)"
	@echo "   ALL  /api/v1/orders/*        # Orders proxy (session required)"
	@echo "   ALL  /api/v1/sessions/*      # Session management (session required)"
	@echo ""
	@echo "$(YELLOW)🧪 Test Commands:$(RESET)"
	@echo "   # Health check"
	@echo "   curl http://localhost:8082/api/health"
	@echo ""
	@echo "   # Hello endpoint"
	@echo "   curl http://localhost:8082/api/hello"
	@echo ""
	@echo "   # Login test"
	@echo "   curl -X POST http://localhost:8082/api/v1/auth/login \\"
	@echo "        -H 'Content-Type: application/json' \\"
	@echo "        -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
	@echo ""

version: ## Show version information
	@echo "$(CYAN)📋 Version Information$(RESET)"
	@echo "========================="
	@echo "Gateway Service: 1.0.0"
	@echo "Go Version: $$(go version)"

# List all targets for tab completion
.PHONY: help start stop restart health test build dev deps lint format install fresh clean clean-all info version 