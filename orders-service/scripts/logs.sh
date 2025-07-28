#!/bin/bash

# Ice Cream Store Orders Service Logs Script
# This script shows logs from the orders service container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

# Change to docker directory
cd "$DOCKER_DIR"

# Check if follow flag is provided
if [ "$1" = "-f" ] || [ "$1" = "--follow" ]; then
    echo "🍦📦 Following Orders Service logs (Ctrl+C to exit)..."
    docker-compose logs -f orders-service
else
    echo "🍦📦 Showing recent Orders Service logs..."
    docker-compose logs --tail=50 orders-service
fi 