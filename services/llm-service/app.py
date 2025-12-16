"""
LLM Service - Industry classification reasoning using Qwen 2.5 0.5B
Deployed on Railway as a microservice

Using 0.5B model for reliable operation within Railway's 8GB memory limit.
Memory footprint: ~1.5GB (vs 6GB for 3B, 14GB for 7B)
Inference time: 2-5 seconds on CPU
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional, Dict, Any
import logging
import time
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
import json
import gc  # For explicit garbage collection

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="LLM Classification Service",
    description="Industry classification reasoning using Qwen 2.5 0.5B",
    version="1.1.0"
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
# Using 0.5B model for reliable operation within Railway's 8GB memory limit
# Memory: ~1GB model + ~0.5GB KV cache + ~0.5GB overhead = ~2GB total
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
logger.info(f"Using device: {DEVICE}")

# Load model and tokenizer on startup (lazy loading)
tokenizer = None
model = None
model_loaded = False

def load_model():
    """Load model and tokenizer (called on first request if not already loaded)"""
    global tokenizer, model, model_loaded
    if model_loaded:
        return
    
    logger.info(f"Loading model: {MODEL_NAME} on {DEVICE}...")
    try:
        # Use float16 for memory efficiency
        # 0.5B model: ~1GB in fp16, ~2GB in fp32
        dtype = torch.float16
        logger.info(f"Using float16 for {DEVICE} inference")
        
        logger.info(f"Loading tokenizer...")
        tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
        
        logger.info(f"Loading model with dtype={dtype}...")
        model = AutoModelForCausalLM.from_pretrained(
            MODEL_NAME,
            torch_dtype=dtype,
            device_map="auto" if DEVICE == "cuda" else None,
            low_cpu_mem_usage=True,
        )
        
        if DEVICE == "cpu":
            model = model.to(DEVICE)
            # CPU optimizations for parallel inference
            if hasattr(torch, 'set_num_threads'):
                torch.set_num_threads(4)  # 0.5B model can use more threads
        
        model.eval()  # Set to evaluation mode
        
        # Clear any cached memory
        gc.collect()
        if hasattr(torch, 'cuda'):
            torch.cuda.empty_cache()
        
        model_loaded = True
        logger.info(f"✅ Model loaded successfully on {DEVICE} with dtype={dtype}!")
        logger.info(f"Memory footprint: ~1.5GB (plenty of headroom in Railway's 8GB)")
    except torch.cuda.OutOfMemoryError as e:
        logger.error(f"❌ CUDA OOM during model loading: {e}")
        model_loaded = False
        raise HTTPException(status_code=503, detail="Model loading failed: Out of memory (CUDA)")
    except MemoryError as e:
        logger.error(f"❌ System OOM during model loading: {e}")
        model_loaded = False
        raise HTTPException(status_code=503, detail="Model loading failed: Out of memory (System)")
    except Exception as e:
        logger.error(f"❌ Failed to load model: {e}", exc_info=True)
        model_loaded = False
        raise

# DO NOT load model on startup - wait for first request
# This prevents OOM crashes during service startup
logger.info("Service starting - model will be loaded on first request to avoid OOM")

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
    max_tokens: Optional[int] = 400  # Smaller model = smaller context window

class ClassificationResponse(BaseModel):
    primary_industry: str
    confidence: float
    reasoning: str
    codes: Dict[str, List[Dict[str, Any]]]
    alternative_classifications: List[str]
    processing_time_ms: int

# Health check endpoint
@app.get("/health")
async def health_check():
    """Health check endpoint for Railway"""
    # Always return 200 OK - Railway will check status field
    # This prevents Railway from restarting the service while model is loading
    status = "healthy" if model_loaded else "model_loading"
    
    # If model is not loaded, return 503 to indicate service not ready
    # But only if we've attempted to load it (to avoid startup issues)
    if not model_loaded:
        # Return 200 with "model_loading" status - Railway health check should be lenient
        # during startup period (start-period in Dockerfile is 180s)
        pass
    
    return {
        "status": status,
        "model": MODEL_NAME,
        "model_size": "0.5B",
        "device": DEVICE,
        "model_loaded": model_loaded,
        "memory_footprint_gb": 1.5,
        "service": "llm-service",
        "version": "1.1.0"
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
    # Ensure model is loaded
    if not model_loaded:
        try:
            load_model()
        except Exception as e:
            logger.error(f"Failed to load model on request: {e}")
            raise HTTPException(status_code=503, detail=f"Model loading failed: {str(e)}")
    
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
    
    # Format for Qwen (ensure tokenizer is loaded)
    if tokenizer is None:
        raise ValueError("Tokenizer not loaded")
    
    prompt = tokenizer.apply_chat_template(
        messages,
        tokenize=False,
        add_generation_prompt=True
    )
    
    return prompt

def generate_classification(
    prompt: str,
    temperature: float = 0.1,
    max_tokens: int = 400
) -> str:
    """Generate classification using LLM."""
    
    # Tokenize
    inputs = tokenizer(prompt, return_tensors="pt").to(DEVICE)
    
    # Store input length for later extraction (before we delete inputs)
    input_ids = inputs['input_ids'][0]
    prompt_decoded = tokenizer.decode(input_ids, skip_special_tokens=True)
    prompt_length = len(prompt_decoded)
    
    # Generate with inference mode (lower memory than no_grad on CPU)
    with torch.inference_mode():
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
    
    # Clear inputs from memory after generation
    del inputs
    del input_ids
    gc.collect()
    
    # Decode full response
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)
    
    # Extract just the assistant's response (after the prompt)
    response = response[prompt_length:].strip()
    
    return response

def normalize_code_entry(code_entry: Any) -> Dict[str, Any]:
    """Normalize a code entry to the expected format: {code, description, confidence}"""
    if isinstance(code_entry, dict):
        # Already a dict, ensure required fields
        return {
            "code": str(code_entry.get("code", "")),
            "description": str(code_entry.get("description", "")),
            "confidence": float(code_entry.get("confidence", 0.8))
        }
    elif isinstance(code_entry, str):
        # Just a code string
        return {
            "code": code_entry,
            "description": "",
            "confidence": 0.8
        }
    elif isinstance(code_entry, (int, float)):
        # Numeric code
        return {
            "code": str(int(code_entry)),
            "description": "",
            "confidence": 0.8
        }
    else:
        return {"code": "", "description": "", "confidence": 0.0}


def normalize_codes(codes: Any) -> List[Dict[str, Any]]:
    """Normalize codes field - handles various malformed outputs from smaller models."""
    if codes is None:
        return []
    
    if isinstance(codes, list):
        # Already a list, normalize each entry
        return [normalize_code_entry(entry) for entry in codes]
    
    elif isinstance(codes, str):
        # Single code as string (common with 0.5B model)
        return [normalize_code_entry(codes)]
    
    elif isinstance(codes, (int, float)):
        # Single numeric code
        return [normalize_code_entry(codes)]
    
    elif isinstance(codes, dict):
        # Single code as dict
        return [normalize_code_entry(codes)]
    
    else:
        return []


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
    
    # Normalize codes - handle various malformed outputs from smaller models
    # The 0.5B model sometimes returns strings instead of lists
    if not isinstance(parsed['codes'], dict):
        parsed['codes'] = {'mcc': [], 'sic': [], 'naics': []}
    
    # Normalize each code type
    parsed['codes']['mcc'] = normalize_codes(parsed['codes'].get('mcc'))
    parsed['codes']['sic'] = normalize_codes(parsed['codes'].get('sic'))
    parsed['codes']['naics'] = normalize_codes(parsed['codes'].get('naics'))
    
    # Ensure alternative_classifications is a list
    alt_class = parsed.get('alternative_classifications', [])
    if isinstance(alt_class, str):
        parsed['alternative_classifications'] = [alt_class]
    elif not isinstance(alt_class, list):
        parsed['alternative_classifications'] = []
    else:
        # Ensure all entries are strings
        parsed['alternative_classifications'] = [str(x) for x in alt_class]
    
    # Validate confidence is in range
    try:
        confidence = float(parsed['confidence'])
        if confidence < 0.0 or confidence > 1.0:
            logger.warning(f"Confidence out of range: {confidence}, clamping to [0, 1]")
            parsed['confidence'] = max(0.0, min(1.0, confidence))
    except (ValueError, TypeError):
        logger.warning(f"Invalid confidence value: {parsed['confidence']}, defaulting to 0.7")
        parsed['confidence'] = 0.7
    
    return parsed

# Info endpoint
@app.get("/info")
async def get_info():
    """Get information about the LLM service"""
    return {
        "model": MODEL_NAME,
        "device": DEVICE,
        "parameters": "0.5B",
        "memory_footprint_gb": 1.5,
        "description": "Industry classification using LLM reasoning (lightweight model)",
        "endpoints": {
            "/classify": "Classify business with reasoning",
            "/health": "Health check",
            "/info": "Service information"
        }
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)

