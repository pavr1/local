#!/bin/bash

# Ice Cream Store Orders Service Startup Script
# This script starts the orders service and ensures database connectivity

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦📦 Starting Ice Cream Store Orders Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if the database network exists (from data-service)
if ! docker network ls | grep -q "docker_icecream_network"; then
    echo "⚠️  docker_icecream_network not found. Please ensure data-service is running first."
    echo "   Run: cd ../data-service && make start"
    exit 1
fi

# Check if data-service is ready
echo "⏳ Waiting for data-service to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8086/api/v1/data/p/health > /dev/null 2>&1; then
        echo "✅ Data-service is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Data-service not ready yet... (trying: http://localhost:8086/api/v1/data/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Data-service failed to start within the expected time"
    echo "   Please ensure data-service is running first: cd ../data-service && make start-docker"
    exit 1
fi

# Stop existing orders service container if it exists
echo "🧹 Cleaning up existing orders service container..."
docker-compose down 2>/dev/null || true

# Start the orders service
echo "🚀 Starting orders service..."
docker-compose up -d

# Wait for orders service to be ready
echo "⏳ Waiting for orders-service API to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8083/api/v1/orders/p/health > /dev/null 2>&1; then
        echo "✅ Orders-service API is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Orders-service API not ready yet... (trying: http://localhost:8083/api/v1/orders/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Orders service failed to start within the expected time"
    echo "   You can check the logs with: docker-compose logs orders-service"
    exit 1
fi

# Show container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "🎉 Orders Service is ready!"
echo ""
echo "📝 Service Details:"
echo "   Orders API: http://localhost:8083"
echo "   Health:     http://localhost:8083/api/v1/orders/p/health"
echo "   Create:     POST http://localhost:8083/api/v1/orders"
echo ""
echo "📋 Useful Commands:"
echo "   Stop service:     ./scripts/stop.sh"
echo "   View logs:        ./scripts/logs.sh"
echo "   Test service:     curl http://localhost:8083/api/v1/orders/p/health"
echo "" 