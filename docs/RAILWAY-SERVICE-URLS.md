# Railway Service URLs - Production

**Date**: January 2025  
**Status**: ✅ **OFFICIAL SERVICE URLs**

---

## Production Service URLs

### Core Services

| Service | URL | Status |
|---------|-----|--------|
| **API Gateway** | `https://api-gateway-service-production-21fd.up.railway.app` | ✅ Active |
| **Risk Assessment Service** | `https://risk-assessment-service-production.up.railway.app` | ⚠️ Check Status |
| **Classification Service** | `https://classification-service-production.up.railway.app` | ✅ Active |
| **Merchant Service** | `https://merchant-service-production.up.railway.app` | ✅ Active |
| **Frontend Service** | `https://frontend-service-production-b225.up.railway.app` | ✅ Active |
| **BI Service** | `https://bi-service-production.up.railway.app` | ✅ Active |
| **Pipeline Service** | `https://pipeline-service-production.up.railway.app` | ✅ Active |
| **Monitoring Service** | `https://monitoring-service-production.up.railway.app` | ✅ Active |
| **Service Discovery** | `https://service-discovery-production-d397.up.railway.app` | ✅ Active |

---

## Important Notes

⚠️ **DO NOT USE OLD URLs**:
- ❌ `kyb-api-gateway-production.up.railway.app` (OLD - DO NOT USE)
- ✅ `api-gateway-service-production-21fd.up.railway.app` (CORRECT)

---

## Usage in Code

### Frontend API Config
```javascript
// web/js/api-config.js
return 'https://api-gateway-service-production-21fd.up.railway.app';
```

### Backend Service Config
```go
// services/api-gateway/internal/config/config.go
RiskAssessmentURL: "https://risk-assessment-service-production.up.railway.app"
```

### Test Scripts
```bash
# scripts/test-risk-endpoints.sh
API_BASE_URL="https://api-gateway-service-production-21fd.up.railway.app"
```

---

## Testing Endpoints

### API Gateway
```bash
# Health
curl "https://api-gateway-service-production-21fd.up.railway.app/health"

# Benchmarks
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/benchmarks?mcc=5411"

# Predictions
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/predictions/test-merchant-123"
```

### Risk Assessment Service (Direct)
```bash
# Health
curl "https://risk-assessment-service-production.up.railway.app/health"

# Benchmarks
curl "https://risk-assessment-service-production.up.railway.app/api/v1/risk/benchmarks?mcc=5411"
```

---

## Last Updated

**Date**: January 2025  
**Verified**: ✅ All URLs confirmed

