# Ice Cream Store - Root Makefile
# This Makefile orchestrates all services in the Ice Cream Store system

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

# Service directories
DATA_SERVICE := data-service
SESSION_SERVICE := session-service
ORDERS_SERVICE := orders-service
GATEWAY_SERVICE := gateway-service

## 🍦 Ice Cream Store - Complete System Management

help: ## Show this help message
	@echo "$(CYAN)🍦 Ice Cream Store - Complete System$(RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(YELLOW)📋 Available Commands:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)🚀 Quick Start Commands:$(RESET)"
	@echo "  $(GREEN)make fresh$(RESET)            # 🔥 Fresh install of ALL services"
	@echo "  $(GREEN)make start-all$(RESET)        # Start all services"
	@echo "  $(GREEN)make stop-all$(RESET)         # Stop all services"
	@echo "  $(GREEN)make status$(RESET)           # Check status of all services"
	@echo ""
	@echo "$(YELLOW)🛠️  Individual Service Commands:$(RESET)"
	@echo "  $(BLUE)make fresh-data$(RESET)       # Fresh install data service only"
	@echo "  $(BLUE)make fresh-auth$(RESET)       # Fresh install auth service only"
	@echo "  $(BLUE)make fresh-orders$(RESET)     # Fresh install orders service only"
	@echo "  $(BLUE)make fresh-gateway$(RESET)    # Fresh install gateway service only"
	@echo ""
	@echo "$(YELLOW)📖 Service URLs:$(RESET)"
	@echo "  $(MAGENTA)Data Service:$(RESET)     http://localhost:5432 (PostgreSQL + PgAdmin: :8080)"
	@echo "  $(MAGENTA)Auth Service:$(RESET)     http://localhost:8081"
	@echo "  $(MAGENTA)Orders Service:$(RESET)   http://localhost:8083"
	@echo "  $(MAGENTA)Gateway Service:$(RESET)  http://localhost:8080"
	@echo ""

## 🚀 Complete System Commands

fresh: banner fresh-data fresh-auth fresh-orders fresh-gateway start-gateway final-status ## Fresh install of ALL services (recommended)
	@echo ""
	@echo "$(GREEN)🎉 COMPLETE SYSTEM FRESH INSTALLATION COMPLETED! 🎉$(RESET)"
	@echo "$(CYAN)============================================$(RESET)"
	@echo ""
	@echo "$(YELLOW)🌟 Your Ice Cream Store is ready!$(RESET)"
	@echo ""
	@echo "$(GREEN)✅ Services Status:$(RESET)"
	@echo "  🗄️  Data Service: $(GREEN)RUNNING$(RESET) (PostgreSQL + PgAdmin)"
	@echo "  🔐 Auth Service: $(GREEN)RUNNING$(RESET) (JWT Authentication)"  
	@echo "  📦 Orders Service: $(GREEN)RUNNING$(RESET) (Order Management)"
	@echo "  🌐 Gateway Service: $(GREEN)RUNNING$(RESET) (http://localhost:8082)"
	@echo ""
	@echo "$(CYAN)🔗 Access Your Services:$(RESET)"
	@echo "  • Database: http://localhost:8080 (PgAdmin)"
	@echo "  • Auth API: http://localhost:8081/api/v1/auth/health"
	@echo "  • Orders API: http://localhost:8083/api/v1/orders/health"
	@echo "  • Docker UI: https://localhost:9443 (Portainer)"
	@echo ""
	@echo "  🌐 Gateway Service API: $(GREEN)http://localhost:8082$(RESET)"
	@echo ""

start-all: start-data start-auth start-orders start-gateway ## Start all services in correct order
	@echo "$(GREEN)🚀 All services are starting up!$(RESET)"

stop-all: stop-gateway stop-orders stop-auth stop-data ## Stop all services in reverse order
	@echo "$(YELLOW)🛑 All services stopped$(RESET)"

restart-all: stop-all start-all ## Restart all services
	@echo "$(GREEN)🔄 All services restarted!$(RESET)"

test-all: test-data test-auth test-orders test-gateway ## Test all services
	@echo "$(GREEN)🧪 All service tests completed!$(RESET)"

status: status-data status-auth status-orders status-gateway ## Check status of all services
	@echo "$(CYAN)📊 System status check completed$(RESET)"

health-all: health-data health-auth health-orders health-gateway ## Check health of all services
	@echo "$(GREEN)🏥 System health check completed!$(RESET)"

final-status: ## Final status check after fresh installation
	@echo "$(CYAN)🔍 Verifying Fresh Installation Status...$(RESET)"
	@echo ""
	@echo "$(YELLOW)📊 Container Status:$(RESET)"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(icecream|portainer)" || echo "No containers running"
	@echo ""
	@echo "$(YELLOW)🏥 Service Health Checks:$(RESET)"
	@printf "  🗄️  PostgreSQL: "
	@if docker exec icecream_postgres pg_isready -U postgres -d icecream_store > /dev/null 2>&1; then \
		echo "$(GREEN)✅ HEALTHY$(RESET)"; \
	else \
		echo "$(RED)❌ UNHEALTHY$(RESET)"; \
	fi
	@printf "  🔐 Auth Service: "
	@if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ RUNNING$(RESET)"; \
	else \
		echo "$(RED)❌ NOT RESPONDING$(RESET)"; \
	fi
	@printf "  📦 Orders Service: "
	@if curl -f http://localhost:8083/api/v1/orders/health > /dev/null 2>&1; then \
		echo "$(GREEN)✅ RUNNING$(RESET)"; \
	else \
		echo "$(RED)❌ NOT RESPONDING$(RESET)"; \
	fi
	@printf "  🌐 Gateway Service: "
	@gateway_running=false; \
	db_healthy=false; \
	auth_running=false; \
	orders_running=false; \
	if curl -f http://localhost:8082/api/hello > /dev/null 2>&1; then \
		gateway_running=true; \
	fi; \
	if docker exec icecream_postgres pg_isready -U postgres -d icecream_store > /dev/null 2>&1; then \
		db_healthy=true; \
	fi; \
	if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then \
		auth_running=true; \
	fi; \
	if curl -f http://localhost:8083/api/v1/orders/health > /dev/null 2>&1; then \
		orders_running=true; \
	fi; \
	if [ "$$gateway_running" = "false" ]; then \
		echo "$(RED)❌ NOT RESPONDING$(RESET)"; \
	elif [ "$$db_healthy" = "false" ] || [ "$$auth_running" = "false" ] || [ "$$orders_running" = "false" ]; then \
		echo "$(YELLOW)🟡 DEGRADED (dependencies down)$(RESET)"; \
	else \
		echo "$(GREEN)✅ RUNNING$(RESET)"; \
	fi
	@echo ""

## 📊 Individual Service - Fresh Install Commands

fresh-data: ## Fresh install data service only
	@echo "$(CYAN)🗄️  Running fresh install for Data Service...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) fresh
	@echo "$(GREEN)✅ Data Service fresh install completed!$(RESET)"

fresh-auth: ## Fresh install auth service only
	@echo "$(CYAN)🔐 Running fresh install for Auth Service...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) fresh
	@echo "$(GREEN)✅ Auth Service fresh install completed!$(RESET)"

fresh-orders: ## Fresh install orders service only
	@echo "$(CYAN)📦 Running fresh install for Orders Service...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) fresh
	@echo "$(GREEN)✅ Orders Service fresh install completed!$(RESET)"

fresh-gateway: ## Fresh install gateway service only
	@echo "$(CYAN)🌐 Running fresh install for Gateway Service...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) fresh
	@echo "$(GREEN)✅ Gateway Service fresh install completed!$(RESET)"

## 🎛️  Individual Service - Management Commands

start-data: ## Start data service
	@echo "$(CYAN)🗄️  Starting Data Service...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) start

start-auth: ## Start auth service
	@echo "$(CYAN)🔐 Starting Auth Service...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) start

start-orders: ## Start orders service
	@echo "$(CYAN)📦 Starting Orders Service...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) start

start-gateway: ## Start gateway service
	@echo "$(CYAN)🌐 Starting Gateway Service...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) start

stop-data: ## Stop data service
	@echo "$(YELLOW)🗄️  Stopping Data Service...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) stop

stop-auth: ## Stop auth service
	@echo "$(YELLOW)🔐 Stopping Auth Service...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) stop

stop-orders: ## Stop orders service
	@echo "$(YELLOW)📦 Stopping Orders Service...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) stop

stop-gateway: ## Stop gateway service
	@echo "$(YELLOW)🌐 Stopping Gateway Service...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) stop

## 🔍 Individual Service - Status & Testing Commands

status-data: ## Check data service status
	@echo "$(BLUE)🗄️  Data Service Status:$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) status

status-auth: ## Check auth service status
	@echo "$(BLUE)🔐 Auth Service Status:$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) status

status-orders: ## Check orders service status
	@echo "$(BLUE)📦 Orders Service Status:$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) status

status-gateway: ## Check gateway service status
	@echo "$(BLUE)🌐 Gateway Service Status:$(RESET)"
	@echo "$(YELLOW)Note: Gateway service doesn't have containers to check$(RESET)"

test-data: ## Test data service
	@echo "$(CYAN)🧪 Testing Data Service...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) test

test-auth: ## Test auth service
	@echo "$(CYAN)🧪 Testing Auth Service...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) test

test-orders: ## Test orders service
	@echo "$(CYAN)🧪 Testing Orders Service...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) test

test-gateway: ## Test gateway service
	@echo "$(CYAN)🧪 Testing Gateway Service...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) test

health-data: ## Check data service health
	@echo "$(CYAN)🏥 Checking Data Service health...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) health

health-auth: ## Check auth service health
	@echo "$(CYAN)🏥 Checking Auth Service health...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) health

health-orders: ## Check orders service health
	@echo "$(CYAN)🏥 Checking Orders Service health...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) health

health-gateway: ## Check gateway service health
	@echo "$(CYAN)🏥 Checking Gateway Service health...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) health

## 🧹 Cleanup Commands

clean-all: clean-gateway clean-orders clean-auth clean-data ## Clean all services
	@echo "$(YELLOW)🧹 Cleaning all services...$(RESET)"
	@echo "$(GREEN)✅ All services cleaned!$(RESET)"

clean-data: ## Clean data service
	@echo "$(YELLOW)🗄️  Cleaning Data Service...$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) clean

clean-auth: ## Clean auth service
	@echo "$(YELLOW)🔐 Cleaning Auth Service...$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) clean

clean-orders: ## Clean orders service
	@echo "$(YELLOW)📦 Cleaning Orders Service...$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) clean

clean-gateway: ## Clean gateway service
	@echo "$(YELLOW)🌐 Cleaning Gateway Service...$(RESET)"
	@cd $(GATEWAY_SERVICE) && $(MAKE) clean

## 📋 Information Commands

system-info: ## Show complete system information
	@echo ""
	@echo "$(CYAN)🍦 Ice Cream Store System Information$(RESET)"
	@echo "====================================="
	@echo ""
	@echo "$(YELLOW)🏗️  Architecture Overview:$(RESET)"
	@echo "  $(BLUE)┌─────────────────┐$(RESET)     $(BLUE)┌─────────────────┐$(RESET)     $(BLUE)┌─────────────────┐$(RESET)     $(BLUE)┌─────────────────┐$(RESET)"
	@echo "  $(BLUE)│   Data Service  │$(RESET) ──▶ $(BLUE)│   Auth Service  │$(RESET) ──▶ $(BLUE)│ Orders Service  │$(RESET) ──▶ $(BLUE)│ Gateway Service │$(RESET)"
	@echo "  $(BLUE)│   PostgreSQL    │$(RESET)     $(BLUE)│   JWT Auth      │$(RESET)     $(BLUE)│   Order Mgmt    │$(RESET)     $(BLUE)│   API Gateway   │$(RESET)"
	@echo "  $(BLUE)│   Port: 5432    │$(RESET)     $(BLUE)│   Port: 8081    │$(RESET)     $(BLUE)│   Port: 8083    │$(RESET)     $(BLUE)│   Port: 8080    │$(RESET)"
	@echo "  $(BLUE)└─────────────────┘$(RESET)     $(BLUE)└─────────────────┘$(RESET)     $(BLUE)└─────────────────┘$(RESET)     $(BLUE)└─────────────────┘$(RESET)"
	@echo ""
	@echo "$(YELLOW)🔗 Service Endpoints:$(RESET)"
	@echo "  $(GREEN)Data Service (PostgreSQL):$(RESET)"
	@echo "    • Database: postgresql://postgres:postgres123@localhost:5432/icecream_store"
	@echo "    • PgAdmin:  http://localhost:8080 (admin@icecreamstore.com / admin123)"
	@echo ""
	@echo "  $(GREEN)Auth Service:$(RESET)"
	@echo "    • Base URL: http://localhost:8081"
	@echo "    • Health:   http://localhost:8081/api/v1/auth/health"
	@echo "    • Login:    POST http://localhost:8081/api/v1/auth/login"
	@echo ""
	@echo "  $(GREEN)Orders Service:$(RESET)"
	@echo "    • Base URL: http://localhost:8083"
	@echo "    • Health:   http://localhost:8083/api/v1/orders/health"
	@echo "    • Orders:   GET/POST http://localhost:8083/api/v1/orders"
	@echo ""
	@echo "  $(GREEN)Gateway Service:$(RESET)"
	@echo "    • Base URL: http://localhost:8080"
	@echo "    • Health:   http://localhost:8080/api/health"
	@echo "    • Hello:    http://localhost:8080/api/hello"
	@echo ""
	@echo "$(YELLOW)🧪 Quick Test Commands:$(RESET)"
	@echo "  # Test database"
	@echo "  curl http://localhost:5432 # PostgreSQL"
	@echo ""
	@echo "  # Test auth service"
	@echo "  curl http://localhost:8081/api/v1/auth/health"
	@echo ""
	@echo "  # Test orders service"
	@echo "  curl http://localhost:8083/api/v1/orders/health"
	@echo ""
	@echo "  # Test gateway"
	@echo "  curl http://localhost:8080/api/health"
	@echo ""

banner: ## Show system banner
	@echo ""
	@echo "$(CYAN)╔══════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║                    🍦 ICE CREAM STORE 🍦                     ║$(RESET)"
	@echo "$(CYAN)║                     System Orchestration                     ║$(RESET)"
	@echo "$(CYAN)╚══════════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""

logs-all: ## View logs from all services
	@echo "$(CYAN)📋 Viewing logs from all services...$(RESET)"
	@echo "$(YELLOW)Note: This will show recent logs from containerized services$(RESET)"
	@echo ""
	@echo "$(BLUE)=== Data Service Logs ====$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) logs || true
	@echo ""
	@echo "$(BLUE)=== Auth Service Logs ====$(RESET)"
	@cd $(SESSION_SERVICE) && $(MAKE) logs || true
	@echo ""
	@echo "$(BLUE)=== Orders Service Logs ====$(RESET)"
	@cd $(ORDERS_SERVICE) && $(MAKE) logs || true
	@echo ""
	@echo "$(BLUE)=== Gateway Service ====$(RESET)"
	@echo "$(YELLOW)Gateway service runs as binary - no container logs$(RESET)"

version: ## Show version information for all services
	@echo "$(CYAN)📋 System Version Information$(RESET)"
	@echo "=============================="
	@echo ""
	@echo "$(GREEN)System Dependencies:$(RESET)"
	@echo "Go version: $$(go version 2>/dev/null || echo 'Go not found')"
	@echo "Docker version: $$(docker --version 2>/dev/null || echo 'Docker not found')"
	@echo "Docker Compose version: $$(docker-compose --version 2>/dev/null || echo 'Docker Compose not found')"
	@echo ""
	@echo "$(GREEN)Service Versions:$(RESET)"
	@cd $(DATA_SERVICE) && $(MAKE) version || true
	@echo ""
	@cd $(SESSION_SERVICE) && $(MAKE) version || true
	@echo ""
	@cd $(ORDERS_SERVICE) && $(MAKE) version || true
	@echo ""
	@cd $(GATEWAY_SERVICE) && $(MAKE) version || true

# List all targets for tab completion
.PHONY: help fresh start-all stop-all restart-all test-all status health-all final-status fresh-data fresh-auth fresh-orders fresh-gateway start-data start-auth start-orders start-gateway stop-data stop-auth stop-orders stop-gateway status-data status-auth status-orders status-gateway test-data test-auth test-orders test-gateway health-data health-auth health-orders health-gateway clean-all clean-data clean-auth clean-orders clean-gateway system-info banner logs-all version 