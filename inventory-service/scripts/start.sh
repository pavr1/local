#!/bin/bash

# Ice Cream Store Inventory Service - Start Script
set -e

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

echo -e "${CYAN}🍦📦 Starting Ice Cream Store Inventory Service...${RESET}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: Docker is not running. Please start Docker and try again.${RESET}"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: docker-compose is not installed or not in PATH.${RESET}"
    exit 1
fi

# Navigate to docker directory
cd "$(dirname "$0")/../docker"

# Build and start the service
echo -e "${YELLOW}🔧 Building inventory service container...${RESET}"
docker-compose build --no-cache

echo -e "${YELLOW}🚀 Starting inventory service...${RESET}"
docker-compose up -d

# Wait for service to be ready
echo -e "${YELLOW}⏳ Waiting for inventory service to be ready...${RESET}"
sleep 10

# Health check
MAX_ATTEMPTS=30
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if curl -f -s http://localhost:8082/api/v1/inventory/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Inventory service is ready and healthy!${RESET}"
        echo -e "${GREEN}🔗 Service URL: http://localhost:8082${RESET}"
        echo -e "${GREEN}🏥 Health check: http://localhost:8082/api/v1/inventory/health${RESET}"
        echo -e "${GREEN}📋 API Documentation: Check /api/v1 endpoints${RESET}"
        echo ""
        echo -e "${CYAN}Available endpoints:${RESET}"
        echo -e "  • Suppliers: /api/v1/suppliers"
        echo -e "  • Ingredients: /api/v1/ingredients"
        echo -e "  • Existences: /api/v1/existences"
        echo -e "  • Runout Reports: /api/v1/runout-reports"
        echo -e "  • Recipe Categories: /api/v1/recipe-categories"
        echo -e "  • Recipes: /api/v1/recipes"
        echo -e "  • Recipe Ingredients: /api/v1/recipe-ingredients"
        echo ""
        echo -e "${YELLOW}💡 Use 'make logs' to follow service logs${RESET}"
        echo -e "${YELLOW}💡 Use 'make stop' to stop the service${RESET}"
        exit 0
    fi
    
    echo -e "${YELLOW}⏳ Attempt $ATTEMPT/$MAX_ATTEMPTS - Waiting for inventory service...${RESET}"
    sleep 2
    ATTEMPT=$((ATTEMPT + 1))
done

echo -e "${RED}❌ Inventory service failed to start within expected time${RESET}"
echo -e "${RED}📋 Check logs with: docker-compose logs inventory-service${RESET}"
exit 1 