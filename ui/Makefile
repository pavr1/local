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
	@echo "$(CYAN)📝 Service will be available at: http://localhost:3000$(RESET)"
	@npm start

stop-locally: ## Stop UI development server (manual process)
	@echo "$(YELLOW)🛑 To stop UI development server:$(RESET)"
	@echo "$(YELLOW)   Press Ctrl+C in the terminal running the server$(RESET)"
	@echo "$(YELLOW)   Or run: pkill -f 'npm start' || pkill -f 'live-server'$(RESET)"

## 🐳 Container Management Commands

start-docker: ## Start UI in Docker container (not implemented)
	@echo "$(YELLOW)⚠️  Docker container mode not yet implemented for UI service$(RESET)"
	@echo "$(CYAN)   Use 'make start-locally' instead$(RESET)"

stop-docker: ## Stop UI Docker container (not implemented)
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