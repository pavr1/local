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
	@go run .

stop-locally: ## Stop local service (if running in background)
	@echo "$(YELLOW)🛑 Stopping local Invoice Service...$(RESET)"
	@pkill -f "go run main.go" 2>/dev/null || echo "$(YELLOW)No local service found running$(RESET)"
	@pkill -f "invoice-service" 2>/dev/null || true
	@echo "$(GREEN)✅ Local service stopped$(RESET)"

## 🐳 Container Management Commands
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

logs:
	@echo "Showing logs for $(SERVICE_NAME)..."
	docker logs $(SERVICE_NAME)

fresh: clean deps build 