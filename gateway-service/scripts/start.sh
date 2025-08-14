#!/bin/bash

# Ice Cream Store Gateway Service - Start Script
# This script starts the gateway service using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🌐 Starting Ice Cream Store Gateway Service...${NC}"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if the shared network exists, create if it doesn't
if ! docker network ls | grep -q "docker_icecream_network"; then
    echo -e "${YELLOW}📡 Creating shared network 'docker_icecream_network'...${NC}"
    docker network create docker_icecream_network
fi

# Stop existing gateway service container if it exists
echo -e "${YELLOW}🧹 Cleaning up existing gateway service container...${NC}"
docker-compose down 2>/dev/null || true

# Start the service
echo -e "${BLUE}🚀 Starting Gateway Service container...${NC}"
docker-compose up -d

# Check if all business services are ready
echo -e "${YELLOW}⏳ Checking if all business services are ready...${NC}"

# Check session service
echo -e "${CYAN}🔍 Checking session service...${NC}"
MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8081/api/v1/sessions/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Session service is ready!${NC}"
        break
    fi
    echo -e "${YELLOW}   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Session service not ready yet... (trying: http://localhost:8081/api/v1/sessions/p/health)${NC}"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Session service failed to start within the expected time${NC}"
    exit 1
fi

# Check orders service
echo -e "${CYAN}🔍 Checking orders service...${NC}"
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8083/api/v1/orders/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Orders service is ready!${NC}"
        break
    fi
    echo -e "${YELLOW}   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Orders service not ready yet... (trying: http://localhost:8083/api/v1/orders/p/health)${NC}"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Orders service failed to start within the expected time${NC}"
    exit 1
fi

# Check inventory service
echo -e "${CYAN}🔍 Checking inventory service...${NC}"
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8084/api/v1/inventory/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Inventory service is ready!${NC}"
        break
    fi
    echo -e "${YELLOW}   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Inventory service not ready yet... (trying: http://localhost:8084/api/v1/inventory/p/health)${NC}"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Inventory service failed to start within the expected time${NC}"
    exit 1
fi

# Check invoice service
echo -e "${CYAN}🔍 Checking invoice service...${NC}"
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8085/api/v1/invoices/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Invoice service is ready!${NC}"
        break
    fi
    echo -e "${YELLOW}   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Invoice service not ready yet... (trying: http://localhost:8085/api/v1/invoices/p/health)${NC}"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Invoice service failed to start within the expected time${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All business services are ready!${NC}"

# Wait for gateway service to be ready
echo -e "${YELLOW}⏳ Waiting for Gateway Service to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:8082/api/v1/gateway/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Gateway Service is ready!${NC}"
        echo ""
        echo -e "${CYAN}📊 Gateway Service Status:${NC}"
        echo -e "   🌐 Gateway API: ${GREEN}http://localhost:8082${NC}"
        echo -e "   🔌 Health Check: ${GREEN}http://localhost:8082/api/v1/gateway/p/health${NC}"
        echo -e "   🔐 Session Proxy: ${GREEN}http://localhost:8082/api/v1/sessions/*${NC}"
        echo -e "   🛒 Orders Proxy: ${GREEN}http://localhost:8082/api/v1/orders/*${NC}"
        echo -e "   📦 Inventory Proxy: ${GREEN}http://localhost:8082/api/v1/inventory/*${NC}"
        echo -e "   📄 Invoice Proxy: ${GREEN}http://localhost:8082/api/v1/invoices/*${NC}"
        echo ""
        echo -e "${GREEN}🎉 Gateway Service started successfully!${NC}"
        exit 0
    fi
    echo -e "${YELLOW}   Attempt $i/30: Still waiting...${NC}"
    sleep 2
done

echo -e "${RED}❌ Gateway Service failed to start within 60 seconds${NC}"
echo -e "${YELLOW}📋 Checking container logs...${NC}"
docker-compose logs gateway-service
exit 1 