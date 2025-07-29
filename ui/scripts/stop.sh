#!/bin/bash

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="UI Service"
CONTAINER_NAME="icecream_ui"

echo -e "${BLUE}🛑 Stopping ${SERVICE_NAME}...${NC}"

# Stop the service
cd docker && docker-compose down

echo -e "${GREEN}✅ ${SERVICE_NAME} stopped successfully!${NC}" 