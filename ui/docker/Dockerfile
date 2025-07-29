# UI Service Dockerfile
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy nginx configuration
COPY docker/nginx.conf /etc/nginx/conf.d/

# Copy config update script
COPY docker/update-config.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/update-config.sh

# Copy UI files to nginx html directory
COPY . /usr/share/nginx/html/

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set ownership of nginx directories
RUN chown -R appuser:appgroup /var/cache/nginx && \
    chown -R appuser:appgroup /var/log/nginx && \
    chown -R appuser:appgroup /etc/nginx/conf.d && \
    chown -R appuser:appgroup /usr/share/nginx/html && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appgroup /var/run/nginx.pid

# Create startup script
RUN echo '#!/bin/sh' > /usr/local/bin/startup.sh && \
    echo '/usr/local/bin/update-config.sh' >> /usr/local/bin/startup.sh && \
    echo 'nginx -g "daemon off;"' >> /usr/local/bin/startup.sh && \
    chmod +x /usr/local/bin/startup.sh

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3000/ || exit 1

# Start with config update then nginx
CMD ["/usr/local/bin/startup.sh"] 