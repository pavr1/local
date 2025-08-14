# Ice Cream Store - UI Service Makefile

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

## 🎨 Ice Cream Store UI Service

help: ## Show this help message
	@echo "$(CYAN)🎨 Ice Cream Store UI Service$(RESET)"
	@echo "================================"
	@echo ""
	@echo "$(YELLOW)Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""

## 🚀 Development Commands

install: ## Install dependencies
	@echo "$(CYAN)📦 Installing UI dependencies...$(RESET)"
	@npm install
	@echo "$(GREEN)✅ Dependencies installed!$(RESET)"

## 🚀 Local Development Commands

start-locally: ## Start UI development server locally
	@echo "$(CYAN)🚀 Starting UI development server locally...$(RESET)"
	@echo "$(YELLOW)🔍 Checking for existing processes on port 3000...$(RESET)"
	@if lsof -ti:3000 >/dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Found process using port 3000, killing it...$(RESET)"; \
		lsof -ti:3000 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	@echo "$(CYAN)📝 Service will be available at: http://localhost:3000$(RESET)"
	@npm start

stop-locally: ## Stop UI development server (manual process)
	@echo "$(YELLOW)🛑 Stopping local UI Service...$(RESET)"
	@echo "$(CYAN)🔍 Finding and killing all related processes...$(RESET)"
	
	# Kill processes on port 3000
	@if lsof -ti:3000 >/dev/null 2>&1; then \
		echo "$(YELLOW)Found service on port 3000, stopping...$(RESET)"; \
		lsof -ti:3000 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	
	# Kill npm and node processes related to UI service
	@pkill -f "npm start" 2>/dev/null || true
	@pkill -f "live-server" 2>/dev/null || true
	@pkill -f "http-server" 2>/dev/null || true
	@pkill -f "serve" 2>/dev/null || true
	@pkill -f "node.*ui" 2>/dev/null || true
	
	# Kill any remaining processes using port 3000
	@lsof -ti:3000 | xargs kill -9 2>/dev/null || true
	
	# Clean up temporary files and build artifacts
	@echo "$(CYAN)🧹 Cleaning up resources...$(RESET)"
	@rm -rf node_modules/.cache 2>/dev/null || true
	@rm -f *.log 2>/dev/null || true
	@rm -rf .parcel-cache 2>/dev/null || true
	@rm -rf dist 2>/dev/null || true
	@rm -rf build 2>/dev/null || true
	
	# Final verification
	@sleep 2
	@if lsof -ti:3000 >/dev/null 2>&1; then \
		echo "$(RED)⚠️  Warning: Port 3000 still in use after cleanup$(RESET)"; \
		lsof -ti:3000 | xargs kill -9 2>/dev/null || true; \
	else \
		echo "$(GREEN)✅ Port 3000 is now free$(RESET)"; \
	fi
	
	@echo "$(GREEN)✅ Local UI service stopped and resources cleaned$(RESET)"

## 🐳 Standardized Docker Commands

build-docker: ## Build Docker image for UI service
	@echo "$(CYAN)🔨 Building UI Service Docker image...$(RESET)"
	@cd docker && docker build -t docker-ui-service -f Dockerfile ..
	@echo "$(GREEN)✅ Docker image built: docker-ui-service$(RESET)"

start-docker: build-docker ## Start UI service in Docker
	@echo "$(CYAN)🎨 Starting UI Service...$(RESET)"
	@./scripts/start.sh

stop-docker: ## Stop UI service Docker container
	@echo "$(YELLOW)🛑 Stopping UI Service containers...$(RESET)"
	@cd docker && docker-compose down
	@echo "$(GREEN)✅ UI service stopped successfully!$(RESET)"

fresh-docker: stop-docker build-docker start-docker ## Fresh start: stop, rebuild, and start Docker container

## 📊 Monitoring & Debugging Commands

health: ## Check service health
	@echo "$(CYAN)🏥 Checking UI service health...$(RESET)"
	@if curl -f http://localhost:3000/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ UI service is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ UI service is not responding$(RESET)"; \
		exit 1; \
	fi

logs: ## View service logs
	@echo "$(CYAN)📋 Viewing UI Service logs...$(RESET)"
	@./scripts/logs.sh

logs-follow: ## Follow service logs in real-time
	@echo "$(CYAN)📋 Following UI Service logs...$(RESET)"
	@./scripts/logs.sh -f

## 🐳 Legacy Container Management Commands

start-container: ## Start UI in Docker container (not implemented)
	@echo "$(YELLOW)⚠️  Docker container mode not yet implemented for UI service$(RESET)"
	@echo "$(CYAN)   Use 'make start-locally' instead$(RESET)"

stop-container: ## Stop UI Docker container (not implemented)
	@echo "$(YELLOW)⚠️  Docker container mode not yet implemented for UI service$(RESET)"

# Legacy aliases for backward compatibility
start: start-locally ## [DEPRECATED] Use start-locally instead
dev: ## [DEPRECATED] Start UI server and open browser - use start-locally instead
	@echo "$(CYAN)🚀 Starting UI development server with browser...$(RESET)"
	@npm run dev

build: ## Build UI (no-op for static files)
	@echo "$(CYAN)🔨 Building UI...$(RESET)"
	@npm run build
	@echo "$(GREEN)✅ Build completed!$(RESET)"

## 🧪 Testing Commands

test: ## Run tests (placeholder)
	@echo "$(CYAN)🧪 Running UI tests...$(RESET)"
	@npm test
	@echo "$(GREEN)✅ Tests completed!$(RESET)"

## 🔍 Status Commands

status: ## Check UI service status
	@echo "$(CYAN)🔍 Checking UI service status...$(RESET)"
	@if curl -f http://localhost:3000 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ UI service is running on http://localhost:3000$(RESET)"; \
	else \
		echo "$(RED)❌ UI service is not running$(RESET)"; \
	fi

health: status ## Check UI service health (alias for status)

## 🧹 Cleanup Commands

clean: ## Clean dependencies and cache
	@echo "$(YELLOW)🧹 Cleaning UI dependencies...$(RESET)"
	@rm -rf node_modules package-lock.json
	@echo "$(GREEN)✅ UI cleaned!$(RESET)"

## 📊 Utility Commands

fresh: clean install start ## Fresh install and start
	@echo "$(GREEN)🎉 Fresh UI setup completed!$(RESET)"

info: ## Show service information
	@echo "$(CYAN)📊 UI Service Information$(RESET)"
	@echo "========================="
	@echo "Service: Ice Cream Store UI"
	@echo "Port: 3000"
	@echo "URL: http://localhost:3000"
	@echo "Type: Static Frontend (HTML/CSS/JS)"
	@echo "Framework: Bootstrap 5 + Vanilla JavaScript"
	@echo ""

logs: ## Show server logs (if running via npm)
	@echo "$(CYAN)📋 Checking for running UI server...$(RESET)"
	@if pgrep -f "http-server.*3000" > /dev/null; then \
		echo "$(GREEN)✅ UI server is running$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️ No UI server process found$(RESET)"; \
	fi

version: ## Show version information
	@echo "$(CYAN)📋 UI Service Version Information$(RESET)"
	@echo "=================================="
	@echo "UI Service: 1.0.0"
	@node --version 2>/dev/null || echo "Node.js: Not installed"
	@npm --version 2>/dev/null || echo "NPM: Not installed"

# List all targets for tab completion
.PHONY: help install start-locally stop-locally start-docker stop-docker start dev build test status health clean fresh info logs version 