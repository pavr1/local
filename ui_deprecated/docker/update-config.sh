#!/bin/sh

# Update config.js with environment variables
CONFIG_FILE="/usr/share/nginx/html/config.js"

# Default values if environment variables are not set
GATEWAY_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8082"}

echo "Updating UI configuration..."
echo "Gateway URL: $GATEWAY_URL"

# Replace the GATEWAY_URL in config.js
sed -i "s|GATEWAY_URL: 'http://localhost:8082'|GATEWAY_URL: '$GATEWAY_URL'|g" "$CONFIG_FILE"

echo "Configuration updated successfully!" 