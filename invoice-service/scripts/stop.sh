#!/bin/bash

# Ice Cream Store Invoice Service Stop Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🛑 Stopping Invoice Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Stop and remove containers
docker-compose down

echo "✅ Invoice service stopped successfully!"
