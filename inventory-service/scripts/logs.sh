#!/bin/bash

# Colors for output
CYAN='\033[0;36m'
RESET='\033[0m'

echo -e "${CYAN}📋 Viewing Inventory Service logs...${RESET}"

# Show logs from docker-compose
cd docker && docker-compose logs "$@"
