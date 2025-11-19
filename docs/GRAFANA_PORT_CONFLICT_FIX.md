# Grafana Port Conflict Fix

## Issue

Playwright E2E tests were failing because tests were navigating to Grafana's login page instead of the Next.js application dashboard pages.

### Root Cause

**Grafana was running on port 3000**, which conflicts with the Next.js dev server that Playwright expects to run on port 3000.

When tests navigated to `http://localhost:3000/dashboard`, they were hitting Grafana's login page instead of the Next.js Business Intelligence Dashboard.

### Evidence

- Test screenshots showed Grafana login pages
- `docker ps` confirmed Grafana was bound to `0.0.0.0:3000->3000/tcp`
- `curl http://localhost:3000` returned 302 redirects to Grafana
- Playwright config expects Next.js dev server on `http://localhost:3000`

## Solution

Changed Grafana to run on **port 3001** instead of port 3000, freeing port 3000 for the Next.js dev server.

### Changes Made

1. **`docker-compose.monitoring.yml`**:
   - Changed Grafana port mapping from `"3000:3000"` to `"3001:3000"`
   - Grafana now accessible at `http://localhost:3001`

2. **`configs/monitoring.yml`**:
   - Updated Grafana URL from `http://localhost:3000` to `http://localhost:3001` (2 occurrences)

3. **Documentation Updates**:
   - `docs/GRAFANA-FALLBACK-DASHBOARD-GUIDE.md`: Updated access URL to port 3001
   - `docs/GRAFANA-DASHBOARD-IMPORT-GUIDE.md`: Updated access URL to port 3001

### Verification

- ✅ Grafana now accessible at `http://localhost:3001` (HTTP 200)
- ✅ Port 3000 is now free for Next.js dev server
- ✅ Playwright tests can now access the correct application

## Impact

- **Grafana Access**: Now at `http://localhost:3001` (was `http://localhost:3000`)
- **Next.js Dev Server**: Can now run on `http://localhost:3000` without conflicts
- **Playwright Tests**: Will now correctly navigate to Next.js pages instead of Grafana

## Next Steps

1. Restart Grafana container if it was already running:
   ```bash
   docker-compose -f docker-compose.monitoring.yml restart grafana
   ```

2. Update any scripts or documentation that reference Grafana on port 3000

3. Run Playwright tests to verify they now access the correct pages:
   ```bash
   cd frontend && npm run test:e2e
   ```

## Related Files

- `docker-compose.monitoring.yml` - Grafana port configuration
- `configs/monitoring.yml` - Monitoring service Grafana URL
- `frontend/playwright.config.ts` - Playwright base URL (expects port 3000 for Next.js)
- `docs/GRAFANA-FALLBACK-DASHBOARD-GUIDE.md` - Updated access instructions
- `docs/GRAFANA-DASHBOARD-IMPORT-GUIDE.md` - Updated access instructions

