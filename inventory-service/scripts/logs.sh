#!/bin/bash

# Ice Cream Store Inventory Service - Logs Script
set -e

# Colors for output
CYAN='\033[0;36m'
RESET='\033[0m'

echo -e "${CYAN}🍦📦 Ice Cream Store Inventory Service Logs${RESET}"
echo -e "${CYAN}===========================================${RESET}"

# Navigate to docker directory
cd "$(dirname "$0")/../docker"

# Show logs
docker-compose logs -f inventory-service 