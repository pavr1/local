#!/bin/bash

# Ice Cream Store Orders Service Stop Script
# This script stops the orders service container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦📦 Stopping Ice Cream Store Orders Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Stop the orders service
docker-compose down

echo "✅ Orders service stopped successfully!" 