# Ice Cream Store Data Service Makefile
# This Makefile provides convenient commands to manage the PostgreSQL database service

.PHONY: help start stop restart reset logs connect status health test test-unit test-integration build run clean install deps fresh

# Default target
.DEFAULT_GOAL := help

# Colors for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

## 🚀 Database Management Commands

create-network: ## Create the shared Docker network for all services
	@echo "$(CYAN)📡 Creating shared Docker network...$(RESET)"
	@if ! docker network ls | grep -q "docker_icecream_network"; then \
		echo "$(YELLOW)🔄 Creating docker_icecream_network...$(RESET)"; \
		docker network create docker_icecream_network; \
		echo "$(GREEN)✅ Network docker_icecream_network created$(RESET)"; \
	else \
		echo "$(GREEN)✅ Network docker_icecream_network already exists$(RESET)"; \
	fi

migrate-up: ## Apply all pending database migrations
	@echo "$(CYAN)📈 Applying database migrations...$(RESET)"
	@./scripts/migrate.sh up
	@echo "$(GREEN)✅ Migrations completed!$(RESET)"

migrate-up-force: ## Force apply migrations with verbose output (for debugging)
	@echo "$(CYAN)📈 Force applying database migrations with verbose output...$(RESET)"
	@echo "$(YELLOW)⚠️  This will show detailed migration output$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "SELECT 'Migration test connection' as status;"
	@echo "$(CYAN)🔍 Checking current table structure...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "\d existences" | head -20
	@echo "$(CYAN)📋 Applying migration manually...$(RESET)"
	@docker cp docker/init/migrations/002_update_existences_pricing_fields.up.sql icecream_postgres:/tmp/migration.sql
	@docker exec icecream_postgres psql -U postgres -d icecream_store -f /tmp/migration.sql
	@echo "$(CYAN)📋 Recording migration...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "INSERT INTO schema_migrations (version) VALUES ('002_update_existences_pricing_fields') ON CONFLICT (version) DO NOTHING;"
	@echo "$(CYAN)🔍 Verifying new table structure...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "\d existences" | head -20
	@echo "$(GREEN)✅ Force migration completed!$(RESET)"

migrate-down: ## Rollback last database migration
	@echo "$(YELLOW)📉 Rolling back last migration...$(RESET)"
	@./scripts/migrate.sh down
	@echo "$(GREEN)✅ Rollback completed!$(RESET)"

migrate-status: ## Show database migration status
	@echo "$(CYAN)📊 Checking migration status...$(RESET)"
	@./scripts/migrate.sh status

migrate-create: ## Create new migration files (usage: make migrate-create description="add new table")
	@echo "$(CYAN)📝 Creating new migration...$(RESET)"
	@./scripts/migrate.sh create "$(description)"
	@echo "$(GREEN)✅ Migration files created!$(RESET)"

migrate-reset: ## Reset migrations and reapply (⚠️  DESTRUCTIVE - removes migration history)
	@echo "$(RED)⚠️  WARNING: This will reset migration history and reapply all migrations!$(RESET)"
	@echo "$(YELLOW)This is useful when migrations get out of sync$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "DROP TABLE IF EXISTS schema_migrations;"
	@echo "$(CYAN)📋 Recreating migrations table...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -c "CREATE TABLE schema_migrations (version VARCHAR(255) PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"
	@echo "$(CYAN)📈 Reapplying all migrations...$(RESET)"
	@./scripts/migrate.sh up
	@echo "$(GREEN)✅ Migration reset completed!$(RESET)"

fresh: create-network clean deps remove-volumes reset start init-manual migrate-up test status info ## Complete fresh installation: clean everything, install, setup, and test
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

start: create-network ## Start PostgreSQL and PgAdmin containers
	@echo "$(CYAN)🍦 Starting Ice Cream Store Data Service...$(RESET)"
	@./scripts/start.sh

stop: ## Stop all database containers
	@echo "$(YELLOW)🛑 Stopping Data Service containers...$(RESET)"
	@./scripts/stop.sh

restart: stop start ## Restart all database containers
	@echo "$(GREEN)🔄 Data Service restarted successfully!$(RESET)"

## 🚀 Local Development Commands

start-locally: start deps ## Start data service containers and HTTP server locally
	@echo "$(CYAN)🚀 Starting Data Service HTTP server locally...$(RESET)"
	@echo "🛠️  Starting Data Service HTTP server locally..."
	@echo "🔍 Checking for existing processes on port 8086..."
	@if lsof -ti:8086 > /dev/null 2>&1; then \
		echo "⚠️  Found process using port 8086, killing it..."; \
		kill -9 $$(lsof -ti:8086) 2>/dev/null || true; \
		sleep 1; \
	fi
	@echo "⚠️  Make sure database containers are running first!"
	@echo "📝 Service will be available at: http://localhost:8086"
	@echo "🚀 Starting HTTP server in background with DB_HOST=localhost..."
	@DB_HOST=localhost nohup go run main.go > service.log 2>&1 & echo $$! > service.pid
	@sleep 2
	@echo "$(GREEN)✅ Data Service HTTP server started successfully$(RESET)"

stop-locally: stop ## Stop data service containers and HTTP server
	@echo "$(YELLOW)🛑 Stopping Data Service HTTP server...$(RESET)"
	@if [ -f service.pid ]; then \
		echo "Found service PID file, stopping..."; \
		kill -9 $$(cat service.pid) 2>/dev/null || true; \
		rm -f service.pid; \
		echo "✅ HTTP server stopped"; \
	else \
		echo "No PID file found, checking port 8086..."; \
		if lsof -ti:8086 > /dev/null 2>&1; then \
			echo "Found service on port 8086, stopping..."; \
			kill -9 $$(lsof -ti:8086) 2>/dev/null || true; \
			echo "✅ Service on port 8086 stopped"; \
		else \
			echo "✅ No HTTP server running"; \
		fi; \
	fi
	@pkill -f "go run main.go" 2>/dev/null || true
	@echo "$(GREEN)✅ Local service stopped$(RESET)"

start-docker: build-docker start ## Start data service containers (DB_HOST=postgres)
	@echo "$(CYAN)🐳 Data service containers started with DB_HOST=postgres$(RESET)"
stop-docker: stop ## [ALIAS] Stop data service containers
restart-docker: restart ## [ALIAS] Restart data service containers

reset: ## Reset database (⚠️  DELETES ALL DATA)
	@echo "$(RED)⚠️  WARNING: This will delete ALL database data!$(RESET)"
	@./scripts/reset.sh

remove-volumes: ## Remove database volumes to force fresh initialization
	@echo "$(YELLOW)🗑️  Removing database volumes...$(RESET)"
	@docker volume rm docker_postgres_data docker_pgadmin_data 2>/dev/null || true
	@echo "$(GREEN)✅ Database volumes removed$(RESET)"

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
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
	@echo "$(GREEN)✅ Build completed: main (Linux binary)$(RESET)"

build-docker: build ## Build the Docker image
	@echo "$(CYAN)🐳 Building Docker image...$(RESET)"
	@docker build -t docker-data-service -f docker/Dockerfile .
	@echo "$(GREEN)✅ Docker image built: docker-data-service$(RESET)"

run: build ## Run the data service application to test database connection
	@echo "$(CYAN)🧪 Running data service application...$(RESET)"
	@./bin/data-service

test: test-unit ## Run unit tests (no database required)
	@echo "$(GREEN)✅ All tests completed!$(RESET)"

test-unit: ## Run unit tests only (no database required)
	@echo "$(CYAN)🧪 Running unit tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✅ Unit tests completed!$(RESET)"

test-integration: health ## Run integration tests (requires database)
	@echo "$(CYAN)🧪 Running integration tests...$(RESET)"
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

deps: ## Download and tidy dependencies
	@echo "$(CYAN)📦 Managing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(RESET)"

init: ## Initialize database with fresh schema (requires containers to be running)
	@echo "$(CYAN)🔄 Reinitializing database schema...$(RESET)"
	@docker exec icecream_postgres psql -U postgres -d icecream_store -f /docker-entrypoint-initdb.d/01-init-database.sql > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Database schema reinitialized$(RESET)" || \
		echo "$(RED)❌ Failed to reinitialize schema. Try 'make reset' instead.$(RESET)"

init-manual: ## Manually initialize database schema (copy script and run it)
	@echo "$(CYAN)🔄 Manually initializing database schema...$(RESET)"
	@docker cp docker/init/01-init-database.sql icecream_postgres:/tmp/init.sql
	@docker exec icecream_postgres psql -U postgres -d icecream_store -f /tmp/init.sql > /dev/null 2>&1 && \
		echo "$(GREEN)✅ Database schema initialized successfully$(RESET)" || \
		(echo "$(RED)❌ Failed to initialize schema$(RESET)" && exit 1)

## 🧹 Cleanup Commands

clean: ## Clean up containers, volumes, and built binaries
	@echo "$(YELLOW)🧹 Cleaning up...$(RESET)"
	@cd docker && docker-compose down -v --remove-orphans 2>/dev/null || true
	@docker volume rm docker_postgres_data docker_pgadmin_data 2>/dev/null || true
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
	@echo "   Migrations:     make migrate-up / make migrate-down / make migrate-status"
	@echo "   Migration Debug: make migrate-up-force (verbose output)"
	@echo "   Migration Reset: make migrate-reset (when migrations are out of sync)"
	@echo "   Cleanup:        make clean"
	@echo ""
	@echo "$(YELLOW)🔧 Environment Variables:$(RESET)"
	@echo "   start-locally:  DB_HOST=localhost (connects to localhost:5432)"
	@echo "   start-docker:   DB_HOST=postgres (connects to container:5432)"
	@echo "" 