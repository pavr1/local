# Ice Cream Store Orders Service Makefile
# This Makefile provides commands for managing the orders service

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

## 🍦 Ice Cream Store Orders Service Management

help: ## Show this help message
	@echo "$(CYAN)🍦📦 Ice Cream Store Orders Service$(RESET)"
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

start: fix-permissions ## Start the orders service container
	@echo "$(CYAN)🚀 Starting Orders Service...$(RESET)"
	@./scripts/start.sh

stop: ## Stop the orders service container
	@echo "$(YELLOW)🛑 Stopping Orders Service...$(RESET)"
	@./scripts/stop.sh

restart: stop start ## Restart the orders service
	@echo "$(GREEN)🔄 Orders Service restarted successfully!$(RESET)"

reset: fix-permissions ## Reset service (rebuild and restart)
	@echo "$(YELLOW)🔄 Resetting Orders Service...$(RESET)"
	@./scripts/reset.sh

## 📊 Monitoring & Debugging Commands

logs: fix-permissions ## View service logs
	@echo "$(CYAN)📋 Viewing Orders Service logs...$(RESET)"
	@./scripts/logs.sh

logs-follow: fix-permissions ## Follow service logs in real-time
	@echo "$(CYAN)📋 Following Orders Service logs...$(RESET)"
	@./scripts/logs.sh -f

status: ## Show container status
	@echo "$(CYAN)📊 Container Status:$(RESET)"
	@cd docker && docker-compose ps

health: ## Check service health
	@echo "$(CYAN)🏥 Checking orders service health...$(RESET)"
	@if curl -f http://localhost:8083/api/v1/orders/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Orders service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Orders service is not responding$(RESET)"; \
		exit 1; \
	fi

## 🧪 Testing Commands

test: ## Run unit tests
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: fix-permissions ## Run integration tests (requires service to be running)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
	@if curl -f http://localhost:8083/api/v1/orders/health > /dev/null 2>&1; then \
		./scripts/test.sh; \
	else \
		echo "$(YELLOW)⚠️  Orders service is not running, skipping integration tests$(RESET)"; \
		echo "$(YELLOW)   Start the service with 'make start' to run integration tests$(RESET)"; \
	fi

test-basic: ## Run basic connectivity test
	@echo "$(CYAN)🧪 Running basic connectivity test...$(RESET)"
	@curl -f http://localhost:8083/api/v1/orders/health > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Basic connectivity test passed!$(RESET)" || \
		(echo "$(RED)❌ Basic connectivity test failed!$(RESET)" && exit 1)

test-orders: ## Test orders endpoint (requires auth)
	@echo "$(CYAN)🧪 Testing orders endpoint...$(RESET)"
	@if [ -z "$$AUTH_TOKEN" ]; then \
		echo "$(YELLOW)⚠️  AUTH_TOKEN not set. Get token from auth service first.$(RESET)"; \
		echo "   Run: AUTH_TOKEN=\$$(curl -s -X POST http://localhost:8081/api/v1/sessions/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}' | grep -o '\"token\":\"[^\"]*\"' | cut -d'\"' -f4)"; \
		exit 1; \
	fi
	@curl -s -H "Authorization: Bearer $$AUTH_TOKEN" http://localhost:8083/api/v1/orders | \
		grep -q '"orders"' && \
		echo "$(GREEN)✅ Orders endpoint test passed!$(RESET)" || \
		(echo "$(RED)❌ Orders endpoint test failed!$(RESET)" && exit 1)

## 🔧 Development Commands

build: fix-permissions ## Build the orders service container
	@echo "$(CYAN)🔨 Building Orders Service...$(RESET)"
	@cd docker && docker-compose build
	@echo "$(GREEN)✅ Build completed!$(RESET)"

build-fresh: fix-permissions ## Force rebuild the orders service container (no cache)
	@echo "$(CYAN)🔨 Force rebuilding Orders Service (no cache)...$(RESET)"
	@cd docker && docker-compose build --no-cache --force-rm
	@echo "$(GREEN)✅ Fresh build completed!$(RESET)"

dev: ## Run the service in development mode (without Docker)
	@echo "$(CYAN)🛠️  Running Orders Service in development mode...$(RESET)"
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
	@echo "$(GREEN)🎉 Orders Service installation completed!$(RESET)"
	@echo "$(CYAN)📝 Service Details:$(RESET)"
	@echo "   Orders API: http://localhost:8083"
	@echo "   Health:     http://localhost:8083/api/v1/orders/health"
	@echo "   List:       GET http://localhost:8083/api/v1/orders"
	@echo "   Create:     POST http://localhost:8083/api/v1/orders"
	@echo ""
	@echo "$(CYAN)🧪 Test Commands:$(RESET)"
	@echo "   make test                      # Run all tests"
	@echo "   make test-orders              # Test orders endpoint"
	@echo "   curl http://localhost:8083/api/v1/orders/health"

fresh: clean deps build-fresh start test status info ## Fresh start: clean, force rebuild, start, test, and show info
	@echo "$(GREEN)🎉 Fresh Orders Service setup completed!$(RESET)"

## 🧹 Cleanup Commands

clean: ## Clean up containers and images
	@echo "$(YELLOW)🧹 Cleaning up Orders Service...$(RESET)"
	@cd docker && docker-compose down -v 2>/dev/null || true
	@docker stop icecream_orders 2>/dev/null || true
	@docker rm icecream_orders 2>/dev/null || true
	@docker image rm orders-service_orders-service 2>/dev/null || true
	@docker image rm temp-orders:latest 2>/dev/null || true
	@docker image rm temp-orders-fixed:latest 2>/dev/null || true
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

clean-all: clean ## Clean up everything including Go build cache
	@echo "$(YELLOW)🧹 Deep cleaning...$(RESET)"
	@go clean -cache -modcache -testcache 2>/dev/null || true
	@echo "$(GREEN)✅ Deep cleanup completed$(RESET)"

## 📋 Information Commands

info: ## Show service information and URLs
	@echo "$(CYAN)📋 Orders Service Information$(RESET)"
	@echo "=================================="
	@echo ""
	@echo "$(YELLOW)🔗 Service URLs:$(RESET)"
	@echo "   Base URL:     http://localhost:8083"
	@echo "   Health Check: http://localhost:8083/api/v1/orders/health"
	@echo "   API Docs:     http://localhost:8083/"
	@echo ""
	@echo "$(YELLOW)📦 API Endpoints:$(RESET)"
	@echo "   GET  /api/v1/orders/health    # Health check"
	@echo "   GET  /api/v1/orders           # List orders (requires token)"
	@echo "   POST /api/v1/orders           # Create order (requires token)"
	@echo "   GET  /api/v1/orders/{id}      # Get order by ID (requires token)"
	@echo "   PUT  /api/v1/orders/{id}      # Update order (requires token)"
	@echo "   DELETE /api/v1/orders/{id}    # Cancel order (requires token)"
	@echo ""
	@echo "$(YELLOW)🧪 Test Commands:$(RESET)"
	@echo "   # Health check"
	@echo "   curl http://localhost:8083/api/v1/orders/health"
	@echo ""
	@echo "   # Get auth token first"
	@echo "   AUTH_TOKEN=\$$(curl -s -X POST http://localhost:8081/api/v1/sessions/login \\"
	@echo "        -H 'Content-Type: application/json' \\"
	@echo "        -d '{\"username\":\"admin\",\"password\":\"admin123\"}' | \\"
	@echo "        grep -o '\"token\":\"[^\"]*\"' | cut -d'\"' -f4)"
	@echo ""
	@echo "   # List orders"
	@echo "   curl -H \"Authorization: Bearer \$$AUTH_TOKEN\" http://localhost:8083/api/v1/orders"
	@echo ""
	@echo "   # Create order"
	@echo "   curl -X POST -H 'Content-Type: application/json' \\"
	@echo "        -H \"Authorization: Bearer \$$AUTH_TOKEN\" \\"
	@echo "        http://localhost:8083/api/v1/orders \\"
	@echo "        -d '{\"customer_id\":1,\"order_type\":\"dine_in\",\"ordered_recipes\":[{\"recipe_id\":1,\"quantity\":2}]}'"
	@echo ""

env: ## Show environment configuration
	@echo "$(CYAN)📋 Environment Configuration$(RESET)"
	@echo "====================================="
	@echo ""
	@echo "$(YELLOW)🔧 Default Configuration:$(RESET)"
	@echo "   Server Host: 0.0.0.0"
	@echo "   Server Port: 8083"
	@echo "   Database Host: postgres (container name)"
	@echo "   Database Port: 5432"
	@echo "   Database Name: icecream_store"
	@echo "   Log Level: info"
	@echo "   Default Tax Rate: 13.0%"
	@echo "   Default Service Rate: 10.0%"
	@echo "   Order Timeout: 30 minutes"
	@echo ""
	@echo "$(YELLOW)📝 Override with environment variables:$(RESET)"
	@echo "   JWT_SECRET, DEFAULT_TAX_RATE, ORDER_TIMEOUT, LOG_LEVEL, etc."
	@echo "   See config.env.example for all options"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information$(RESET)"
	@echo "========================="
	@echo "Orders Service: 1.0.0"
	@echo "Go Version: $$(go version)"
	@echo "Docker: $$(docker --version 2>/dev/null || echo 'Not available')"
	@echo "Docker Compose: $$(docker-compose --version 2>/dev/null || echo 'Not available')"

# List all targets for tab completion
.PHONY: help start stop restart reset logs logs-follow status health test test-integration test-basic test-orders build dev deps lint format install fresh clean clean-all info env version fix-permissions 