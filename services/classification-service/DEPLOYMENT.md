# Classification Service Deployment Guide

**Version**: 3.0  
**Last Updated**: 2025-11-26  
**Status**: ✅ Production Ready

---

## Quick Start

### Railway Deployment

1. **Set Root Directory**: `services/classification-service`
2. **Set Environment Variables** (see below)
3. **Deploy**: Push to main branch or use Railway CLI

### Environment Variables

```bash
# Required
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
PORT=8081

# Multi-Strategy Classifier (v3.0+)
MULTI_STRATEGY_ENABLED=true
CONFIDENCE_CALIBRATION_ENABLED=true
CACHE_ENABLED=true
CACHE_TTL=5m
```

---

## Multi-Strategy Classifier Overview

The Classification Service v3.0+ uses a **multi-strategy classifier** that combines:

1. **Keyword-Based** (40% weight) - Enhanced keyword matching with synonym, stemming, fuzzy
2. **Entity-Based** (25% weight) - Named Entity Recognition (NER)
3. **Topic-Based** (20% weight) - TF-IDF topic modeling
4. **Co-Occurrence-Based** (15% weight) - Pattern matching

### Performance

- ✅ **Average Response Time**: ~1.4s (72% faster than 5s target)
- ✅ **Max Response Time**: ~4.8s (under 5s target)
- ✅ **Success Rate**: 100%
- ✅ **Accuracy**: 95% target with confidence calibration

---

## API Endpoints

### POST /v1/classify

Classify a business using multi-strategy classifier.

**Request**:
```json
{
  "business_name": "Microsoft Corporation",
  "description": "Software development",
  "website_url": "https://microsoft.com"
}
```

**Response**:
```json
{
  "success": true,
  "business_name": "Microsoft Corporation",
  "confidence_score": 0.8073,
  "primary_industry": "Technology",
  "method": "multi_strategy",
  "keywords": ["tech", "platform", ...],
  "processing_time": "1.29s"
}
```

### GET /health

Health check endpoint.

---

## Database Requirements

### Required Tables

- `industries` - Industry definitions
- `industry_keywords` - Industry-specific keywords
- `classification_codes` - MCC, NAICS, SIC codes
- `code_keywords` - Keywords associated with codes
- `industry_patterns` - Co-occurrence patterns
- `keyword_weights` - Keyword relevance weights

### Data Population

Run these SQL scripts in Supabase:

1. `scripts/populate_all_classification_codes_comprehensive.sql`
2. `scripts/populate_code_keywords_comprehensive.sql`

---

## Configuration

See [Full Deployment Guide](../../docs/classification-service-deployment-guide.md) for:
- Complete environment variable reference
- Performance tuning recommendations
- Troubleshooting guide
- Monitoring setup

---

**Last Updated**: 2025-11-26  
**Version**: 3.0

