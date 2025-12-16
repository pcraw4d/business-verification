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

# Load model and tokenizer on startup (lazy loading)
tokenizer = None
model = None
model_loaded = False

def load_model():
    """Load model and tokenizer (called on first request if not already loaded)"""
    global tokenizer, model, model_loaded
    if model_loaded:
        return
    
    logger.info(f"Loading model: {MODEL_NAME}...")
    try:
        tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
        model = AutoModelForCausalLM.from_pretrained(
            MODEL_NAME,
            torch_dtype=torch.float16 if DEVICE == "cuda" else torch.float32,
            device_map="auto" if DEVICE == "cuda" else None,
        )
        if DEVICE == "cpu":
            model = model.to(DEVICE)
        
        model.eval()  # Set to evaluation mode
        model_loaded = True
        logger.info(f"Model loaded successfully on {DEVICE}!")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise

# Attempt to load model on startup, but don't block or crash if it fails
# This allows the service to start even if model download/loading has issues
try:
    load_model()
except Exception as e:
    logger.warning(f"Model could not be loaded during startup: {e}. It will be loaded on first request.")

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
        "status": "healthy" if model_loaded else "model_loading",
        "model": MODEL_NAME,
        "device": DEVICE,
        "model_loaded": model_loaded,
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

