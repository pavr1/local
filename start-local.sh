#!/bin/bash

# Ice Cream Store - Local Development Script
# This script runs all services locally (non-containerized) for better performance
# Only the database remains containerized

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
RESET='\033[0m'

# Service configuration
DATA_SERVICE_DIR="data-service"
SESSION_SERVICE_DIR="session-service"
ORDERS_SERVICE_DIR="orders-service"
GATEWAY_SERVICE_DIR="gateway-service"
INVENTORY_SERVICE_DIR="inventory-service"
UI_SERVICE_DIR="ui"

# Ports
SESSION_PORT=8081
GATEWAY_PORT=8082
ORDERS_PORT=8083
INVENTORY_PORT=8084
UI_PORT=3000
DB_PORT=5432
PGADMIN_PORT=8080

# PID storage directory
PID_DIR=".local-pids"
mkdir -p "$PID_DIR"

# Function to print banner
print_banner() {
    echo -e "${CYAN}"
    echo "🍦 ========================================"
    echo "   ICE CREAM STORE - LOCAL DEVELOPMENT"
    echo "   Running services locally for performance"
    echo "========================================${RESET}"
    echo ""
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}❌ Docker is not running. Please start Docker first.${RESET}"
        exit 1
    fi
}

# Function to start database (containerized)
start_database() {
    echo -e "${YELLOW}🗄️  Starting Database Service (containerized)...${RESET}"
    cd "$DATA_SERVICE_DIR/docker" || exit 1
    
    # Stop any existing containers
    docker-compose down > /dev/null 2>&1
    
    # Start database
    if docker-compose up -d; then
        echo -e "${GREEN}✅ Database service started${RESET}"
        
        # Wait for database to be ready
        echo -e "${YELLOW}⏳ Waiting for database to be ready...${RESET}"
        for i in {1..30}; do
            if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
                echo -e "${GREEN}✅ Database is ready${RESET}"
                break
            fi
            echo -n "."
            sleep 1
        done
        echo ""
    else
        echo -e "${RED}❌ Failed to start database service${RESET}"
        exit 1
    fi
    cd - > /dev/null
}

# Function to build Go service
build_service() {
    local service_dir=$1
    local service_name=$2
    
    echo -e "${YELLOW}🔨 Building $service_name...${RESET}"
    cd "$service_dir" || return 1
    
    if go build -o "${service_name}" .; then
        echo -e "${GREEN}✅ $service_name built successfully${RESET}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}❌ Failed to build $service_name${RESET}"
        cd - > /dev/null
        return 1
    fi
}

# Function to start Go service
start_go_service() {
    local service_dir=$1
    local service_name=$2
    local port=$3
    local binary_name=$4
    
    echo -e "${YELLOW}🚀 Starting $service_name on port $port...${RESET}"
    cd "$service_dir" || return 1
    
    # Kill any existing process on this port
    kill_process_on_port "$port"
    
    # Start the service in background
    nohup "./$binary_name" > "../logs/${service_name}.log" 2>&1 &
    local pid=$!
    echo "$pid" > "../$PID_DIR/${service_name}.pid"
    
    # Wait a moment and check if it started
    sleep 2
    if kill -0 "$pid" 2>/dev/null; then
        echo -e "${GREEN}✅ $service_name started (PID: $pid)${RESET}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}❌ Failed to start $service_name${RESET}"
        cd - > /dev/null
        return 1
    fi
}

# Function to start UI service
start_ui_service() {
    echo -e "${YELLOW}🎨 Starting UI Service on port $UI_PORT...${RESET}"
    
    # Check if npm is available
    if ! command -v npm &> /dev/null; then
        echo -e "${YELLOW}⚠️  npm not found. Skipping UI service.${RESET}"
        echo -e "${YELLOW}💡 Install Node.js to enable UI: brew install node${RESET}"
        return 0
    fi
    
    cd "$UI_SERVICE_DIR" || return 1
    
    # Kill any existing process on this port
    kill_process_on_port "$UI_PORT"
    
    # Check if we need to install dependencies
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}📦 Installing UI dependencies...${RESET}"
        if ! npm install; then
            echo -e "${RED}❌ Failed to install UI dependencies${RESET}"
            cd - > /dev/null
            return 1
        fi
    fi
    
    # Start the UI service
    nohup npm start > "../logs/ui-service.log" 2>&1 &
    local pid=$!
    echo "$pid" > "../$PID_DIR/ui-service.pid"
    
    sleep 3
    if kill -0 "$pid" 2>/dev/null; then
        echo -e "${GREEN}✅ UI service started (PID: $pid)${RESET}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}❌ Failed to start UI service${RESET}"
        cd - > /dev/null
        return 1
    fi
}

# Function to kill process on specific port
kill_process_on_port() {
    local port=$1
    local pid=$(lsof -ti:"$port" 2>/dev/null)
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}🔄 Killing existing process on port $port (PID: $pid)${RESET}"
        kill -9 "$pid" 2>/dev/null
        sleep 1
    fi
}

# Function to check service health
check_service_health() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    
    echo -e "${YELLOW}🏥 Checking $service_name health...${RESET}"
    for i in $(seq 1 $max_attempts); do
        if curl -s "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $service_name is healthy${RESET}"
            return 0
        fi
        echo -n "."
        sleep 1
    done
    echo -e "${RED}❌ $service_name health check failed${RESET}"
    return 1
}

# Function to show service status
show_status() {
    echo -e "${CYAN}📊 Service Status:${RESET}"
    echo "=================================="
    
    # Database
    if docker ps --format "table {{.Names}}" | grep -q "icecream_postgres"; then
        echo -e "🗄️  Database:        ${GREEN}RUNNING${RESET} (Docker)"
    else
        echo -e "🗄️  Database:        ${RED}STOPPED${RESET}"
    fi
    
    # Go services
    for service in "session-service:$SESSION_PORT" "orders-service:$ORDERS_PORT" "gateway-service:$GATEWAY_PORT" "inventory-service:$INVENTORY_PORT"; do
        IFS=':' read -r name port <<< "$service"
        if [ -f "$PID_DIR/${name}.pid" ] && kill -0 "$(cat "$PID_DIR/${name}.pid")" 2>/dev/null; then
            echo -e "🔧 ${name}:${GREEN}RUNNING${RESET} (Port: $port)"
        else
            echo -e "🔧 ${name}:${RED}STOPPED${RESET}"
        fi
    done
    
    # UI service
    if [ -f "$PID_DIR/ui-service.pid" ] && kill -0 "$(cat "$PID_DIR/ui-service.pid")" 2>/dev/null; then
        echo -e "🎨 UI Service:        ${GREEN}RUNNING${RESET} (Port: $UI_PORT)"
    else
        echo -e "🎨 UI Service:        ${RED}STOPPED${RESET}"
    fi
    
    echo ""
    echo -e "${CYAN}🔗 Service URLs:${RESET}"
    echo "  • UI Application:    http://localhost:$UI_PORT"
    echo "  • Gateway API:       http://localhost:$GATEWAY_PORT"
    echo "  • Session API:       http://localhost:$SESSION_PORT"
    echo "  • Orders API:        http://localhost:$ORDERS_PORT"
    echo "  • Inventory API:     http://localhost:$INVENTORY_PORT"
    echo "  • Database Admin:    http://localhost:$PGADMIN_PORT"
    echo ""
}

# Function to stop all services
stop_all_services() {
    echo -e "${YELLOW}🛑 Stopping all services...${RESET}"
    
    # Stop Go services
    for service in "session-service" "orders-service" "gateway-service" "inventory-service" "ui-service"; do
        if [ -f "$PID_DIR/${service}.pid" ]; then
            local pid=$(cat "$PID_DIR/${service}.pid")
            if kill -0 "$pid" 2>/dev/null; then
                echo -e "${YELLOW}🔄 Stopping $service (PID: $pid)...${RESET}"
                kill "$pid" 2>/dev/null
                # Wait for graceful shutdown
                for i in {1..5}; do
                    if ! kill -0 "$pid" 2>/dev/null; then
                        break
                    fi
                    sleep 1
                done
                # Force kill if still running
                if kill -0 "$pid" 2>/dev/null; then
                    kill -9 "$pid" 2>/dev/null
                fi
            fi
            rm -f "$PID_DIR/${service}.pid"
        fi
    done
    
    # Stop database
    echo -e "${YELLOW}🗄️  Stopping database...${RESET}"
    cd "$DATA_SERVICE_DIR/docker" && docker-compose down > /dev/null 2>&1
    cd - > /dev/null
    
    # Clean up port conflicts
    for port in $SESSION_PORT $ORDERS_PORT $GATEWAY_PORT $INVENTORY_PORT $UI_PORT; do
        kill_process_on_port "$port"
    done
    
    echo -e "${GREEN}✅ All services stopped${RESET}"
}

# Function to show logs
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        echo -e "${YELLOW}Available logs:${RESET}"
        ls -1 logs/ 2>/dev/null | sed 's/^/  /'
        return
    fi
    
    if [ -f "logs/${service}.log" ]; then
        tail -f "logs/${service}.log"
    else
        echo -e "${RED}❌ Log file not found: logs/${service}.log${RESET}"
    fi
}

# Main script logic
case "${1:-start}" in
    "start")
        # Create logs directory
        mkdir -p logs
        
        print_banner
        check_docker
        
        echo -e "${CYAN}🚀 Starting Ice Cream Store (Local Mode)...${RESET}"
        echo ""
        
        # Start database first
        start_database
        echo ""
        
        # Build and start Go services
        echo -e "${CYAN}🔧 Building and starting Go services...${RESET}"
        
        build_service "$SESSION_SERVICE_DIR" "session-service" && \
        start_go_service "$SESSION_SERVICE_DIR" "session-service" "$SESSION_PORT" "session-service"
        echo ""
        
        build_service "$ORDERS_SERVICE_DIR" "orders-service" && \
        start_go_service "$ORDERS_SERVICE_DIR" "orders-service" "$ORDERS_PORT" "orders-service"
        echo ""
        
        build_service "$INVENTORY_SERVICE_DIR" "inventory-service" && \
        start_go_service "$INVENTORY_SERVICE_DIR" "inventory-service" "$INVENTORY_PORT" "inventory-service"
        echo ""
        
        build_service "$GATEWAY_SERVICE_DIR" "gateway-service" && \
        start_go_service "$GATEWAY_SERVICE_DIR" "gateway-service" "$GATEWAY_PORT" "gateway-service"
        echo ""
        
        # Start UI service
        start_ui_service
        echo ""
        
        # Health checks
        echo -e "${CYAN}🏥 Performing health checks...${RESET}"
        sleep 5
        check_service_health "Session Service" "http://localhost:$SESSION_PORT/api/v1/sessions/p/health"
        check_service_health "Orders Service" "http://localhost:$ORDERS_PORT/api/v1/orders/p/health"
        check_service_health "Inventory Service" "http://localhost:$INVENTORY_PORT/api/v1/inventory/p/health"
        check_service_health "Gateway Service" "http://localhost:$GATEWAY_PORT/api/health"
        echo ""
        
        show_status
        
        echo -e "${GREEN}🎉 Ice Cream Store is running locally!${RESET}"
        echo -e "${YELLOW}💡 Use './start-local.sh stop' to stop all services${RESET}"
        echo -e "${YELLOW}💡 Use './start-local.sh status' to check status${RESET}"
        echo -e "${YELLOW}💡 Use './start-local.sh logs <service-name>' to view logs${RESET}"
        ;;
        
    "stop")
        print_banner
        stop_all_services
        ;;
        
    "restart")
        print_banner
        stop_all_services
        echo ""
        exec "$0" start
        ;;
        
    "status")
        print_banner
        show_status
        ;;
        
    "logs")
        show_logs "$2"
        ;;
        
    "health")
        print_banner
        echo -e "${CYAN}🏥 Health Check Results:${RESET}"
        check_service_health "Session Service" "http://localhost:$SESSION_PORT/api/v1/sessions/p/health"
        check_service_health "Orders Service" "http://localhost:$ORDERS_PORT/api/v1/orders/p/health"  
        check_service_health "Inventory Service" "http://localhost:$INVENTORY_PORT/api/v1/inventory/p/health"
        check_service_health "Gateway Service" "http://localhost:$GATEWAY_PORT/api/health"
        ;;
        
    "help"|"-h"|"--help")
        print_banner
        echo -e "${YELLOW}Usage:${RESET}"
        echo "  ./start-local.sh [command]"
        echo ""
        echo -e "${YELLOW}Commands:${RESET}"
        echo "  start    - Start all services locally (default)"
        echo "  stop     - Stop all services"
        echo "  restart  - Restart all services"
        echo "  status   - Show service status"
        echo "  health   - Check service health"
        echo "  logs     - Show available logs or tail specific service log"
        echo "  help     - Show this help message"
        echo ""
        echo -e "${YELLOW}Examples:${RESET}"
        echo "  ./start-local.sh"
        echo "  ./start-local.sh stop"
        echo "  ./start-local.sh logs gateway-service"
        ;;
        
    *)
        echo -e "${RED}❌ Unknown command: $1${RESET}"
        echo "Use './start-local.sh help' for usage information"
        exit 1
        ;;
esac 