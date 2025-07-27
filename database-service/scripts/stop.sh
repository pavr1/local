#!/bin/bash

# Ice Cream Store Database Stop Script
# This script stops the PostgreSQL database containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦 Stopping Ice Cream Store Database..."

# Change to docker directory
cd "$DOCKER_DIR"

# Stop the containers
echo "🛑 Stopping database containers..."
docker-compose down

echo "✅ Database containers stopped successfully!"
echo ""
echo "📋 To start again: ./scripts/start.sh" 