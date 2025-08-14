#!/bin/bash

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

echo -e "${YELLOW}🛑 Stopping Session Service containers...${RESET}"

# Stop containers
cd docker && docker-compose down

echo -e "${GREEN}✅ Session service stopped successfully!${RESET}"
