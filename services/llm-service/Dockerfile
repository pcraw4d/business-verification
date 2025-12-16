FROM nvidia/cuda:12.1.0-base-ubuntu22.04

# Install Python 3.11
RUN apt-get update && apt-get install -y \
    python3.11 \
    python3.11-dev \
    python3-pip \
    wget \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy requirements
COPY services/llm-service/requirements.txt .

# Install PyTorch with CUDA 12.1 support first
RUN pip3 install --no-cache-dir torch>=2.2.0,<2.3.0 --index-url https://download.pytorch.org/whl/cu121

# Install remaining Python dependencies
RUN pip3 install --no-cache-dir -r requirements.txt

# Download model at build time (caches in image, ~14GB)
# Note: This will take a while but ensures model is available on startup
# Skip model download at build time - download on first startup instead
# This avoids build timeout issues and allows the service to start faster
# RUN python3 -c "from transformers import AutoModelForCausalLM, AutoTokenizer; \
#     AutoTokenizer.from_pretrained('Qwen/Qwen2.5-7B-Instruct'); \
#     AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct', device_map='auto')"

# Copy application
COPY services/llm-service/app.py .

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=180s --retries=3 \
    CMD python3 -c "import requests; requests.get('http://localhost:8000/health', timeout=10)"

# Run application
CMD ["python3", "-m", "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]

