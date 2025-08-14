#!/bin/bash

# Ice Cream Store Data Service Startup Script
# This script starts the data service with PostgreSQL and PgAdmin

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🗄️  Starting Ice Cream Store Data Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Stop existing containers if they exist
echo "🧹 Cleaning up existing containers..."
docker-compose down 2>/dev/null || true

# Start the data service
echo "🚀 Starting data service..."
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker exec icecream_postgres pg_isready -U postgres -d icecream_store > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - PostgreSQL not ready yet..."
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ PostgreSQL failed to start within the expected time"
    echo "   You can check the logs with: docker-compose logs postgres"
    exit 1
fi

# Wait for data-service API to be ready
echo "⏳ Waiting for data-service API to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    	if curl -f http://localhost:8086/api/v1/data/p/health > /dev/null 2>&1; then
        echo "✅ Data-service API is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Data-service API not ready yet... (trying: http://localhost:8086/api/v1/data/p/health)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Data-service API failed to start within the expected time"
    echo "   You can check the logs with: docker-compose logs data-service"
    exit 1
fi

# Show container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "🎉 Data Service is ready!"
echo ""
echo "📝 Service Details:"
echo "   PostgreSQL:    localhost:5432"
echo "   PgAdmin:       http://localhost:8080 (admin@icecreamstore.com / admin123)"
echo "   Data API:      http://localhost:8086"
echo "   Health:        http://localhost:8086/api/v1/data/p/health"
echo ""
echo "📋 Useful Commands:"
echo "   Stop service:     ./scripts/stop.sh"
echo "   View logs:        ./scripts/logs.sh"
echo "   Test database:    docker exec icecream_postgres pg_isready -U postgres -d icecream_store"
echo "   Test API:         curl http://localhost:8086/api/v1/data/p/health"
echo "" 