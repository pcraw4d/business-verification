# Using railway.env for DATABASE_URL

The `DATABASE_URL` has been added to your `railway.env` file in the **DATABASE CONFIGURATION** section.

## Loading railway.env

### Option 1: Source the file directly
```bash
source railway.env
```

### Option 2: Export variables from railway.env
```bash
export $(cat railway.env | grep -v '^#' | xargs)
```

### Option 3: Use with your server
If your server loads environment variables automatically, make sure to:
1. **Update the password** in `railway.env`:
   ```env
   DATABASE_URL=postgresql://postgres:YOUR_ACTUAL_PASSWORD@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres?sslmode=require
   ```

2. **Load before starting the server**:
   ```bash
   source railway.env
   go run cmd/railway-server/main.go
   ```

## Important: Update the Password

The `DATABASE_URL` in `railway.env` currently has a placeholder:
```
your_supabase_db_password_here
```

**Replace this with your actual Supabase database password** to make it work.

## Getting Your Password

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Go to **Project Settings** > **Database**
4. Find your database password (or reset it if needed)
5. Update `railway.env` with the actual password

## Verifying It Works

After updating the password and loading the file:
```bash
# Load the environment
source railway.env

# Check if DATABASE_URL is set
echo $DATABASE_URL

# Test the connection
psql $DATABASE_URL -c "SELECT version();"

# Or run the migration
./scripts/run-migration-010.sh
```

## For Railway Deployment

If you're deploying to Railway, you can:
1. Copy the `DATABASE_URL` value from `railway.env`
2. Add it as an environment variable in Railway dashboard
3. Or use Railway's CLI to set it:
   ```bash
   railway variables set DATABASE_URL="postgresql://..."
   ```

