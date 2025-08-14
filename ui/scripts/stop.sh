#!/bin/bash

# Ice Cream Store UI Service - Stop Script
# This script stops the UI service Docker containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🎨 Stopping Ice Cream Store UI Service...${NC}"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running.${NC}"
    exit 1
fi

# Stop the containers
echo -e "${YELLOW}🛑 Stopping UI Service containers...${NC}"
docker-compose down

echo -e "${GREEN}✅ UI service stopped successfully!${NC}"
