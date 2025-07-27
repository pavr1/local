#!/bin/bash

# Ice Cream Store Database Reset Script
# This script completely resets the database by removing all data

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦 Resetting Ice Cream Store Database..."
echo "⚠️  WARNING: This will delete ALL database data!"
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Reset cancelled."
    exit 1
fi

# Change to docker directory
cd "$DOCKER_DIR"

# Stop and remove containers and volumes
echo "🧹 Stopping containers and removing volumes..."
docker-compose down -v --remove-orphans

# Remove any orphaned volumes
echo "🗑️  Removing orphaned volumes..."
docker volume prune -f

echo "✅ Database reset completed!"
echo ""
echo "📋 To start fresh: ./scripts/start.sh" 