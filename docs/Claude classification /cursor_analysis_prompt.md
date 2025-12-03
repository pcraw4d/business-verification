# Cursor Analysis Prompt: Classification System Evaluation

## Your Task

You are tasked with performing a comprehensive analysis of the existing industry classification implementation in this codebase. Your goal is to compare the current implementation against a proposed ideal-state hybrid classification system and identify opportunities to leverage existing infrastructure.

## Context: Proposed Ideal State

The proposed system is a **hybrid 3-layer classification approach**:

1. **Layer 1: Rule-Based Classification** (Fast path)
   - Uses keyword matching and trigram analysis
   - Queries Supabase tables: `keywords`, `trigram`, `codes`
   - Target: 90%+ confidence threshold
   - Performance: <100ms

2. **Layer 2: Embedding-Based Similarity** (Medium path)
   - Vector similarity search using pgvector in Supabase
   - Pre-computed code embeddings (384-dim, all-MiniLM-L6-v2)
   - Target: 85-90% confidence threshold
   - Performance: 200-500ms

3. **Layer 3: LLM Classification** (Slow path, high accuracy)
   - Open-source LLM (Qwen 7B or Mistral 7B) on Railway
   - Structured JSON output with reasoning
   - Handles complex/ambiguous cases
   - Performance: 2-5s

**Key Infrastructure:**
- Supabase: PostgreSQL with pgvector, existing tables (codes, keywords, trigram, crosswalks)
- Railway: API gateway, multiple microservices
- Services: Classification orchestrator, website scraper, embedding service, LLM service
- Caching: 30-day classification cache in Supabase

## Your Analysis Instructions

### Step 1: Code Discovery
First, explore the codebase to understand the current implementation:

1. Find all files related to classification/industry identification
2. Identify database models, schemas, and migrations
3. Locate API endpoints that handle classification
4. Find any existing ML/LLM integration code
5. Identify configuration files, environment variables
6. Look for any existing caching mechanisms

### Step 2: Architecture Analysis
Document the current architecture:

1. What classification approach is currently used? (pure LLM, rules, embeddings, hybrid?)
2. What services/components exist? (monolith vs microservices)
3. Which databases are used and for what purpose?
4. What external APIs or models are integrated?
5. How is website scraping handled?
6. What deployment platform is used?

### Step 3: Database Analysis
Examine the existing Supabase schema:

1. List all tables related to classification
2. Document the structure of: `codes`, `keywords`, `trigram`, `crosswalks`, and any others
3. Are embeddings already stored? If so, what dimension and model?
4. Is pgvector enabled?
5. Are there any existing RPC functions or database functions?
6. What indexes exist?

### Step 4: Gap Analysis
Compare current vs proposed implementation:

1. **What exists and can be reused?**
   - List components/code that align with the proposed architecture
   - Identify tables/schemas that match or can be extended
   - Note any existing services that fit the microservices pattern

2. **What exists but needs modification?**
   - Components that could be refactored to fit the 3-layer approach
   - Database schemas that need extensions (e.g., adding embedding columns)
   - API endpoints that need restructuring

3. **What's missing entirely?**
   - Components from the proposed system that don't exist
   - Required infrastructure not yet set up
   - Missing database tables or functions

### Step 5: Performance & Cost Analysis
Evaluate current performance:

1. What's the average classification time?
2. What are the cost drivers? (API calls, compute, storage)
3. Are there any performance bottlenecks?
4. Is caching implemented? How effective is it?

### Step 6: Opportunity Assessment
Identify strategic opportunities:

1. **Quick Wins**: What can be improved with minimal changes?
2. **High-Value Additions**: What proposed features would have the biggest impact?
3. **Technical Debt**: What existing issues should be addressed during refactoring?
4. **Data Assets**: What existing data could be leveraged for improvement? (e.g., historical classifications for training)

## Output Format

Please produce your analysis in the following markdown format:

---

# Classification System Analysis Report

**Date:** [Current Date]
**Analyzer:** Cursor AI
**Codebase Version:** [Git commit or version if available]

## Executive Summary

[2-3 paragraph overview of findings]

- Current approach: [brief description]
- Key gaps: [3-5 bullet points]
- Reusable components: [percentage or count]
- Recommended path forward: [high-level recommendation]

---

## 1. Current Implementation Overview

### 1.1 Architecture
```
[Create an ASCII diagram of current architecture]
```

**Description:**
[Detailed explanation of how classification currently works]

### 1.2 Technology Stack
- **Backend Framework:** [e.g., FastAPI, Flask, Express]
- **Database:** [e.g., Supabase PostgreSQL]
- **ML/AI:** [e.g., OpenAI API, Hugging Face models]
- **Deployment:** [e.g., Railway, Vercel]
- **Other Services:** [list any other key services]

### 1.3 Key Components
| Component | File Path | Purpose | Status |
|-----------|-----------|---------|--------|
| [Name] | `/path/to/file` | [Description] | ✅ Working / ⚠️ Needs Work / ❌ Broken |

### 1.4 API Endpoints
```
POST /api/classify
GET /api/classification/:id
...
```

**Request/Response Examples:**
```json
// Include actual examples from code
```

---

## 2. Database Schema Analysis

### 2.1 Current Tables

#### Table: `codes`
```sql
[Provide actual schema or describe columns]
```
**Usage:** [How this table is used]
**Alignment with Proposal:** ✅ Matches / ⚠️ Needs Extension / ❌ Different Structure

#### Table: `keywords`
```sql
[Provide actual schema]
```
**Usage:** [How this table is used]
**Alignment with Proposal:** [Assessment]

#### Table: `trigram`
```sql
[Provide actual schema]
```
**Usage:** [How this table is used]
**Alignment with Proposal:** [Assessment]

#### Table: `crosswalks`
```sql
[Provide actual schema]
```
**Usage:** [How this table is used]
**Alignment with Proposal:** [Assessment]

#### [List any other relevant tables]

### 2.2 Missing Tables/Features
- [ ] `code_embeddings` table with pgvector
- [ ] `classification_cache` table
- [ ] `classification_feedback` table
- [ ] [Other missing elements]

### 2.3 Database Functions & Indexes
**Existing:**
```sql
[List any existing RPC functions, triggers, indexes]
```

**Missing (from proposal):**
- `match_code_embeddings()` RPC function for vector search
- [Other missing functions]

---

## 3. Classification Logic Analysis

### 3.1 Current Classification Flow
```
[Step-by-step breakdown of how classification currently works]

1. Receive request with URL
2. [Next step]
3. [Next step]
...
```

### 3.2 Classification Methods in Use

#### Method 1: [e.g., Pure LLM]
**Files:** `[path/to/files]`
**How it works:** [Description]
**Performance:** [Speed, cost, accuracy if known]
**Pros:** [List strengths]
**Cons:** [List weaknesses]

#### Method 2: [if multiple methods exist]
[Same structure as above]

### 3.3 Website Scraping Implementation
**Current approach:** [Description]
**Libraries used:** [e.g., Playwright, BeautifulSoup]
**Content extraction:** [How content is extracted and processed]
**Alignment with proposal:** [Assessment]

### 3.4 Confidence Scoring
**How is confidence calculated?** [Explain current method]
**Is it accurate?** [Your assessment]
**Improvements needed:** [Suggestions]

### 3.5 Explainability/Auditability
**Current explanation generation:** [How explanations are created]
**Quality assessment:** [Good/Poor/Missing]
**Gap vs proposal:** [What's missing]

---

## 4. Gap Analysis: Current vs Proposed

### 4.1 Layer 1: Rule-Based Classification

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Keyword matching | [Exists/Partial/Missing] | Required | [Describe gap] |
| Trigram analysis | [Status] | Required | [Describe gap] |
| Fast path routing | [Status] | Required (90%+ confidence) | [Describe gap] |
| Performance target | [Current: Xms] | Target: <100ms | [Describe gap] |

**Reusable Code:**
```
[List files/functions that can be reused]
- /path/to/file.py:function_name() - [description]
```

**Required Changes:**
- [ ] [Specific change needed]
- [ ] [Another change]

### 4.2 Layer 2: Embedding-Based Similarity

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Code embeddings | [Status] | Required (384-dim) | [Describe gap] |
| Vector search | [Status] | Required (pgvector) | [Describe gap] |
| Embedding service | [Status] | Required | [Describe gap] |
| Performance target | [Current] | Target: 200-500ms | [Describe gap] |

**Reusable Code:**
[List what can be reused]

**Required Changes:**
- [ ] [Specific change needed]

### 4.3 Layer 3: LLM Classification

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| LLM integration | [Status] | Open-source on Railway | [Describe gap] |
| Structured output | [Status] | JSON with reasoning | [Describe gap] |
| Prompt engineering | [Status] | Optimized prompts | [Describe gap] |
| Performance target | [Current] | Target: 2-5s | [Describe gap] |

**Current LLM Setup:**
- Model: [e.g., GPT-4, Claude, Llama]
- Hosting: [API / Self-hosted]
- Cost per classification: [if known]
- Latency: [average response time]

**Reusable Code:**
[List what can be reused]

**Required Changes:**
- [ ] [Specific change needed]

### 4.4 Orchestration & Decision Logic

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Multi-layer routing | [Status] | Required | [Describe gap] |
| Confidence thresholds | [Status] | Layer-specific thresholds | [Describe gap] |
| Fallback logic | [Status] | Cascading layers | [Describe gap] |

**Required Changes:**
- [ ] [Specific change needed]

### 4.5 Caching & Performance

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Classification cache | [Status] | 30-day cache | [Describe gap] |
| Content hashing | [Status] | SHA-256 hashing | [Describe gap] |
| Cache invalidation | [Status] | TTL-based | [Describe gap] |

**Current Caching:**
[Describe if/how caching is implemented]

**Required Changes:**
- [ ] [Specific change needed]

### 4.6 Services & Deployment

| Service | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Classification API | [Status] | FastAPI on Railway | [Describe gap] |
| Scraper Service | [Status] | Playwright on Railway | [Describe gap] |
| Embedding Service | [Status] | Required on Railway | [Describe gap] |
| LLM Service | [Status] | Qwen/Mistral on Railway | [Describe gap] |

**Current Deployment:**
- Platform: [e.g., Railway, Vercel]
- Architecture: [Monolith / Microservices]
- Services count: [number]

**Required Changes:**
- [ ] [Specific change needed]

---

## 5. Data Assets & Opportunities

### 5.1 Existing Data That Can Be Leveraged

#### Historical Classifications
- **Exists:** [Yes/No]
- **Volume:** [Number of records]
- **Quality:** [Assessment]
- **Opportunity:** [How this can be used - e.g., training data, validation set]

#### Keyword/Trigram Data
- **Quality:** [Assessment of current keywords/trigrams]
- **Coverage:** [How complete is the data]
- **Opportunity:** [How to improve or leverage]

#### Code Mappings
- **Completeness:** [Assessment of crosswalks table]
- **Opportunity:** [How to enhance]

### 5.2 Quick Wins

List specific improvements that could be implemented quickly:

1. **[Quick Win #1]**
   - **Effort:** [Hours/Days]
   - **Impact:** [High/Medium/Low]
   - **Description:** [What to do]
   - **Files to modify:** [Specific paths]

2. **[Quick Win #2]**
   [Same structure]

### 5.3 High-Value Additions

List proposed features with the biggest impact:

1. **[Addition #1]**
   - **Effort:** [Weeks]
   - **Impact:** [High/Medium/Low]
   - **Value:** [Why this matters]
   - **Dependencies:** [What's needed first]

---

## 6. Performance & Cost Assessment

### 6.1 Current Performance Metrics
- **Average classification time:** [Xms/s]
- **95th percentile:** [Xms/s]
- **Cache hit rate:** [X%]
- **Error rate:** [X%]

### 6.2 Current Cost Structure
- **Per classification:** [estimated cost]
- **Monthly volume:** [estimated]
- **Monthly cost:** [estimated]
- **Main cost drivers:** [e.g., LLM API calls, compute]

### 6.3 Projected Performance (with proposed changes)
- **Expected average time:** [estimate based on 3-layer approach]
- **Expected cost reduction:** [percentage]
- **Expected accuracy improvement:** [if measurable]

---

## 7. Technical Debt & Issues

### 7.1 Current Issues
List any bugs, inefficiencies, or problems in current implementation:

1. **[Issue #1]**
   - **Severity:** [High/Medium/Low]
   - **Impact:** [Description]
   - **Should fix during refactor:** [Yes/No/Maybe]

### 7.2 Code Quality Observations
- **Test coverage:** [percentage if available]
- **Documentation:** [Good/Fair/Poor]
- **Code organization:** [Assessment]
- **Areas needing refactor:** [List specific areas]

---

## 8. Migration Path Recommendation

### 8.1 Recommended Approach

**Option A: Incremental Enhancement** (Recommended if current system is working)
- Keep existing implementation running
- Add new layers one at a time
- Gradually shift traffic to new system
- Timeline: [X weeks]

**Option B: Full Rebuild** (Recommended if current system has major issues)
- Build new system in parallel
- Migrate data and historical records
- Cut over all at once
- Timeline: [X weeks]

**Option C: Hybrid** (Recommended if [specific condition])
- [Description]
- Timeline: [X weeks]

### 8.2 Implementation Phases

**Phase 1: Foundation (Week 1-2)**
- [ ] Task 1
- [ ] Task 2

**Phase 2: Layer 1 (Week 3-4)**
- [ ] Task 1
- [ ] Task 2

**Phase 3: Layer 2 (Week 5-6)**
- [ ] Task 1
- [ ] Task 2

**Phase 4: Layer 3 (Week 7-8)**
- [ ] Task 1
- [ ] Task 2

**Phase 5: Testing & Optimization (Week 9-10)**
- [ ] Task 1
- [ ] Task 2

### 8.3 Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [Risk 1] | High/Med/Low | High/Med/Low | [How to mitigate] |

---

## 9. Code Reusability Matrix

List all significant code files and assess reusability:

| File Path | Current Purpose | Reusable? | For Which Layer? | Modifications Needed |
|-----------|-----------------|-----------|------------------|---------------------|
| `/path/to/file.py` | [Purpose] | ✅ Yes / ⚠️ Partial / ❌ No | Layer 1/2/3/Orchestrator | [Brief description] |

---

## 10. Recommendations Summary

### 10.1 Top Priorities
1. **[Priority #1]** - [Why this first]
2. **[Priority #2]** - [Why this second]
3. **[Priority #3]** - [Why this third]

### 10.2 Technologies to Add
- [ ] pgvector extension in Supabase
- [ ] sentence-transformers library
- [ ] [Other tech]

### 10.3 Code to Preserve
- **Definitely keep:** [List critical code]
- **Refactor and keep:** [List code that needs work but has value]
- **Consider deprecating:** [List code that might not fit new architecture]

### 10.4 Estimated Effort
- **Total development time:** [X weeks]
- **Team size needed:** [X developers]
- **Complexity level:** [Low/Medium/High]

---

## 11. Appendices

### Appendix A: Full File Tree
```
[Provide relevant portions of the file tree]
```

### Appendix B: Environment Variables
```
[List all relevant env vars]
```

### Appendix C: Code Snippets
[Include any important code snippets that illustrate current implementation]

```python
# Example: Current classification function
[paste relevant code]
```

---

## Questions for Clarification

[List any questions you have about the codebase or requirements that need human input]

1. [Question 1]
2. [Question 2]

---

**End of Analysis Report**

