FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Pre-download hrequests native library during build
# This prevents runtime download failures (timeouts, corrupted files)
# The library is downloaded from GitHub releases during import, which can fail in containerized environments
RUN python -c "import hrequests; print('Library pre-loaded successfully')" || \
    (echo "Pre-downloading hrequests native library..." && \
     mkdir -p /usr/local/lib/python3.11/site-packages/hrequests/bin && \
     curl -L --max-time 60 --retry 3 -o /usr/local/lib/python3.11/site-packages/hrequests/bin/hrequests-cgo-3.1-linux-amd64.so \
     https://github.com/daijro/hrequests/releases/download/v3.1/hrequests-cgo-3.1-linux-amd64.so && \
     python -c "import hrequests; print('Library verified successfully')" || \
     (echo "Warning: Library download failed, but continuing build. Service may fail at runtime." && exit 0))

# Copy application code
COPY app.py .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health').read()"

# Run the application
CMD ["python", "app.py"]



