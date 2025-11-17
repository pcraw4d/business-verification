# Database Setup Guide - Quick Start

## What You Need

Based on your `railway.env` file, I can see you have:
- **Supabase Project**: `qpqhuqqmkjxsltzshfam.supabase.co`
- **Database Host**: `db.qpqhuqqmkjxsltzshfam.supabase.co`

## Step 1: Get Your Database Connection String

### Option A: From Supabase Dashboard (Recommended)

1. Go to: https://app.supabase.com/project/qpqhuqqmkjxsltzshfam
2. Navigate to: **Settings** → **Database**
3. Scroll to **Connection string** section
4. Select **URI** format
5. Copy the connection string

It will look like:
```
postgresql://postgres:[YOUR-PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

**Important**: Replace `[YOUR-PASSWORD]` with your actual database password.

### Option B: Construct from Known Values

If you know your database password, you can construct it:
```
postgresql://postgres:YOUR_PASSWORD@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

## Step 2: Run the Setup Script

I've created an interactive setup script that will:
- ✅ Guide you through configuration
- ✅ Test the connection
- ✅ Run the migration
- ✅ Verify the schema
- ✅ Save to .env file

**Run it now:**
```bash
./scripts/setup_database.sh
```

The script will:
1. Ask for your DATABASE_URL
2. Test the connection
3. Check if the table exists
4. Run the migration if needed
5. Verify everything works
6. Optionally save to .env file

## Step 3: Manual Setup (Alternative)

If you prefer to set it up manually:

```bash
# 1. Set DATABASE_URL
export DATABASE_URL="postgresql://postgres:YOUR_PASSWORD@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres"

# 2. Test connection
./scripts/test_database_connection.sh

# 3. Run migration
psql $DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql

# 4. Verify schema
./scripts/verify_database_schema.sh

# 5. Save to .env (optional)
echo "DATABASE_URL=$DATABASE_URL" >> .env
```

## Step 4: Verify Everything Works

After setup:

```bash
# 1. Validate configuration
./scripts/validate_config.sh

# 2. Test connection
./scripts/test_database_connection.sh

# 3. Verify schema
./scripts/verify_database_schema.sh

# 4. Start server and test
go run cmd/railway-server/main.go
# Look for: "✅ Database connection established for new API routes"
```

## Troubleshooting

### "Connection refused"
- Check your database password
- Verify Supabase project is active
- Check if IP restrictions are enabled

### "Table does not exist"
- Run the migration: `psql $DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql`

### "psql: command not found"
- Install PostgreSQL client tools
- Or use the API endpoint to verify (if server is running)

## Next Steps

Once database is configured:
1. ✅ Test all endpoints: `./test/restoration_tests.sh`
2. ✅ Test persistence: `./test/test_database_persistence.sh`
3. ✅ Run pre-deployment check: `./scripts/pre_deployment_check.sh`

---

**Ready to start?** Run: `./scripts/setup_database.sh`

