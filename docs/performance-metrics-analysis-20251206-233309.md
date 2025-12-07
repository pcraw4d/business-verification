# Performance Metrics Analysis

**Date:** Sat Dec  6 23:33:12 EST 2025  
**Source:** Classification Service Logs

---

## Summary

### Key Metrics

- **Total ClassifyBusinessByContextualKeywords Calls:** 65
- **Context Deadline Violations:** 289
- **Average Classification Duration:** 36.39s
- **Average Enhanced Score Duration:** 0.00s

---

## Detailed Metrics

### ClassifyBusinessByContextualKeywords Entry Times

```
-3.728707631s"
-2.19178074s"
-1.075397s"
-10.020294667s"
-7.831802191s"
-6.667848243s"
-7.404268891s"
-7.483677676s"
-5.986621593s"
-5.978668831s"
-4.28788309s"
-7.148245934s"
-2.878937865s"
-10.36083204s"
-4.115245131s"
-4.126251246s"
-2.270672148s"
-2.369095875s"
-8.543044184s"
-9.196083163s"
```

### Duration Breakdowns

```

```

### Parallel Query Metrics

```

```

---

## Analysis

### Performance Improvements

1. **Classification Duration:** 
   - Target: < 10 seconds
   - Current: See metrics above

2. **Context Deadline Violations:**
   - Target: < 5%
   - Current: 289 violations found

3. **Enhanced Score Duration:**
   - Target: < 8 seconds
   - Current: See metrics above

---

## Recommendations

Based on the metrics above, review:
1. extractKeywords duration (may be the bottleneck)
2. Context deadline management
3. Cache hit rates
4. Parallel query effectiveness

