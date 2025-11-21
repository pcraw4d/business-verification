# Manual Testing Quick Start Guide

## ✅ Dev Server Status
**The dev server is now running!**

- **Frontend URL**: http://localhost:3000
- **Backend API**: http://localhost:8080 (default)

## Quick Access

1. **Open your browser** and navigate to:
   ```
   http://localhost:3000
   ```

2. **To test merchant details**, navigate to:
   ```
   http://localhost:3000/merchant-portfolio/[merchant-id]
   ```
   Replace `[merchant-id]` with an actual merchant ID from your database.

3. **Open DevTools Console** (F12 or Cmd+Option+I) to see:
   - Development logs with `[ComponentName]` prefixes
   - Error codes in error messages
   - API request/response logging

## Environment Setup

### Current Configuration
- ✅ Dev server running on port 3000
- ⚠️ API base URL defaults to `http://localhost:8080` if not set

### If Backend API is on Different Port

Create `frontend/.env.local` file:
```bash
cd frontend
echo "NEXT_PUBLIC_API_BASE_URL=http://localhost:8080" > .env.local
```

Or if your backend is on a different port:
```bash
echo "NEXT_PUBLIC_API_BASE_URL=http://localhost:YOUR_PORT" > .env.local
```

Then restart the dev server:
```bash
# Stop current server (Ctrl+C)
npm run dev
```

## Manual Testing Checklist

Use the checklist at: `.cursor/plans/phase2-manual-test-checklist.md`

### Quick Test Steps

1. **Open Merchant Details Page**
   - Navigate to a merchant details page
   - Check browser console for development logs

2. **Test Error Codes**
   - Look for error messages with format: "Error PC-001: ..."
   - Verify error codes appear in error states

3. **Test CTA Buttons**
   - "Run Risk Assessment" button
   - "Refresh Data" button
   - "Start Risk Assessment" button
   - "Enrich Data" button

4. **Test Data Display**
   - Verify all available data fields are displayed
   - Check for data completeness indicators
   - Verify charts and visualizations render

## Troubleshooting

### Dev Server Not Starting
```bash
cd frontend
npm run dev
```

### Port Already in Use
```bash
# Kill process on port 3000
lsof -ti:3000 | xargs kill -9

# Or use different port
PORT=3001 npm run dev
```

### API Connection Issues
- Check if backend API is running on port 8080
- Verify `NEXT_PUBLIC_API_BASE_URL` in `.env.local`
- Check browser console for API errors

### Missing Environment Variables
The dev server will work with defaults, but for production builds you need:
- `NEXT_PUBLIC_API_BASE_URL` (defaults to http://localhost:8080)

## Next Steps

1. ✅ Dev server is running
2. ⏭️ Open http://localhost:3000 in your browser
3. ⏭️ Navigate to a merchant details page
4. ⏭️ Follow the manual testing checklist
5. ⏭️ Report any issues found

## Useful Commands

```bash
# Start dev server (from frontend directory)
cd frontend && npm run dev

# Check if server is running
curl http://localhost:3000

# View logs
# Check terminal where npm run dev is running

# Stop server
# Press Ctrl+C in the terminal running npm run dev
```


