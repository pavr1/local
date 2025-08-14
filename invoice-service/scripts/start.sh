#!/bin/bash

# Ice Cream Store Invoice Service Startup Script
# This script starts the invoice service and ensures database connectivity

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦📄 Starting Ice Cream Store Invoice Service..."

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
    echo "   Run: cd ../data-service && make start-docker"
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

# Stop existing invoice service container if it exists
echo "🧹 Cleaning up existing invoice service container..."
docker-compose down 2>/dev/null || true

# Start the invoice service
echo "🚀 Starting invoice service..."
docker-compose up -d

# Wait for invoice service to be ready
echo "⏳ Waiting for invoice-service API to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8085/api/v1/invoices/p/health > /dev/null 2>&1; then
        echo "✅ Invoice-service API is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Invoice-service API not ready yet... (trying: http://localhost:8085/api/v1/invoices/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Invoice-service API failed to start within the expected time"
    echo "   You can check the logs with: docker-compose logs invoice-service"
    exit 1
fi

# Show container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "🎉 Invoice Service is ready!"
echo ""
echo "📝 Service Details:"
echo "   Invoice API: http://localhost:8085"
echo "   Health:     http://localhost:8085/api/v1/invoices/p/health"
echo "   Create:     POST http://localhost:8085/api/v1/invoices"
echo ""
echo "📋 Useful Commands:"
echo "   Stop service:     ./scripts/stop.sh"
echo "   View logs:        ./scripts/logs.sh"
echo "   Test service:     curl http://localhost:8085/api/v1/invoices/p/health"
echo ""
