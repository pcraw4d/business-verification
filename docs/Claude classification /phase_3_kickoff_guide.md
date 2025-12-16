# Phase 3 Kick-Off Guide: Add Layer 2 (Embeddings)
## Weeks 5-6: From 80-85% to 85-90% Accuracy

**Goal:** Add embedding-based similarity search to handle edge cases, ambiguous classifications, and novel business types that Layer 1 struggles with.

---

## Phase 2 Success Validation

Before starting Phase 3, verify your Phase 2 results:

✅ **Checklist:**
- [ ] Returns top 3 codes per type (MCC/SIC/NAICS)
- [ ] Confidence scores calibrated (70-95% range)
- [ ] Fast path handles 60-70% of requests in <100ms
- [ ] Structured explanations generated
- [ ] "General Business" <10% of results
- [ ] Accuracy at 80-85% on test set

**If all checked:** You're ready for Phase 3! 🎉

**If issues remain:** Address them before proceeding. Layer 2 builds on Layer 1.

---

## Phase 3 Overview

### What We're Adding

**Current State (After Phase 2):**
- ✅ Layer 1 (Multi-strategy) working well for common cases
- ✅ Fast path handles obvious classifications
- ✅ 80-85% accuracy
- ❌ Struggles with edge cases and novel businesses
- ❌ No semantic understanding (relies on exact keyword matches)
- ❌ Misses industry classifications with different terminology

**Target State (After Phase 3):**
- ✅ Layer 2 (Embeddings) catches edge cases
- ✅ Semantic understanding via vector similarity
- ✅ Handles novel businesses and non-standard terminology
- ✅ 2-layer routing (Layer 1 → Layer 2 for low confidence)
- ✅ 85-90% accuracy

### Why Embeddings?

**Problem with Layer 1:**
```
Business: "Cloud-native DevOps consultancy specializing in Kubernetes orchestration"
Layer 1 keywords: No match for "consultancy", weak match for "cloud"
Result: Low confidence or wrong classification

With Layer 2 embeddings:
Embedding captures semantic meaning → Matches "Technology Consulting" with 0.87 similarity
Result: High confidence, correct classification
```

**Embeddings understand:**
- Synonyms: "attorney" = "lawyer"
- Related concepts: "DevOps" relates to "Software Development"
- Industry jargon: "Kubernetes" → Technology sector
- Context: Not just keywords, but overall meaning

### Implementation Timeline

**Week 5:**
- Day 1: Enable pgvector in Supabase
- Day 2: Pre-compute code embeddings (one-time script)
- Day 3: Create embedding service (Python on Railway)
- Day 4: Add vector search RPC function
- Day 5: Test embedding service and vector search

**Week 6:**
- Day 1-2: Build Go embedding classifier
- Day 3: Add Layer 2 routing logic
- Day 4: Integration testing
- Day 5: Full test suite validation

---

## Week 5: Infrastructure Setup

### Task 1: Enable pgvector in Supabase (Day 1)

**What is pgvector?**
PostgreSQL extension that enables vector similarity search. Allows storing 384-dimensional embeddings and querying for similar vectors.

**File:** `supabase-migrations/050_enable_pgvector.sql`

```sql
-- Migration: Enable pgvector and create embeddings infrastructure

-- Step 1: Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Step 2: Create code_embeddings table
CREATE TABLE code_embeddings (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL,
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    extended_description TEXT,
    industry_context TEXT,
    embedding vector(384), -- all-MiniLM-L6-v2 produces 384-dim vectors
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_code_embedding UNIQUE(code_type, code)
);

-- Step 3: Create indexes for fast similarity search
-- IVFFlat index for approximate nearest neighbor search
CREATE INDEX idx_code_embeddings_vector ON code_embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Additional indexes for filtering
CREATE INDEX idx_code_embeddings_type ON code_embeddings(code_type);
CREATE INDEX idx_code_embeddings_code ON code_embeddings(code);
CREATE INDEX idx_code_embeddings_updated ON code_embeddings(updated_at);

-- Step 4: Create function for similarity search
CREATE OR REPLACE FUNCTION match_code_embeddings(
    query_embedding vector(384),
    code_type_filter text,
    match_threshold float DEFAULT 0.7,
    match_count int DEFAULT 5
)
RETURNS TABLE (
    code text,
    code_type text,
    description text,
    extended_description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ce.code,
        ce.code_type,
        ce.description,
        ce.extended_description,
        1 - (ce.embedding <=> query_embedding) as similarity
    FROM code_embeddings ce
    WHERE ce.code_type = code_type_filter
        AND 1 - (ce.embedding <=> query_embedding) > match_threshold
    ORDER BY ce.embedding <=> query_embedding
    LIMIT match_count;
END;
$$;

-- Step 5: Create function to search across all code types
CREATE OR REPLACE FUNCTION match_code_embeddings_all_types(
    query_embedding vector(384),
    match_threshold float DEFAULT 0.7,
    match_count_per_type int DEFAULT 5
)
RETURNS TABLE (
    code text,
    code_type text,
    description text,
    extended_description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    (
        SELECT * FROM match_code_embeddings(query_embedding, 'MCC', match_threshold, match_count_per_type)
        UNION ALL
        SELECT * FROM match_code_embeddings(query_embedding, 'SIC', match_threshold, match_count_per_type)
        UNION ALL
        SELECT * FROM match_code_embeddings(query_embedding, 'NAICS', match_threshold, match_count_per_type)
    )
    ORDER BY similarity DESC;
END;
$$;

-- Step 6: Grant permissions
GRANT SELECT ON code_embeddings TO authenticated;
GRANT SELECT ON code_embeddings TO anon;
GRANT EXECUTE ON FUNCTION match_code_embeddings TO authenticated;
GRANT EXECUTE ON FUNCTION match_code_embeddings TO anon;
GRANT EXECUTE ON FUNCTION match_code_embeddings_all_types TO authenticated;
GRANT EXECUTE ON FUNCTION match_code_embeddings_all_types TO anon;

-- Step 7: Add helpful comments
COMMENT ON TABLE code_embeddings IS 'Pre-computed embeddings for all industry codes (MCC/SIC/NAICS) using all-MiniLM-L6-v2';
COMMENT ON COLUMN code_embeddings.embedding IS '384-dimensional embedding vector from sentence-transformers/all-MiniLM-L6-v2';
COMMENT ON FUNCTION match_code_embeddings IS 'Find similar codes using vector similarity search (cosine distance)';

-- Step 8: Create trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_code_embeddings_updated_at
    BEFORE UPDATE ON code_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Run Migration:**
```bash
# Option 1: Via psql
psql $SUPABASE_DB_URL -f supabase-migrations/050_enable_pgvector.sql

# Option 2: Via Supabase CLI
supabase db push

# Option 3: Via Supabase Dashboard
# Go to SQL Editor → Paste migration → Run
```

**Verify Installation:**
```sql
-- Check pgvector is enabled
SELECT * FROM pg_extension WHERE extname = 'vector';

-- Check table exists
\d code_embeddings

-- Check indexes exist
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'code_embeddings';

-- Test vector operations work
SELECT vector_dims('[1,2,3]'::vector); -- Should return 3
```

**Expected Output:**
```
pgvector extension: ✅ Enabled
code_embeddings table: ✅ Created
Indexes: ✅ 4 indexes created
Functions: ✅ 2 functions created
```

---

### Task 2: Pre-compute Code Embeddings (Day 2)

**Why pre-compute?**
Generating embeddings is slow (50-100ms each). Pre-computing for all codes (once) enables fast lookups (<5ms).

**File:** `scripts/precompute_embeddings.py`

```python
#!/usr/bin/env python3
"""
Pre-compute embeddings for all classification codes.
Run once after enabling pgvector.
"""

import os
import sys
from typing import List, Dict
from sentence_transformers import SentenceTransformer
from supabase import create_client, Client
from tqdm import tqdm
import time

# Configuration
MODEL_NAME = 'sentence-transformers/all-MiniLM-L6-v2'
BATCH_SIZE = 50
SUPABASE_URL = os.getenv('SUPABASE_URL')
SUPABASE_KEY = os.getenv('SUPABASE_SERVICE_KEY')  # Use service key for admin access

def main():
    print("=" * 60)
    print("CODE EMBEDDINGS PRE-COMPUTATION")
    print("=" * 60)
    
    # Validate environment
    if not SUPABASE_URL or not SUPABASE_KEY:
        print("❌ Error: SUPABASE_URL and SUPABASE_SERVICE_KEY must be set")
        sys.exit(1)
    
    # Initialize Supabase client
    print("\n1. Connecting to Supabase...")
    supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)
    print("✅ Connected")
    
    # Initialize embedding model
    print(f"\n2. Loading embedding model: {MODEL_NAME}...")
    model = SentenceTransformer(MODEL_NAME)
    print(f"✅ Model loaded (embedding dimension: {model.get_sentence_embedding_dimension()})")
    
    # Fetch all codes
    print("\n3. Fetching classification codes from database...")
    codes = fetch_all_codes(supabase)
    print(f"✅ Fetched {len(codes)} codes")
    print(f"   - MCC: {sum(1 for c in codes if c['code_type'] == 'MCC')}")
    print(f"   - SIC: {sum(1 for c in codes if c['code_type'] == 'SIC')}")
    print(f"   - NAICS: {sum(1 for c in codes if c['code_type'] == 'NAICS')}")
    
    # Enrich descriptions
    print("\n4. Enriching code descriptions with context...")
    enriched_codes = enrich_codes_with_context(codes, supabase)
    print("✅ Descriptions enriched")
    
    # Generate embeddings
    print(f"\n5. Generating embeddings (batch size: {BATCH_SIZE})...")
    embeddings_data = generate_embeddings_batch(enriched_codes, model, BATCH_SIZE)
    print(f"✅ Generated {len(embeddings_data)} embeddings")
    
    # Insert into database
    print("\n6. Inserting embeddings into database...")
    insert_embeddings(embeddings_data, supabase, BATCH_SIZE)
    print("✅ All embeddings inserted")
    
    # Verify
    print("\n7. Verifying insertion...")
    verify_embeddings(supabase)
    
    print("\n" + "=" * 60)
    print("✅ EMBEDDINGS PRE-COMPUTATION COMPLETE!")
    print("=" * 60)

def fetch_all_codes(supabase: Client) -> List[Dict]:
    """Fetch all classification codes from database."""
    response = supabase.table('classification_codes').select('*').execute()
    return response.data

def enrich_codes_with_context(codes: List[Dict], supabase: Client) -> List[Dict]:
    """Enrich code descriptions with keywords and additional context."""
    enriched = []
    
    for code in tqdm(codes, desc="Enriching"):
        code_type = code['code_type']
        code_value = code['code']
        description = code['description']
        
        # Build extended description for better embeddings
        extended_parts = [description]
        
        # Add industry context if available
        if code.get('industry'):
            extended_parts.append(f"Industry: {code['industry']}")
        
        # Fetch related keywords
        keywords_response = supabase.table('code_keywords').select('keyword').eq(
            'code_type', code_type
        ).eq('code', code_value).limit(15).execute()
        
        if keywords_response.data:
            keywords = [kw['keyword'] for kw in keywords_response.data]
            extended_parts.append(f"Related terms: {', '.join(keywords)}")
        
        # Fetch examples if available
        # (Adjust based on your schema)
        if code.get('examples'):
            extended_parts.append(f"Examples: {code['examples']}")
        
        # Add business type hints
        business_type_hints = get_business_type_hints(code_type, code_value)
        if business_type_hints:
            extended_parts.append(business_type_hints)
        
        enriched.append({
            'code_type': code_type,
            'code': code_value,
            'description': description,
            'extended_description': '. '.join(extended_parts),
            'industry_context': code.get('industry', ''),
        })
    
    return enriched

def get_business_type_hints(code_type: str, code: str) -> str:
    """Add contextual hints for better embeddings."""
    hints_map = {
        # MCC hints
        ('MCC', '5812'): "Restaurants, cafes, dining establishments, food service",
        ('MCC', '5814'): "Fast food restaurants, quick service, QSR",
        ('MCC', '5411'): "Grocery stores, supermarkets, food retail",
        ('MCC', '7372'): "Software development, programming services, IT consulting",
        ('MCC', '8011'): "Medical doctors, physicians, healthcare providers",
        ('MCC', '8021'): "Dental offices, dentists, orthodontists",
        # Add more hints for common codes
    }
    
    return hints_map.get((code_type, code), "")

def generate_embeddings_batch(
    codes: List[Dict],
    model: SentenceTransformer,
    batch_size: int
) -> List[Dict]:
    """Generate embeddings in batches for efficiency."""
    embeddings_data = []
    
    # Process in batches
    for i in tqdm(range(0, len(codes), batch_size), desc="Generating embeddings"):
        batch = codes[i:i+batch_size]
        
        # Extract texts for embedding
        texts = [code['extended_description'] for code in batch]
        
        # Generate embeddings for batch
        embeddings = model.encode(texts, show_progress_bar=False)
        
        # Combine with metadata
        for code, embedding in zip(batch, embeddings):
            embeddings_data.append({
                'code_type': code['code_type'],
                'code': code['code'],
                'description': code['description'],
                'extended_description': code['extended_description'],
                'industry_context': code['industry_context'],
                'embedding': embedding.tolist(),
            })
    
    return embeddings_data

def insert_embeddings(
    embeddings_data: List[Dict],
    supabase: Client,
    batch_size: int
):
    """Insert embeddings into database in batches."""
    total = len(embeddings_data)
    
    for i in tqdm(range(0, total, batch_size), desc="Inserting"):
        batch = embeddings_data[i:i+batch_size]
        
        try:
            supabase.table('code_embeddings').insert(batch).execute()
            time.sleep(0.1)  # Brief pause to avoid rate limits
        except Exception as e:
            print(f"\n❌ Error inserting batch {i//batch_size + 1}: {e}")
            # Try inserting one by one for this batch
            for item in batch:
                try:
                    supabase.table('code_embeddings').insert(item).execute()
                except Exception as e2:
                    print(f"   ❌ Failed to insert {item['code_type']} {item['code']}: {e2}")

def verify_embeddings(supabase: Client):
    """Verify embeddings were inserted correctly."""
    # Count total
    response = supabase.table('code_embeddings').select('*', count='exact').execute()
    total_count = response.count
    
    print(f"   Total embeddings in database: {total_count}")
    
    # Count by type
    for code_type in ['MCC', 'SIC', 'NAICS']:
        response = supabase.table('code_embeddings').select(
            '*', count='exact'
        ).eq('code_type', code_type).execute()
        print(f"   - {code_type}: {response.count}")
    
    # Test a similarity search
    print("\n   Testing similarity search...")
    test_embedding = [0.0] * 384  # Dummy embedding
    test_embedding[0] = 1.0
    
    try:
        result = supabase.rpc(
            'match_code_embeddings',
            {
                'query_embedding': test_embedding,
                'code_type_filter': 'MCC',
                'match_threshold': 0.0,
                'match_count': 3
            }
        ).execute()
        
        print(f"   ✅ Similarity search working (returned {len(result.data)} results)")
    except Exception as e:
        print(f"   ❌ Similarity search test failed: {e}")

if __name__ == '__main__':
    main()
```

**File:** `scripts/requirements.txt`

```txt
sentence-transformers==2.2.2
supabase==2.3.0
tqdm==4.66.1
torch==2.1.0
```

**Run Pre-computation:**
```bash
# Install dependencies
cd scripts
pip install -r requirements.txt

# Set environment variables
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_KEY="your-service-key"  # Use service key, not anon key

# Run script (takes 10-30 minutes depending on code count)
python precompute_embeddings.py
```

**Expected Output:**
```
============================================================
CODE EMBEDDINGS PRE-COMPUTATION
============================================================

1. Connecting to Supabase...
✅ Connected

2. Loading embedding model: sentence-transformers/all-MiniLM-L6-v2...
✅ Model loaded (embedding dimension: 384)

3. Fetching classification codes from database...
✅ Fetched 1523 codes
   - MCC: 487
   - SIC: 521
   - NAICS: 515

4. Enriching code descriptions with context...
Enriching: 100%|████████████████████| 1523/1523 [00:45<00:00, 33.84it/s]
✅ Descriptions enriched

5. Generating embeddings (batch size: 50)...
Generating embeddings: 100%|████████| 31/31 [02:15<00:00,  4.36s/it]
✅ Generated 1523 embeddings

6. Inserting embeddings into database...
Inserting: 100%|█████████████████████| 31/31 [00:23<00:00,  1.32it/s]
✅ All embeddings inserted

7. Verifying insertion...
   Total embeddings in database: 1523
   - MCC: 487
   - SIC: 521
   - NAICS: 515
   
   Testing similarity search...
   ✅ Similarity search working (returned 3 results)

============================================================
✅ EMBEDDINGS PRE-COMPUTATION COMPLETE!
============================================================
```

**This is a ONE-TIME operation.** You only need to re-run if:
- You add new codes to the database
- You want to update the embedding model
- The extended descriptions change significantly

---

### Task 3: Create Embedding Service (Day 3)

**Why a separate service?**
The embedding model (sentence-transformers) is Python-based. Your main classification service is Go. Rather than mixing languages, we create a dedicated Python microservice.

**File:** `services/embedding-service/app.py`

```python
"""
Embedding Service - Generate semantic embeddings for text
Deployed on Railway as a microservice
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from typing import List, Optional
import logging
import time

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Embedding Service",
    description="Generate 384-dimensional embeddings using all-MiniLM-L6-v2",
    version="1.0.0"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Load embedding model on startup
MODEL_NAME = 'sentence-transformers/all-MiniLM-L6-v2'
logger.info(f"Loading embedding model: {MODEL_NAME}...")
model = SentenceTransformer(MODEL_NAME)
logger.info(f"Model loaded successfully! Dimension: {model.get_sentence_embedding_dimension()}")

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
    return {
        "status": "healthy",
        "model": MODEL_NAME,
        "dimension": model.get_sentence_embedding_dimension(),
        "service": "embedding-service",
        "version": "1.0.0"
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
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

**File:** `services/embedding-service/requirements.txt`

```txt
fastapi==0.109.0
uvicorn==0.27.0
sentence-transformers==2.2.2
torch==2.1.0
pydantic==2.5.3
```

**File:** `services/embedding-service/Dockerfile`

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Download model at build time (caches in image)
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')"

# Copy application
COPY app.py .

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Run application
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

**File:** `services/embedding-service/.dockerignore`

```
__pycache__
*.pyc
*.pyo
*.pyd
.Python
env/
venv/
.venv/
pip-log.txt
pip-delete-this-directory.txt
.tox/
.coverage
.coverage.*
.cache
nosetests.xml
coverage.xml
*.cover
*.log
.git
.gitignore
.mypy_cache
.pytest_cache
.hypothesis
```

**Deploy to Railway:**

1. **Via Railway Dashboard:**
```bash
# Push to GitHub
git add services/embedding-service/
git commit -m "Add embedding service for Layer 2"
git push origin phase-3-embeddings

# In Railway Dashboard:
# 1. Create new service: "embedding-service"
# 2. Connect to GitHub repo
# 3. Set root directory: services/embedding-service
# 4. Railway auto-detects Dockerfile and builds
# 5. Wait for deployment (~3-5 minutes)
# 6. Note the service URL: https://embedding-service-production.up.railway.app
```

2. **Configure Resources:**
```yaml
# In Railway service settings:
Memory: 2GB (model needs ~1.5GB)
CPU: 1 vCPU
Health Check Path: /health
Port: 8000
```

3. **Test Deployment:**
```bash
# Get service URL from Railway
EMBEDDING_SERVICE_URL="https://embedding-service-production.up.railway.app"

# Test health endpoint
curl $EMBEDDING_SERVICE_URL/health

# Expected response:
# {
#   "status": "healthy",
#   "model": "sentence-transformers/all-MiniLM-L6-v2",
#   "dimension": 384,
#   "service": "embedding-service",
#   "version": "1.0.0"
# }

# Test embedding generation
curl -X POST $EMBEDDING_SERVICE_URL/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "Italian restaurant serving pizza and pasta"}'

# Expected response:
# {
#   "embedding": [0.123, -0.456, 0.789, ...], // 384 numbers
#   "dimension": 384,
#   "processing_time_ms": 45
# }
```

4. **Add to Environment Variables:**
```bash
# Add to your classification service environment
EMBEDDING_SERVICE_URL=https://embedding-service-production.up.railway.app
```

---

### Task 4: Test Infrastructure (Day 4-5)

**Test vector search in Supabase:**

```sql
-- Test 1: Basic similarity search
SELECT * FROM match_code_embeddings(
    (SELECT embedding FROM code_embeddings WHERE code = '5812' AND code_type = 'MCC' LIMIT 1),
    'MCC',
    0.7,
    5
);

-- Should return codes similar to 5812 (Eating Places)
-- Expected: 5814 (Fast Food), 5813 (Drinking Places), etc.

-- Test 2: Cross-type search
SELECT * FROM match_code_embeddings_all_types(
    (SELECT embedding FROM code_embeddings WHERE code = '5812' AND code_type = 'MCC' LIMIT 1),
    0.7,
    3
);

-- Should return similar codes across all types (MCC/SIC/NAICS)

-- Test 3: Performance test
EXPLAIN ANALYZE
SELECT * FROM match_code_embeddings(
    (SELECT embedding FROM code_embeddings WHERE code = '5812' AND code_type = 'MCC' LIMIT 1),
    'MCC',
    0.7,
    10
);

-- Should show: <10ms execution time with index usage
```

**Test embedding service performance:**

```bash
# Test single embedding
time curl -X POST $EMBEDDING_SERVICE_URL/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "Software development and consulting services"}'

# Expected: <100ms response time

# Test batch embeddings
time curl -X POST $EMBEDDING_SERVICE_URL/embed/batch \
  -H "Content-Type: application/json" \
  -d '{
    "texts": [
      "Restaurant serving Italian cuisine",
      "Software development company",
      "Dental office and clinic",
      "Auto repair and maintenance",
      "Law firm specializing in corporate law"
    ]
  }'

# Expected: <300ms for 5 texts
```

**Test end-to-end flow:**

```python
# Test script: test_embedding_flow.py
import requests
import json

# 1. Get embedding from service
text = "Cloud-native DevOps consulting specializing in Kubernetes"
response = requests.post(
    "https://embedding-service-production.up.railway.app/embed",
    json={"text": text}
)
embedding = response.json()["embedding"]

print(f"Generated embedding: {len(embedding)} dimensions")

# 2. Query Supabase for similar codes
from supabase import create_client
supabase = create_client(SUPABASE_URL, SUPABASE_KEY)

result = supabase.rpc(
    'match_code_embeddings',
    {
        'query_embedding': embedding,
        'code_type_filter': 'MCC',
        'match_threshold': 0.7,
        'match_count': 5
    }
).execute()

print(f"\nTop {len(result.data)} similar codes:")
for item in result.data:
    print(f"  {item['code']}: {item['description']} (similarity: {item['similarity']:.3f})")

# Expected output:
# Generated embedding: 384 dimensions
# 
# Top 5 similar codes:
#   7372: Computer Programming Services (similarity: 0.872)
#   7379: Computer Related Services (similarity: 0.845)
#   7373: Computer Integrated Systems Design (similarity: 0.823)
#   8742: Management Consulting Services (similarity: 0.801)
#   7371: Computer Programming (similarity: 0.798)
```

---

## Week 6: Integration

### Task 5: Build Go Embedding Classifier (Day 1-2)

**File:** `internal/classification/embedding_classifier.go`

```go
package classification

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "strings"
    "time"
)

type EmbeddingClassifier struct {
    embeddingServiceURL string
    supabaseRepo        Repository
    httpClient          *http.Client
}

func NewEmbeddingClassifier(
    embeddingServiceURL string,
    repo Repository,
) *EmbeddingClassifier {
    return &EmbeddingClassifier{
        embeddingServiceURL: embeddingServiceURL,
        supabaseRepo:        repo,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

type EmbeddingClassificationResult struct {
    MCC             []CodeResult
    SIC             []CodeResult
    NAICS           []CodeResult
    Confidence      float64
    Method          string
    TopMatch        string  // Description of top match
    TopSimilarity   float64
    ProcessingTimeMs int64
}

// Main classification method
func (e *EmbeddingClassifier) ClassifyByEmbedding(
    ctx context.Context,
    content *ScrapedContent,
) (*EmbeddingClassificationResult, error) {
    
    startTime := time.Now()
    
    slog.Info("Starting embedding-based classification",
        "domain", content.Domain)
    
    // Step 1: Prepare text for embedding
    text := e.prepareTextForEmbedding(content)
    
    if len(text) < 50 {
        return nil, fmt.Errorf("insufficient text for embedding: %d chars", len(text))
    }
    
    slog.Debug("Prepared text for embedding",
        "length", len(text),
        "preview", text[:min(100, len(text))])
    
    // Step 2: Generate embedding
    embedding, err := e.getEmbedding(ctx, text)
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }
    
    slog.Info("Generated embedding", "dimension", len(embedding))
    
    // Step 3: Search for similar codes (each type)
    mccMatches, err := e.searchSimilarCodes(ctx, embedding, "MCC", 0.70, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to search MCC codes: %w", err)
    }
    
    sicMatches, err := e.searchSimilarCodes(ctx, embedding, "SIC", 0.70, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to search SIC codes: %w", err)
    }
    
    naicsMatches, err := e.searchSimilarCodes(ctx, embedding, "NAICS", 0.70, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to search NAICS codes: %w", err)
    }
    
    // Step 4: Select top 3 from each type
    result := &EmbeddingClassificationResult{
        MCC:    e.selectTopCodes(mccMatches, 3),
        SIC:    e.selectTopCodes(sicMatches, 3),
        NAICS:  e.selectTopCodes(naicsMatches, 3),
        Method: "embedding_similarity",
    }
    
    // Step 5: Calculate overall confidence
    if len(mccMatches) > 0 {
        result.TopMatch = mccMatches[0].Description
        result.TopSimilarity = mccMatches[0].Confidence
        result.Confidence = e.calculateConfidence(mccMatches, sicMatches, naicsMatches)
    } else {
        result.Confidence = 0.0
    }
    
    result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    
    slog.Info("Embedding classification complete",
        "confidence", result.Confidence,
        "top_match", result.TopMatch,
        "top_similarity", result.TopSimilarity,
        "duration_ms", result.ProcessingTimeMs)
    
    return result, nil
}

// Prepare text from scraped content for embedding
func (e *EmbeddingClassifier) prepareTextForEmbedding(content *ScrapedContent) string {
    parts := []string{}
    
    // Priority order: Title > Meta > About > Headings > Navigation
    
    // Title (highest signal - include 2x)
    if content.Title != "" {
        parts = append(parts, content.Title)
        parts = append(parts, content.Title) // Repeat for emphasis
    }
    
    // Meta description (high signal)
    if content.MetaDesc != "" {
        parts = append(parts, content.MetaDesc)
    }
    
    // About section (most contextual info)
    if content.AboutText != "" {
        // Limit to 500 chars to keep embedding focused
        aboutText := content.AboutText
        if len(aboutText) > 500 {
            aboutText = aboutText[:500]
        }
        parts = append(parts, aboutText)
    }
    
    // Top headings (good signal)
    if len(content.Headings) > 0 {
        // Take first 5 headings
        headingCount := min(5, len(content.Headings))
        headings := strings.Join(content.Headings[:headingCount], ". ")
        parts = append(parts, headings)
    }
    
    // Navigation (indicates business areas)
    if len(content.NavMenu) > 0 {
        // Take first 10 nav items
        navCount := min(10, len(content.NavMenu))
        nav := strings.Join(content.NavMenu[:navCount], ", ")
        parts = append(parts, nav)
    }
    
    // Combine
    combined := strings.Join(parts, ". ")
    
    // Truncate to 5000 chars (model limit)
    if len(combined) > 5000 {
        combined = combined[:5000]
    }
    
    return combined
}

// Get embedding from embedding service
func (e *EmbeddingClassifier) getEmbedding(ctx context.Context, text string) ([]float64, error) {
    reqBody := map[string]interface{}{
        "text":             text,
        "truncate_length":  5000,
    }
    
    reqBodyJSON, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        e.embeddingServiceURL+"/embed",
        bytes.NewReader(reqBodyJSON),
    )
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := e.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
    }
    
    var result struct {
        Embedding       []float64 `json:"embedding"`
        Dimension       int       `json:"dimension"`
        ProcessingTimeMs int       `json:"processing_time_ms"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    slog.Debug("Embedding service response",
        "dimension", result.Dimension,
        "processing_time_ms", result.ProcessingTimeMs)
    
    return result.Embedding, nil
}

// Search for similar codes using vector search
func (e *EmbeddingClassifier) searchSimilarCodes(
    ctx context.Context,
    embedding []float64,
    codeType string,
    threshold float64,
    limit int,
) ([]CodeResult, error) {
    
    // Call Supabase RPC function
    type Match struct {
        Code        string  `json:"code"`
        Description string  `json:"description"`
        ExtendedDesc string `json:"extended_description"`
        Similarity  float64 `json:"similarity"`
    }
    
    // Use your repository method to call the RPC
    matches, err := e.supabaseRepo.MatchCodeEmbeddings(
        ctx,
        embedding,
        codeType,
        threshold,
        limit,
    )
    if err != nil {
        return nil, err
    }
    
    // Convert to CodeResult
    results := make([]CodeResult, 0, len(matches))
    for _, match := range matches {
        results = append(results, CodeResult{
            Code:        match.Code,
            Description: match.Description,
            Confidence:  match.Similarity,
            Source:      "embedding_similarity",
        })
    }
    
    slog.Debug("Vector search results",
        "code_type", codeType,
        "matches", len(results),
        "top_similarity", func() float64 {
            if len(results) > 0 {
                return results[0].Confidence
            }
            return 0.0
        }())
    
    return results, nil
}

// Select top N codes from matches
func (e *EmbeddingClassifier) selectTopCodes(matches []CodeResult, limit int) []CodeResult {
    if len(matches) == 0 {
        return []CodeResult{}
    }
    
    // Already sorted by similarity from database
    if len(matches) > limit {
        return matches[:limit]
    }
    
    return matches
}

// Calculate overall confidence from matches
func (e *EmbeddingClassifier) calculateConfidence(
    mccMatches, sicMatches, naicsMatches []CodeResult,
) float64 {
    
    // Start with top MCC match similarity
    baseConfidence := 0.0
    if len(mccMatches) > 0 {
        baseConfidence = mccMatches[0].Confidence
    }
    
    // Boost if we have strong matches across all types
    if len(mccMatches) > 0 && len(sicMatches) > 0 && len(naicsMatches) > 0 {
        // Check if top matches are all high similarity
        mccTop := mccMatches[0].Confidence
        sicTop := sicMatches[0].Confidence
        naicsTop := naicsMatches[0].Confidence
        
        if mccTop > 0.85 && sicTop > 0.85 && naicsTop > 0.85 {
            baseConfidence *= 1.10 // +10% boost for strong agreement
        } else if mccTop > 0.80 && sicTop > 0.80 && naicsTop > 0.80 {
            baseConfidence *= 1.05 // +5% boost for good agreement
        }
    }
    
    // Check agreement between top 3 MCC matches
    if len(mccMatches) >= 3 {
        similarities := []float64{
            mccMatches[0].Confidence,
            mccMatches[1].Confidence,
            mccMatches[2].Confidence,
        }
        
        // If top 3 are all similar (tight cluster), boost confidence
        maxDiff := similarities[0] - similarities[2]
        if maxDiff < 0.10 {
            baseConfidence *= 1.08 // +8% boost for tight cluster
        }
    }
    
    // Cap at 0.92 (embeddings alone shouldn't claim >92% confidence)
    if baseConfidence > 0.92 {
        baseConfidence = 0.92
    }
    
    return baseConfidence
}

// Helper function
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**Add Repository Method:**

**File:** `internal/classification/repository/supabase_repository.go`

```go
type CodeMatch struct {
    Code        string
    Description string
    Similarity  float64
}

func (r *SupabaseKeywordRepository) MatchCodeEmbeddings(
    ctx context.Context,
    embedding []float64,
    codeType string,
    threshold float64,
    limit int,
) ([]CodeMatch, error) {
    
    // Call Supabase RPC function
    var matches []struct {
        Code        string  `json:"code"`
        Description string  `json:"description"`
        Similarity  float64 `json:"similarity"`
    }
    
    err := r.supabaseClient.Rpc(
        "match_code_embeddings",
        map[string]interface{}{
            "query_embedding":   embedding,
            "code_type_filter": codeType,
            "match_threshold":  threshold,
            "match_count":      limit,
        },
    ).Execute(&matches)
    
    if err != nil {
        return nil, fmt.Errorf("failed to call match_code_embeddings: %w", err)
    }
    
    // Convert to CodeMatch
    results := make([]CodeMatch, len(matches))
    for i, m := range matches {
        results[i] = CodeMatch{
            Code:        m.Code,
            Description: m.Description,
            Similarity:  m.Similarity,
        }
    }
    
    return results, nil
}
```

---

### Task 6: Add Layer 2 Routing (Day 3)

**File:** `internal/classification/service.go`

```go
type IndustryDetectionService struct {
    // Existing fields...
    embeddingClassifier *EmbeddingClassifier // NEW
}

func NewIndustryDetectionService(
    // ... existing params ...
    embeddingServiceURL string,
) *IndustryDetectionService {
    return &IndustryDetectionService{
        // ... existing initialization ...
        embeddingClassifier: NewEmbeddingClassifier(
            embeddingServiceURL,
            repo,
        ),
    }
}

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    startTime := time.Now()
    
    slog.Info("Starting industry detection", "url", websiteURL)
    
    // Scrape website
    content, err := s.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, fmt.Errorf("scraping failed: %w", err)
    }
    
    // Layer 1: Multi-strategy classification
    layer1Result, err := s.multiStrategyClassifier.ClassifyWithMultiStrategy(
        ctx,
        businessName,
        description,
        websiteURL,
    )
    if err != nil {
        return nil, fmt.Errorf("layer 1 classification failed: %w", err)
    }
    
    slog.Info("Layer 1 complete",
        "industry", layer1Result.Industry,
        "confidence", layer1Result.Confidence,
        "method", layer1Result.Method)
    
    // Decision: Use Layer 1 or try Layer 2?
    
    // High confidence from Layer 1 → Use it
    if layer1Result.Confidence >= 0.90 {
        slog.Info("High confidence from Layer 1, using result",
            "confidence", layer1Result.Confidence)
        
        result := s.buildResult(layer1Result, "layer1")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Medium-high confidence (0.80-0.90) → Use Layer 1 but log for monitoring
    if layer1Result.Confidence >= 0.80 {
        slog.Info("Good confidence from Layer 1, using result",
            "confidence", layer1Result.Confidence)
        
        result := s.buildResult(layer1Result, "layer1_medium_conf")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Lower confidence (<0.80) → Try Layer 2 (Embeddings)
    slog.Info("Layer 1 confidence below threshold, trying Layer 2",
        "layer1_confidence", layer1Result.Confidence)
    
    layer2Result, err := s.embeddingClassifier.ClassifyByEmbedding(ctx, content)
    if err != nil {
        slog.Warn("Layer 2 failed, falling back to Layer 1",
            "error", err)
        
        result := s.buildResult(layer1Result, "layer1_fallback")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    slog.Info("Layer 2 complete",
        "top_match", layer2Result.TopMatch,
        "confidence", layer2Result.Confidence,
        "similarity", layer2Result.TopSimilarity)
    
    // Compare Layer 1 vs Layer 2
    if layer2Result.Confidence > layer1Result.Confidence + 0.05 {
        // Layer 2 is meaningfully better
        slog.Info("Using Layer 2 result",
            "layer2_confidence", layer2Result.Confidence,
            "layer1_confidence", layer1Result.Confidence,
            "improvement", layer2Result.Confidence - layer1Result.Confidence)
        
        result := s.buildResultFromEmbedding(layer2Result, "layer2_embedding")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    } else {
        // Layer 1 and Layer 2 similar, prefer Layer 1 (more explainable)
        slog.Info("Layer 1 and Layer 2 similar, using Layer 1",
            "layer1_confidence", layer1Result.Confidence,
            "layer2_confidence", layer2Result.Confidence)
        
        result := s.buildResult(layer1Result, "layer1_validated")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
}

func (s *IndustryDetectionService) buildResultFromEmbedding(
    embResult *EmbeddingClassificationResult,
    layer string,
) *IndustryDetectionResult {
    
    // Derive primary industry from top MCC match
    primaryIndustry := "Unknown"
    if len(embResult.MCC) > 0 {
        primaryIndustry = s.getIndustryFromCode("MCC", embResult.MCC[0].Code)
    }
    
    return &IndustryDetectionResult{
        Classification: ClassificationData{
            PrimaryIndustry: primaryIndustry,
            Confidence:      embResult.Confidence,
            Method:          embResult.Method,
            Keywords:        []string{}, // Embeddings don't use keywords
        },
        Codes: ClassificationCodes{
            MCC:   embResult.MCC,
            SIC:   embResult.SIC,
            NAICS: embResult.NAICS,
        },
        Explanation: ClassificationExplanation{
            PrimaryReason: fmt.Sprintf(
                "Semantic similarity analysis matched '%s' with %.0f%% confidence",
                embResult.TopMatch,
                embResult.TopSimilarity*100,
            ),
            SupportingFactors: []string{
                "Vector embedding-based similarity search",
                fmt.Sprintf("Top semantic match: %s", embResult.TopMatch),
                "Handles industry terminology and context",
            },
            KeyTermsFound:     []string{},
            MethodUsed:        embResult.Method,
            ProcessingPath:    layer,
        },
        ProcessingTimeMs: embResult.ProcessingTimeMs,
    }
}

func (s *IndustryDetectionService) getIndustryFromCode(codeType, code string) string {
    // Query industries table or use description
    industry := s.repo.GetIndustryByCode(codeType, code)
    if industry != nil {
        return industry.Name
    }
    
    // Fallback: use code description
    codeInfo := s.repo.GetCodeInfo(codeType, code)
    if codeInfo != nil {
        return codeInfo.Description
    }
    
    return "Unknown"
}
```

---

### Task 7: Full Testing (Day 4-5)

**Test Layer 2 directly:**

```bash
# Test case 1: Novel business terminology
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "Cloud-native DevOps consultancy",
    "description": "Kubernetes orchestration and CI/CD pipeline automation"
  }'

# Expected:
# - Layer 1 confidence: ~0.65 (struggles with jargon)
# - Layer 2 triggers
# - Top match: "Computer Programming Services" or "Management Consulting"
# - Final confidence: 0.82-0.88

# Test case 2: Standard business (Layer 1 should handle)
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "Joe'\''s Pizza Restaurant",
    "website_url": "https://joespizza.com"
  }'

# Expected:
# - Layer 1 confidence: 0.92+
# - Layer 2 NOT triggered (confidence already high)
# - Method: "fast_path_keyword" or "multi_strategy"

# Test case 3: Ambiguous business
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "Smith & Associates",
    "description": "Professional services firm"
  }'

# Expected:
# - Layer 1 confidence: ~0.70 (generic)
# - Layer 2 triggers
# - Embedding provides more specific classification
# - Final confidence: 0.78-0.85
```

**Run full test suite:**

```bash
# Test on your complete test set
# Track:
# - How often Layer 2 is triggered (expect 20-30%)
# - Accuracy improvement from Layer 2
# - Processing time distribution

# Expected results:
# - Overall accuracy: 85-90%
# - Layer 1 handles: 70-80% of cases
# - Layer 2 improves: 15-25% of cases
# - Layer 2 avg latency: 400-800ms (includes embedding generation + vector search)
```

**Performance benchmarks:**

```
Layer 1 (Fast Path): <100ms, 60-70% of requests
Layer 1 (Full):      200-500ms, 15-20% of requests  
Layer 2 (Embedding): 400-800ms, 15-25% of requests

Overall p50: <200ms
Overall p90: <600ms
Overall p95: <900ms
```

---

## Phase 3 Success Criteria

Before moving to Phase 4, verify:

### Infrastructure
- [ ] ✅ pgvector enabled in Supabase
- [ ] ✅ Code embeddings pre-computed (1500+ codes)
- [ ] ✅ Embedding service deployed on Railway
- [ ] ✅ Vector search RPC functions working
- [ ] ✅ Embedding service <100ms response time

### Integration
- [ ] ✅ Go embedding classifier implemented
- [ ] ✅ Layer 2 routing logic added
- [ ] ✅ Can call embedding service from Go
- [ ] ✅ Vector search queries working

### Performance
- [ ] ✅ Layer 2 triggers for 15-25% of requests
- [ ] ✅ Layer 2 latency <800ms (p95)
- [ ] ✅ Overall p95 latency <900ms
- [ ] ✅ No performance degradation for Layer 1 cases

### Quality
- [ ] ✅ Accuracy improved to 85-90% on test set
- [ ] ✅ Edge cases handled better
- [ ] ✅ Novel terminology classifications improved
- [ ] ✅ Semantic understanding working

---

## Expected Improvement Summary

**Before Phase 3:**
- Accuracy: 80-85%
- Handles: Common cases well
- Struggles: Edge cases, novel businesses, non-standard terminology
- Layers: 1 layer (multi-strategy)

**After Phase 3:**
- Accuracy: ✅ 85-90%
- Handles: ✅ Common + edge cases
- Struggles: ✅ Much improved on edge cases
- Layers: ✅ 2 layers with intelligent routing

**Key Wins:**
- Semantic understanding via embeddings
- Catches cases that Layer 1 misses
- Handles industry jargon and new business models
- Foundation for Layer 3 (LLM) ready

---

## Next Steps: Phase 4

Once Phase 3 is complete:
- **Phase 4 (Weeks 7-8):** Add Layer 3 (LLM for complex cases)
- **Expected accuracy:** 90-95%
- **What it adds:** Reasoning for truly complex/ambiguous cases

**Phase 4 Guide will be provided once Phase 3 is complete.**

---

## Troubleshooting

**Issue: pgvector not installing**
```sql
-- Check PostgreSQL version (need 11+)
SELECT version();

-- Try manual installation
CREATE EXTENSION vector;

-- If fails, contact Supabase support
```

**Issue: Embedding service slow**
```bash
# Check Railway logs
railway logs -s embedding-service

# Verify model loaded
curl https://your-embedding-service.up.railway.app/info

# Check memory (needs 2GB)
```

**Issue: Vector search returns no results**
```sql
-- Check embeddings exist
SELECT COUNT(*) FROM code_embeddings;

-- Test with lower threshold
SELECT * FROM match_code_embeddings(
    (SELECT embedding FROM code_embeddings LIMIT 1),
    'MCC',
    0.5,  -- Lower threshold
    10
);
```

**Issue: Layer 2 never triggers**
- Check layer1Result.Confidence values in logs
- Verify routing threshold (should be 0.80)
- Check embedding service URL is set correctly

**Issue: Layer 2 too slow**
- Verify embedding service is on Railway (not local)
- Check pgvector index exists
- Consider increasing threshold to 0.75 (trigger less often)

---

## Questions During Phase 3?

Check in after:
- Day 2 (embeddings pre-computed)
- Day 3 (embedding service deployed)
- Day 5 (Phase 3 complete)

You're building something sophisticated now - a true 2-layer hybrid system! 🚀
