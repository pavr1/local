#!/bin/bash

# Ice Cream Store UI Service - Logs Script
# This script shows the UI service logs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}📋 Viewing Ice Cream Store UI Service logs...${NC}"

# Change to docker directory
cd "$(dirname "$0")/../docker"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running.${NC}"
    exit 1
fi

# Show logs
if [ "$1" = "-f" ]; then
    echo -e "${YELLOW}📋 Following UI Service logs (Ctrl+C to stop)...${NC}"
    docker-compose logs -f ui
else
    docker-compose logs ui
fi
