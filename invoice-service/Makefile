# Variables
BUILD_DIR=bin
SERVICE_NAME=invoice-service
BINARY_NAME=invoice-service
MAIN_FILE=main.go

# Docker variables
DOCKER_IMAGE=icecream-invoice-service

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build targets
.PHONY: all build clean test deps run start-locally stop-locally docker-build docker-run start stop logs fresh

all: deps build

build:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests for $(SERVICE_NAME)..."
	$(GOTEST) -v ./...
	@echo "✅ Tests completed"

deps:
	@echo "Downloading dependencies for $(SERVICE_NAME)..."
	$(GOMOD) download
	$(GOMOD) tidy

run: build
	@echo "Running $(SERVICE_NAME)..."
	 ./$(BUILD_DIR)/$(BINARY_NAME)

## 🚀 Local Development Commands

start-locally: deps ## Run the service locally (no Docker)
	@echo "$(CYAN)🛠️  Starting Invoice Service locally...$(RESET)"
	@echo "$(YELLOW)🔍 Checking for existing processes on port 8085...$(RESET)"
	@if lsof -ti:8085 >/dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Found process using port 8085, killing it...$(RESET)"; \
		lsof -ti:8085 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	@echo "$(YELLOW)⚠️  Make sure data-service is running first!$(RESET)"
	@echo "$(CYAN)📝 Service will be available at: http://localhost:8085$(RESET)"
	@echo "$(CYAN)📝 Loading configuration from config.env...$(RESET)"
	@echo "$(CYAN)🚀 Starting service in background...$(RESET)"
	@set -a && [ -f config.env ] && . ./config.env && set +a && nohup go run . > service.log 2>&1 & echo $$! > service.pid
	@sleep 2
	@echo "$(GREEN)✅ Service started successfully$(RESET)"

stop-locally: ## Stop local service (if running in background)
	@echo "$(YELLOW)🛑 Stopping local Invoice Service...$(RESET)"
	@echo "$(CYAN)🔍 Finding and killing all related processes...$(RESET)"
	
	# Kill processes on port 8085
	@if lsof -ti:8085 >/dev/null 2>&1; then \
		echo "$(YELLOW)Found service on port 8085, stopping...$(RESET)"; \
		lsof -ti:8085 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	
	# Kill Go processes related to invoice service
	@pkill -f "go run.*main.go" 2>/dev/null || true
	@pkill -f "invoice-service" 2>/dev/null || true
	@pkill -f "expense-service" 2>/dev/null || true
	@pkill -f "main.go" 2>/dev/null || true
	@pkill -f "/go-build/.*invoice" 2>/dev/null || true
	@pkill -f "/go-build/.*expense" 2>/dev/null || true
	
	# Kill any remaining processes using port 8085
	@lsof -ti:8085 | xargs kill -9 2>/dev/null || true
	
	# Clean up temporary files and build artifacts
	@echo "$(CYAN)🧹 Cleaning up resources...$(RESET)"
	@rm -f invoice-service 2>/dev/null || true
	@rm -f expense-service 2>/dev/null || true
	@rm -f service.pid 2>/dev/null || true
	@rm -f service.log 2>/dev/null || true
	@rm -f *.out 2>/dev/null || true
	@rm -f *.log 2>/dev/null || true
	@rm -f coverage.html 2>/dev/null || true
	@rm -rf /tmp/go-build-* 2>/dev/null || true
	
	# Final verification
	@sleep 2
	@if lsof -ti:8085 >/dev/null 2>&1; then \
		echo "$(RED)⚠️  Warning: Port 8085 still in use after cleanup$(RESET)"; \
		lsof -ti:8085 | xargs kill -9 2>/dev/null || true; \
	else \
		echo "$(GREEN)✅ Port 8085 is now free$(RESET)"; \
	fi
	
	@echo "$(GREEN)✅ Local service stopped and resources cleaned$(RESET)"

## 🐳 Standardized Container Commands

build-container: ## Build Docker image for invoice service
	@echo "$(CYAN)🔨 Building Invoice Service...$(RESET)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .
	@echo "$(GREEN)✅ Build completed: main (Linux binary)$(RESET)"
	@echo "$(CYAN)🐳 Building Docker image...$(RESET)"
	@cd docker && docker build -t docker-invoice-service -f Dockerfile ..
	@echo "$(GREEN)✅ Docker image built: docker-invoice-service$(RESET)"

start-container: build-container ## Start invoice service in Docker
	@echo "$(CYAN)🍦 Starting Invoice Service...$(RESET)"
	@echo "$(CYAN)🍦 Starting Ice Cream Store Invoice Service...$(RESET)"
	@./scripts/start.sh

stop-container: ## Stop invoice service Docker container
	@echo "$(YELLOW)🛑 Stopping Invoice Service containers...$(RESET)"
	@cd docker && docker-compose down
	@echo "$(GREEN)✅ Invoice service stopped successfully!$(RESET)"

fresh-container: stop-container build-container start-container ## Fresh start: stop, rebuild, and start Docker container

## 📊 Monitoring & Debugging Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking invoice service health...$(RESET)"
	@if curl -f http://localhost:8085/api/v1/invoices/p/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Invoice service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Invoice service is not responding$(RESET)"; \
		exit 1; \
	fi

logs: ## View service logs
	@echo "$(CYAN)📋 Viewing Invoice Service logs...$(RESET)"
	@./scripts/logs.sh

logs-follow: ## Follow service logs in real-time
	@echo "$(CYAN)📋 Following Invoice Service logs...$(RESET)"
	@./scripts/logs.sh -f

## 🐳 Legacy Container Management Commands
docker-build:
	@echo "Building Docker image for $(SERVICE_NAME)..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "Running $(SERVICE_NAME) in Docker..."
	docker run --rm -p 8085:8085 --name $(SERVICE_NAME) $(DOCKER_IMAGE)

start: docker-run

stop:
	@echo "Stopping $(SERVICE_NAME)..."
	docker stop $(SERVICE_NAME) || true

fresh: clean deps build 