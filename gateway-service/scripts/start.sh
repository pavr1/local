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

# Start the service
echo -e "${BLUE}🚀 Starting Gateway Service container...${NC}"
docker-compose up -d

# Wait for service to be ready
echo -e "${YELLOW}⏳ Waiting for Gateway Service to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:8082/api/hello > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Gateway Service is ready!${NC}"
        echo ""
        echo -e "${CYAN}📊 Gateway Service Status:${NC}"
        echo -e "   🌐 Gateway API: ${GREEN}http://localhost:8082${NC}"
        echo -e "   🔌 Health Check: ${GREEN}http://localhost:8082/api/hello${NC}"
        echo -e "   🔐 Auth Proxy: ${GREEN}http://localhost:8082/api/v1/auth/*${NC}"
        echo -e "   🛒 Orders Proxy: ${GREEN}http://localhost:8082/api/v1/orders/*${NC}"
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