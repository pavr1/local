#!/bin/bash

# Ice Cream Store Auth Service Stop Script
# This script stops the authentication service containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🔐 Stopping Ice Cream Store Auth Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Stop containers
echo "🛑 Stopping auth service containers..."
docker-compose down

echo "✅ Auth service stopped successfully!" 