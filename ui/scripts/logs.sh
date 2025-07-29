#!/bin/bash

# Colors for output
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="UI Service"
CONTAINER_NAME="icecream_ui"

echo -e "${BLUE}📋 Viewing ${SERVICE_NAME} logs...${NC}"
echo -e "${BLUE}Press Ctrl+C to exit log view${NC}"
echo ""

# Follow logs
docker logs -f ${CONTAINER_NAME} 2>/dev/null || {
    echo "Container ${CONTAINER_NAME} not running. Showing last logs:"
    docker logs ${CONTAINER_NAME} 2>/dev/null || echo "No logs available"
} 