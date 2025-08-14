#!/bin/bash

# Ice Cream Store Data Service Stop Script
# This script stops the data service containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🗄️  Stopping Ice Cream Store Data Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Stop the data service
docker-compose down

echo "✅ Data service stopped successfully!" 