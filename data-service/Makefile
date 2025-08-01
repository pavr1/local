# Ice Cream Store Data Service Makefile
# This Makefile provides convenient commands to manage the PostgreSQL database service

.PHONY: help start stop restart reset logs connect status health test build run clean install deps fresh

# Default target
.DEFAULT_GOAL := help

# Colors for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

## 🚀 Database Management Commands

fresh: clean deps start init test status info ## Complete fresh installation: clean everything, install, setup, and test
	@echo "$(GREEN)🎉 Fresh installation completed successfully!$(RESET)"
	@echo "$(CYAN)📋 Quick verification:$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "SELECT 'Fresh install verified!' as status, COUNT(*) as total_tables FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null || echo "$(RED)❌ Database verification failed$(RESET)"
	@echo ""
	@echo "$(YELLOW)🎯 Next Steps:$(RESET)"
	@echo "   • Access PgAdmin: http://localhost:8080"
	@echo "   • Run: make connect (for CLI access)"
	@echo "   • Import in other services: import \"data-service/pkg/database\""
	@echo ""
	@echo "$(CYAN)🔐 PostgreSQL Login Information:$(RESET)"
	@echo "============================================="
	@echo "$(GREEN)Direct Database Connection:$(RESET)"
	@echo "   Host/URL: localhost"
	@echo "   Port: 5432"
	@echo "   Database: icecream_store"
	@echo "   Username: postgres"
	@echo "   Password: postgres123"
	@echo "   Connection String: postgresql://postgres:postgres123@localhost:5432/icecream_store"
	@echo ""
	@echo "$(GREEN)PgAdmin Web Interface:$(RESET)"
	@echo "   URL: http://localhost:8080"
	@echo "   Login Email: admin@icecreamstore.com"
	@echo "   Login Password: admin123"
	@echo ""
	@echo "$(GREEN)Application Admin User:$(RESET)"
	@echo "   Username: admin"
	@echo "   Password: admin123"
	@echo "   Role: super_admin (full system access)"

## 🐳 Container Management Commands (Data Service Always Uses Containers)

start: ## Start PostgreSQL and PgAdmin containers
	@echo "$(CYAN)🍦 Starting Ice Cream Store Data Service...$(RESET)"
	@./scripts/start.sh

stop: ## Stop all database containers
	@echo "$(YELLOW)🛑 Stopping Data Service containers...$(RESET)"
	@./scripts/stop.sh

restart: stop start ## Restart all database containers
	@echo "$(GREEN)🔄 Data Service restarted successfully!$(RESET)"

# Aliases for consistency with other services (data service always uses containers)
start-locally: start ## [ALIAS] Start data service (always uses containers)
stop-locally: stop ## [ALIAS] Stop data service containers

start-docker: start ## [ALIAS] Start data service containers
stop-docker: stop ## [ALIAS] Stop data service containers
restart-docker: restart ## [ALIAS] Restart data service containers

reset: ## Reset database (⚠️  DELETES ALL DATA)
	@echo "$(RED)⚠️  WARNING: This will delete ALL database data!$(RESET)"
	@./scripts/reset.sh

## 📊 Monitoring & Debugging Commands

logs: ## View logs from all containers
	@echo "$(CYAN)📋 Viewing Data Service logs...$(RESET)"
	@./scripts/logs.sh

logs-db: ## View PostgreSQL logs only
	@echo "$(CYAN)📋 Viewing PostgreSQL logs...$(RESET)"
	@./scripts/logs.sh postgres

logs-admin: ## View PgAdmin logs only
	@echo "$(CYAN)📋 Viewing PgAdmin logs...$(RESET)"
	@./scripts/logs.sh pgadmin

connect: ## Connect to PostgreSQL database via CLI
	@echo "$(GREEN)🔌 Connecting to database...$(RESET)"
	@./scripts/connect.sh

status: ## Show container status
	@echo "$(CYAN)📊 Container Status:$(RESET)"
	@cd docker && docker-compose ps

health: ## Check database health
	@echo "$(CYAN)🏥 Checking database health...$(RESET)"
	@if docker exec icecream_postgres pg_isready -U postgres -d icecream_store > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Database is healthy and ready!$(RESET)"; \
	else \
		echo "$(RED)❌ Database is not responding$(RESET)"; \
		exit 1; \
	fi

## 🔧 Development Commands

build: ## Build the data service application
	@echo "$(CYAN)🔨 Building data service application...$(RESET)"
	@go build -o bin/data-service .
	@echo "$(GREEN)✅ Build completed: bin/data-service$(RESET)"

run: build ## Run the data service application to test database connection
	@echo "$(CYAN)🧪 Running data service application...$(RESET)"
	@./bin/data-service

test: health ## Run basic database connectivity tests
	@echo "$(CYAN)🧪 Running connectivity tests...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "SELECT 'Connection test successful!' as result;" > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Database connectivity test passed!$(RESET)" || \
		(echo "$(RED)❌ Database connectivity test failed!$(RESET)" && exit 1)
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "SELECT COUNT(*) FROM roles;" > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Schema validation test passed!$(RESET)" || \
		(echo "$(RED)❌ Schema validation test failed!$(RESET)" && exit 1)

## 📦 Setup & Installation Commands

install: deps start ## Complete setup: install dependencies and start services
	@echo "$(GREEN)🎉 Data Service installation completed!$(RESET)"
	@echo "$(CYAN)📝 Connection Details:$(RESET)"
	@echo "   Database: postgresql://postgres:postgres123@localhost:5432/icecream_store"
	@echo "   PgAdmin:  http://localhost:8080 (admin@icecreamstore.com / admin123)"

deps: ## Install Go dependencies
	@echo "$(CYAN)📦 Installing Go dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(RESET)"

init: ## Initialize database with fresh schema (requires containers to be running)
	@echo "$(CYAN)🔄 Reinitializing database schema...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -f /docker-entrypoint-initdb.d/01-init-database.sql > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Database schema reinitialized$(RESET)" || \
		echo "$(RED)❌ Failed to reinitialize schema. Try 'make reset' instead.$(RESET)"

## 🧹 Cleanup Commands

clean: ## Clean up containers, volumes, and built binaries
	@echo "$(YELLOW)🧹 Cleaning up...$(RESET)"
	@cd docker && docker-compose down -v --remove-orphans 2>/dev/null || true
	@docker volume prune -f 2>/dev/null || true
	@rm -rf bin/
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

clean-all: clean ## Complete cleanup including Docker images
	@echo "$(YELLOW)🧹 Performing complete cleanup...$(RESET)"
	@docker image prune -f 2>/dev/null || true
	@echo "$(GREEN)✅ Complete cleanup finished$(RESET)"

## 🔍 Information Commands

info: ## Show database connection information
	@echo "$(CYAN)📝 Ice Cream Store Data Service Information$(RESET)"
	@echo "=============================================="
	@echo "$(GREEN)Database Connection:$(RESET)"
	@echo "   Host: localhost"
	@echo "   Port: 5432"
	@echo "   Database: icecream_store"
	@echo "   Username: postgres"
	@echo "   Password: postgres123"
	@echo ""
	@echo "$(GREEN)PgAdmin Web Interface:$(RESET)"
	@echo "   URL: http://localhost:8080"
	@echo "   Email: admin@icecreamstore.com"
	@echo "   Password: admin123"
	@echo ""
	@echo "$(GREEN)Default Admin User:$(RESET)"
	@echo "   Username: admin"
	@echo "   Password: admin123"

env: ## Show environment variables template
	@echo "$(CYAN)📋 Environment Variables Template:$(RESET)"
	@cat config.env.example

version: ## Show service version and dependencies
	@echo "$(CYAN)📋 Data Service Version Information:$(RESET)"
	@echo "=============================================="
	@echo "Go version: $(shell go version)"
	@echo "Docker version: $(shell docker --version 2>/dev/null || echo 'Docker not found')"
	@echo "Docker Compose version: $(shell docker-compose --version 2>/dev/null || echo 'Docker Compose not found')"
	@echo ""
	@echo "$(GREEN)Go Dependencies:$(RESET)"
	@go list -m all

## 🆘 Help Command

help: ## Show this help message
	@echo "$(CYAN)🍦 Ice Cream Store Data Service - Available Commands$(RESET)"
	@echo "============================================================="
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { \
		printf "$(GREEN)%-15s$(RESET) %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)📋 Quick Start:$(RESET)"
	@echo "   make fresh      # 🚀 COMPLETE FRESH INSTALL (recommended for first time)"
	@echo "   make install    # Complete setup and start services"
	@echo "   make start      # Start database containers"
	@echo "   make test       # Test database connectivity"
	@echo "   make run        # Run data service application"
	@echo ""
	@echo "$(YELLOW)📋 Common Workflows:$(RESET)"
	@echo "   First Time:     make fresh (complete setup from scratch)"
	@echo "   Development:    make start → make test → make run"
	@echo "   Reset Data:     make reset → make start"
	@echo "   Fresh Start:    make fresh (when something is broken)"
	@echo "   Cleanup:        make clean"
	@echo "" 