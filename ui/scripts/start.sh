#!/bin/bash

# Ice Cream Store UI Service - Start Script
# This script starts the UI service using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🎨 Starting Ice Cream Store UI Service...${NC}"

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

# Check if gateway service is ready (UI depends on gateway for API calls)
echo -e "${YELLOW}⏳ Checking if gateway service is ready...${NC}"
MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8082/api/v1/gateway/p/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Gateway service is ready!${NC}"
        break
    fi
    echo -e "${YELLOW}   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Gateway service not ready yet... (trying: http://localhost:8082/api/v1/gateway/p/health)${NC}"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Gateway service failed to start within the expected time${NC}"
    echo -e "${YELLOW}⚠️  UI service requires gateway service to be running for API calls${NC}"
    exit 1
fi

# Stop existing UI service container if it exists
echo -e "${YELLOW}🧹 Cleaning up existing UI service container...${NC}"
docker-compose down 2>/dev/null || true

# Start the service
echo -e "${BLUE}🚀 Starting UI Service container...${NC}"
docker-compose up -d

# Wait for service to be ready
echo -e "${YELLOW}⏳ Waiting for UI Service to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:3000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ UI Service is ready!${NC}"
        echo ""
        echo -e "${CYAN}📊 UI Service Status:${NC}"
        echo -e "   🎨 UI Application: ${GREEN}http://localhost:3000${NC}"
        echo -e "   🔌 Health Check: ${GREEN}http://localhost:3000/health${NC}"
        echo -e "   🔐 Login Page: ${GREEN}http://localhost:3000/login.html${NC}"
        echo -e "   📊 Dashboard: ${GREEN}http://localhost:3000/dashboard.html${NC}"
        echo -e "   🛒 Orders Page: ${GREEN}http://localhost:3000/orders.html${NC}"
        echo -e "   🌐 Gateway API: ${GREEN}http://localhost:8082${NC}"
        echo ""
        echo -e "${GREEN}🎉 UI Service started successfully!${NC}"
        exit 0
    fi
    echo -e "${YELLOW}   Attempt $i/30: Still waiting...${NC}"
    sleep 2
done

echo -e "${RED}❌ UI Service failed to start within 60 seconds${NC}"
echo -e "${YELLOW}📋 Checking container logs...${NC}"
docker-compose logs ui
exit 1
