FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements (from service directory)
COPY requirements.txt .

# Install critical dependencies first with exact versions to prevent upgrades
# These must be installed before sentence-transformers to lock versions
RUN pip install --no-cache-dir \
    numpy==1.26.4 \
    torch==2.1.0 \
    transformers==4.35.2

# Downgrade huggingface-hub to a version compatible with sentence-transformers==2.2.2
# sentence-transformers 2.2.2 requires huggingface-hub < 0.20.0 (needs cached_download)
RUN pip install --no-cache-dir --force-reinstall "huggingface-hub<0.20.0"

# Install application dependencies
RUN pip install --no-cache-dir \
    fastapi==0.109.0 \
    uvicorn==0.27.0 \
    pydantic==2.5.3

# Install sentence-transformers dependencies that don't conflict
RUN pip install --no-cache-dir \
    tqdm \
    torchvision==0.16.0 \
    scikit-learn \
    scipy \
    nltk \
    sentencepiece

# Install sentence-transformers without dependencies (using already-installed versions)
RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2

# Download model at build time (caches in image)
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')"

# Copy application (from service directory)
COPY app.py .

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Run application
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]

