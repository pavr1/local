# Ice Cream Store Inventory Service Makefile
# This Makefile provides commands for managing the inventory service

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

## 🍦 Ice Cream Store Inventory Service Management

help: ## Show this help message
	@echo "$(CYAN)🍦📦 Ice Cream Store Inventory Service$(RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(YELLOW)📋 Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🚀 Common Workflows:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start, test"
	@echo "  $(GREEN)make test$(RESET)             # Run comprehensive API tests"
	@echo "  $(GREEN)make logs -f$(RESET)          # Follow live logs"
	@echo ""
	@echo "$(YELLOW)📖 Prerequisites:$(RESET)"
	@echo "  • Docker and Docker Compose must be running"
	@echo "  • Data service must be running (cd ../data-service && make start)"
	@echo ""

## 🚀 Service Management Commands

start: fix-permissions ## Start the inventory service container
	@echo "$(CYAN)🚀 Starting Inventory Service...$(RESET)"
	@./scripts/start.sh

stop: ## Stop the inventory service container
	@echo "$(YELLOW)🛑 Stopping Inventory Service...$(RESET)"
	@./scripts/stop.sh

restart: stop start ## Restart the inventory service

logs: ## Show service logs (use -f to follow)
	@echo "$(CYAN)📋 Showing Inventory Service Logs...$(RESET)"
	@./scripts/logs.sh

status: ## Check service status and health
	@echo "$(CYAN)📊 Inventory Service Status$(RESET)"
	@echo "=========================="
	@if docker ps | grep -q "inventory-service"; then \
		echo "$(GREEN)✅ Service is running$(RESET)"; \
		echo ""; \
		echo "$(YELLOW)Container Status:$(RESET)"; \
		docker ps --filter "name=inventory-service" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"; \
		echo ""; \
		echo "$(YELLOW)Health Check:$(RESET)"; \
		if curl -f -s http://localhost:8082/api/v1/inventory/health > /dev/null 2>&1; then \
			echo "$(GREEN)✅ Service is healthy$(RESET)"; \
		else \
			echo "$(RED)❌ Service is not responding$(RESET)"; \
		fi; \
	else \
		echo "$(RED)❌ Service is not running$(RESET)"; \
	fi

## 🧪 Testing Commands

test: fix-permissions ## Run API tests against the service
	@echo "$(CYAN)🧪 Running Inventory Service Tests...$(RESET)"
	@./scripts/test.sh

health: ## Quick health check
	@echo "$(CYAN)🏥 Health Check$(RESET)"
	@if curl -f -s http://localhost:8082/api/v1/inventory/health; then \
		echo ""; \
		echo "$(GREEN)✅ Service is healthy$(RESET)"; \
	else \
		echo "$(RED)❌ Service is not responding$(RESET)"; \
		exit 1; \
	fi

## 🛠️ Development Commands

build: fix-permissions ## Build the service binary locally
	@echo "$(CYAN)🔨 Building Inventory Service...$(RESET)"
	@go mod tidy
	@go build -o inventory-service .
	@echo "$(GREEN)✅ Build completed$(RESET)"

clean: ## Clean up containers and images
	@echo "$(YELLOW)🧹 Cleaning up...$(RESET)"
	@./scripts/reset.sh

fresh: clean start test ## Fresh start: clean, build, start, and test

## 📦 Docker Commands

docker-build: fix-permissions ## Build Docker image
	@echo "$(CYAN)📦 Building Docker Image...$(RESET)"
	@cd docker && docker-compose build --no-cache

docker-shell: ## Open shell in running container
	@echo "$(CYAN)🐚 Opening shell in inventory service container...$(RESET)"
	@docker exec -it inventory-service /bin/sh

docker-logs: logs ## Alias for logs command

## 📊 Monitoring Commands

ps: ## Show running containers
	@echo "$(CYAN)📊 Running Containers$(RESET)"
	@docker ps --filter "name=inventory-service" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.RunningFor}}"

top: ## Show container resource usage
	@echo "$(CYAN)📊 Container Resource Usage$(RESET)"
	@docker stats inventory-service --no-stream

## 🗄️ Database Commands

db-status: ## Check database connection
	@echo "$(CYAN)🗄️ Database Connection Status$(RESET)"
	@if docker exec inventory-service wget --no-verbose --tries=1 --spider http://localhost:8082/api/v1/inventory/health 2>/dev/null; then \
		echo "$(GREEN)✅ Database connection is working$(RESET)"; \
	else \
		echo "$(RED)❌ Database connection failed$(RESET)"; \
	fi

## 🎯 Quick Actions

install: start ## Complete installation and startup
	@echo "$(GREEN)✅ Inventory Service installation completed$(RESET)"

uninstall: clean ## Complete uninstallation
	@echo "$(GREEN)✅ Inventory Service uninstalled$(RESET)"

reset: clean ## Alias for clean command

## 📈 API Testing

api-test: ## Test specific API endpoints
	@echo "$(CYAN)🔍 Testing API Endpoints...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Testing Health Endpoint:$(RESET)"
	@curl -s http://localhost:8082/api/v1/inventory/health | jq '.' || echo "Service not responding"
	@echo ""
	@echo "$(YELLOW)Testing Service Info:$(RESET)"
	@curl -s http://localhost:8082/ | jq '.' || echo "Service not responding"

suppliers-test: ## Test suppliers endpoints
	@echo "$(CYAN)👥 Testing Suppliers Endpoints...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Listing suppliers:$(RESET)"
	@curl -s http://localhost:8082/api/v1/suppliers | jq '.' || echo "Service not responding"

ingredients-test: ## Test ingredients endpoints
	@echo "$(CYAN)🥄 Testing Ingredients Endpoints...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Listing ingredients:$(RESET)"
	@curl -s http://localhost:8082/api/v1/ingredients | jq '.' || echo "Service not responding" 