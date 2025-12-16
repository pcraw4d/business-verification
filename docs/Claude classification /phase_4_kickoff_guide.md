# Phase 4 Kick-Off Guide: Add Layer 3 (LLM)
## Weeks 7-8: From 85-90% to 90-95% Accuracy

**Goal:** Add LLM-based reasoning for truly complex, ambiguous, or novel cases that Layers 1 and 2 can't confidently classify.

---

## Phase 3 Success Validation

Before starting Phase 4, verify your Phase 3 results:

✅ **Checklist:**
- [ ] pgvector enabled with 1500+ embeddings
- [ ] Embedding service deployed (<100ms response)
- [ ] Layer 2 triggers for 20-30% of requests
- [ ] Layer 2 routing logic working
- [ ] Accuracy at 85-90% on test set

**If all checked:** You're ready for Phase 4! 🎉

**If issues remain:** Address them before proceeding. Layer 3 is the final refinement.

---

## Phase 4 Overview

### What We're Adding

**Current State (After Phase 3):**
- ✅ Layer 1 (Multi-strategy) handles common cases (60-70%)
- ✅ Layer 2 (Embeddings) handles edge cases (15-25%)
- ✅ 85-90% accuracy
- ❌ Still struggles with truly ambiguous businesses
- ❌ Can't reason about complex industry classifications
- ❌ No explanation for why a classification was chosen

**Target State (After Phase 4):**
- ✅ Layer 3 (LLM) handles complex reasoning (5-15%)
- ✅ Can explain its reasoning in natural language
- ✅ Handles truly novel business models
- ✅ 3-layer orchestration with fallback logic
- ✅ 90-95% accuracy

### Why LLM (Large Language Model)?

**The Gap Layers 1-2 Can't Fill:**

```
Business: "We provide fractional CFO services to seed-stage B2B SaaS 
companies, combining strategic financial planning with hands-on 
controller functions and investor relations support."

Layer 1 (Keywords): 
- Finds "financial", "CFO" → matches "Accounting Services"
- Confidence: 0.62 (uncertain about "fractional", "seed-stage", "SaaS")

Layer 2 (Embeddings):
- Semantic match to "Management Consulting" or "Accounting Services"
- Confidence: 0.73 (can't reason about the combined services)

Layer 3 (LLM):
- Reasons: "This is a specialized consulting service combining CFO 
  expertise with startup-specific needs. Primary classification: 
  Management Consulting (MCC 8742), with secondary classification 
  of Accounting Services (MCC 8721)."
- Confidence: 0.88
- Provides detailed reasoning
```

**LLM Capabilities:**
- **Reasoning:** Understands complex business models
- **Context:** Considers industry trends and business evolution
- **Explanation:** Can articulate why a classification makes sense
- **Novel cases:** Handles business types not in training data
- **Ambiguity resolution:** Weighs multiple factors to decide

### Implementation Timeline

**Week 7:**
- Day 1-2: Deploy LLM service (Qwen 2.5 7B on Railway)
- Day 3: Build structured prompts
- Day 4: Test LLM service directly
- Day 5: Build Go LLM classifier

**Week 8:**
- Day 1-2: Add Layer 3 routing logic
- Day 3: Integration testing
- Day 4: Optimize prompts and confidence
- Day 5: Full test suite validation

---

## Week 7: LLM Service Setup

### Task 1: Deploy LLM Service (Day 1-2)

**Why Qwen 2.5 7B?**
- Better reasoning than older models (GPT-2, DistilBART)
- Instruction-tuned for following prompts
- 7B params: Large enough for reasoning, small enough for Railway
- Open source: No API costs
- JSON mode: Can output structured responses

**File:** `services/llm-service/app.py`

```python
"""
LLM Service - Industry classification reasoning using Qwen 2.5 7B
Deployed on Railway as a microservice
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional, Dict
import logging
import time
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
import json

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="LLM Classification Service",
    description="Industry classification reasoning using Qwen 2.5 7B",
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

# Model configuration
MODEL_NAME = "Qwen/Qwen2.5-7B-Instruct"
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
logger.info(f"Using device: {DEVICE}")

# Load model and tokenizer on startup
logger.info(f"Loading model: {MODEL_NAME}...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForCausalLM.from_pretrained(
    MODEL_NAME,
    torch_dtype=torch.float16 if DEVICE == "cuda" else torch.float32,
    device_map="auto" if DEVICE == "cuda" else None,
)
if DEVICE == "cpu":
    model = model.to(DEVICE)

model.eval()  # Set to evaluation mode
logger.info(f"Model loaded successfully on {DEVICE}!")

# Request/Response models
class ClassificationContext(BaseModel):
    business_name: str
    description: Optional[str] = ""
    website_content: Optional[str] = ""
    layer1_result: Optional[Dict] = None
    layer2_result: Optional[Dict] = None

class ClassificationRequest(BaseModel):
    context: ClassificationContext
    temperature: Optional[float] = 0.1
    max_tokens: Optional[int] = 800

class ClassificationResponse(BaseModel):
    primary_industry: str
    confidence: float
    reasoning: str
    codes: Dict[str, List[Dict[str, any]]]
    alternative_classifications: List[str]
    processing_time_ms: int

# Health check endpoint
@app.get("/health")
async def health_check():
    """Health check endpoint for Railway"""
    return {
        "status": "healthy",
        "model": MODEL_NAME,
        "device": DEVICE,
        "service": "llm-service",
        "version": "1.0.0"
    }

# Classification endpoint
@app.post("/classify", response_model=ClassificationResponse)
async def classify_business(request: ClassificationRequest):
    """
    Classify a business using LLM reasoning.
    
    Example:
        POST /classify
        {
            "context": {
                "business_name": "Acme Corp",
                "description": "We provide X services...",
                "website_content": "...",
                "layer1_result": {...},
                "layer2_result": {...}
            }
        }
    """
    start_time = time.time()
    
    try:
        # Build prompt
        prompt = build_classification_prompt(request.context)
        
        logger.info(f"Classifying business: {request.context.business_name}")
        logger.debug(f"Prompt length: {len(prompt)} chars")
        
        # Generate response
        response = generate_classification(
            prompt,
            temperature=request.temperature,
            max_tokens=request.max_tokens
        )
        
        logger.info(f"LLM response generated (length: {len(response)} chars)")
        
        # Parse structured response
        result = parse_classification_response(response)
        
        processing_time = int((time.time() - start_time) * 1000)
        result["processing_time_ms"] = processing_time
        
        logger.info(f"Classification complete: {result['primary_industry']} "
                   f"(confidence: {result['confidence']:.2f}, time: {processing_time}ms)")
        
        return result
        
    except Exception as e:
        logger.error(f"Error during classification: {e}")
        raise HTTPException(status_code=500, detail=f"Classification failed: {str(e)}")

def build_classification_prompt(context: ClassificationContext) -> str:
    """Build structured prompt for LLM classification."""
    
    # System message
    system_msg = """You are an expert business classification system. Your task is to classify businesses into industry codes (MCC, SIC, NAICS) based on their description and operations.

You will receive information about a business and must determine:
1. The primary industry classification
2. Confidence level (0.0 to 1.0)
3. Detailed reasoning for your classification
4. Top 3 industry codes for each type (MCC, SIC, NAICS)
5. Alternative classifications if applicable

Be thorough in your reasoning, considering:
- Core business activities
- Revenue generation methods
- Industry standards and categorizations
- Similar business types
- Edge cases and ambiguities

Respond ONLY with a valid JSON object in this exact format:
{
    "primary_industry": "Industry name",
    "confidence": 0.85,
    "reasoning": "Detailed explanation of classification...",
    "codes": {
        "mcc": [
            {"code": "1234", "description": "...", "confidence": 0.90},
            {"code": "1235", "description": "...", "confidence": 0.85},
            {"code": "1236", "description": "...", "confidence": 0.78}
        ],
        "sic": [...],
        "naics": [...]
    },
    "alternative_classifications": ["Alternative 1", "Alternative 2"]
}"""
    
    # Build user message with context
    user_msg_parts = []
    
    user_msg_parts.append(f"BUSINESS INFORMATION:")
    user_msg_parts.append(f"Name: {context.business_name}")
    
    if context.description:
        user_msg_parts.append(f"\nDescription: {context.description}")
    
    if context.website_content:
        # Truncate to 2000 chars to fit in context
        content = context.website_content[:2000]
        user_msg_parts.append(f"\nWebsite Content:\n{content}")
    
    # Add Layer 1 hints if available
    if context.layer1_result:
        user_msg_parts.append(f"\nLayer 1 Classification (keyword-based):")
        user_msg_parts.append(f"  Industry: {context.layer1_result.get('industry', 'Unknown')}")
        user_msg_parts.append(f"  Confidence: {context.layer1_result.get('confidence', 0.0):.2f}")
        if context.layer1_result.get('keywords'):
            keywords = ', '.join(context.layer1_result['keywords'][:5])
            user_msg_parts.append(f"  Keywords found: {keywords}")
    
    # Add Layer 2 hints if available
    if context.layer2_result:
        user_msg_parts.append(f"\nLayer 2 Classification (embedding-based):")
        user_msg_parts.append(f"  Top match: {context.layer2_result.get('top_match', 'Unknown')}")
        user_msg_parts.append(f"  Similarity: {context.layer2_result.get('top_similarity', 0.0):.2f}")
        user_msg_parts.append(f"  Confidence: {context.layer2_result.get('confidence', 0.0):.2f}")
    
    user_msg_parts.append("\nBased on this information, classify this business into the appropriate industry codes.")
    user_msg_parts.append("Remember to respond with ONLY a valid JSON object, no additional text.")
    
    user_msg = "\n".join(user_msg_parts)
    
    # Combine into chat format
    messages = [
        {"role": "system", "content": system_msg},
        {"role": "user", "content": user_msg}
    ]
    
    # Format for Qwen
    prompt = tokenizer.apply_chat_template(
        messages,
        tokenize=False,
        add_generation_prompt=True
    )
    
    return prompt

def generate_classification(
    prompt: str,
    temperature: float = 0.1,
    max_tokens: int = 800
) -> str:
    """Generate classification using LLM."""
    
    # Tokenize
    inputs = tokenizer(prompt, return_tensors="pt").to(DEVICE)
    
    # Generate
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_new_tokens=max_tokens,
            temperature=temperature,
            do_sample=temperature > 0,
            top_p=0.9,
            repetition_penalty=1.1,
            pad_token_id=tokenizer.pad_token_id,
            eos_token_id=tokenizer.eos_token_id,
        )
    
    # Decode
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)
    
    # Extract just the assistant's response (after the prompt)
    # The response includes the full prompt + generation
    # We want just the new generation
    prompt_length = len(tokenizer.decode(inputs['input_ids'][0], skip_special_tokens=True))
    response = response[prompt_length:].strip()
    
    return response

def parse_classification_response(response: str) -> Dict:
    """Parse LLM response into structured format."""
    
    # Try to extract JSON from response
    # Sometimes LLM adds extra text before/after JSON
    
    # Find JSON object
    start_idx = response.find('{')
    end_idx = response.rfind('}') + 1
    
    if start_idx == -1 or end_idx == 0:
        raise ValueError(f"No JSON object found in response: {response[:200]}")
    
    json_str = response[start_idx:end_idx]
    
    try:
        parsed = json.loads(json_str)
    except json.JSONDecodeError as e:
        logger.error(f"Failed to parse JSON: {json_str[:200]}")
        raise ValueError(f"Invalid JSON in response: {e}")
    
    # Validate required fields
    required_fields = ['primary_industry', 'confidence', 'reasoning', 'codes']
    for field in required_fields:
        if field not in parsed:
            raise ValueError(f"Missing required field: {field}")
    
    # Ensure codes have MCC/SIC/NAICS
    if 'mcc' not in parsed['codes']:
        parsed['codes']['mcc'] = []
    if 'sic' not in parsed['codes']:
        parsed['codes']['sic'] = []
    if 'naics' not in parsed['codes']:
        parsed['codes']['naics'] = []
    
    # Ensure alternative_classifications exists
    if 'alternative_classifications' not in parsed:
        parsed['alternative_classifications'] = []
    
    # Validate confidence is in range
    confidence = float(parsed['confidence'])
    if confidence < 0.0 or confidence > 1.0:
        logger.warning(f"Confidence out of range: {confidence}, clamping to [0, 1]")
        parsed['confidence'] = max(0.0, min(1.0, confidence))
    
    return parsed

# Info endpoint
@app.get("/info")
async def get_info():
    """Get information about the LLM service"""
    return {
        "model": MODEL_NAME,
        "device": DEVICE,
        "parameters": "7B",
        "description": "Industry classification using LLM reasoning",
        "endpoints": {
            "/classify": "Classify business with reasoning",
            "/health": "Health check",
            "/info": "Service information"
        }
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

**File:** `services/llm-service/requirements.txt`

```txt
fastapi==0.109.0
uvicorn==0.27.0
transformers==4.36.0
torch==2.1.0
pydantic==2.5.3
accelerate==0.25.0
sentencepiece==0.1.99
protobuf==4.25.1
```

**File:** `services/llm-service/Dockerfile`

```dockerfile
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
COPY requirements.txt .

# Install Python dependencies
RUN pip3 install --no-cache-dir -r requirements.txt

# Download model at build time (caches in image, ~14GB)
RUN python3 -c "from transformers import AutoModelForCausalLM, AutoTokenizer; \
    AutoTokenizer.from_pretrained('Qwen/Qwen2.5-7B-Instruct'); \
    AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct', device_map='auto')"

# Copy application
COPY app.py .

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=180s --retries=3 \
    CMD python3 -c "import requests; requests.get('http://localhost:8000/health', timeout=10)"

# Run application
CMD ["python3", "-m", "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

**Deploy to Railway:**

```bash
# 1. Push to GitHub
git add services/llm-service/
git commit -m "Add LLM service for Layer 3"
git push origin phase-4-llm

# 2. In Railway Dashboard:
# - Create new service: "llm-service"
# - Connect to GitHub repo
# - Set root directory: services/llm-service
# - IMPORTANT: Configure resources

# 3. Resource Configuration (Critical!)
Memory: 8GB (minimum for 7B model)
CPU: 2 vCPU
GPU: Optional (speeds up 5-10x, but adds cost)
Timeout: 300 seconds (model loading takes time)
Health Check Path: /health
Port: 8000

# 4. Wait for build (~15-20 minutes, model is 14GB)
# Railway will:
# - Build Docker image
# - Download Qwen 2.5 7B model (~14GB)
# - Start service
# - Run health check

# 5. Note service URL
LLM_SERVICE_URL="https://llm-service-production.up.railway.app"
```

**Railway Cost Note:**
- Without GPU: ~$20-30/month (8GB RAM)
- With GPU: ~$100-150/month (faster inference)
- For development: Start without GPU, add later if needed

**Test Deployment:**

```bash
# Test health
curl https://llm-service-production.up.railway.app/health

# Expected:
# {
#   "status": "healthy",
#   "model": "Qwen/Qwen2.5-7B-Instruct",
#   "device": "cpu",  # or "cuda" if GPU enabled
#   "service": "llm-service"
# }

# Test classification
curl -X POST https://llm-service-production.up.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{
    "context": {
      "business_name": "Acme Restaurant",
      "description": "Italian restaurant serving pizza and pasta"
    }
  }'

# Expected: JSON response with classification after 2-5 seconds
```

---

### Task 2: Build Structured Prompts (Day 3)

**Why Prompt Engineering Matters:**

Poor prompt → Poor results:
```
"Classify this business: Acme Corp provides financial advisory services"
→ Generic response, low confidence, no reasoning
```

Good prompt → Great results:
```
System: You are an expert classifier...
Context: Business name, description, website, Layer 1/2 results
Format: JSON with specific fields
→ Specific classification, high confidence, detailed reasoning
```

**Prompt Components:**

**1. System Message** (Sets role and expectations)
```
You are an expert business classification system specializing in 
merchant underwriting and KYB (Know Your Business) processes.

Your task: Classify businesses into industry codes with high accuracy.

Requirements:
- Provide detailed reasoning
- Consider edge cases
- Output structured JSON
- Be confident when appropriate, cautious when uncertain
```

**2. Context** (Give LLM all available information)
```
Business Name: [name]
Description: [description]
Website Content: [first 2000 chars]

Layer 1 Result: [keyword-based classification]
Layer 2 Result: [embedding-based classification]

Based on ALL this information, classify the business.
```

**3. Format Instructions** (Exact output structure)
```json
{
    "primary_industry": "Specific Industry Name",
    "confidence": 0.85,
    "reasoning": "I classified this as X because...",
    "codes": {
        "mcc": [{"code": "1234", "description": "...", "confidence": 0.90}, ...],
        "sic": [...],
        "naics": [...]
    },
    "alternative_classifications": ["Alt 1", "Alt 2"]
}
```

**Test Prompts:**

```python
# Test script: test_llm_prompts.py
import requests
import json

LLM_URL = "https://llm-service-production.up.railway.app"

# Test 1: Simple case
simple_case = {
    "context": {
        "business_name": "Pizza Palace",
        "description": "Italian restaurant specializing in wood-fired pizza"
    }
}

response = requests.post(f"{LLM_URL}/classify", json=simple_case)
result = response.json()

print("Test 1: Simple Restaurant")
print(f"  Industry: {result['primary_industry']}")
print(f"  Confidence: {result['confidence']}")
print(f"  Reasoning: {result['reasoning'][:100]}...")
print()

# Test 2: Ambiguous case with Layer 1/2 context
ambiguous_case = {
    "context": {
        "business_name": "Smith & Associates",
        "description": "Professional services firm providing strategic advisory",
        "layer1_result": {
            "industry": "Business Services",
            "confidence": 0.68,
            "keywords": ["professional", "services", "advisory"]
        },
        "layer2_result": {
            "top_match": "Management Consulting Services",
            "confidence": 0.74,
            "top_similarity": 0.78
        }
    }
}

response = requests.post(f"{LLM_URL}/classify", json=ambiguous_case)
result = response.json()

print("Test 2: Ambiguous Consulting")
print(f"  Industry: {result['primary_industry']}")
print(f"  Confidence: {result['confidence']}")
print(f"  Codes: MCC={result['codes']['mcc'][0]['code']}")
print(f"  Reasoning: {result['reasoning'][:150]}...")
print()

# Test 3: Novel business model
novel_case = {
    "context": {
        "business_name": "FractionalCFO.ai",
        "description": "AI-powered fractional CFO services for seed-stage SaaS companies",
        "website_content": "We combine financial expertise with AI automation to provide affordable CFO services to startups. Our platform offers strategic financial planning, investor relations support, and automated bookkeeping."
    }
}

response = requests.post(f"{LLM_URL}/classify", json=novel_case)
result = response.json()

print("Test 3: Novel Business Model")
print(f"  Industry: {result['primary_industry']}")
print(f"  Confidence: {result['confidence']}")
print(f"  Alternatives: {result['alternative_classifications']}")
print(f"  Reasoning: {result['reasoning'][:150]}...")
```

**Expected Output:**
```
Test 1: Simple Restaurant
  Industry: Restaurants and Eating Places
  Confidence: 0.92
  Reasoning: This is clearly a restaurant business based on the description of serving Italian food and wood-fired pizza. The core...

Test 2: Ambiguous Consulting
  Industry: Management Consulting Services
  Confidence: 0.85
  Codes: MCC=8742
  Reasoning: Based on the description and both Layer 1 and Layer 2 results pointing toward consulting services, this business is best...

Test 3: Novel Business Model
  Industry: Management Consulting Services
  Confidence: 0.82
  Alternatives: ["Accounting Services", "Computer Programming Services"]
  Reasoning: This represents a hybrid business model combining financial consulting (CFO services) with technology (AI automation). The primary...
```

---

### Task 3: Build Go LLM Classifier (Day 4-5)

**File:** `internal/classification/llm_classifier.go`

```go
package classification

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "time"
)

type LLMClassifier struct {
    llmServiceURL string
    httpClient    *http.Client
}

func NewLLMClassifier(llmServiceURL string) *LLMClassifier {
    return &LLMClassifier{
        llmServiceURL: llmServiceURL,
        httpClient: &http.Client{
            Timeout: 60 * time.Second, // LLM can be slow
        },
    }
}

type LLMClassificationResult struct {
    PrimaryIndustry          string
    Confidence               float64
    Reasoning                string
    MCC                      []CodeResult
    SIC                      []CodeResult
    NAICS                    []CodeResult
    AlternativeClassifications []string
    ProcessingTimeMs         int64
}

// Main classification method
func (l *LLMClassifier) ClassifyWithLLM(
    ctx context.Context,
    content *ScrapedContent,
    businessName string,
    description string,
    layer1Result *ClassificationResult,
    layer2Result *EmbeddingClassificationResult,
) (*LLMClassificationResult, error) {
    
    startTime := time.Now()
    
    slog.Info("Starting LLM-based classification",
        "business", businessName)
    
    // Prepare website content (truncate for context limits)
    websiteContent := l.prepareWebsiteContent(content)
    
    // Build request payload
    reqBody := map[string]interface{}{
        "context": map[string]interface{}{
            "business_name":     businessName,
            "description":       description,
            "website_content":   websiteContent,
        },
        "temperature": 0.1, // Low temperature for consistency
        "max_tokens":  800,
    }
    
    // Add Layer 1 context if available
    if layer1Result != nil {
        reqBody["context"].(map[string]interface{})["layer1_result"] = map[string]interface{}{
            "industry":   layer1Result.Industry,
            "confidence": layer1Result.Confidence,
            "keywords":   layer1Result.Keywords,
        }
    }
    
    // Add Layer 2 context if available
    if layer2Result != nil {
        reqBody["context"].(map[string]interface{})["layer2_result"] = map[string]interface{}{
            "top_match":      layer2Result.TopMatch,
            "confidence":     layer2Result.Confidence,
            "top_similarity": layer2Result.TopSimilarity,
        }
    }
    
    // Call LLM service
    response, err := l.callLLMService(ctx, reqBody)
    if err != nil {
        return nil, fmt.Errorf("LLM service call failed: %w", err)
    }
    
    // Parse response
    result := &LLMClassificationResult{
        PrimaryIndustry:            response["primary_industry"].(string),
        Confidence:                 response["confidence"].(float64),
        Reasoning:                  response["reasoning"].(string),
        AlternativeClassifications: l.parseAlternatives(response["alternative_classifications"]),
        ProcessingTimeMs:           time.Since(startTime).Milliseconds(),
    }
    
    // Parse codes
    codes := response["codes"].(map[string]interface{})
    result.MCC = l.parseCodes(codes["mcc"])
    result.SIC = l.parseCodes(codes["sic"])
    result.NAICS = l.parseCodes(codes["naics"])
    
    slog.Info("LLM classification complete",
        "industry", result.PrimaryIndustry,
        "confidence", result.Confidence,
        "duration_ms", result.ProcessingTimeMs)
    
    return result, nil
}

func (l *LLMClassifier) prepareWebsiteContent(content *ScrapedContent) string {
    parts := []string{}
    
    // Title
    if content.Title != "" {
        parts = append(parts, fmt.Sprintf("Title: %s", content.Title))
    }
    
    // Meta description
    if content.MetaDesc != "" {
        parts = append(parts, fmt.Sprintf("Description: %s", content.MetaDesc))
    }
    
    // About section (truncated)
    if content.AboutText != "" {
        about := content.AboutText
        if len(about) > 800 {
            about = about[:800] + "..."
        }
        parts = append(parts, fmt.Sprintf("About: %s", about))
    }
    
    // Top headings
    if len(content.Headings) > 0 {
        headings := content.Headings
        if len(headings) > 5 {
            headings = headings[:5]
        }
        parts = append(parts, fmt.Sprintf("Headings: %s", strings.Join(headings, ", ")))
    }
    
    combined := strings.Join(parts, "\n")
    
    // Truncate to 2000 chars total
    if len(combined) > 2000 {
        combined = combined[:2000] + "..."
    }
    
    return combined
}

func (l *LLMClassifier) callLLMService(
    ctx context.Context,
    reqBody map[string]interface{},
) (map[string]interface{}, error) {
    
    reqBodyJSON, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        l.llmServiceURL+"/classify",
        bytes.NewReader(reqBodyJSON),
    )
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    slog.Debug("Calling LLM service", "url", l.llmServiceURL)
    
    resp, err := l.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("LLM service returned status %d", resp.StatusCode)
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    slog.Debug("LLM service response received",
        "processing_time_ms", result["processing_time_ms"])
    
    return result, nil
}

func (l *LLMClassifier) parseCodes(codesInterface interface{}) []CodeResult {
    if codesInterface == nil {
        return []CodeResult{}
    }
    
    codesList, ok := codesInterface.([]interface{})
    if !ok {
        return []CodeResult{}
    }
    
    results := make([]CodeResult, 0, len(codesList))
    for _, item := range codesList {
        codeMap, ok := item.(map[string]interface{})
        if !ok {
            continue
        }
        
        results = append(results, CodeResult{
            Code:        codeMap["code"].(string),
            Description: codeMap["description"].(string),
            Confidence:  codeMap["confidence"].(float64),
            Source:      "llm_reasoning",
        })
    }
    
    return results
}

func (l *LLMClassifier) parseAlternatives(altInterface interface{}) []string {
    if altInterface == nil {
        return []string{}
    }
    
    altList, ok := altInterface.([]interface{})
    if !ok {
        return []string{}
    }
    
    results := make([]string, 0, len(altList))
    for _, item := range altList {
        if str, ok := item.(string); ok {
            results = append(results, str)
        }
    }
    
    return results
}
```

---

## Week 8: Integration & Optimization

### Task 4: Add Layer 3 Routing (Day 1-2)

**File:** `internal/classification/service.go`

```go
type IndustryDetectionService struct {
    // Existing fields...
    llmClassifier *LLMClassifier // NEW
}

func NewIndustryDetectionService(
    // ... existing params ...
    llmServiceURL string,
) *IndustryDetectionService {
    return &IndustryDetectionService{
        // ... existing initialization ...
        llmClassifier: NewLLMClassifier(llmServiceURL),
    }
}

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    startTime := time.Now()
    
    slog.Info("Starting industry detection with 3-layer orchestration",
        "business", businessName,
        "url", websiteURL)
    
    // Scrape website
    content, err := s.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, fmt.Errorf("scraping failed: %w", err)
    }
    
    // ========================================
    // LAYER 1: Multi-Strategy Classification
    // ========================================
    
    layer1Result, err := s.multiStrategyClassifier.ClassifyWithMultiStrategy(
        ctx,
        businessName,
        description,
        websiteURL,
    )
    if err != nil {
        return nil, fmt.Errorf("layer 1 failed: %w", err)
    }
    
    slog.Info("Layer 1 complete",
        "industry", layer1Result.Industry,
        "confidence", layer1Result.Confidence,
        "method", layer1Result.Method)
    
    // Decision Point 1: High confidence from Layer 1?
    if layer1Result.Confidence >= 0.90 {
        slog.Info("High confidence from Layer 1 (≥0.90), using result",
            "confidence", layer1Result.Confidence)
        
        result := s.buildResult(layer1Result, "layer1_high_conf")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Decision Point 2: Good confidence from Layer 1?
    if layer1Result.Confidence >= 0.85 {
        slog.Info("Good confidence from Layer 1 (≥0.85), using result",
            "confidence", layer1Result.Confidence)
        
        result := s.buildResult(layer1Result, "layer1_good_conf")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // ========================================
    // LAYER 2: Embedding Similarity Search
    // ========================================
    
    slog.Info("Layer 1 confidence below 0.85, trying Layer 2",
        "layer1_confidence", layer1Result.Confidence)
    
    layer2Result, err := s.embeddingClassifier.ClassifyByEmbedding(ctx, content)
    if err != nil {
        slog.Warn("Layer 2 failed, will try Layer 3",
            "error", err)
        layer2Result = nil // Continue to Layer 3
    } else {
        slog.Info("Layer 2 complete",
            "top_match", layer2Result.TopMatch,
            "confidence", layer2Result.Confidence,
            "similarity", layer2Result.TopSimilarity)
        
        // Decision Point 3: High confidence from Layer 2?
        if layer2Result.Confidence >= 0.88 {
            slog.Info("High confidence from Layer 2 (≥0.88), using result",
                "confidence", layer2Result.Confidence)
            
            result := s.buildResultFromEmbedding(layer2Result, "layer2_high_conf")
            result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
            return result, nil
        }
        
        // Decision Point 4: Layer 2 better than Layer 1?
        if layer2Result.Confidence > layer1Result.Confidence + 0.05 {
            slog.Info("Layer 2 meaningfully better than Layer 1, using Layer 2",
                "layer2_conf", layer2Result.Confidence,
                "layer1_conf", layer1Result.Confidence,
                "improvement", layer2Result.Confidence - layer1Result.Confidence)
            
            result := s.buildResultFromEmbedding(layer2Result, "layer2_better")
            result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
            return result, nil
        }
    }
    
    // ========================================
    // LAYER 3: LLM Reasoning
    // ========================================
    
    slog.Info("Layer 2 inconclusive, trying Layer 3 (LLM)",
        "layer1_confidence", layer1Result.Confidence,
        "layer2_confidence", func() float64 {
            if layer2Result != nil {
                return layer2Result.Confidence
            }
            return 0.0
        }())
    
    layer3Result, err := s.llmClassifier.ClassifyWithLLM(
        ctx,
        content,
        businessName,
        description,
        layer1Result,
        layer2Result,
    )
    if err != nil {
        slog.Warn("Layer 3 failed, falling back to best available result",
            "error", err)
        
        // Fallback: Use best of Layer 1 vs Layer 2
        if layer2Result != nil && layer2Result.Confidence > layer1Result.Confidence {
            result := s.buildResultFromEmbedding(layer2Result, "layer2_fallback")
            result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
            return result, nil
        }
        
        result := s.buildResult(layer1Result, "layer1_fallback")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    slog.Info("Layer 3 complete",
        "industry", layer3Result.PrimaryIndustry,
        "confidence", layer3Result.Confidence,
        "duration_ms", layer3Result.ProcessingTimeMs)
    
    // Decision Point 5: Use Layer 3 result?
    
    // If Layer 3 confidence is high, use it
    if layer3Result.Confidence >= 0.85 {
        slog.Info("Layer 3 confidence high (≥0.85), using result",
            "confidence", layer3Result.Confidence)
        
        result := s.buildResultFromLLM(layer3Result, "layer3_high_conf")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // If Layer 3 is better than both Layer 1 and Layer 2, use it
    bestPreviousConf := layer1Result.Confidence
    if layer2Result != nil && layer2Result.Confidence > bestPreviousConf {
        bestPreviousConf = layer2Result.Confidence
    }
    
    if layer3Result.Confidence > bestPreviousConf + 0.03 {
        slog.Info("Layer 3 better than previous layers, using result",
            "layer3_conf", layer3Result.Confidence,
            "best_previous", bestPreviousConf)
        
        result := s.buildResultFromLLM(layer3Result, "layer3_better")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Layer 3 didn't add value, use best previous result
    slog.Info("Layer 3 didn't improve, using best previous result")
    
    if layer2Result != nil && layer2Result.Confidence > layer1Result.Confidence {
        result := s.buildResultFromEmbedding(layer2Result, "layer2_final")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    result := s.buildResult(layer1Result, "layer1_final")
    result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    return result, nil
}

func (s *IndustryDetectionService) buildResultFromLLM(
    llmResult *LLMClassificationResult,
    layer string,
) *IndustryDetectionResult {
    
    return &IndustryDetectionResult{
        Classification: ClassificationData{
            PrimaryIndustry: llmResult.PrimaryIndustry,
            Confidence:      llmResult.Confidence,
            Method:          "llm_reasoning",
            Keywords:        []string{}, // LLM doesn't use keywords
        },
        Codes: ClassificationCodes{
            MCC:   llmResult.MCC,
            SIC:   llmResult.SIC,
            NAICS: llmResult.NAICS,
        },
        Explanation: ClassificationExplanation{
            PrimaryReason:     llmResult.Reasoning,
            SupportingFactors: []string{
                "Advanced LLM reasoning with context understanding",
                "Considers business model complexity and nuance",
                "Provides detailed rationale for classification",
            },
            KeyTermsFound:     []string{},
            MethodUsed:        "llm_reasoning",
            ProcessingPath:    layer,
            AlternativeClassifications: llmResult.AlternativeClassifications,
        },
        ProcessingTimeMs: llmResult.ProcessingTimeMs,
    }
}
```

**Routing Decision Tree:**

```
┌─────────────────────┐
│  Start Detection    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│    Layer 1:         │
│  Multi-Strategy     │
└──────────┬──────────┘
           │
           ▼
      Confidence ≥ 0.90? ──YES──> ✅ Use Layer 1 (60-70% of cases)
           │
           NO
           │
           ▼
      Confidence ≥ 0.85? ──YES──> ✅ Use Layer 1 (10-15% of cases)
           │
           NO
           │
           ▼
┌─────────────────────┐
│    Layer 2:         │
│  Embedding Search   │
└──────────┬──────────┘
           │
           ▼
      Confidence ≥ 0.88? ──YES──> ✅ Use Layer 2 (5-10% of cases)
           │
           NO
           │
           ▼
   Better than Layer 1? ──YES──> ✅ Use Layer 2 (5-10% of cases)
           │
           NO
           │
           ▼
┌─────────────────────┐
│    Layer 3:         │
│   LLM Reasoning     │
└──────────┬──────────┘
           │
           ▼
      Confidence ≥ 0.85? ──YES──> ✅ Use Layer 3 (3-8% of cases)
           │
           NO
           │
           ▼
   Better than L1/L2?  ──YES──> ✅ Use Layer 3 (2-5% of cases)
           │
           NO
           │
           ▼
      ✅ Use best of Layer 1/2 (2-5% of cases)
```

---

### Task 5: Testing & Validation (Day 3-5)

**Test Layer 3 directly:**

```bash
# Test Case 1: Complex business model
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "FractionalCFO.ai",
    "description": "AI-powered fractional CFO services combining financial expertise with automation for seed-stage SaaS companies"
  }'

# Expected:
# - Layer 1 confidence: ~0.60-0.70 (uncertain)
# - Layer 2 confidence: ~0.75-0.80 (better but still uncertain)
# - Layer 3 triggers
# - Layer 3 provides detailed reasoning
# - Final confidence: 0.85-0.90
# - Method: "llm_reasoning"

# Test Case 2: Hybrid business
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "TechHealth Innovations",
    "description": "Healthcare technology company providing both software development and medical device manufacturing"
  }'

# Expected:
# - Ambiguous (software vs medical devices)
# - Layer 3 provides nuanced classification
# - May suggest alternative classifications
# - Reasoning explains the dual nature

# Test Case 3: Standard case (should not trigger Layer 3)
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "Pizza Restaurant",
    "website_url": "https://joespizza.com"
  }'

# Expected:
# - Layer 1 high confidence (0.92+)
# - Layer 2 NOT triggered
# - Layer 3 NOT triggered
# - Fast response (<200ms)
```

**Performance benchmarks:**

```
Layer 1 only (Fast path):     <100ms   (60-70% of requests)
Layer 1 only (Full):          200-500ms (10-15% of requests)
Layer 1 → Layer 2:            400-800ms (10-15% of requests)
Layer 1 → Layer 2 → Layer 3:  2-5s      (5-10% of requests)

Overall distribution:
- p50: <300ms
- p90: <1000ms
- p95: <3000ms
- p99: <5000ms
```

**Full test suite:**

```bash
# Run on complete test set
# Track:
# - Overall accuracy (target: 90-95%)
# - Layer distribution
# - Performance by layer
# - Complex cases improved

# Expected results:
# - Accuracy: 90-95%
# - Layer 1 only: 70-85% of cases
# - Layer 2: 10-20% of cases
# - Layer 3: 5-10% of cases
# - Complex/novel cases: 85-95% accuracy (up from 70-80%)
```

---

## Phase 4 Success Criteria

Before declaring Phase 4 complete, verify:

### Infrastructure
- [ ] ✅ LLM service deployed on Railway (8GB RAM)
- [ ] ✅ LLM responding with valid JSON (<5s)
- [ ] ✅ Structured prompts working
- [ ] ✅ Go LLM classifier implemented

### Integration
- [ ] ✅ 3-layer orchestration logic implemented
- [ ] ✅ Proper fallback handling
- [ ] ✅ Layer 3 triggers for 5-10% of requests
- [ ] ✅ Decision tree working as designed

### Performance
- [ ] ✅ Layer 1/2 performance not degraded
- [ ] ✅ Layer 3 latency <5s (p95)
- [ ] ✅ Overall p95 <3s
- [ ] ✅ No timeouts or failures

### Quality
- [ ] ✅ **Accuracy at 90-95% on test set**
- [ ] ✅ Complex cases handled well
- [ ] ✅ Novel business models classified correctly
- [ ] ✅ Detailed reasoning provided
- [ ] ✅ Alternative classifications useful

---

## Expected Results

**Before Phase 4:**
- Accuracy: 85-90%
- Struggles: Complex, hybrid, novel businesses
- Explanation: Limited
- Layers: 2

**After Phase 4:**
- Accuracy: ✅ 90-95%
- Handles: ✅ Complex, hybrid, novel cases
- Explanation: ✅ Detailed LLM reasoning
- Layers: ✅ 3-layer intelligent routing

**Layer Distribution:**
```
Layer 1 only:       70-85% of requests (fast, confident)
Layer 2 only:       10-20% of requests (semantic understanding)
Layer 3:            5-10% of requests (complex reasoning)

Accuracy by path:
- Layer 1 path:     92-96% accurate
- Layer 2 path:     88-93% accurate
- Layer 3 path:     87-92% accurate
- Overall:          90-95% accurate
```

---

## Cost Analysis

**Monthly Costs (Railway):**

```
Classification Service (Go):  $7
Playwright Scraper:           $5
Embedding Service (2GB):      $15
LLM Service (8GB, CPU):       $30
Database (Supabase):          $0-25

Total: ~$60-80/month

For 10,000 classifications:
Cost per classification: $0.006-0.008

vs. API-based LLM:
GPT-4 @ $0.03/request = $300/month
Savings: 80-85%
```

**With GPU (faster inference):**
```
LLM Service (8GB + GPU):      $120
Total: ~$150-170/month

But 5-10x faster Layer 3 (<1s instead of 2-5s)
```

---

## Troubleshooting

**Issue: LLM service out of memory**
- Check Railway: Service needs 8GB minimum
- Model uses ~7GB when loaded
- Solution: Increase to 8GB or use quantized model

**Issue: LLM taking >10s per request**
- Without GPU: Expected 2-5s
- If >10s: Check Railway logs for errors
- Solution: Consider GPU or use smaller model (3B)

**Issue: LLM returning invalid JSON**
- Check prompt format
- Verify system message includes JSON instructions
- Add JSON validation in parsing
- Log raw LLM response to debug

**Issue: Layer 3 never triggers**
- Check routing thresholds in service.go
- Verify Layer 1/2 confidence scores
- Add debug logging to decision points
- Should trigger when confidence <0.85

**Issue: Layer 3 confidence too low**
- Review prompt quality
- Check if context is being passed correctly
- Verify Layer 1/2 hints are included
- May need prompt tuning

---

## Optimization Tips

**Prompt Optimization:**
- Include Layer 1/2 results as hints
- Provide concrete examples in system message
- Use temperature=0.1 for consistency
- Request reasoning before conclusion

**Performance Optimization:**
- Cache LLM results (30-day cache)
- Use GPU for 5-10x speedup
- Consider smaller model (Qwen 2.5 3B)
- Aggressive Layer 1/2 thresholds (reduce Layer 3 usage)

**Cost Optimization:**
- Start without GPU, add if needed
- Use higher thresholds (Layer 3 for <5% of requests)
- Cache aggressively
- Monitor usage patterns

---

## Next Steps: Phase 5

Once Phase 4 is complete:
- **Phase 5 (Week 9):** UI integration, caching, monitoring
- **Expected:** Production-ready system
- **Final accuracy:** 90-95% validated

**Phase 5 Guide will be provided once Phase 4 is complete.**

---

You're building a sophisticated 3-layer AI system! Layer 3 is the most advanced piece - LLM reasoning for the hardest cases. 🚀

Ready to deploy the LLM service and push to 90-95% accuracy?
