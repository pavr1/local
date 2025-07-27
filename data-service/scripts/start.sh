#!/bin/bash

# Ice Cream Store Database Startup Script
# This script starts the PostgreSQL database container and waits for it to be ready

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦 Starting Ice Cream Store Data Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Stop existing containers if they exist
echo "🧹 Cleaning up existing containers..."
docker-compose down --remove-orphans 2>/dev/null || true

# Start the containers
echo "🚀 Starting database containers..."
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

# Show container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "🎉 Database is ready!"
echo ""
echo "📝 Connection Details:"
echo "   Host: localhost"
echo "   Port: 5432"
echo "   Database: icecream_store"
echo "   Username: postgres"
echo "   Password: postgres123"
echo ""
echo "🔧 PgAdmin (Web Interface):"
echo "   URL: http://localhost:8080"
echo "   Email: admin@icecreamstore.com"
echo "   Password: admin123"
echo ""
echo "📋 Useful Commands:"
echo "   Stop database:     ./scripts/stop.sh"
echo "   Reset database:    ./scripts/reset.sh"
echo "   View logs:         ./scripts/logs.sh"
echo "   Connect via CLI:   ./scripts/connect.sh"
echo "" 