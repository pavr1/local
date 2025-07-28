#!/bin/bash

# Ice Cream Store Gateway Service - Stop Script
# This script stops the gateway service containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🛑 Stopping Ice Cream Store Gateway Service...${NC}"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running.${NC}"
    exit 1
fi

# Stop the containers
echo -e "${YELLOW}🛑 Stopping Gateway Service containers...${NC}"
docker-compose down

# Check if containers are stopped
if docker ps | grep -q "icecream_gateway"; then
    echo -e "${RED}❌ Failed to stop Gateway Service containers${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Gateway Service stopped successfully!${NC}"
fi 