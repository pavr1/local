# Single-stage nginx build for static UI files
FROM nginx:alpine

# Install curl for health checks
RUN apk add --no-cache curl

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy nginx configuration
COPY docker/nginx.conf /etc/nginx/nginx.conf
RUN chmod 644 /etc/nginx/nginx.conf

# Copy static files to nginx serve directory
COPY *.html /usr/share/nginx/html/
COPY partials/ /usr/share/nginx/html/partials/
COPY shared/ /usr/share/nginx/html/shared/

# Copy JS and CSS files if they exist
RUN find . -maxdepth 1 -name "*.js" -exec cp {} /usr/share/nginx/html/ \; 2>/dev/null || true
RUN find . -maxdepth 1 -name "*.css" -exec cp {} /usr/share/nginx/html/ \; 2>/dev/null || true

# config.js will be mounted as volume in docker-compose.yml

# Create nginx directories and set permissions
RUN mkdir -p /var/cache/nginx /var/log/nginx /tmp && \
    chown -R appuser:appgroup /usr/share/nginx/html /var/cache/nginx /var/log/nginx /tmp

# Run as root for nginx PID file access
# USER appuser

# Expose port
EXPOSE 3000

# Health check with faster intervals
HEALTHCHECK --interval=1s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Set timezone
ENV TZ=UTC

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
