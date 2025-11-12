# Next Steps After Setting DATABASE_URL

You've successfully updated `DATABASE_URL` in `railway.env`. Here's what to do next:

## 1. Load the Environment Variables

Before running the server, load the environment variables:

```bash
source railway.env
```

Or if you prefer to export them:
```bash
export $(cat railway.env | grep -v '^#' | xargs)
```

## 2. Verify DATABASE_URL is Loaded

Check that the variable is set:
```bash
echo $DATABASE_URL
```

You should see your connection string (password will be visible, so be careful in shared terminals).

## 3. Test the Database Connection

Test that the connection works:
```bash
# Using psql
psql $DATABASE_URL -c "SELECT version();"

# Or using the migration script
./scripts/run-migration-010.sh
```

## 4. Run the Database Migration (If Not Already Done)

If you haven't run the migration yet:
```bash
source railway.env
./scripts/run-migration-010.sh
```

## 5. Start the Server

Start your server with the environment loaded:
```bash
source railway.env
go run cmd/railway-server/main.go
```

You should see:
```
✅ Database connection established for new API routes
✅ New API routes registered:
   - GET /api/v1/merchants/{merchantId}/analytics
   - GET /api/v1/merchants/{merchantId}/website-analysis
   - POST /api/v1/risk/assess
   - GET /api/v1/risk/assess/{assessmentId}
```

## 6. Test the Endpoints

Use the Postman or Insomnia collections:
- `tests/api/merchant-details/postman-collection.json`
- `tests/api/merchant-details/insomnia-collection.json`

Or test with curl:
```bash
# Get merchant analytics
curl -X GET "http://localhost:8080/api/v1/merchants/{merchantId}/analytics" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Start risk assessment
curl -X POST "http://localhost:8080/api/v1/risk/assess" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"merchantId": "merchant-123", "options": {"includeHistory": true}}'
```

## Troubleshooting

### "DATABASE_URL not set"
- Make sure you've run `source railway.env`
- Check that the variable exists: `grep DATABASE_URL railway.env`

### "Database connection failed"
- Verify your password is correct in the connection string
- Check that your IP is allowed in Supabase (if using IP restrictions)
- Ensure the connection string format is correct

### "Routes not registered"
- Check server logs for database connection messages
- Verify `DATABASE_URL` is set when the server starts
- Look for "⚠️ Skipping new API route registration" warnings

## For Railway Deployment

If deploying to Railway:
1. Copy the `DATABASE_URL` value from `railway.env`
2. Add it as an environment variable in Railway dashboard
3. Or use Railway CLI:
   ```bash
   railway variables set DATABASE_URL="postgresql://..."
   ```

