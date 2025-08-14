# Colors for output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
RESET=\033[0m

# Default target
.DEFAULT_GOAL := help

## 🚀 Quick Start Commands

help: ## Show this help message
	@echo "$(CYAN)Gateway Service - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  $(GREEN)make install$(RESET)          # Complete setup and start service"
	@echo "  $(GREEN)make fresh$(RESET)            # Fresh start: clean, build, start, test"
	@echo "  $(GREEN)make test$(RESET)             # Run unit tests"
	@echo "  $(GREEN)make dev$(RESET)              # Run in development mode"
	@echo ""
	@echo "$(GREEN)Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

## 🧪 Testing Commands

test: ## Run unit tests
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: ## Run integration tests (requires services to be running)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
	@if curl -f http://localhost:8080/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Gateway is running, running integration tests...$(RESET)"; \
		./scripts/test_all_services.sh; \
	else \
		echo "$(YELLOW)⚠️  Gateway service is not running, skipping integration tests$(RESET)"; \
		echo "$(YELLOW)   Start the service with 'make start' to run integration tests$(RESET)"; \
	fi

test-coverage: ## Run tests with coverage
	@echo "$(CYAN)🧪 Running tests with coverage...$(RESET)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

## 🔧 Development Commands

build: ## Build the gateway service binary
	@echo "$(CYAN)🔨 Building Gateway Service...$(RESET)"
	@go build -o bin/gateway-service .
	@echo "$(GREEN)✅ Build completed: bin/gateway-service$(RESET)"

## 🚀 Local Development Commands

start-locally: deps ## Run the service locally (no Docker)
	@echo "$(CYAN)🛠️  Starting Gateway Service locally...$(RESET)"
	@echo "$(YELLOW)🔍 Checking for existing processes on port 8082...$(RESET)"
	@if lsof -ti:8082 >/dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Found process using port 8082, killing it...$(RESET)"; \
		lsof -ti:8082 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	@echo "$(YELLOW)⚠️  Make sure all backend services are running first!$(RESET)"
	@echo "$(CYAN)📝 Service will be available at: http://localhost:8082$(RESET)"
	@go run .

stop-locally: ## Stop local service (if running in background)
	@echo "$(YELLOW)🛑 Stopping local Gateway Service...$(RESET)"
	@echo "$(CYAN)🔍 Finding and killing all related processes...$(RESET)"
	
	# Kill processes on port 8082
	@if lsof -ti:8082 >/dev/null 2>&1; then \
		echo "$(YELLOW)Found service on port 8082, stopping...$(RESET)"; \
		lsof -ti:8082 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	
	# Kill Go processes related to gateway service
	@pkill -f "go run.*main.go" 2>/dev/null || true
	@pkill -f "gateway-service" 2>/dev/null || true
	@pkill -f "main.go" 2>/dev/null || true
	@pkill -f "/go-build/.*gateway" 2>/dev/null || true
	
	# Kill any remaining processes using port 8082
	@lsof -ti:8082 | xargs kill -9 2>/dev/null || true
	
	# Clean up temporary files and build artifacts
	@echo "$(CYAN)🧹 Cleaning up resources...$(RESET)"
	@rm -f gateway-service 2>/dev/null || true
	@rm -f *.out 2>/dev/null || true
	@rm -f *.log 2>/dev/null || true
	@rm -f coverage.html 2>/dev/null || true
	@rm -rf /tmp/go-build-* 2>/dev/null || true
	
	# Final verification
	@sleep 2
	@if lsof -ti:8082 >/dev/null 2>&1; then \
		echo "$(RED)⚠️  Warning: Port 8082 still in use after cleanup$(RESET)"; \
		lsof -ti:8082 | xargs kill -9 2>/dev/null || true; \
	else \
		echo "$(GREEN)✅ Port 8082 is now free$(RESET)"; \
	fi
	
	@echo "$(GREEN)✅ Local service stopped and resources cleaned$(RESET)"

## 🐳 Standardized Docker Commands

build-docker: ## Build Docker image for gateway service
	@echo "$(CYAN)🔨 Building Gateway Service...$(RESET)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .
	@echo "$(GREEN)✅ Build completed: main (Linux binary)$(RESET)"
	@echo "$(CYAN)🐳 Building Docker image...$(RESET)"
	@cd docker && docker build -t docker-gateway-service -f Dockerfile ..
	@echo "$(GREEN)✅ Docker image built: docker-gateway-service$(RESET)"

start-docker: build-docker ## Start gateway service in Docker
	@echo "$(CYAN)🌐 Starting Gateway Service...$(RESET)"
	@echo "$(CYAN)🌐 Starting Ice Cream Store Gateway Service...$(RESET)"
	@./scripts/start.sh

stop-docker: ## Stop gateway service Docker container
	@echo "$(YELLOW)🛑 Stopping Gateway Service containers...$(RESET)"
	@cd docker && docker-compose down
	@echo "$(GREEN)✅ Gateway service stopped successfully!$(RESET)"

fresh-docker: stop-docker build-docker start-docker ## Fresh start: stop, rebuild, and start Docker container

## 📊 Monitoring & Debugging Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking gateway service health...$(RESET)"
	@if curl -f http://localhost:8082/api/v1/gateway/p/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Gateway service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Gateway service is not responding$(RESET)"; \
		exit 1; \
	fi

logs: ## View service logs
	@echo "$(CYAN)📋 Viewing Gateway Service logs...$(RESET)"
	@./scripts/logs.sh

logs-follow: ## Follow service logs in real-time
	@echo "$(CYAN)📋 Following Gateway Service logs...$(RESET)"
	@./scripts/logs.sh -f

## 🐳 Legacy Container Management Commands

start-container: ## Start the gateway service in container (not implemented)
	@echo "$(YELLOW)⚠️  Container mode not yet implemented for gateway service$(RESET)"
	@echo "$(CYAN)   Use 'make start-locally' instead$(RESET)"

stop-container: ## Stop the gateway service container (not implemented)
	@echo "$(YELLOW)⚠️  Container mode not yet implemented for gateway service$(RESET)"

# Legacy aliases for backward compatibility
run: build ## [DEPRECATED] Use start-locally instead - Run the gateway service locally
	@echo "$(CYAN)🚀 Running Gateway Service locally...$(RESET)"
	@./bin/gateway-service

dev: start-locally ## [DEPRECATED] Use start-locally instead

start: start-locally ## [DEPRECATED] Use start-locally instead

stop: stop-locally ## [DEPRECATED] Use stop-locally instead

restart: stop-locally start-locally ## [DEPRECATED] Use stop-locally then start-locally instead
	@echo "$(GREEN)✅ Gateway Service restarted!$(RESET)"

clean: ## Clean build artifacts and cache
	@echo "$(CYAN)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f gateway-service local-server
	@go clean -cache
	@echo "$(GREEN)✅ Clean completed!$(RESET)"

## 📦 Dependencies

deps: ## Download and tidy dependencies
	@echo "$(CYAN)📦 Managing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

deps-update: ## Update all dependencies
	@echo "$(CYAN)📦 Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

## 🔍 Code Quality

lint: ## Run linters
	@echo "$(CYAN)🔍 Running linters...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed, running go vet instead$(RESET)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✅ Linting completed!$(RESET)"

format: ## Format code
	@echo "$(CYAN)✨ Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted!$(RESET)"

## 🚀 Lifecycle Commands

install: deps build ## Complete setup and start service
	@echo "$(GREEN)🎉 Gateway Service is ready!$(RESET)"
	@echo "$(CYAN)Run 'make dev' to start in development mode$(RESET)"

fresh: clean deps build test ## Fresh start: clean, build, test
	@echo "$(GREEN)🎉 Fresh gateway service build completed!$(RESET)"

## 🧪 Service Testing Commands

test-session: ## Test session flow
	@echo "$(CYAN)🧪 Testing session flow...$(RESET)"
	@if [ -f ./test_session_flow.sh ]; then \
		./test_session_flow.sh; \
	else \
		echo "$(YELLOW)Session flow test script not found$(RESET)"; \
	fi

test-services: ## Test all service connections
	@echo "$(CYAN)🧪 Testing service connections...$(RESET)"
	@echo "Testing Session Service..." && curl -f http://localhost:8081/health || echo "$(RED)Session service not responding$(RESET)"
	@echo "Testing Orders Service..." && curl -f http://localhost:8083/health || echo "$(RED)Orders service not responding$(RESET)"
	@echo "Testing Inventory Service..." && curl -f http://localhost:8084/health || echo "$(RED)Inventory service not responding$(RESET)"
	@echo "Testing Data Service..." && curl -f http://localhost:8082/health || echo "$(RED)Data service not responding$(RESET)"
	@echo "Testing UI Service..." && curl -f http://localhost:3000 || echo "$(RED)UI service not responding$(RESET)"

## 📊 Information Commands

info: ## Show service information
	@echo "$(CYAN)📊 Gateway Service Information:$(RESET)"
	@echo "  Service: gateway-service"
	@echo "  Port: 8080"
	@echo "  Go Version: $(shell go version)"
	@echo "  Module: $(shell go list -m)"
	@echo ""

env: ## Show environment variables
	@echo "$(CYAN)🌍 Environment Variables:$(RESET)"
	@echo "  SESSION_SERVICE_URL: $(or $(SESSION_SERVICE_URL),not set (default: http://localhost:8081))"
	@echo "  ORDERS_SERVICE_URL: $(or $(ORDERS_SERVICE_URL),not set (default: http://localhost:8083))"
	@echo "  INVENTORY_SERVICE_URL: $(or $(INVENTORY_SERVICE_URL),not set (default: http://localhost:8084))"
	@echo "  DATA_SERVICE_URL: $(or $(DATA_SERVICE_URL),not set (default: http://localhost:8082))"
	@echo "  UI_SERVICE_URL: $(or $(UI_SERVICE_URL),not set (default: http://localhost:3000))"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information:$(RESET)"
	@echo "  Gateway Service: 1.0.0"
	@echo "  Go Version: $(shell go version)"
	@echo "  Build Date: $(shell date)"

## 🔧 Utility Commands

check-deps: ## Check if required dependencies are installed
	@echo "$(CYAN)🔍 Checking dependencies...$(RESET)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go is not installed$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Go is installed$(RESET)"

health: ## Check gateway health
	@echo "$(CYAN)🔍 Checking gateway health...$(RESET)"
	@curl -f http://localhost:8080/health || echo "$(RED)❌ Gateway is not responding$(RESET)"

## 📋 Service Logs Commands

logs: ## Show logs for a specific service (usage: make logs SERVICE=service-name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Error: SERVICE parameter is required$(RESET)"; \
		echo "$(CYAN)Usage: make logs SERVICE=service-name$(RESET)"; \
		echo "$(CYAN)Available services: gateway-service, session-service, orders-service, inventory-service, invoice-service, data-service$(RESET)"; \
		exit 1; \
	fi
	@echo "$(CYAN)📋 Loading logs for $(SERVICE)...$(RESET)"
	@case "$(SERVICE)" in \
		"gateway-service") \
			if [ -f "../gateway-service/service.log" ]; then \
				tail -n 50 "../gateway-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for gateway-service$(RESET)"; \
			fi ;; \
		"session-service") \
			if [ -f "../session-service/service.log" ]; then \
				tail -n 50 "../session-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for session-service$(RESET)"; \
			fi ;; \
		"orders-service") \
			if [ -f "../orders-service/service.log" ]; then \
				tail -n 50 "../orders-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for orders-service$(RESET)"; \
			fi ;; \
		"inventory-service") \
			if [ -f "../inventory-service/service.log" ]; then \
				tail -n 50 "../inventory-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for inventory-service$(RESET)"; \
			fi ;; \
		"invoice-service") \
			if [ -f "../invoice-service/service.log" ]; then \
				tail -n 50 "../invoice-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for invoice-service$(RESET)"; \
			fi ;; \
		"data-service") \
			if [ -f "../data-service/service.log" ]; then \
				tail -n 50 "../data-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for data-service$(RESET)"; \
			fi ;; \
		*) \
			echo "$(RED)❌ Unknown service: $(SERVICE)$(RESET)"; \
			echo "$(CYAN)Available services: gateway-service, session-service, orders-service, inventory-service, invoice-service, data-service$(RESET)" ;; \
	esac

logs-tail: ## Tail logs for a specific service (usage: make logs-tail SERVICE=service-name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Error: SERVICE parameter is required$(RESET)"; \
		echo "$(CYAN)Usage: make logs-tail SERVICE=service-name$(RESET)"; \
		exit 1; \
	fi
	@echo "$(CYAN)📋 Tailing logs for $(SERVICE)...$(RESET)"
	@case "$(SERVICE)" in \
		"gateway-service") \
			if [ -f "../gateway-service/service.log" ]; then \
				tail -f "../gateway-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for gateway-service$(RESET)"; \
			fi ;; \
		"session-service") \
			if [ -f "../session-service/service.log" ]; then \
				tail -f "../session-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for session-service$(RESET)"; \
			fi ;; \
		"orders-service") \
			if [ -f "../orders-service/service.log" ]; then \
				tail -f "../orders-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for orders-service$(RESET)"; \
			fi ;; \
		"inventory-service") \
			if [ -f "../inventory-service/service.log" ]; then \
				tail -f "../inventory-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for inventory-service$(RESET)"; \
			fi ;; \
		"invoice-service") \
			if [ -f "../invoice-service/service.log" ]; then \
				tail -f "../invoice-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for invoice-service$(RESET)"; \
			fi ;; \
		"data-service") \
			if [ -f "../data-service/service.log" ]; then \
				tail -f "../data-service/service.log"; \
			else \
				echo "$(YELLOW)⚠️  No log file found for data-service$(RESET)"; \
			fi ;; \
		*) \
			echo "$(RED)❌ Unknown service: $(SERVICE)$(RESET)" ;; \
	esac

## 📝 Documentation

docs: ## Generate documentation
	@echo "$(CYAN)📝 Generating documentation...$(RESET)"
	@go doc -all > docs.txt
	@echo "$(GREEN)✅ Documentation generated: docs.txt$(RESET)"

# List all targets for tab completion
.PHONY: help test test-integration test-coverage build run dev start stop restart clean deps deps-update lint format install fresh test-session test-services info env version check-deps health docs 