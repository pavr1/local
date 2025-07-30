#!/bin/bash

# Ice Cream Store Inventory Service - Stop Script
set -e

# Colors for output
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RESET='\033[0m'

echo -e "${CYAN}🍦📦 Stopping Ice Cream Store Inventory Service...${RESET}"

# Navigate to docker directory
cd "$(dirname "$0")/../docker"

# Stop the service
echo -e "${YELLOW}🛑 Stopping inventory service...${RESET}"
docker-compose down

echo -e "${CYAN}✅ Inventory service stopped successfully${RESET}" 