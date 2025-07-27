#!/bin/bash

# Ice Cream Store Auth Service Logs Script
# This script shows logs from the authentication service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

# Change to docker directory
cd "$DOCKER_DIR"

# Check if we should follow logs
FOLLOW=""
if [[ "$1" == "-f" || "$1" == "--follow" ]]; then
    FOLLOW="-f"
    shift
fi

# Determine which service to show logs for
SERVICE="${1:-auth-service}"

case "$SERVICE" in
    "auth" | "auth-service")
        echo "📋 Viewing Auth Service logs..."
        if [[ -n "$FOLLOW" ]]; then
            echo "   (Following logs - press Ctrl+C to stop)"
        fi
        docker-compose logs $FOLLOW auth-service
        ;;
    "all")
        echo "📋 Viewing all service logs..."
        if [[ -n "$FOLLOW" ]]; then
            echo "   (Following logs - press Ctrl+C to stop)"
        fi
        docker-compose logs $FOLLOW
        ;;
    *)
        echo "📋 Available log options:"
        echo "   ./scripts/logs.sh [auth|all] [-f|--follow]"
        echo ""
        echo "Examples:"
        echo "   ./scripts/logs.sh              # Show auth service logs"
        echo "   ./scripts/logs.sh -f           # Follow auth service logs"
        echo "   ./scripts/logs.sh all          # Show all logs"
        echo "   ./scripts/logs.sh all -f       # Follow all logs"
        exit 1
        ;;
esac 