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
RUN pip3 install --no-cache-dir "torch>=2.2.0,<2.3.0" --index-url https://download.pytorch.org/whl/cu121

# Install remaining Python dependencies
RUN pip3 install --no-cache-dir -r requirements.txt

# Pre-download model at build time (0.5B model is small enough ~1GB)
# This makes first request much faster
RUN python3 -c "from transformers import AutoModelForCausalLM, AutoTokenizer; \
    print('Downloading Qwen 2.5 0.5B model...'); \
    AutoTokenizer.from_pretrained('Qwen/Qwen2.5-0.5B-Instruct'); \
    AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-0.5B-Instruct', low_cpu_mem_usage=True); \
    print('Model downloaded successfully!')"

# Copy application
COPY services/llm-service/app.py .

# Expose port
EXPOSE 8000

# Health check (reduced start period since 0.5B model loads fast)
HEALTHCHECK --interval=30s --timeout=30s --start-period=60s --retries=3 \
    CMD python3 -c "import requests; requests.get('http://localhost:8000/health', timeout=10)"

# Run application
CMD ["python3", "-m", "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]

