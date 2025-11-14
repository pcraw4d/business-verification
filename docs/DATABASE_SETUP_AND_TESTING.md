# Database Setup and Endpoint Testing Guide

## Current Status

### Docker Status
⚠️ **Docker is not currently running**. To start the database containers, run:
```bash
docker-compose -f docker-compose.test.yml up -d
```

### Server Status
✅ **Server is running** on port 8080
- Health endpoint: `http://localhost:8080/health` ✅ Working
- Server version: 4.0.0-CACHE-BUST-REBUILD

## Database Connection Behavior

Based on the recent bug fix, the server now handles database unavailability gracefully:

1. **When database ping fails**:
   - Connection is closed and set to `nil`
   - Repositories are not initialized
   - Services are not initialized
   - Routes still register with in-memory threshold fallback
   - Server logs: "⚠️ Using in-memory threshold storage (database unavailable)"

2. **When database is available**:
   - Repositories are initialized with database connection
   - Services are initialized with repositories
   - ThresholdManager loads thresholds from database
   - Server logs: "✅ Database connection established for new API routes"

## Expected Routes

According to `cmd/railway-server/main.go` (lines 1012-1023), the following routes should be registered:

### Merchant Analytics Routes
- `GET /api/v1/merchants/analytics` - Portfolio-level analytics
- `GET /api/v1/merchants/{merchantId}/analytics` - Merchant-specific analytics
- `GET /api/v1/merchants/{merchantId}/website-analysis` - Website analysis

### Risk Assessment Routes
- `POST /api/v1/risk/assess` - Start risk assessment
- `GET /api/v1/risk/assess/{assessmentId}` - Get assessment status

### Enhanced Risk Routes (Public)
- `GET /v1/risk/factors` - Get risk factors
- `GET /v1/risk/categories` - Get risk categories
- `GET /v1/risk/thresholds` - Get risk thresholds

### Enhanced Risk Routes (Admin)
- `POST /v1/admin/risk/thresholds` - Create threshold
- `GET /v1/admin/risk/system/health` - System health
- `GET /v1/admin/risk/system/metrics` - System metrics

## Testing Steps

### 1. Start Docker Containers
```bash
docker-compose -f docker-compose.test.yml up -d
```

### 2. Verify Containers are Running
```bash
docker ps | grep -E "(kyb-test-postgres|kyb-test-redis)"
```

### 3. Start Server with Database Connection
```bash
export DATABASE_URL="postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable"
export PORT=8080
go run ./cmd/railway-server/main.go
```

### 4. Test Endpoints

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Risk Thresholds (should work with in-memory fallback)
```bash
curl http://localhost:8080/v1/risk/thresholds
```

#### Risk Factors
```bash
curl http://localhost:8080/v1/risk/factors
```

#### Risk Categories
```bash
curl http://localhost:8080/v1/risk/categories
```

#### Create Threshold (Admin)
```bash
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Threshold",
    "category": "financial",
    "risk_levels": {
      "low": 25.0,
      "medium": 50.0,
      "high": 75.0
    }
  }'
```

#### Merchant Analytics
```bash
curl http://localhost:8080/api/v1/merchants/analytics
```

## Troubleshooting

### Routes Returning 404

If routes are returning 404, check:

1. **Server logs** - Look for route registration messages:
   ```
   ✅ New API routes registered:
      - GET /v1/risk/thresholds
      - GET /v1/risk/factors
      ...
   ```

2. **Database connection status** - Check logs for:
   - "✅ Database connection established" (database available)
   - "⚠️ Using in-memory threshold storage" (database unavailable, but routes should still work)

3. **Route conflicts** - Ensure no duplicate route registrations

### Database Connection Issues

If database connection fails:

1. **Check Docker is running**:
   ```bash
   docker ps
   ```

2. **Check container health**:
   ```bash
   docker-compose -f docker-compose.test.yml ps
   ```

3. **Check database URL**:
   ```bash
   echo $DATABASE_URL
   # Should be: postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable
   ```

4. **Test database connection**:
   ```bash
   psql "postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable" -c "SELECT 1;"
   ```

## Next Steps

1. **Start Docker** (if not already running)
2. **Restart server** with database connection
3. **Test endpoints** to verify:
   - Routes are registered correctly
   - Database persistence works (when available)
   - In-memory fallback works (when database unavailable)
   - Threshold CRUD operations work
   - Merchant analytics work

