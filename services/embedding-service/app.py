"""
Embedding Service - Generate semantic embeddings for text
Deployed on Railway as a microservice
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from typing import List, Optional
from contextlib import asynccontextmanager
import logging
import time
import os

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global model variable
model = None
MODEL_NAME = 'sentence-transformers/all-MiniLM-L6-v2'

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Load model on startup, cleanup on shutdown"""
    global model
    # Startup: Load model
    logger.info(f"Loading embedding model: {MODEL_NAME}...")
    model = SentenceTransformer(MODEL_NAME)
    logger.info(f"Model loaded successfully! Dimension: {model.get_sentence_embedding_dimension()}")
    yield
    # Shutdown: Cleanup (if needed)
    logger.info("Shutting down embedding service")

# Initialize FastAPI app with lifespan events
app = FastAPI(
    title="Embedding Service",
    description="Generate 384-dimensional embeddings using all-MiniLM-L6-v2",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Request/Response models
class EmbedRequest(BaseModel):
    text: str
    truncate_length: Optional[int] = 5000

class EmbedResponse(BaseModel):
    embedding: List[float]
    dimension: int
    processing_time_ms: int

class EmbedBatchRequest(BaseModel):
    texts: List[str]
    truncate_length: Optional[int] = 5000

class EmbedBatchResponse(BaseModel):
    embeddings: List[List[float]]
    count: int
    processing_time_ms: int

# Health check endpoint
@app.get("/health")
async def health_check():
    """Health check endpoint for Railway"""
    # Always return 200 OK - Railway will check status field
    # This prevents Railway from restarting the service while model is loading
    # Railway's start-period (90s) gives enough time for model loading
    status = "healthy" if model is not None else "model_loading"
    
    try:
        if model is not None:
            dimension = model.get_sentence_embedding_dimension()
            return {
                "status": status,
                "model": MODEL_NAME,
                "dimension": dimension,
                "service": "embedding-service",
                "version": "1.0.0"
            }
        else:
            # Model still loading - return 200 with loading status
            # Railway's start-period will prevent premature healthcheck failures
            return {
                "status": status,
                "model": MODEL_NAME,
                "service": "embedding-service",
                "version": "1.0.0",
                "message": "Model is loading, please wait"
            }
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        # Return 200 with error status instead of 503
        # This prevents Railway from restarting during transient errors
        return {
            "status": "error",
            "model": MODEL_NAME,
            "service": "embedding-service",
            "version": "1.0.0",
            "error": str(e)
        }

# Single text embedding endpoint
@app.post("/embed", response_model=EmbedResponse)
async def create_embedding(request: EmbedRequest):
    """
    Generate embedding for a single text.
    
    Example:
        POST /embed
        {
            "text": "Restaurant serving Italian cuisine",
            "truncate_length": 5000
        }
    """
    start_time = time.time()
    
    try:
        # Truncate text if too long
        text = request.text
        if len(text) > request.truncate_length:
            text = text[:request.truncate_length]
            logger.warning(f"Text truncated from {len(request.text)} to {request.truncate_length} chars")
        
        # Validate text
        if not text or len(text.strip()) == 0:
            raise HTTPException(status_code=400, detail="Text cannot be empty")
        
        # Check if model is loaded
        if model is None:
            raise HTTPException(status_code=503, detail="Model not loaded yet")
        
        # Generate embedding
        embedding = model.encode(text, show_progress_bar=False)
        
        processing_time = int((time.time() - start_time) * 1000)
        
        logger.info(f"Generated embedding for text (length: {len(text)}, time: {processing_time}ms)")
        
        return EmbedResponse(
            embedding=embedding.tolist(),
            dimension=len(embedding),
            processing_time_ms=processing_time
        )
        
    except Exception as e:
        logger.error(f"Error generating embedding: {e}")
        raise HTTPException(status_code=500, detail=f"Error generating embedding: {str(e)}")

# Batch embedding endpoint
@app.post("/embed/batch", response_model=EmbedBatchResponse)
async def create_embeddings_batch(request: EmbedBatchRequest):
    """
    Generate embeddings for multiple texts in batch.
    More efficient than calling /embed multiple times.
    
    Example:
        POST /embed/batch
        {
            "texts": [
                "Restaurant serving Italian cuisine",
                "Software development company",
                "Dental office and clinic"
            ]
        }
    """
    start_time = time.time()
    
    try:
        # Validate
        if not request.texts or len(request.texts) == 0:
            raise HTTPException(status_code=400, detail="Texts list cannot be empty")
        
        if len(request.texts) > 100:
            raise HTTPException(status_code=400, detail="Maximum 100 texts per batch")
        
        # Truncate texts if needed
        texts = []
        for text in request.texts:
            if len(text) > request.truncate_length:
                texts.append(text[:request.truncate_length])
            else:
                texts.append(text)
        
        # Check if model is loaded
        if model is None:
            raise HTTPException(status_code=503, detail="Model not loaded yet")
        
        # Generate embeddings in batch (more efficient)
        embeddings = model.encode(texts, show_progress_bar=False)
        
        processing_time = int((time.time() - start_time) * 1000)
        
        logger.info(f"Generated {len(embeddings)} embeddings in batch (time: {processing_time}ms)")
        
        return EmbedBatchResponse(
            embeddings=[emb.tolist() for emb in embeddings],
            count=len(embeddings),
            processing_time_ms=processing_time
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error generating batch embeddings: {e}")
        raise HTTPException(status_code=500, detail=f"Error generating embeddings: {str(e)}")

# Info endpoint
@app.get("/info")
async def get_info():
    """Get information about the embedding service"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded yet")
    
    return {
        "model": MODEL_NAME,
        "dimension": model.get_sentence_embedding_dimension(),
        "max_sequence_length": model.max_seq_length,
        "description": "Generates semantic embeddings for text using sentence-transformers",
        "endpoints": {
            "/embed": "Generate single embedding",
            "/embed/batch": "Generate multiple embeddings",
            "/health": "Health check",
            "/info": "Service information"
        }
    }

if __name__ == "__main__":
    import uvicorn
    # Use Railway's PORT environment variable, fallback to 8000
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)

