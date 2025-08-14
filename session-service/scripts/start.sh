#!/bin/bash

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

echo -e "${CYAN}🍦 Starting Ice Cream Store Session Service...${RESET}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${RESET}"
    exit 1
fi

# Check if the shared network exists
if ! docker network ls | grep -q "docker_icecream_network"; then
    echo -e "${YELLOW}⚠️  Creating shared network docker_icecream_network...${RESET}"
    docker network create docker_icecream_network
fi

# Check if data-service is running (which manages PostgreSQL)
if ! docker ps | grep -q "icecream_data_service"; then
    echo -e "${YELLOW}⚠️  Data-service is not running. Starting it...${RESET}"
    cd ../data-service && ./scripts/start.sh
    cd ../session-service
fi

# Wait for data-service to be ready (which verifies PostgreSQL)
echo -e "${CYAN}⏳ Waiting for data-service to be ready...${RESET}"
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8086/api/v1/data/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Data-service is ready!${RESET}"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Data-service API not ready yet... (trying: http://localhost:8086/api/v1/data/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Data-service failed to start within the expected time${RESET}"
    exit 1
fi

# Stop only session-service containers (don't touch data-service)
echo -e "${YELLOW}🧹 Cleaning up existing session containers...${RESET}"
cd docker && docker-compose down 2>/dev/null || true

# Start the service
echo -e "${CYAN}🚀 Starting session service...${RESET}"
docker-compose up -d

# Wait for the service to be ready
echo -e "${CYAN}⏳ Waiting for session-service API to be ready...${RESET}"
RETRY_COUNT=0
MAX_RETRIES=30

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8081/api/v1/sessions/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Session-service API is ready!${RESET}"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Session-service API not ready yet... (trying: http://localhost:8081/api/v1/sessions/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Session-service API failed to start within the expected time${RESET}"
    exit 1
fi

# Show container status
echo -e "${CYAN}📊 Container Status:${RESET}"
docker-compose ps

echo ""
echo -e "${GREEN}🎉 Session Service is ready!${RESET}"
echo ""
echo -e "${CYAN}📝 Service Details:${RESET}"
echo "   PostgreSQL:    localhost:5432"
echo "   Session API:   http://localhost:8081"
echo "   Health:        http://localhost:8081/api/v1/sessions/p/health"
echo ""
echo -e "${CYAN}📋 Useful Commands:${RESET}"
echo "   Stop service:     ./scripts/stop.sh"
echo "   View logs:        ./scripts/logs.sh"
echo "   Test database:    docker exec icecream_postgres pg_isready -U postgres -d icecream_store"
echo "   Test API:         curl http://localhost:8081/api/v1/sessions/p/health"
echo ""
echo -e "${CYAN}🐳 Session service containers started with DB_HOST=postgres${RESET}"
