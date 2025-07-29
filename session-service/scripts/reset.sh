#!/bin/bash

# Ice Cream Store Auth Service Reset Script
# This script stops, rebuilds, and starts the authentication service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🔐 Resetting Ice Cream Store Auth Service..."

# Change to docker directory
cd "$DOCKER_DIR"

echo "🛑 Stopping existing containers..."
docker-compose down --remove-orphans

echo "🧹 Removing old images..."
docker-compose down --rmi local 2>/dev/null || true

echo "🔨 Rebuilding auth service..."
docker-compose build --no-cache

echo "🚀 Starting auth service..."
docker-compose up -d

echo "⏳ Waiting for service to be ready..."
sleep 5

# Check if service is healthy
MAX_RETRIES=20
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then
        echo "✅ Auth service is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Auth service not ready yet..."
    sleep 3
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Auth service failed to start. Check logs with: docker-compose logs"
    exit 1
fi

echo ""
echo "🎉 Auth service reset completed successfully!"
echo "   Service URL: http://localhost:8081"
echo "   Health check: http://localhost:8081/api/v1/auth/health" 