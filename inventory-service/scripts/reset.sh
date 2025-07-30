#!/bin/bash

# Ice Cream Store Inventory Service - Reset Script
set -e

# Colors for output
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
GREEN='\033[0;32m'
RESET='\033[0m'

echo -e "${CYAN}🍦📦 Resetting Ice Cream Store Inventory Service...${RESET}"

# Navigate to docker directory
cd "$(dirname "$0")/../docker"

# Warning message
echo -e "${RED}⚠️  WARNING: This will stop the service and remove all containers${RESET}"
echo -e "${YELLOW}💡 This will NOT affect the database data${RESET}"
echo ""

# Stop and remove containers
echo -e "${YELLOW}🛑 Stopping and removing inventory service containers...${RESET}"
docker-compose down --remove-orphans

# Remove any dangling images
echo -e "${YELLOW}🧹 Cleaning up unused Docker resources...${RESET}"
docker system prune -f

echo -e "${GREEN}✅ Inventory service reset completed${RESET}"
echo -e "${YELLOW}💡 Use 'make start' to rebuild and start the service${RESET}" 