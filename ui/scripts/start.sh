#!/bin/bash

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="UI Service"
CONTAINER_NAME="icecream_ui"
PORT="3000"

echo -e "${BLUE}🚀 Starting ${SERVICE_NAME}...${NC}"
echo -e "${BLUE}🌐 Starting Ice Cream Store ${SERVICE_NAME}...${NC}"

# Stop existing ui service container if it exists
echo -e "${YELLOW}🧹 Cleaning up existing ${SERVICE_NAME} container...${NC}"
cd docker && docker-compose down 2>/dev/null || true

# Build the ui service (CRITICAL: Always build before starting)
echo -e "${BLUE}🔨 Building ${SERVICE_NAME}...${NC}"
docker-compose build --no-cache

# Start the service
echo -e "${BLUE}🚀 Starting ${SERVICE_NAME} container...${NC}"
docker-compose up -d

# Wait for service to be ready
echo -e "${YELLOW}⏳ Waiting for ${SERVICE_NAME} to be ready...${NC}"
for i in {1..30}; do
    if curl -f http://localhost:${PORT}/ > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ ${SERVICE_NAME} failed to start within 30 seconds${NC}"
        exit 1
    fi
    echo -e "${YELLOW}   Attempt $i/30 - ${SERVICE_NAME} not ready yet...${NC}"
    sleep 1
done

echo -e "${GREEN}✅ ${SERVICE_NAME} is ready!${NC}"

# Show container status
echo ""
echo -e "${BLUE}📊 Container Status:${NC}"
docker ps --filter "name=${CONTAINER_NAME}" \
  --format "table {{.Names}}\t{{.Image}}\t{{.Command}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo -e "${GREEN}🎉 ${SERVICE_NAME} is ready!${NC}"
echo ""
echo -e "${BLUE}📝 Service Details:${NC}"
echo -e "   UI Application: http://localhost:${PORT}"
echo -e "   Health Check:   http://localhost:${PORT}/"
echo ""
echo -e "${BLUE}📋 Useful Commands:${NC}"
echo -e "   Stop service:     ./scripts/stop.sh"
echo -e "   View logs:        ./scripts/logs.sh"
echo -e "   Open application: open http://localhost:${PORT}"
echo "" 