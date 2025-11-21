# Manual Testing - Dev Environment Setup

## Issue
The `npm run dev` command doesn't work from the root directory because the dev script is in `frontend/package.json`.

## Solution

### Option 1: Run from Frontend Directory (Recommended)

```bash
# Navigate to frontend directory
cd frontend

# Start the dev server
npm run dev
```

The server will start on `http://localhost:3000`

### Option 2: Run from Root with Full Path

```bash
# From root directory
cd frontend && npm run dev
```

## Environment Variables Check

Before starting the dev server, ensure you have the required environment variables set. Check if `.env.local` exists in the `frontend/` directory:

```bash
cd frontend
ls -la .env*
```

If missing, you may need to create `.env.local` with:
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

## Quick Start Commands

```bash
# 1. Navigate to frontend
cd /Users/petercrawford/New\ tool/frontend

# 2. Check environment variables
cat .env.local 2>/dev/null || echo "Create .env.local file"

# 3. Start dev server
npm run dev
```

## Verify Server is Running

Once started, you should see:
- Server running on `http://localhost:3000`
- Next.js compilation messages
- Ready message

Then navigate to:
- `http://localhost:3000/merchant-portfolio/[merchant-id]` for merchant details
- Open browser DevTools console to see development logs

## Troubleshooting

If you get errors about missing environment variables:
1. Check `frontend/scripts/verify-build-env.js` for required variables
2. Create `frontend/.env.local` with required values
3. Restart the dev server


