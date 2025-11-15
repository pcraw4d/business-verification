# Supabase Database Connection for Integration Tests

## Issue

The integration tests need a PostgreSQL connection string to connect to Supabase. The `SUPABASE_SERVICE_ROLE_KEY` is a JWT token for API authentication, not a database password, so it cannot be used directly in a PostgreSQL connection string.

## Solutions

### Option 1: Use DATABASE_URL (Recommended)

Add the `DATABASE_URL` to `.env.railway.full`. You can find this in your Supabase dashboard:

1. Go to your Supabase project dashboard
2. Navigate to **Settings** → **Database**
3. Find the **Connection string** section
4. Copy the **URI** connection string (it will look like: `postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT_REF].supabase.co:5432/postgres`)
5. Add it to `.env.railway.full`:

```bash
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

**Note:** Replace `[YOUR-PASSWORD]` with your actual database password.

### Option 2: Use Connection Pooler with Service Role Key

If you want to use the service role key, you need to use Supabase's connection pooler. The format is:

```
postgres://postgres.[PROJECT_REF]:[SERVICE_ROLE_KEY]@aws-0-[REGION].pooler.supabase.com:6543/postgres
```

However, you need to know your Supabase region. You can find this in your Supabase dashboard under **Settings** → **Database** → **Connection Pooling**.

### Option 3: Get Database Password

1. Go to Supabase dashboard
2. Navigate to **Settings** → **Database**
3. Find your database password (you may need to reset it if you don't have it)
4. Use it in the connection string format:
   ```
   postgresql://postgres:[PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
   ```

## Current Status

The test runner is configured to:
1. ✅ Load credentials from `.env.railway.full`
2. ✅ Check for `DATABASE_URL` first (if available)
3. ✅ Fall back to constructing connection string from `SUPABASE_URL` and `SUPABASE_SERVICE_ROLE_KEY`
4. ⚠️ The constructed connection string may not work without the correct format/region

## Recommended Action

**Add `DATABASE_URL` to `.env.railway.full`** with the connection string from your Supabase dashboard. This is the most reliable method.

## Testing the Connection

Once you've added `DATABASE_URL`, you can test the connection:

```bash
bash test/integration/run_weeks_24_tests.sh
```

The tests should now connect successfully to your Supabase database.

