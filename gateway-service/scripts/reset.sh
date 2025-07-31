#!/bin/bash

# Ice Cream Store Gateway Service - Reset Script
# This script stops, removes containers, and restarts the gateway service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🔄 Resetting Ice Cream Store Gateway Service...${NC}"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Stop and remove containers
echo -e "${YELLOW}🛑 Stopping and removing Gateway Service containers...${NC}"
docker-compose down

# Remove any existing containers
echo -e "${YELLOW}🗑️  Removing any existing Gateway Service containers...${NC}"
docker container rm icecream_gateway 2>/dev/null || true

# Remove Gateway Service images to force rebuild
echo -e "${YELLOW}🗑️  Removing Gateway Service images...${NC}"
docker rmi $(docker images | grep "gateway-service" | awk '{print $3}') 2>/dev/null || true

# Rebuild and start
echo -e "${BLUE}🔨 Building and starting Gateway Service...${NC}"
docker-compose up -d --build

# Wait for service to be ready
echo -e "${YELLOW}⏳ Waiting for Gateway Service to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:8082/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Gateway Service is ready!${NC}"
        echo ""
        echo -e "${CYAN}📊 Gateway Service Status:${NC}"
        echo -e "   🌐 Gateway API: ${GREEN}http://localhost:8082${NC}"
        echo -e "   🔌 Health Check: ${GREEN}http://localhost:8082/api/health${NC}"
        echo -e "   🔐 Session Proxy: ${GREEN}http://localhost:8082/api/v1/sessions/*${NC}"
        echo -e "   🛒 Orders Proxy: ${GREEN}http://localhost:8082/api/v1/orders/*${NC}"
        echo ""
        echo -e "${GREEN}🎉 Gateway Service reset completed successfully!${NC}"
        exit 0
    fi
    echo -e "${YELLOW}   Attempt $i/30: Still waiting...${NC}"
    sleep 2
done

echo -e "${RED}❌ Gateway Service failed to start within 60 seconds${NC}"
echo -e "${YELLOW}📋 Checking container logs...${NC}"
docker-compose logs gateway-service
exit 1 