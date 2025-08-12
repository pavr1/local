# Centralized Logging with Fluentd + ELK Stack

This directory contains the centralized logging infrastructure for the Ice Cream Store system.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   All Services  │───▶│     Fluentd     │───▶│  Elasticsearch  │◀───│     Kibana      │
│   (Port 24224)  │    │   (Port 24224)  │    │   (Port 9200)   │    │   (Port 5601)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### 1. Start the Logging Stack
```bash
make logs-start
```

### 2. Access the Tools
- **Kibana (Web UI)**: http://localhost:5601
- **Elasticsearch (API)**: http://localhost:9200
- **Dashboard Logs**: Use the "View Logs" buttons in the UI

### 3. Stop the Logging Stack
```bash
make logs-stop
```

## How It Works

### 1. Services Send Logs to Fluentd
All services use the shared logger to send logs to Fluentd:

```go
// In any service
logger := logger.InitLogger("service-name", "localhost", 24224)
logger.Info("Service started", map[string]interface{}{
    "port": 8080,
    "version": "1.0.0",
})
```

### 2. Fluentd Processes and Forwards
Fluentd receives logs, processes them, and forwards to Elasticsearch.

### 3. Elasticsearch Stores
Elasticsearch indexes and stores all logs for fast searching.

### 4. Dashboard Queries
The dashboard queries Elasticsearch through the gateway API.

## Dashboard Features

- **Service Selection**: Choose which service logs to view
- **Level Filtering**: Filter by log level (DEBUG, INFO, WARN, ERROR, FATAL)
- **Text Search**: Search within log messages
- **Real-time Updates**: Auto-refresh every second
- **Fallback Support**: Falls back to local logs if centralized logging is down

## Integration Steps

### 1. Add Shared Logger to Your Service
```go
import "shared/logger"

func main() {
    // Initialize logger
    logger := logger.InitLogger("your-service-name", "localhost", 24224)
    defer logger.Close()
    
    // Use logger
    logger.Info("Service started", map[string]interface{}{
        "port": 8080,
    })
}
```

### 2. Environment Variables
Set these environment variables in your service:
```bash
FLUENTD_HOST=localhost
FLUENTD_PORT=24224
```

### 3. Test Integration
1. Start the logging stack: `make logs-start`
2. Start your service
3. Check logs in the dashboard or Kibana

## Troubleshooting

### Fluentd Not Receiving Logs
- Check if Fluentd is running: `make logs-status`
- Verify port 24224 is accessible
- Check service logs for connection errors

### Elasticsearch Issues
- Check Elasticsearch health: `curl http://localhost:9200/_cluster/health`
- Verify enough memory (Elasticsearch needs at least 512MB)

### Dashboard Not Showing Logs
- Check if centralized logging is working: `curl http://localhost:8082/api/v1/logs`
- Verify Elasticsearch has data: `curl http://localhost:9200/ice-cream-logs-*/_search`

## Advanced Configuration

### Custom Fluentd Configuration
Edit `fluentd/conf/fluent.conf` to customize:
- Log parsing rules
- Output formats
- Filtering rules

### Elasticsearch Configuration
Edit `docker-compose.logging.yml` to customize:
- Memory settings
- Index settings
- Security settings

## Monitoring

### Check Logging Stack Health
```bash
make logs-status
```

### View Fluentd Logs
```bash
cd fluentd && docker-compose -f docker-compose.logging.yml logs fluentd
```

### View Elasticsearch Logs
```bash
cd fluentd && docker-compose -f docker-compose.logging.yml logs elasticsearch
``` 