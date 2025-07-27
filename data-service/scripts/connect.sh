#!/bin/bash

# Ice Cream Store Database Connect Script
# This script connects to the PostgreSQL database via command line

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

echo "🍦 Connecting to Ice Cream Store Data Service..."

# Change to docker directory
cd "$DOCKER_DIR"

# Check if container is running
if ! docker ps | grep -q icecream_postgres; then
    echo "❌ PostgreSQL container is not running."
    echo "   Start it with: ./scripts/start.sh"
    exit 1
fi

echo "🔌 Opening PostgreSQL CLI..."
echo "   Database: icecream_store"
echo "   User: postgres"
echo ""
echo "💡 Useful Commands:"
echo "   \\l          - List databases"
echo "   \\dt         - List tables"
echo "   \\d [table]  - Describe table structure"
echo "   \\q          - Quit"
echo ""

# Connect to the database
docker exec -it icecream_postgres psql -U postgres -d icecream_store 