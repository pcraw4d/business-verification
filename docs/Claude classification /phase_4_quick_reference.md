# Phase 4 Quick Reference
## 85-90% → 90-95% Accuracy in 2 Weeks

**Status:** Phase 3 ✅ Complete | Phase 4 ⏳ In Progress

---

## 📋 10-Day Execution Plan

### Week 7: LLM Service (The Reasoning Engine)
**Day 1-2:** Deploy LLM service (Qwen 2.5 7B on Railway)  
**Day 3:** Build structured prompts  
**Day 4:** Test LLM service directly  
**Day 5:** Build Go LLM classifier

### Week 8: Integration (The Final Layer)
**Day 6-7:** Add Layer 3 routing logic  
**Day 8:** Integration testing  
**Day 9:** Prompt optimization & tuning  
**Day 10:** Full test suite validation

---

## 🎯 What Phase 4 Adds

| Component | What It Does | Why It Matters |
|-----------|--------------|----------------|
| **LLM service** | Qwen 2.5 7B for reasoning | Understands complex cases |
| **Structured prompts** | Guide LLM to classify | Consistent, accurate output |
| **Layer 3 routing** | Use LLM for complex cases | Handles what Layers 1-2 miss |
| **Detailed reasoning** | Explains classification | Transparency and trust |

**The Magic:** LLM can **reason** about complex, ambiguous, or novel business models.

```
Case: "Fractional CFO services for SaaS startups with AI automation"

Layer 1 (Keywords): "financial", "CFO" → Accounting (0.62 confidence)
Layer 2 (Embeddings): → Management Consulting (0.73 confidence)
Layer 3 (LLM): "This hybrid service combines financial consulting 
               with technology. Primary: Management Consulting. 
               Secondary: Accounting Services." (0.88 confidence)
```

---

## 📁 Files You'll Create

```
services/llm-service/
├── app.py                            [NEW] FastAPI LLM service
├── requirements.txt                  [NEW] transformers, torch
├── Dockerfile                        [NEW] Container config (8GB RAM!)
└── .dockerignore                     [NEW] Build exclusions

internal/classification/
├── llm_classifier.go                 [NEW] Layer 3 implementation
└── service.go                        [MODIFY] Add 3-layer routing
```

---

## ✅ Daily Checklist

### Day 1-2: Deploy LLM Service
- [ ] Create `services/llm-service/app.py`
- [ ] Add FastAPI app with `/classify` endpoint
- [ ] Build structured prompts (system + user messages)
- [ ] Add JSON response parsing
- [ ] Create Dockerfile (downloads Qwen 2.5 7B at build)
- [ ] Push to GitHub: `git push origin phase-4-llm`
- [ ] Deploy to Railway
- [ ] **CRITICAL:** Configure 8GB RAM (model needs ~7GB)
- [ ] Wait for build (~20 minutes, 14GB model download)
- [ ] Test: `curl https://llm-service.../health`
- [ ] Verify: Response in 2-5 seconds ✅

### Day 3: Build Structured Prompts
- [ ] Define system message (role + expectations)
- [ ] Build context formatter (name, description, website, Layer 1/2 hints)
- [ ] Add JSON format instructions
- [ ] Test with simple cases
- [ ] Test with complex cases
- [ ] Verify JSON output is valid
- [ ] Check reasoning quality ✅

### Day 4: Test LLM Service
- [ ] Test Case 1: Simple restaurant (should be fast, confident)
- [ ] Test Case 2: Ambiguous consulting (should provide reasoning)
- [ ] Test Case 3: Novel business model (should handle complexity)
- [ ] Measure latency (should be 2-5s on CPU, <1s on GPU)
- [ ] Validate JSON structure
- [ ] Check confidence calibration ✅

### Day 5: Build Go LLM Classifier
- [ ] Create `internal/classification/llm_classifier.go`
- [ ] Implement `ClassifyWithLLM()` method
- [ ] Add `prepareWebsiteContent()` (truncate to 2000 chars)
- [ ] Add `callLLMService()` (HTTP client with 60s timeout)
- [ ] Add JSON response parsing
- [ ] Handle errors gracefully
- [ ] Test directly (bypass routing) ✅

### Day 6-7: Add Layer 3 Routing
- [ ] Modify `internal/classification/service.go`
- [ ] Add LLM classifier to service struct
- [ ] Implement 3-layer decision tree:
  - Layer 1 confidence ≥0.90 → Use Layer 1
  - Layer 1 confidence ≥0.85 → Use Layer 1
  - Try Layer 2
  - Layer 2 confidence ≥0.88 → Use Layer 2
  - Layer 2 better than Layer 1 → Use Layer 2
  - Try Layer 3
  - Layer 3 confidence ≥0.85 → Use Layer 3
  - Layer 3 better than L1/L2 → Use Layer 3
  - Else → Use best of L1/L2
- [ ] Add `buildResultFromLLM()` method
- [ ] Add comprehensive logging
- [ ] Test routing logic ✅

### Day 8: Integration Testing
- [ ] Test standard case (should use Layer 1 only)
- [ ] Test edge case (should use Layer 2)
- [ ] Test complex case (should use Layer 3)
- [ ] Verify layer distribution: ~5-10% Layer 3
- [ ] Check end-to-end performance
- [ ] Validate fallback logic ✅

### Day 9: Prompt Optimization
- [ ] Review Layer 3 results on test set
- [ ] Identify cases where reasoning is weak
- [ ] Refine prompt with better examples
- [ ] Adjust temperature (0.05-0.2)
- [ ] Test confidence calibration
- [ ] Fine-tune decision thresholds ✅

### Day 10: Full Validation
- [ ] Run full test set (all test cases)
- [ ] Calculate accuracy: **Target 90-95%**
- [ ] Measure layer usage:
  - Layer 1: 70-85%
  - Layer 2: 10-20%
  - Layer 3: 5-10%
- [ ] Measure performance:
  - p50: <300ms
  - p90: <1000ms
  - p95: <3000ms
- [ ] **Accuracy ≥90%** ✅

---

## 🚀 Day 1 Quick Start

**Your immediate action:**

```bash
# 1. Create LLM service directory
mkdir -p services/llm-service
cd services/llm-service

# 2. Create app.py (see full code in Kick-Off Guide)
cat > app.py << 'EOF'
from fastapi import FastAPI, HTTPException
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

app = FastAPI()

# Load Qwen 2.5 7B
MODEL_NAME = "Qwen/Qwen2.5-7B-Instruct"
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, device_map="auto")

@app.get("/health")
async def health():
    return {"status": "healthy", "model": MODEL_NAME}

@app.post("/classify")
async def classify(request: dict):
    # Build prompt, generate response, parse JSON
    # (Full implementation in Kick-Off Guide)
    pass
EOF

# 3. Create requirements.txt
cat > requirements.txt << 'EOF'
fastapi==0.109.0
uvicorn==0.27.0
transformers==4.36.0
torch==2.1.0
accelerate==0.25.0
EOF

# 4. Create Dockerfile
cat > Dockerfile << 'EOF'
FROM nvidia/cuda:12.1.0-base-ubuntu22.04
# Install Python 3.11, dependencies, download model
# (Full Dockerfile in Kick-Off Guide)
EOF

# 5. Deploy to Railway
git add services/llm-service/
git commit -m "Add LLM service for Layer 3"
git push origin phase-4-llm

# In Railway:
# - Create service "llm-service"
# - Set RAM: 8GB (CRITICAL!)
# - Wait for build (~20 min)
# - Note URL
```

**Expected:** LLM service deployed, responding with JSON ✅

---

## 📊 Success Metrics

### Before Phase 4 (After Phase 3)
```
Accuracy: 85-90%
Layers: 2 (Multi-strategy + Embeddings)
Complex cases: 70-80% accurate
Novel models: Poor
Reasoning: Limited
Latency p95: ~900ms
```

### After Phase 4
```
Accuracy: ✅ 90-95%
Layers: ✅ 3 (Multi-strategy + Embeddings + LLM)
Complex cases: ✅ 85-92% accurate
Novel models: ✅ Good handling
Reasoning: ✅ Detailed explanations
Latency p95: ✅ <3s
```

**Layer 3 Impact:**
- Triggers: 5-10% of requests
- Improves: Most complex/ambiguous cases
- Adds latency: 2-5s (when used)
- Accuracy gain: +5-10 percentage points

---

## 🧪 Testing Commands

### Test LLM Service

**Health check:**
```bash
LLM_URL="https://llm-service-production.up.railway.app"

curl $LLM_URL/health
# Expected: {"status": "healthy", "model": "Qwen/Qwen2.5-7B-Instruct"}
```

**Simple classification:**
```bash
curl -X POST $LLM_URL/classify \
  -H "Content-Type: application/json" \
  -d '{
    "context": {
      "business_name": "Pizza Palace",
      "description": "Italian restaurant serving wood-fired pizza"
    }
  }'

# Expected response (2-5s):
# {
#   "primary_industry": "Restaurants and Eating Places",
#   "confidence": 0.92,
#   "reasoning": "This is clearly a restaurant...",
#   "codes": {
#     "mcc": [{"code": "5812", "description": "...", "confidence": 0.92}, ...],
#     "sic": [...],
#     "naics": [...]
#   },
#   "processing_time_ms": 2340
# }
```

**Complex case:**
```bash
curl -X POST $LLM_URL/classify \
  -d '{
    "context": {
      "business_name": "TechHealth AI",
      "description": "Healthcare software with integrated medical devices",
      "layer1_result": {
        "industry": "Computer Programming",
        "confidence": 0.65
      },
      "layer2_result": {
        "top_match": "Medical Equipment Manufacturing",
        "confidence": 0.72
      }
    }
  }'

# Expected: Nuanced classification considering both software and devices
```

### Test Classification

**Standard case (Layer 1 only):**
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Pizza Restaurant"}'

# Expected:
# - layer: "layer1_high_conf"
# - confidence: 0.92+
# - processing_time_ms: <200
```

**Complex case (Layer 3):**
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{
    "business_name": "Fractional CFO Services",
    "description": "AI-powered fractional CFO for SaaS startups"
  }'

# Expected:
# - layer: "layer3_high_conf"
# - confidence: 0.85-0.90
# - reasoning: Detailed explanation
# - processing_time_ms: 2000-5000
```

---

## 💡 Key Concepts

### What Is an LLM?

**Large Language Model** = Neural network trained on massive text to understand and generate language.

**Qwen 2.5 7B:**
- 7 billion parameters
- Instruction-tuned (follows prompts)
- Reasoning capabilities
- JSON output support

**Why Self-Hosted?**
- Cost: $0.006/request vs $0.03+ for API
- Privacy: Your data stays on your infrastructure
- Control: Customize prompts, no rate limits
- Latency: 2-5s vs API round-trip

### Layer 1 vs Layer 2 vs Layer 3

**Layer 1 (Keywords/Trigrams):**
- Speed: <500ms
- Best for: Common, clear cases
- Method: Pattern matching
- Accuracy: 92-96% (when confident)
- Handles: 70-85% of requests

**Layer 2 (Embeddings):**
- Speed: 400-800ms
- Best for: Semantic similarity
- Method: Vector search
- Accuracy: 88-93% (when triggered)
- Handles: 10-20% of requests

**Layer 3 (LLM):**
- Speed: 2-5s (CPU), <1s (GPU)
- Best for: Complex reasoning
- Method: LLM generation
- Accuracy: 87-92% (when triggered)
- Handles: 5-10% of requests

**Why 3 Layers?**
- Fast path for easy cases (Layer 1)
- Semantic understanding for medium cases (Layer 2)
- Deep reasoning for hard cases (Layer 3)
- Optimal cost vs accuracy vs latency

### The 3-Layer Decision Tree

```
Request
  ↓
Layer 1 (Multi-Strategy)
  ↓
  Confidence ≥ 90%? → YES → ✅ DONE (70% of cases)
  ↓ NO
  Confidence ≥ 85%? → YES → ✅ DONE (10% of cases)
  ↓ NO
Layer 2 (Embeddings)
  ↓
  Confidence ≥ 88%? → YES → ✅ DONE (5% of cases)
  ↓ NO
  Better than L1? → YES → ✅ DONE (5% of cases)
  ↓ NO
Layer 3 (LLM)
  ↓
  Confidence ≥ 85%? → YES → ✅ DONE (3% of cases)
  ↓ NO
  Better than L1/L2? → YES → ✅ DONE (2% of cases)
  ↓ NO
✅ Use best of Layer 1/2 (5% of cases)
```

---

## 🎓 Pro Tips

**Tip 1: Prompt Engineering Is Critical**

Bad prompt:
```
"Classify this business"
```

Good prompt:
```
System: You are an expert classifier...
Context: Business name, description, website, L1/L2 hints
Format: Respond with JSON {...}
Requirements: Provide reasoning, top 3 codes, alternatives
```

**Tip 2: Include Layer 1/2 Context**

LLM performs better when it knows what previous layers found:
```json
{
  "layer1_result": {
    "industry": "Management Consulting",
    "confidence": 0.68,
    "keywords": ["consulting", "advisory"]
  },
  "layer2_result": {
    "top_match": "Computer Programming Services",
    "confidence": 0.74
  }
}
```

This helps LLM reason: "Both layers suggest professional services, but disagree on tech vs consulting. Let me analyze the business description more carefully..."

**Tip 3: Temperature Matters**

```
Temperature = 0.0: Deterministic (same input → same output)
Temperature = 0.1: Mostly consistent (our default)
Temperature = 0.5: More creative
Temperature = 1.0: Very random
```

For classification: Use 0.1 (consistent but not completely rigid)

**Tip 4: Set Realistic Thresholds**

Layer 3 is expensive (2-5s). Don't trigger it unless needed:
```go
// Good thresholds:
Layer 1: Use if ≥ 0.90
Layer 2: Use if ≥ 0.88
Layer 3: Use if ≥ 0.85

// Too aggressive (overuses Layer 3):
Layer 1: Use if ≥ 0.95  // Too high, triggers L3 too often
```

**Tip 5: Monitor Layer Usage**

Track what % of requests use each layer:
```
Expected distribution:
- Layer 1: 70-85%
- Layer 2: 10-20%
- Layer 3: 5-10%

If Layer 3 > 15%: Thresholds too strict, costs too high
If Layer 3 < 3%: Thresholds too loose, missing value
```

---

## ⚠️ Common Issues

**Issue: LLM service out of memory**
```bash
# Check Railway logs
railway logs -s llm-service

# Common cause: Not enough RAM
# Solution: Increase to 8GB minimum
# Model needs ~7GB when loaded
```

**Issue: LLM returning invalid JSON**
```python
# Check raw response
response = generate_classification(prompt)
print("Raw response:", response)

# Look for:
# - Text before/after JSON
# - Incomplete JSON
# - Missing fields

# Solution: Improve prompt with clearer instructions
```

**Issue: LLM too slow (>10s)**
```bash
# Without GPU: 2-5s is normal
# With GPU: <1s expected

# If >10s on CPU:
# - Check Railway CPU allocation
# - Consider smaller model (Qwen 2.5 3B)
# - Or add GPU (5-10x speedup)
```

**Issue: Layer 3 never triggers**
```go
// Check logs for routing decisions
slog.Info("Routing decision",
    "layer1_conf", layer1Result.Confidence,
    "layer2_conf", layer2Result.Confidence,
    "threshold", 0.85)

// Verify thresholds:
if layer1Result.Confidence >= 0.90  // Should be 0.90, not higher
```

**Issue: Layer 3 confidence too low**
```
# Common causes:
1. Prompt not clear enough
2. Not enough context provided
3. Business genuinely ambiguous

# Solutions:
1. Add examples to system message
2. Include Layer 1/2 results as hints
3. Provide more website content
4. Accept that some businesses are ambiguous
```

**Issue: High Railway costs**
```
# Without GPU: ~$30/month for 8GB
# With GPU: ~$120/month

# To reduce:
1. Increase Layer 1/2 thresholds (trigger L3 less)
2. Cache Layer 3 results aggressively
3. Start without GPU, add if needed
4. Use smaller model (3B instead of 7B)
```

---

## 📈 Progress Tracker

| Day | Task | Status | Accuracy | Notes |
|-----|------|--------|----------|-------|
| 1-2 | LLM service deployed | ⬜ | - | Railway URL? |
| 3 | Prompts built | ⬜ | - | JSON valid? |
| 4 | LLM service tested | ⬜ | - | 2-5s latency? |
| 5 | Go classifier built | ⬜ | - | Parsing works? |
| 6-7 | Layer 3 routing | ⬜ | - | Triggers 5-10%? |
| 8 | Integration tested | ⬜ | - | All layers work? |
| 9 | Prompts optimized | ⬜ | - | Reasoning good? |
| 10 | **Full validation** | ⬜ | **?%** | **≥90%?** ⭐ |

---

## 💰 Cost Analysis

### Monthly Costs (Railway)

**Without GPU:**
```
Classification Service:  $7
Playwright:              $5
Embedding Service:       $15
LLM Service (8GB CPU):   $30
Supabase:                $0-25
──────────────────────────
Total:                   ~$60-80/month

Per classification:      $0.006-0.008
```

**With GPU (faster):**
```
Classification Service:  $7
Playwright:              $5
Embedding Service:       $15
LLM Service (8GB + GPU): $120
Supabase:                $0-25
──────────────────────────
Total:                   ~$150-170/month

Per classification:      $0.015-0.017
But 5-10x faster Layer 3 (<1s vs 2-5s)
```

**vs. API-Based:**
```
GPT-4 API @ $0.03:       $300/month (10k requests)
Claude API @ $0.025:     $250/month
Gemini @ $0.02:          $200/month

Savings: 70-85%
```

---

## 📞 Checkpoints

**Good times to validate:**
- After Day 2: "LLM service responding"
- After Day 4: "Prompts generating good JSON"
- After Day 7: "3-layer routing working"
- After Day 9: "Accuracy improving"
- After Day 10: "90-95% accuracy achieved"

**Red flags:**
- Day 2: LLM service not deploying (check RAM = 8GB)
- Day 4: Invalid JSON responses (fix prompts)
- Day 7: Layer 3 triggers >20% (thresholds too strict)
- Day 9: Reasoning quality poor (optimize prompts)
- Day 10: Accuracy <88% (troubleshoot Layer 3)

---

## 🎉 Phase 4 Complete!

**You'll know it's done when:**
- ✅ LLM service deployed and responding (<5s)
- ✅ 3-layer routing working correctly
- ✅ Layer 3 triggers for 5-10% of requests
- ✅ Complex cases handled with detailed reasoning
- ✅ **Accuracy at 90-95% on test set**

**Then you're ready for Phase 5:** Production polish (caching, monitoring, UI)

---

## 🔥 Build the Final Layer!

You've got 2 solid layers working (multi-strategy + embeddings). Now add the intelligence layer (LLM reasoning) and reach 90-95% accuracy!

**Start with Day 1:** Deploy LLM service to Railway. The complete code is in Phase 4 Kick-Off Guide.

**Expected time:** 2 weeks (10 working days)

Ready to add reasoning and hit 90-95%? 💪
