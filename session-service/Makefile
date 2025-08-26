# Session Service Makefile

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# Variables
SERVICE_NAME := session-service
BINARY_NAME := session-service
BUILD_DIR := bin
MAIN_FILE := main.go

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

# Default target
.DEFAULT_GOAL := help

## 🚀 Local Development Commands

start-locally: deps ## Run the service locally (no Docker)
	@echo "$(CYAN)🛠️  Starting Session Service locally...$(RESET)"
	@echo "$(YELLOW)🔍 Checking for existing processes on port 8081...$(RESET)"
	@if lsof -ti:8081 >/dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Found process using port 8081, killing it...$(RESET)"; \
		lsof -ti:8081 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	@echo "$(YELLOW)⚠️  Make sure data-service is running first!$(RESET)"
	@echo "$(CYAN)📝 Service will be available at: http://localhost:8081$(RESET)"
	@echo "$(CYAN)🚀 Starting service in background...$(RESET)"
	@nohup go run . > service.log 2>&1 & echo $$! > service.pid
	@sleep 2
	@echo "$(GREEN)✅ Service started successfully$(RESET)"

stop-locally: ## Stop local service (if running in background)
	@echo "$(YELLOW)🛑 Stopping local Session Service...$(RESET)"
	@echo "$(CYAN)🔍 Finding and killing all related processes...$(RESET)"
	
	# Kill processes on port 8081
	@if lsof -ti:8081 >/dev/null 2>&1; then \
		echo "$(YELLOW)Found service on port 8081, stopping...$(RESET)"; \
		lsof -ti:8081 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	
	# Kill Go processes related to session service
	@pkill -f "go run.*main.go" 2>/dev/null || true
	@pkill -f "session-service" 2>/dev/null || true
	@pkill -f "main.go" 2>/dev/null || true
	@pkill -f "/go-build/.*session" 2>/dev/null || true
	
	# Kill any remaining processes using port 8081
	@lsof -ti:8081 | xargs kill -9 2>/dev/null || true
	
	# Clean up temporary files and build artifacts
	@echo "$(CYAN)🧹 Cleaning up resources...$(RESET)"
	@rm -f session-service 2>/dev/null || true
	@rm -f service.pid 2>/dev/null || true
	@rm -f service.log 2>/dev/null || true
	@rm -f *.out 2>/dev/null || true
	@rm -f *.log 2>/dev/null || true
	@rm -f coverage.html 2>/dev/null || true
	@rm -rf /tmp/go-build-* 2>/dev/null || true
	
	# Final verification
	@sleep 2
	@if lsof -ti:8081 >/dev/null 2>&1; then \
		echo "$(RED)⚠️  Warning: Port 8081 still in use after cleanup$(RESET)"; \
		lsof -ti:8081 | xargs kill -9 2>/dev/null || true; \
	else \
		echo "$(GREEN)✅ Port 8081 is now free$(RESET)"; \
	fi
	
	@echo "$(GREEN)✅ Local service stopped and resources cleaned$(RESET)"

## 🔧 Build and Development Commands

# Build the application
.PHONY: build
build:
	@echo "$(CYAN)Building $(SERVICE_NAME)...$(RESET)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o main .
	@echo "$(GREEN)Build complete: main (Linux binary)$(RESET)"

build-container: build ## Build the Docker image
	@echo "$(CYAN)🐳 Building Docker image...$(RESET)"
	@docker build -t docker-session-service -f docker/Dockerfile .
	@echo "$(GREEN)✅ Docker image built: docker-session-service$(RESET)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean complete$(RESET)"

# Run tests
.PHONY: test
test:
	@echo "$(CYAN)Running tests...$(RESET)"
	$(GOTEST) -v ./...
	@echo "$(GREEN)✅ Tests completed!$(RESET)"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

# Install dependencies
.PHONY: deps
deps:
	@echo "$(CYAN)Installing dependencies...$(RESET)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

# Run the application
.PHONY: run
run:
	@echo "$(CYAN)Running $(SERVICE_NAME)...$(RESET)"
	$(GOCMD) run $(MAIN_FILE)

# Start the service (build and run)
.PHONY: start
start: build
	@echo "$(CYAN)Starting $(SERVICE_NAME)...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Stop the service
.PHONY: stop
stop:
	@echo "$(YELLOW)Stopping $(SERVICE_NAME)...$(RESET)"
	@pkill -f $(BINARY_NAME) || true

# Restart the service
.PHONY: restart
restart: stop start

# Container commands
start-container: build-container ## Start session service containers (DB_HOST=postgres)
	@echo "$(CYAN)🍦 Starting Session Service...$(RESET)"
	@./scripts/start.sh
	@echo "$(CYAN)🐳 Session service containers started with DB_HOST=postgres$(RESET)"

stop-container: ## Stop session service containers
	@echo "$(YELLOW)🛑 Stopping Session Service containers...$(RESET)"
	@./scripts/stop.sh

restart-container: stop-container start-container ## Restart session service containers

# Fresh start (clean, build, run)
.PHONY: fresh
fresh: clean build start

# Development mode (run with hot reload)
.PHONY: dev
dev:
	@echo "$(CYAN)Starting $(SERVICE_NAME) in development mode...$(RESET)"
	$(GOCMD) run $(MAIN_FILE)

# Format code
.PHONY: fmt
fmt:
	@echo "$(CYAN)Formatting code...$(RESET)"
	$(GOCMD) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "$(CYAN)Linting code...$(RESET)"
	$(GOCMD) vet ./...

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "$(CYAN)Checking for security vulnerabilities...$(RESET)"
	$(GOCMD) list -json -deps ./... | grep -i "vulnerability" || echo "No known vulnerabilities found"

# Generate documentation
.PHONY: docs
docs:
	@echo "$(CYAN)Generating documentation...$(RESET)"
	$(GOCMD) doc -all ./...

# Install the service
.PHONY: install
install: build
	@echo "$(CYAN)Installing $(SERVICE_NAME)...$(RESET)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ || echo "Installation requires sudo privileges"

# Uninstall the service
.PHONY: uninstall
uninstall:
	@echo "$(CYAN)Uninstalling $(SERVICE_NAME)...$(RESET)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME) || echo "Uninstallation requires sudo privileges"

# Show help
.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)🔐 Session Service - Available Commands:$(RESET)"
	@echo "====================================="
	@echo ""
	@echo "$(YELLOW)📋 Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🚀 Common Workflows:$(RESET)"
	@echo "  $(GREEN)make start-locally$(RESET)     # Start service locally (no Docker)"
	@echo "  $(GREEN)make stop-locally$(RESET)      # Stop local service"
	@echo "  $(GREEN)make fresh$(RESET)             # Fresh start: clean, build, start"
	@echo "  $(GREEN)make test$(RESET)              # Run unit tests"
	@echo ""
	@echo "$(YELLOW)📖 Prerequisites:$(RESET)"
	@echo "  • Data service must be running (cd ../data-service && make start)"
	@echo "" 