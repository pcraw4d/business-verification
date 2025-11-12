# How to Get Supabase PostgreSQL Connection String

## Method 1: From Supabase Dashboard

1. Go to https://supabase.com/dashboard
2. Select your project (qpqhuqqmkjxsltzshfam)
3. Navigate to: **Project Settings** > **Database**
4. Scroll to **Connection string** section
5. Select **URI** format
6. Copy the connection string

It will look like:
```
postgresql://postgres:[YOUR-PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

## Method 2: Using Connection Pooler (Recommended for migrations)

1. Same steps as above
2. Select **Connection pooling** tab
3. Use the **Session mode** connection string

It will look like:
```
postgresql://postgres.qpqhuqqmkjxsltzshfam:[YOUR-PASSWORD]@aws-0-[region].pooler.supabase.com:6543/postgres
```

## Method 3: Construct Manually

If you know your password:
```
postgresql://postgres:[PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

Replace:
- `[PASSWORD]` with your database password
- Port `5432` for direct connection, or `6543` for connection pooler

## Usage

Once you have the connection string:

```bash
export DATABASE_URL='postgresql://postgres:your-password@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres'
./scripts/run-migration-010.sh
```

## Security Note

⚠️ **Never commit your database password to version control!**

Use environment variables or a secrets manager.
