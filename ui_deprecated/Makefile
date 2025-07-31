# UI Service Makefile
# Ice Cream Store - UI Service Management

# Service configuration
SERVICE_NAME = ui-service
CONTAINER_NAME = icecream_ui
DOCKER_IMAGE = docker-ui-service
PORT = 3000

# Colors for output
BLUE = \033[0;34m
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: help build start stop restart logs clean fresh info test health status

# Default target
help: ## Show this help message
	@echo "🍦 Ice Cream Store - UI Service"
	@echo "Available commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-12s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the UI service Docker image
	@echo -e "$(BLUE)🔨 Building UI Service...$(NC)"
	cd docker && docker-compose build --no-cache

start: ## Start the UI service
	@./scripts/start.sh

stop: ## Stop the UI service
	@./scripts/stop.sh

restart: ## Restart the UI service (stop + start)
	@echo -e "$(YELLOW)🔄 Restarting UI Service...$(NC)"
	@$(MAKE) stop
	@$(MAKE) start

logs: ## View service logs
	@./scripts/logs.sh

clean: ## Clean up containers and images
	@echo -e "$(YELLOW)🧹 Cleaning up UI Service...$(NC)"
	cd docker && docker-compose down --rmi all --volumes --remove-orphans
	@echo -e "$(GREEN)✅ Cleanup complete!$(NC)"

fresh: ## Fresh start (stop + clean + build + start)
	@echo -e "$(BLUE)🔄 Fresh start for UI Service...$(NC)"
	@$(MAKE) stop
	@$(MAKE) clean
	@$(MAKE) build
	@$(MAKE) start

info: ## Show service information
	@echo -e "$(BLUE)📊 UI Service Information:$(NC)"
	@echo "  Service: $(SERVICE_NAME)"
	@echo "  Container: $(CONTAINER_NAME)"
	@echo "  Port: $(PORT)"
	@echo "  URL: http://localhost:$(PORT)"
	@echo ""
	@echo -e "$(BLUE)🐳 Docker Status:$(NC)"
	@docker ps --filter "name=$(CONTAINER_NAME)" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

test: ## Run health check test
	@echo -e "$(BLUE)🧪 Testing UI Service...$(NC)"
	@curl -f http://localhost:$(PORT)/ > /dev/null 2>&1 && \
		echo -e "$(GREEN)✅ UI Service is responding$(NC)" || \
		echo -e "$(RED)❌ UI Service is not responding$(NC)"

health: ## Check service health
	@echo -e "$(BLUE)🏥 Checking UI Service Health...$(NC)"
	@docker exec $(CONTAINER_NAME) wget --no-verbose --tries=1 --spider http://localhost:$(PORT)/ 2>/dev/null && \
		echo -e "$(GREEN)✅ UI Service is healthy$(NC)" || \
		echo -e "$(RED)❌ UI Service is unhealthy$(NC)"

status: ## Show detailed service status
	@echo -e "$(BLUE)📊 UI Service Status:$(NC)"
	@echo ""
	@echo -e "$(BLUE)Container Status:$(NC)"
	@docker ps -a --filter "name=$(CONTAINER_NAME)" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"
	@echo ""
	@echo -e "$(BLUE)Service Health:$(NC)"
	@$(MAKE) test

# Development commands
dev: ## Start in development mode (with file watching)
	@echo -e "$(BLUE)🚀 Starting UI Service in development mode...$(NC)"
	@echo "Note: Currently using production nginx. For development with hot reload, consider using a dev server."
	@$(MAKE) start 