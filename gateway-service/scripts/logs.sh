#!/bin/bash

# Ice Cream Store Gateway Service - Logs Script
# This script shows logs from the gateway service containers

set -e

# Colors for output
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}📋 Gateway Service Logs${NC}"
echo "================================"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Follow logs if -f flag is provided
if [ "$1" = "-f" ] || [ "$1" = "--follow" ]; then
    echo -e "${CYAN}📋 Following Gateway Service logs (Ctrl+C to stop)...${NC}"
    docker-compose logs -f gateway-service
else
    echo -e "${CYAN}📋 Showing recent Gateway Service logs...${NC}"
    docker-compose logs --tail=100 gateway-service
fi 