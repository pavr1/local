#!/bin/bash

# Ice Cream Store Database Logs Script
# This script shows logs from the database containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

# Change to docker directory
cd "$DOCKER_DIR"

echo "🍦 Ice Cream Store Database Logs"
echo "================================="

# Check if specific service is requested
if [ "$1" != "" ]; then
    case $1 in
        "postgres"|"pg"|"db")
            echo "📊 PostgreSQL Logs:"
            docker-compose logs -f postgres
            ;;
        "pgadmin"|"admin")
            echo "🔧 PgAdmin Logs:"
            docker-compose logs -f pgadmin
            ;;
        *)
            echo "❌ Unknown service: $1"
            echo "   Available services: postgres, pgadmin"
            exit 1
            ;;
    esac
else
    echo "📊 All Container Logs:"
    echo "Press Ctrl+C to exit"
    echo ""
    docker-compose logs -f
fi 