# Use a simple Alpine image
FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache ca-certificates curl

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary (we'll build it locally)
COPY main .

# Make binary executable
RUN chmod +x main

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Add health check
HEALTHCHECK --interval=1s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8081/api/v1/sessions/p/health || exit 1

# Set timezone
ENV TZ=UTC

# Run the application
CMD ["./main"]
