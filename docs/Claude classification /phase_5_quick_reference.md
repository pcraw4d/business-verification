# Phase 5 Quick Reference
## Production Ready in 7 Days

**Status:** Phase 4 ✅ Complete | Phase 5 ⏳ In Progress

---

## 📋 7-Day Execution Plan

**Day 1:** Add 30-day classification cache  
**Day 2:** Build monitoring dashboard  
**Day 3:** UI integration (show explanations)  
**Day 4:** Performance optimization  
**Day 5-6:** Testing and validation  
**Day 7:** Production deployment

---

## 🎯 What Phase 5 Adds

| Component | What It Does | Why It Matters |
|-----------|--------------|----------------|
| **Classification cache** | 30-day result caching | 98% faster for repeats |
| **Monitoring dashboard** | Real-time metrics | Track performance/usage |
| **UI explanations** | Show reasoning & codes | Transparency & trust |
| **Performance tuning** | Rate limiting, pooling | Handle production load |

**The Magic:** Transform from working prototype → production system.

---

## 📁 Files You'll Create/Modify

```
supabase-migrations/
├── 060_add_classification_cache.sql      [NEW] Cache tables
└── 061_add_analytics_tables.sql          [NEW] Metrics tables

internal/classification/
├── cache.go                               [NEW] Cache logic
├── service.go                             [MODIFY] Integrate cache
└── repository/supabase_repository.go      [MODIFY] Cache queries

services/classification-service/
└── internal/handlers/dashboard.go         [NEW] Dashboard API

frontend/
└── templates/classification_result.html   [MODIFY] Show explanations
```

---

## ✅ Daily Checklist

### Day 1: Classification Cache

**Morning (2-3 hours):**
- [ ] Create migration `060_add_classification_cache.sql`
- [ ] Add tables: `classification_cache`
- [ ] Add functions: `get_cached_classification()`, `set_cached_classification()`
- [ ] Add indexes on `content_hash`, `expires_at`
- [ ] Run migration: `psql $DB_URL -f 060_add_classification_cache.sql`
- [ ] Verify: `SELECT COUNT(*) FROM classification_cache;` returns 0

**Afternoon (2-3 hours):**
- [ ] Create `internal/classification/cache.go`
- [ ] Implement `GenerateCacheKey()` (SHA-256 of content)
- [ ] Implement `Get()` (fetch from cache)
- [ ] Implement `Set()` (store in cache)
- [ ] Add repository methods in `supabase_repository.go`
- [ ] Integrate cache in `service.go` (check before classification)

**Testing:**
```bash
# First request (miss)
time curl -X POST localhost:8080/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'
# Expected: 2-3s, from_cache: false

# Second request (hit)
time curl -X POST localhost:8080/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'
# Expected: <100ms, from_cache: true ✅

# Check cache stats
psql $DB_URL -c "SELECT COUNT(*), AVG(access_count) FROM classification_cache;"
```

**Success Criteria:**
- [ ] Cache table created ✅
- [ ] First request: ~2-3s ✅
- [ ] Second request: <100ms ✅
- [ ] `from_cache: true` in response ✅

---

### Day 2: Monitoring Dashboard

**Morning (2-3 hours):**
- [ ] Create migration `061_add_analytics_tables.sql`
- [ ] Add table: `classification_metrics`
- [ ] Add materialized view: `classification_dashboard`
- [ ] Add function: `get_dashboard_summary()`
- [ ] Run migration
- [ ] Verify: `SELECT * FROM classification_dashboard;` works

**Afternoon (2-3 hours):**
- [ ] Add metrics logging in `service.go` (async, non-blocking)
- [ ] Create `handlers/dashboard.go`
- [ ] Add API routes: `/api/dashboard/summary`, `/api/dashboard/timeseries`
- [ ] Test dashboard API

**Testing:**
```bash
# Generate some test data
for i in {1..50}; do
  curl -s -X POST localhost:8080/v1/classify \
    -d "{\"business_name\": \"Test $i\"}" > /dev/null
done

# Check dashboard
curl localhost:8080/api/dashboard/summary

# Expected response:
{
  "metrics": [
    {"metric": "total_classifications", "value": 50},
    {"metric": "cache_hit_rate", "value": 45.2},
    {"metric": "avg_confidence", "value": 0.89},
    {"metric": "layer1_percentage", "value": 76.0},
    {"metric": "layer3_percentage", "value": 8.0}
  ]
}
```

**Success Criteria:**
- [ ] Metrics table created ✅
- [ ] All classifications logged ✅
- [ ] Dashboard API returns data ✅
- [ ] Materialized view working ✅

---

### Day 3: UI Integration

**Morning (2-3 hours):**
- [ ] Update frontend to show full explanation
- [ ] Display: primary reason, supporting factors, key terms
- [ ] Show processing path badge (layer1/2/3)
- [ ] Display cache status if from cache

**Afternoon (2-3 hours):**
- [ ] Show all industry codes (MCC, SIC, NAICS)
- [ ] Display code confidence scores
- [ ] Show alternative classifications
- [ ] Add styling for confidence levels (high/medium/low)

**Testing:**
```bash
# Classify and check UI
curl -X POST localhost:8080/v1/classify \
  -d '{"website_url": "https://example.com"}' | jq

# Verify response has all fields:
# - explanation.primary_reason
# - explanation.supporting_factors
# - explanation.key_terms_found
# - explanation.processing_path
# - codes.mcc (array of 3)
# - codes.sic (array of 3)
# - codes.naics (array of 3)
```

**UI Checklist:**
- [ ] Shows primary industry prominently ✅
- [ ] Displays confidence with color coding ✅
- [ ] Shows detailed explanation ✅
- [ ] Lists all codes with descriptions ✅
- [ ] Indicates cache status ✅
- [ ] Shows processing layer used ✅

---

### Day 4: Performance Optimization

**Tasks:**
- [ ] Add request timeout (30s) to HTTP server
- [ ] Add rate limiting (100 req/s)
- [ ] Optimize DB connection pool (25 max, 10 idle)
- [ ] Add HTTP client connection pooling
- [ ] Test with load testing tool

**Rate Limiting:**
```go
import "golang.org/x/time/rate"

rateLimiter := rate.NewLimiter(100, 200) // 100/sec, burst 200
router.Use(rateLimiterMiddleware(rateLimiter))
```

**Connection Pooling:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Testing:**
```bash
# Install hey (HTTP load testing tool)
go install github.com/rakyll/hey@latest

# Load test
hey -n 1000 -c 50 -m POST \
  -d '{"business_name":"Test"}' \
  http://localhost:8080/v1/classify

# Expected:
# - Success rate: >99%
# - p50: <500ms
# - p95: <3000ms
# - Rate limit errors: <1% (for 50 concurrent)
```

**Success Criteria:**
- [ ] Rate limiting working ✅
- [ ] No timeouts under load ✅
- [ ] p95 latency <3s ✅
- [ ] DB connections stable ✅

---

### Day 5-6: Testing & Validation

**Accuracy Testing:**
```python
# Run validation script (50-100 test cases)
python3 validate_accuracy.py

# Target results:
# Overall accuracy: ≥90%
# Layer 1 accuracy: ≥92%
# Layer 2 accuracy: ≥88%
# Layer 3 accuracy: ≥87%
```

**Performance Testing:**
```bash
# Cache performance
./performance_test.sh

# Expected:
# Cache hit rate: >60%
# Cache hit latency: <100ms
# Layer distribution: L1 70-85%, L2 10-20%, L3 5-10%
```

**Integration Testing:**
```bash
# Test full workflow
1. Clear cache: DELETE FROM classification_cache;
2. Classify business (cache miss)
3. Classify same business (cache hit)
4. Check dashboard metrics
5. Verify UI shows all fields
6. Test error handling (invalid URL)
7. Test rate limiting (>100 req/s)
```

**Checklist:**
- [ ] Accuracy ≥90% on test set ✅
- [ ] Cache hit rate >60% ✅
- [ ] All layers working correctly ✅
- [ ] Dashboard showing accurate metrics ✅
- [ ] UI displaying all information ✅
- [ ] Error handling tested ✅
- [ ] Performance targets met ✅

---

### Day 7: Production Deployment

**Pre-Deployment Checklist:**
- [ ] All tests passing ✅
- [ ] Accuracy validated ≥90% ✅
- [ ] Performance benchmarked ✅
- [ ] Security reviewed ✅
- [ ] Documentation complete ✅
- [ ] Backup plan ready ✅

**Deployment Steps:**
1. [ ] Tag release: `git tag v1.0.0`
2. [ ] Push to production branch: `git push origin main`
3. [ ] Verify Railway deployments (all services)
4. [ ] Run smoke tests on production
5. [ ] Monitor logs for 1 hour
6. [ ] Enable monitoring alerts
7. [ ] Update documentation with prod URLs

**Post-Deployment:**
```bash
# Smoke tests
curl https://api.auguste.ai/health
curl https://api.auguste.ai/api/dashboard/summary

# Test classification
curl -X POST https://api.auguste.ai/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'

# Monitor Railway metrics
# - Memory usage: <80% of limit
# - CPU usage: <70% average
# - Error rate: <0.1%
# - Response time p95: <3s
```

**Success Criteria:**
- [ ] All services healthy ✅
- [ ] API responding correctly ✅
- [ ] Dashboard accessible ✅
- [ ] Metrics being collected ✅
- [ ] No critical errors ✅
- [ ] **SYSTEM LIVE IN PRODUCTION** 🎉

---

## 🧪 Quick Testing Commands

**Test Cache:**
```bash
# Miss then hit
curl -X POST localhost:8080/v1/classify \
  -d '{"url":"https://test.com"}' | jq '.from_cache'
# false

curl -X POST localhost:8080/v1/classify \
  -d '{"url":"https://test.com"}' | jq '.from_cache'
# true
```

**Test Dashboard:**
```bash
# Summary
curl localhost:8080/api/dashboard/summary | jq

# Specific metric
curl localhost:8080/api/dashboard/summary | \
  jq '.metrics[] | select(.metric=="cache_hit_rate")'
```

**Test Accuracy:**
```bash
# Known business
curl -X POST localhost:8080/v1/classify \
  -d '{"url":"https://mcdonalds.com"}' | \
  jq '.classification.primary_industry'
# Expected: "Restaurants" or similar
```

**Check Logs:**
```bash
# Railway logs
railway logs -s classification-service

# Look for:
# "Cache hit" - Good!
# "Layer 1 high confidence" - Most common
# "Layer 3 triggered" - Should be 5-10%
```

---

## 📊 Success Metrics

### Before Phase 5 (After Phase 4)
```
Accuracy: 90-95% ✅
Cache: None ❌
Monitoring: None ❌
UI: Basic ❌
Production Ready: No ❌
```

### After Phase 5 (Complete)
```
Accuracy: 90-95% ✅
Cache: 30-day, 60-70% hit rate ✅
Monitoring: Dashboard + metrics ✅
UI: Full explanations + codes ✅
Production Ready: YES ✅
```

**Performance Comparison:**
```
                Before    After    Improvement
─────────────────────────────────────────────
Repeat request  2-3s      <100ms   30x faster
Monthly cost    $60       $60      Same
User experience Basic     Rich     Much better
Observability   None      Full     Measurable
```

---

## 💡 Pro Tips

**Tip 1: Cache Invalidation**

When to clear cache:
- Business changes website significantly
- Classification codes updated
- Model/logic improved

```sql
-- Clear specific cache entry
DELETE FROM classification_cache 
WHERE content_hash = 'abc123...';

-- Clear old entries (>30 days)
SELECT cleanup_expired_cache();

-- Clear everything (careful!)
TRUNCATE classification_cache;
```

**Tip 2: Dashboard Refresh**

Materialized view updates:
```sql
-- Manual refresh
SELECT refresh_classification_dashboard();

-- Or setup cron (refresh every hour)
-- In Supabase: Database > Extensions > pg_cron
```

**Tip 3: Monitoring Alerts**

Set up alerts for:
- Accuracy drops below 85%
- Cache hit rate drops below 50%
- Layer 3 usage exceeds 15%
- Error rate exceeds 1%
- p95 latency exceeds 5s

**Tip 4: Cost Monitoring**

Track Railway usage:
```bash
# Check current usage
railway status

# Expected monthly costs:
# Classification: $7
# Playwright: $5
# Embeddings: $15
# LLM: $30
# Total: ~$60/month
```

---

## ⚠️ Common Issues

**Issue: Cache not working**
```bash
# Check table exists
psql $DB_URL -c "\dt classification_cache"

# Check functions exist
psql $DB_URL -c "\df *cached*"

# Check logs
railway logs -s classification-service | grep -i cache

# Verify content hash generation
curl -X POST localhost:8080/v1/classify \
  -d '{"url":"https://test.com"}' -v | grep content_hash
```

**Issue: Dashboard showing no data**
```bash
# Check metrics table
psql $DB_URL -c "SELECT COUNT(*) FROM classification_metrics;"

# Check if logging is working
# Look for "Logging metrics" in service logs

# Manually refresh materialized view
psql $DB_URL -c "REFRESH MATERIALIZED VIEW classification_dashboard;"
```

**Issue: UI not showing explanations**
```bash
# Check API response has explanation field
curl -X POST localhost:8080/v1/classify \
  -d '{"url":"https://test.com"}' | jq '.explanation'

# Should have:
# - primary_reason
# - supporting_factors
# - key_terms_found
# - processing_path
```

**Issue: Performance degraded**
```bash
# Check cache hit rate
curl localhost:8080/api/dashboard/summary | \
  jq '.metrics[] | select(.metric=="cache_hit_rate")'

# Should be >60%
# If lower, check cache expiration settings

# Check layer distribution
# Layer 3 should be 5-10%, not 30%+
# If high, increase Layer 1/2 thresholds
```

---

## 📈 Progress Tracker

| Day | Task | Status | Notes |
|-----|------|--------|-------|
| 1 | Cache implemented | ⬜ | Hit rate target: >60% |
| 2 | Dashboard built | ⬜ | All metrics tracked? |
| 3 | UI updated | ⬜ | Shows explanations? |
| 4 | Performance optimized | ⬜ | Rate limiting working? |
| 5-6 | Testing complete | ⬜ | Accuracy ≥90%? |
| 7 | **PRODUCTION DEPLOY** | ⬜ | **LIVE?** 🎉 |

---

## 🎉 Completion Criteria

Before declaring Phase 5 complete:

### Infrastructure
- [ ] ✅ Cache table created and indexed
- [ ] ✅ Metrics table collecting data
- [ ] ✅ Dashboard API responding
- [ ] ✅ All migrations applied

### Functionality
- [ ] ✅ Cache working (>60% hit rate)
- [ ] ✅ Dashboard showing metrics
- [ ] ✅ UI displaying explanations
- [ ] ✅ Performance optimized

### Quality
- [ ] ✅ Accuracy ≥90% on test set
- [ ] ✅ Performance p95 <3s (uncached)
- [ ] ✅ Performance p95 <100ms (cached)
- [ ] ✅ Error rate <1%
- [ ] ✅ All tests passing

### Production
- [ ] ✅ Deployed to production
- [ ] ✅ Smoke tests passed
- [ ] ✅ Monitoring configured
- [ ] ✅ Documentation complete
- [ ] ✅ **SYSTEM LIVE** 🚀

---

## 🎯 Final System Stats

**Performance:**
```
Accuracy:           90-95%
Cache hit rate:     60-70%
Latency (cached):   <100ms
Latency (p50):      <500ms
Latency (p95):      <3000ms
Uptime:             99.9%
```

**Layer Distribution:**
```
Layer 1:            70-85% of requests
Layer 2:            10-20% of requests
Layer 3:            5-10% of requests
```

**Cost:**
```
Monthly:            $60-80
Per classification: $0.006-0.008
vs API (GPT-4):     80% cheaper
```

---

## 🚀 You're Done!

**What you've built:**
- ✅ 90-95% accurate classification system
- ✅ 3-layer AI orchestration (Multi-strategy + Embeddings + LLM)
- ✅ 30-day intelligent caching (60-70% hit rate)
- ✅ Real-time monitoring dashboard
- ✅ Beautiful UI with detailed explanations
- ✅ Production-ready, scalable infrastructure
- ✅ 80% cost savings vs API-based solutions

**Your system can:**
- Classify any business in 2-3s (or <100ms cached)
- Handle complex, ambiguous, novel business models
- Explain its reasoning with full transparency
- Scale to thousands of requests/day
- Cost effectively compete with expensive APIs

**Congratulations!** 🎉

You've transformed from 50-60% accuracy (Phase 1) to a production-ready 90-95% system in just 9 weeks!

---

## 📞 Quick Reference

**Most Used Commands:**

```bash
# Deploy
git push origin main

# Test classification
curl -X POST localhost:8080/v1/classify \
  -d '{"website_url": "https://example.com"}'

# Check dashboard
curl localhost:8080/api/dashboard/summary

# View logs
railway logs -s classification-service

# Check database
psql $DB_URL -c "SELECT * FROM classification_dashboard;"

# Clear cache
psql $DB_URL -c "TRUNCATE classification_cache;"
```

**Most Important Metrics:**
- Accuracy: ≥90%
- Cache hit rate: ≥60%
- Layer 3 usage: 5-10%
- p95 latency: <3s

**Go build something amazing!** 💪
