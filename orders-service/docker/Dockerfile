# Single-stage Alpine build for faster builds
FROM alpine:latest

# Install ca-certificates and curl for health checks
RUN apk add --no-cache ca-certificates curl

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary (will be built externally)
COPY main .

# Make binary executable
RUN chmod +x main

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8083

# Health check with faster intervals
HEALTHCHECK --interval=1s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8083/api/v1/orders/p/health || exit 1

# Set timezone
ENV TZ=UTC

# Run the application
CMD ["./main"] 