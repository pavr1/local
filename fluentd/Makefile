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
	@echo "$(CYAN)Fluentd Service - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  $(GREEN)make start-locally$(RESET)    # 🚀 Start centralized logging stack locally"
	@echo "  $(GREEN)make stop-locally$(RESET)     # 🛑 Stop centralized logging stack"
	@echo "  $(GREEN)make restart-locally$(RESET)  # 🔄 Restart centralized logging stack"
	@echo "  $(GREEN)make status$(RESET)           # Check status of logging stack"
	@echo ""
	@echo "$(GREEN)Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

## 🚀 Local Development Commands

start-locally: ## Start centralized logging stack locally (Docker containers)
	@echo "$(CYAN)📊 Starting centralized logging stack locally...$(RESET)"
	@echo "$(YELLOW)🔍 Checking for existing containers...$(RESET)"
	@if docker-compose -f docker-compose.logging.yml ps | grep -q "Up"; then \
		echo "$(YELLOW)⚠️  Found running containers, stopping them first...$(RESET)"; \
		docker-compose -f docker-compose.logging.yml down; \
		sleep 2; \
	fi
	@echo "$(CYAN)📝 Starting logging stack containers...$(RESET)"
	@docker-compose -f docker-compose.logging.yml up -d
	@echo "$(GREEN)✅ Logging stack started successfully!$(RESET)"
	@echo "$(CYAN)📊 Kibana available at: http://localhost:5601$(RESET)"
	@echo "$(CYAN)🔍 Elasticsearch available at: http://localhost:9200$(RESET)"
	@echo "$(CYAN)📋 Fluentd listening on port 24224$(RESET)"

stop-locally: ## Stop centralized logging stack
	@echo "$(CYAN)📊 Stopping centralized logging stack...$(RESET)"
	@echo "$(YELLOW)🔍 Finding and stopping containers...$(RESET)"
	@docker-compose -f docker-compose.logging.yml down
	@echo "$(GREEN)✅ Logging stack stopped!$(RESET)"

restart-locally: stop-locally start-locally ## Restart centralized logging stack
	@echo "$(GREEN)✅ Logging stack restarted!$(RESET)"

## 🐳 Container Management Commands

start: start-locally ## [DEPRECATED] Use start-locally instead
	@echo "$(YELLOW)⚠️  'make start' is deprecated, use 'make start-locally' instead$(RESET)"

stop: stop-locally ## [DEPRECATED] Use stop-locally instead
	@echo "$(YELLOW)⚠️  'make stop' is deprecated, use 'make stop-locally' instead$(RESET)"

restart: restart-locally ## [DEPRECATED] Use restart-locally instead
	@echo "$(YELLOW)⚠️  'make restart' is deprecated, use 'make restart-locally' instead$(RESET)"

## 🔍 Monitoring Commands

status: ## Check status of centralized logging stack
	@echo "$(CYAN)📊 Checking logging stack status...$(RESET)"
	@docker-compose -f docker-compose.logging.yml ps

logs: ## View logs from all logging stack containers
	@echo "$(CYAN)📋 Viewing logs from logging stack...$(RESET)"
	@docker-compose -f docker-compose.logging.yml logs

logs-follow: ## Follow logs from all logging stack containers
	@echo "$(CYAN)📋 Following logs from logging stack...$(RESET)"
	@docker-compose -f docker-compose.logging.yml logs -f

logs-fluentd: ## View Fluentd logs
	@echo "$(CYAN)📋 Viewing Fluentd logs...$(RESET)"
	@docker-compose -f docker-compose.logging.yml logs fluentd

logs-elasticsearch: ## View Elasticsearch logs
	@echo "$(CYAN)📋 Viewing Elasticsearch logs...$(RESET)"
	@docker-compose -f docker-compose.logging.yml logs elasticsearch

logs-kibana: ## View Kibana logs
	@echo "$(CYAN)📋 Viewing Kibana logs...$(RESET)"
	@docker-compose -f docker-compose.logging.yml logs kibana

## 🧪 Testing Commands

test-elasticsearch: ## Test Elasticsearch connectivity
	@echo "$(CYAN)🧪 Testing Elasticsearch connectivity...$(RESET)"
	@curl -f http://localhost:9200/_cluster/health || echo "$(RED)❌ Elasticsearch is not responding$(RESET)"

test-kibana: ## Test Kibana connectivity
	@echo "$(CYAN)🧪 Testing Kibana connectivity...$(RESET)"
	@curl -f http://localhost:5601/api/status || echo "$(RED)❌ Kibana is not responding$(RESET)"

test-fluentd: ## Test Fluentd connectivity
	@echo "$(CYAN)🧪 Testing Fluentd connectivity...$(RESET)"
	@nc -z localhost 24224 && echo "$(GREEN)✅ Fluentd is listening on port 24224$(RESET)" || echo "$(RED)❌ Fluentd is not listening on port 24224$(RESET)"

test-all: test-elasticsearch test-kibana test-fluentd ## Test all logging stack components

## 🔧 Configuration Commands

config-validate: ## Validate Fluentd configuration
	@echo "$(CYAN)🔧 Validating Fluentd configuration...$(RESET)"
	@docker-compose -f docker-compose.logging.yml run --rm fluentd fluentd -c /fluentd/etc/fluent.conf --dry-run

config-reload: ## Reload Fluentd configuration
	@echo "$(CYAN)🔧 Reloading Fluentd configuration...$(RESET)"
	@docker-compose -f docker-compose.logging.yml restart fluentd
	@echo "$(GREEN)✅ Fluentd configuration reloaded!$(RESET)"

## 🧹 Cleanup Commands

clean: ## Clean up logging stack containers and volumes
	@echo "$(CYAN)🧹 Cleaning up logging stack...$(RESET)"
	@docker-compose -f docker-compose.logging.yml down -v
	@echo "$(GREEN)✅ Logging stack cleaned up!$(RESET)"

clean-logs: ## Clean up log files
	@echo "$(CYAN)🧹 Cleaning up log files...$(RESET)"
	@rm -rf logs/*.log logs/*.pos 2>/dev/null || true
	@echo "$(GREEN)✅ Log files cleaned up!$(RESET)"

## 📊 Information Commands

info: ## Show logging stack information
	@echo "$(CYAN)📊 Logging Stack Information:$(RESET)"
	@echo "  Service: fluentd-logging-stack"
	@echo "  Fluentd Port: 24224"
	@echo "  Elasticsearch Port: 9200"
	@echo "  Kibana Port: 5601"
	@echo "  Docker Compose: docker-compose.logging.yml"
	@echo ""

health: ## Check health of all logging stack components
	@echo "$(CYAN)🔍 Checking logging stack health...$(RESET)"
	@curl -s http://localhost:8087/p/health | jq '.' 2>/dev/null || echo "$(RED)Health check service not responding$(RESET)"

health-verbose: ## Check health of all logging stack components (verbose)
	@echo "$(CYAN)🔍 Checking logging stack health (verbose)...$(RESET)"
	@echo "$(YELLOW)Elasticsearch:$(RESET)"
	@curl -s http://localhost:9200/_cluster/health | jq '.status' 2>/dev/null || echo "$(RED)Not responding$(RESET)"
	@echo "$(YELLOW)Kibana:$(RESET)"
	@curl -s http://localhost:5601/api/status | jq '.status.overall.level' 2>/dev/null || echo "$(RED)Not responding$(RESET)"
	@echo "$(YELLOW)Fluentd:$(RESET)"
	@nc -z localhost 24224 && echo "$(GREEN)Listening$(RESET)" || echo "$(RED)Not listening$(RESET)"

## 🚀 Lifecycle Commands

install: ## Complete setup and start logging stack
	@echo "$(CYAN)📦 Installing logging stack...$(RESET)"
	@make start-locally
	@echo "$(GREEN)🎉 Logging stack is ready!$(RESET)"
	@echo "$(CYAN)Access Kibana at: http://localhost:5601$(RESET)"

fresh: clean install ## Fresh start: clean, install, start
	@echo "$(GREEN)🎉 Fresh logging stack setup completed!$(RESET)"

## 📋 Utility Commands

check-deps: ## Check if required dependencies are installed
	@echo "$(CYAN)🔍 Checking dependencies...$(RESET)"
	@command -v docker >/dev/null 2>&1 || (echo "$(RED)❌ Docker is not installed$(RESET)" && exit 1)
	@command -v docker-compose >/dev/null 2>&1 || (echo "$(RED)❌ Docker Compose is not installed$(RESET)" && exit 1)
	@echo "$(GREEN)✅ All dependencies are installed$(RESET)"

version: ## Show version information
	@echo "$(CYAN)📋 Version Information:$(RESET)"
	@echo "  Fluentd: $(shell docker-compose -f docker-compose.logging.yml images fluentd --format "table {{.Tag}}" | tail -n +2)"
	@echo "  Elasticsearch: $(shell docker-compose -f docker-compose.logging.yml images elasticsearch --format "table {{.Tag}}" | tail -n +2)"
	@echo "  Kibana: $(shell docker-compose -f docker-compose.logging.yml images kibana --format "table {{.Tag}}" | tail -n +2)"
	@echo "  Docker: $(shell docker --version)"
	@echo "  Docker Compose: $(shell docker-compose --version)"

## 📝 Documentation

docs: ## Generate documentation
	@echo "$(CYAN)📝 Generating documentation...$(RESET)"
	@echo "# Fluentd Logging Stack Documentation" > docs.txt
	@echo "" >> docs.txt
	@echo "## Services:" >> docs.txt
	@echo "- Fluentd: Log collector and aggregator" >> docs.txt
	@echo "- Elasticsearch: Log storage and search engine" >> docs.txt
	@echo "- Kibana: Log visualization and management" >> docs.txt
	@echo "" >> docs.txt
	@echo "## Ports:" >> docs.txt
	@echo "- Fluentd: 24224" >> docs.txt
	@echo "- Elasticsearch: 9200" >> docs.txt
	@echo "- Kibana: 5601" >> docs.txt
	@echo "$(GREEN)✅ Documentation generated: docs.txt$(RESET)"

# List all targets for tab completion
.PHONY: help start-locally stop-locally restart-locally start stop restart status logs logs-follow logs-fluentd logs-elasticsearch logs-kibana test-elasticsearch test-kibana test-fluentd test-all config-validate config-reload clean clean-logs info health health-verbose install fresh check-deps version docs 