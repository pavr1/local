#!/bin/bash

# Ice Cream Store Invoice Service Logs Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_ROOT/docker"

# Change to docker directory
cd "$DOCKER_DIR"

# Show logs with optional follow flag
if [[ "$1" == "-f" ]]; then
    docker-compose logs -f invoice-service
else
    docker-compose logs invoice-service
fi
