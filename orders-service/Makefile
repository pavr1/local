# Ice Cream Store - Orders Service Makefile
# Gestión de Ingresos - Órdenes y Recetas

# Colors for output
RED     := \033[31m
GREEN   := \033[32m
YELLOW  := \033[33m
BLUE    := \033[34m
MAGENTA := \033[35m
CYAN    := \033[36m
WHITE   := \033[37m
RESET   := \033[0m

# Project configuration
SERVICE_NAME := orders-service
BINARY_NAME := orders-service
DOCKER_IMAGE := docker-orders-service
DOCKER_CONTAINER := icecream_orders
PORT := 8083

# Build configuration
BUILD_DIR := bin
MAIN_FILE := main.go

# Environment files
ENV_FILE := config.env
ENV_EXAMPLE := config.env.example

.PHONY: help build clean deps test start stop status info logs fresh docker-build docker-run docker-stop docker-clean

## 🎯 Main Commands
help: ## Show this help message
	@echo "$(CYAN)🍦 Ice Cream Store - Orders Service$(RESET)"
	@echo "$(YELLOW)Income Management - Orders and Recipes$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "$(BOLD)Usage: make [target]$(RESET)\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""

## 🔨 Development Commands

deps: ## Install Go dependencies
	@echo "$(BLUE)📦 Installing Go dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(RESET)"

build: deps ## Build the orders service binary
	@echo "$(BLUE)🔨 Building Orders Service...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

clean: ## Clean build artifacts
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "$(GREEN)✅ Clean completed$(RESET)"

test: ## Run tests
	@echo "$(BLUE)🧪 Running tests...$(RESET)"
	@go test -v ./...
	@echo "$(GREEN)✅ Tests completed$(RESET)"

## 🚀 Service Management

start: build ## Start the orders service
	@echo "$(GREEN)🚀 Starting Orders Service...$(RESET)"
	@echo "$(CYAN)📡 Service will be available at: http://localhost:$(PORT)$(RESET)"
	@echo "$(CYAN)🏥 Health check: http://localhost:$(PORT)/api/v1/orders/health$(RESET)"
	@echo "$(YELLOW)💡 Press Ctrl+C to stop the service$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

stop: ## Stop the orders service
	@echo "$(YELLOW)🛑 Stopping Orders Service...$(RESET)"
	@pkill -f $(BINARY_NAME) || true
	@echo "$(GREEN)✅ Orders Service stopped$(RESET)"

status: ## Check orders service status
	@echo "$(CYAN)📊 Orders Service Status:$(RESET)"
	@if pgrep -f $(BINARY_NAME) > /dev/null; then \
		echo "$(GREEN)✅ Service is running$(RESET)"; \
		echo "$(CYAN)📡 Available at: http://localhost:$(PORT)$(RESET)"; \
	else \
		echo "$(RED)❌ Service is not running$(RESET)"; \
	fi

info: ## Show service information
	@echo "$(CYAN)📋 Orders Service Information:$(RESET)"
	@echo "$(WHITE)Service Name:$(RESET) $(SERVICE_NAME)"
	@echo "$(WHITE)Binary Name:$(RESET) $(BINARY_NAME)"
	@echo "$(WHITE)Port:$(RESET) $(PORT)"
	@echo "$(WHITE)Build Directory:$(RESET) $(BUILD_DIR)"
	@echo ""
	@echo "$(CYAN)🔗 Service Endpoints:$(RESET)"
	@echo "  $(GREEN)Health:$(RESET) http://localhost:$(PORT)/api/v1/orders/health"
	@echo "  $(GREEN)Orders:$(RESET) http://localhost:$(PORT)/api/v1/orders"
	@echo "  $(GREEN)Create Order:$(RESET) POST http://localhost:$(PORT)/api/v1/orders"
	@echo "  $(GREEN)Get Order:$(RESET) GET http://localhost:$(PORT)/api/v1/orders/{id}"
	@echo "  $(GREEN)Update Order:$(RESET) PUT http://localhost:$(PORT)/api/v1/orders/{id}"
	@echo "  $(GREEN)Cancel Order:$(RESET) POST http://localhost:$(PORT)/api/v1/orders/{id}/cancel"
	@echo "  $(GREEN)Order Summary:$(RESET) GET http://localhost:$(PORT)/api/v1/orders/summary"
	@echo "  $(GREEN)Payment Stats:$(RESET) GET http://localhost:$(PORT)/api/v1/orders/stats/payment-methods"
	@echo ""
	@echo "$(CYAN)🔐 Authentication Required:$(RESET)"
	@echo "  $(YELLOW)⚠️ All endpoints (except health) require JWT token$(RESET)"
	@echo "  $(YELLOW)⚠️ Use Authorization: Bearer <token> header$(RESET)"
	@echo ""
	@echo "$(CYAN)🎯 Example Usage:$(RESET)"
	@echo "  $(WHITE)# Create order$(RESET)"
	@echo "  curl -X POST http://localhost:$(PORT)/api/v1/orders \\"
	@echo "    -H \"Authorization: Bearer <token>\" \\"
	@echo "    -H \"Content-Type: application/json\" \\"
	@echo "    -d '{\"payment_method\":\"cash\",\"items\":[{\"recipe_id\":\"...\",\"quantity\":1,\"unit_price\":5.00}]}'"
	@echo ""
	@echo "  $(WHITE)# Get orders$(RESET)"
	@echo "  curl -H \"Authorization: Bearer <token>\" http://localhost:$(PORT)/api/v1/orders"

## 🐳 Docker Commands

docker-build: ## Build Docker image
	@echo "$(BLUE)🐳 Building Docker image...$(RESET)"
	@docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)✅ Docker image built: $(DOCKER_IMAGE)$(RESET)"

docker-run: docker-build ## Run service in Docker container
	@echo "$(GREEN)🐳 Starting Orders Service in Docker...$(RESET)"
	@docker run -d \
		--name $(DOCKER_CONTAINER) \
		--network icecream_network \
		-p $(PORT):$(PORT) \
		-e DB_HOST=icecream_postgres \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=postgres123 \
		-e DB_NAME=icecream_store \
		-e SERVER_PORT=$(PORT) \
		$(DOCKER_IMAGE)
	@echo "$(GREEN)✅ Docker container started: $(DOCKER_CONTAINER)$(RESET)"
	@echo "$(CYAN)📡 Service available at: http://localhost:$(PORT)$(RESET)"

docker-stop: ## Stop Docker container
	@echo "$(YELLOW)🛑 Stopping Docker container...$(RESET)"
	@docker stop $(DOCKER_CONTAINER) || true
	@docker rm $(DOCKER_CONTAINER) || true
	@echo "$(GREEN)✅ Docker container stopped$(RESET)"

docker-clean: docker-stop ## Clean Docker resources
	@echo "$(YELLOW)🧹 Cleaning Docker resources...$(RESET)"
	@docker rmi $(DOCKER_IMAGE) || true
	@echo "$(GREEN)✅ Docker cleanup completed$(RESET)"

## 📊 Monitoring Commands

logs: ## Show service logs (if running in Docker)
	@echo "$(CYAN)📋 Orders Service Logs:$(RESET)"
	@if docker ps | grep -q $(DOCKER_CONTAINER); then \
		docker logs -f $(DOCKER_CONTAINER); \
	else \
		echo "$(RED)❌ Docker container not running$(RESET)"; \
		echo "$(YELLOW)💡 Run 'make docker-run' first$(RESET)"; \
	fi

## 🔄 Complete Workflow Commands

fresh: clean deps build test info ## Fresh start: clean, build, test, and show info
	@echo "$(GREEN)🎉 Fresh Orders Service setup completed!$(RESET)"
	@echo "$(YELLOW)💡 Service binary is ready. Run 'make start' to launch the service$(RESET)"
	@echo "$(CYAN)🧪 All tests passed successfully$(RESET)"

## 🧹 Cleanup Commands

clean-all: clean docker-clean ## Clean everything (build artifacts and Docker)
	@echo "$(GREEN)🎉 Complete cleanup finished!$(RESET)"

## 📋 Development Helpers

dev-setup: ## Setup development environment
	@echo "$(BLUE)🔧 Setting up development environment...$(RESET)"
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)📝 Creating config file from example...$(RESET)"; \
		cp $(ENV_EXAMPLE) $(ENV_FILE) 2>/dev/null || echo "$(YELLOW)⚠️ No example config found$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Development environment ready$(RESET)"
	@echo "$(CYAN)💡 Edit $(ENV_FILE) to customize configuration$(RESET)"

## 🔍 Health Check Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking Orders Service health...$(RESET)"
	@if curl -s http://localhost:$(PORT)/api/v1/orders/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Service is healthy$(RESET)"; \
		curl -s http://localhost:$(PORT)/api/v1/orders/health | jq . 2>/dev/null || curl -s http://localhost:$(PORT)/api/v1/orders/health; \
	else \
		echo "$(RED)❌ Service is not responding$(RESET)"; \
	fi 