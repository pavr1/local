#!/bin/bash

# Ice Cream Store Auth Service Startup Script
# This script starts the authentication service and ensures database connectivity

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦🔐 Starting Ice Cream Store Auth Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if the database network exists (from data-service)
if ! docker network ls | grep -q "icecream_network"; then
    echo "⚠️  icecream_network not found. Please ensure data-service is running first."
    echo "   Run: cd ../data-service && make start"
    exit 1
fi

# Check if database is accessible
echo "🔍 Checking database connectivity..."
if ! docker run --rm --network icecream_network postgres:15-alpine pg_isready -h postgres -p 5432 -U postgres > /dev/null 2>&1; then
    echo "❌ Cannot connect to database. Please ensure data-service is running."
    echo "   Run: cd ../data-service && make start"
    exit 1
fi

echo "✅ Database connectivity confirmed!"

# Stop existing auth service container if it exists
echo "🧹 Cleaning up existing auth service container..."
docker-compose down --remove-orphans 2>/dev/null || true

# Build and start the auth service
echo "🔨 Building auth service..."
docker-compose build --no-cache

echo "🚀 Starting auth service..."
docker-compose up -d

# Wait for auth service to be ready
echo "⏳ Waiting for auth service to be ready..."

MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8081/api/v1/auth/health > /dev/null 2>&1; then
        echo "✅ Auth service is ready!"
        break
    fi
    
    echo "   Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES - Auth service not ready yet..."
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Auth service failed to start within the expected time"
    echo "   You can check the logs with: docker-compose logs auth-service"
    exit 1
fi

# Show container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "🎉 Auth Service is ready!"
echo ""
echo "📝 Service Details:"
echo "   Auth API: http://localhost:8081"
echo "   Health:   http://localhost:8081/api/v1/auth/health"
echo "   Login:    POST http://localhost:8081/api/v1/auth/login"
echo ""
echo "📋 Useful Commands:"
echo "   Stop service:     ./scripts/stop.sh"
echo "   View logs:        ./scripts/logs.sh"
echo "   Test login:       curl -X POST http://localhost:8081/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
echo "" 