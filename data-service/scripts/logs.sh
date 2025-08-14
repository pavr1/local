#!/bin/bash

# Ice Cream Store Data Service Logs Script
# This script shows logs from the data service containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🗄️  Data Service Logs..."

# Change to docker directory
cd "$DOCKER_DIR"

# Show logs
docker-compose logs "$@" 