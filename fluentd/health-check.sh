#!/bin/sh

# Health Check Service for Fluentd Logging Stack
# This script provides HTTP endpoints to check the health of Elasticsearch, Kibana, and Fluentd

# Function to check Elasticsearch health
check_elasticsearch() {
    if curl -s -f "$ELASTICSEARCH_URL/_cluster/health" > /dev/null 2>&1; then
        echo "healthy"
    else
        echo "unhealthy"
    fi
}

# Function to check Kibana health
check_kibana() {
    if curl -s -f "$KIBANA_URL/api/status" > /dev/null 2>&1; then
        echo "healthy"
    else
        echo "unhealthy"
    fi
}

# Function to check Fluentd health
check_fluentd() {
    if nc -z "$FLUENTD_HOST" "$FLUENTD_PORT" 2>/dev/null; then
        echo "healthy"
    else
        echo "unhealthy"
    fi
}

# Function to generate JSON response
generate_response() {
    local elasticsearch_status=$1
    local kibana_status=$2
    local fluentd_status=$3
    
    # Determine overall status
    local overall_status="healthy"
    if [ "$elasticsearch_status" != "healthy" ] || [ "$kibana_status" != "healthy" ] || [ "$fluentd_status" != "healthy" ]; then
        overall_status="unhealthy"
    fi
    
    cat << EOF
{
  "status": "$overall_status",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "services": {
    "elasticsearch": {
      "status": "$elasticsearch_status",
      "url": "$ELASTICSEARCH_URL"
    },
    "kibana": {
      "status": "$kibana_status",
      "url": "$KIBANA_URL"
    },
    "fluentd": {
      "status": "$fluentd_status",
      "host": "$FLUENTD_HOST",
      "port": "$FLUENTD_PORT"
    }
  },
  "service_name": "fluentd-logging-stack",
  "version": "1.0.0"
}
EOF
}

# Function to handle HTTP requests
handle_request() {
    local method=$1
    local path=$2
    
    case "$path" in
        "/health")
            if [ "$method" = "GET" ]; then
                local es_status=$(check_elasticsearch)
                local kibana_status=$(check_kibana)
                local fluentd_status=$(check_fluentd)
                
                echo "HTTP/1.1 200 OK"
                echo "Content-Type: application/json"
                echo "Access-Control-Allow-Origin: *"
                echo "Access-Control-Allow-Methods: GET, OPTIONS"
                echo "Access-Control-Allow-Headers: Content-Type"
                echo ""
                generate_response "$es_status" "$kibana_status" "$fluentd_status"
            else
                echo "HTTP/1.1 405 Method Not Allowed"
                echo "Content-Type: text/plain"
                echo ""
                echo "Method not allowed"
            fi
            ;;
        "/p/health")
            if [ "$method" = "GET" ]; then
                local es_status=$(check_elasticsearch)
                local kibana_status=$(check_kibana)
                local fluentd_status=$(check_fluentd)
                
                echo "HTTP/1.1 200 OK"
                echo "Content-Type: application/json"
                echo "Access-Control-Allow-Origin: *"
                echo "Access-Control-Allow-Methods: GET, OPTIONS"
                echo "Access-Control-Allow-Headers: Content-Type"
                echo ""
                generate_response "$es_status" "$kibana_status" "$fluentd_status"
            else
                echo "HTTP/1.1 405 Method Not Allowed"
                echo "Content-Type: text/plain"
                echo ""
                echo "Method not allowed"
            fi
            ;;
        "/")
            if [ "$method" = "GET" ]; then
                echo "HTTP/1.1 200 OK"
                echo "Content-Type: text/plain"
                echo ""
                echo "Fluentd Logging Stack Health Check Service"
                echo "Available endpoints:"
                echo "  GET /health     - Health check (requires authentication)"
                echo "  GET /p/health   - Public health check"
            else
                echo "HTTP/1.1 405 Method Not Allowed"
                echo "Content-Type: text/plain"
                echo ""
                echo "Method not allowed"
            fi
            ;;
        *)
            echo "HTTP/1.1 404 Not Found"
            echo "Content-Type: text/plain"
            echo ""
            echo "Not found"
            ;;
    esac
}

# Main server loop
echo "Starting Fluentd Health Check Service on port 8087..." >&2

while true; do
    # Read HTTP request
    read -r request_line || break
    
    # Parse method and path
    method=$(echo "$request_line" | cut -d' ' -f1)
    path=$(echo "$request_line" | cut -d' ' -f2)
    
    # Skip headers
    while read -r line; do
        [ -z "$line" ] || [ "$line" = $'\r' ] && break
    done
    
    # Handle the request
    handle_request "$method" "$path"
    
    # Small delay to prevent busy waiting
    sleep 0.1
done 