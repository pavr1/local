# Ice Cream Store Session Service Makefile
# This Makefile provides commands for managing the session service

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

# Ensure scripts are executable
.PHONY: fix-permissions
fix-permissions:
	@chmod +x scripts/*.sh

## 🍦 Ice Cream Store Session Service Management

help: ## Show this help message
	@echo "$(CYAN)🍦🔐 Ice Cream Store Session Service$(RESET)"
	@echo "====================================="
	@echo ""
	@echo "$(YELLOW)📋 Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🚀 Common Workflows:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start, test"
	@echo "  $(GREEN)make test$(RESET)             # Run unit tests"
	@echo "  $(GREEN)make logs -f$(RESET)          # Follow live logs"
	@echo ""
	@echo "$(YELLOW)📖 Prerequisites:$(RESET)"
	@echo "  • Docker and Docker Compose must be running"
	@echo "  • Data service must be running (cd ../data-service && make start)"
	@echo ""

## 🚀 Service Management Commands

start: fix-permissions ## Start the auth service container
	@echo "$(CYAN)🚀 Starting Auth Service...$(RESET)"
	@./scripts/start.sh

stop: ## Stop the auth service container
	@echo "$(YELLOW)🛑 Stopping Auth Service...$(RESET)"
	@./scripts/stop.sh

restart: stop start ## Restart the auth service
	@echo "$(GREEN)🔄 Auth Service restarted successfully!$(RESET)"

reset: fix-permissions ## Reset service (rebuild and restart)
	@echo "$(YELLOW)🔄 Resetting Auth Service...$(RESET)"
	@./scripts/reset.sh

## 📊 Monitoring & Debugging Commands

logs: fix-permissions ## View service logs
	@echo "$(CYAN)📋 Viewing Auth Service logs...$(RESET)"
	@./scripts/logs.sh

logs-follow: fix-permissions ## Follow service logs in real-time
	@echo "$(CYAN)📋 Following Auth Service logs...$(RESET)"
	@./scripts/logs.sh -f

status: ## Show container status
	@echo "$(CYAN)📊 Container Status:$(RESET)"
	@cd docker && docker-compose ps

health: ## Check service health
	@echo "$(CYAN)🏥 Checking auth service health...$(RESET)"
	@if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Auth service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Auth service is not responding$(RESET)"; \
		exit 1; \
	fi

## 🧪 Testing Commands

test: ## Run unit tests
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: fix-permissions ## Run integration tests (requires service to be running)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
	@if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then \
		./scripts/test.sh; \
	else \
		echo "$(YELLOW)⚠️  Session service is not running, skipping integration tests$(RESET)"; \
		echo "$(YELLOW)   Start the service with 'make start' to run integration tests$(RESET)"; \
	fi

test-basic: ## Run basic connectivity test
	@echo "$(CYAN)🧪 Running basic connectivity test...$(RESET)"
	@curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Basic connectivity test passed!$(RESET)" || \
		(echo "$(RED)❌ Basic connectivity test failed!$(RESET)" && exit 1)

test-login: ## Test admin login
	@echo "$(CYAN)🧪 Testing admin login...$(RESET)"
	@curl -s -X POST http://localhost:8081/api/v1/auth/login \
		-H 'Content-Type: application/json' \
		-d '{"username":"admin","password":"admin123"}' | \
		grep -q '"token"' && \
		echo "$(GREEN)✅ Admin login test passed!$(RESET)" || \
		(echo "$(RED)❌ Admin login test failed!$(RESET)" && exit 1)

## 🔧 Development Commands

build: fix-permissions ## Build the auth service container
	@echo "$(CYAN)🔨 Building Auth Service...$(RESET)"
	@cd docker && docker-compose build
	@echo "$(GREEN)✅ Build completed!$(RESET)"

build-fresh: fix-permissions ## Force rebuild the auth service container (no cache)
	@echo "$(CYAN)🔨 Force rebuilding Auth Service (no cache)...$(RESET)"
	@cd docker && docker-compose build --no-cache --force-rm
	@echo "$(GREEN)✅ Fresh build completed!$(RESET)"

dev: ## Run the service in development mode (without Docker)
	@echo "$(CYAN)🛠️  Running Auth Service in development mode...$(RESET)"
	@echo "$(YELLOW)⚠️  Make sure data-service is running first!$(RESET)"
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

install: deps start ## Complete setup: install dependencies and start service
	@echo "$(GREEN)🎉 Auth Service installation completed!$(RESET)"
	@echo "$(CYAN)📝 Service Details:$(RESET)"
	@echo "   Auth API: http://localhost:8081"
	@echo "   Health:   http://localhost:8081/api/v1/auth/health"
	@echo "   Login:    POST http://localhost:8081/api/v1/auth/login"
	@echo ""
	@echo "$(CYAN)🧪 Test Commands:$(RESET)"
	@echo "   make test                      # Run all tests"
	@echo "   make test-login               # Test admin login"
	@echo "   curl -X POST http://localhost:8081/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"

fresh: clean deps build-fresh start test status info ## Fresh start: clean, force rebuild, start, test, and show info
	@echo "$(GREEN)🎉 Fresh Auth Service setup completed!$(RESET)"

## 🧹 Cleanup Commands

clean: ## Clean up containers and images
	@echo "$(YELLOW)🧹 Cleaning up Auth Service...$(RESET)"
	@cd docker && docker-compose down -v 2>/dev/null || true
	@docker image rm session-service_session-service 2>/dev/null || true
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

clean-all: clean ## Clean up everything including Go build cache
	@echo "$(YELLOW)🧹 Deep cleaning...$(RESET)"
	@go clean -cache -modcache -testcache 2>/dev/null || true
	@echo "$(GREEN)✅ Deep cleanup completed$(RESET)"

## 📋 Information Commands

info: ## Show service information and URLs
	@echo "$(CYAN)📋 Auth Service Information$(RESET)"
	@echo "=================================="
	@echo ""
	@echo "$(YELLOW)🔗 Service URLs:$(RESET)"
	@echo "   Base URL:     http://localhost:8081"
	@echo "   Health Check: http://localhost:8081/api/v1/auth/health"
	@echo "   API Docs:     http://localhost:8081/"
	@echo ""
	@echo "$(YELLOW)🔑 API Endpoints:$(RESET)"
	@echo "   POST /api/v1/auth/login      # User login"
	@echo "   POST /api/v1/auth/logout     # User logout (requires token)"
	@echo "   POST /api/v1/auth/refresh    # Refresh token (requires token)"
	@echo "   GET  /api/v1/auth/validate   # Validate token (requires token)"
	@echo "   GET  /api/v1/auth/profile    # Get user profile (requires token)"
	@echo ""
	@echo "$(YELLOW)🧪 Test Commands:$(RESET)"
	@echo "   # Health check"
	@echo "   curl http://localhost:8081/api/v1/auth/health"
	@echo ""
	@echo "   # Login"
	@echo "   curl -X POST http://localhost:8081/api/v1/auth/login \\"
	@echo "        -H 'Content-Type: application/json' \\"
	@echo "        -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
	@echo ""

env: ## Show environment configuration
	@echo "$(CYAN)📋 Environment Configuration$(RESET)"
	@echo "====================================="
	@echo ""
	@echo "$(YELLOW)🔧 Default Configuration:$(RESET)"
	@echo "   Server Host: 0.0.0.0"
	@echo "   Server Port: 8081"
	@echo "   JWT Expiration: 10 minutes"
	@echo "   JWT Refresh Threshold: 2 minutes"
	@echo "   Database Host: postgres (container name)"
	@echo "   Database Port: 5432"
	@echo "   Database Name: icecream_store"
	@echo "   Log Level: info"
	@echo ""
	@echo "$(YELLOW)📝 Override with environment variables:$(RESET)"
	@echo "   JWT_SECRET, JWT_EXPIRATION_TIME, LOG_LEVEL, etc."
	@echo "   See config.env.example for all options"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information$(RESET)"
	@echo "========================="
	@echo "Auth Service: 1.0.0"
	@echo "Go Version: $$(go version)"
	@echo "Docker: $$(docker --version 2>/dev/null || echo 'Not available')"
	@echo "Docker Compose: $$(docker-compose --version 2>/dev/null || echo 'Not available')"

# List all targets for tab completion
.PHONY: help start stop restart reset logs logs-follow status health test test-integration test-basic test-login build dev deps lint format install fresh clean clean-all info env version fix-permissions 