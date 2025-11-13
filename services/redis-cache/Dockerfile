# Redis Cache Service for KYB Platform
FROM redis:7-alpine

# Copy custom configuration
# When build context is repository root, use full path
COPY services/redis-cache/redis.conf /usr/local/etc/redis/redis.conf

# Create data directory
RUN mkdir -p /data

# Expose Redis port
EXPOSE 6379

# Start Redis with custom configuration
CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]
