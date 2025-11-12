# Setting DATABASE_URL Environment Variable

The `DATABASE_URL` environment variable is required for the new merchant analytics and async risk assessment API routes to work.

## Quick Start

### Option 1: Using .env File (Recommended)

1. **Create a `.env` file** in the project root:
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env`** and add your database connection string:
   ```env
   DATABASE_URL=postgresql://postgres:your-password@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
   ```

3. **Load the environment variables**:
   ```bash
   # Using source (bash/zsh)
   source .env
   
   # Or using export
   export $(cat .env | xargs)
   ```

4. **Run the server**:
   ```bash
   go run cmd/railway-server/main.go
   ```

### Option 2: Using Setup Script

Run the interactive setup script:
```bash
./scripts/setup-env.sh
```

This will prompt you for the required values and create a `.env` file automatically.

### Option 3: Export Directly (Temporary)

For a single terminal session:
```bash
export DATABASE_URL='postgresql://postgres:your-password@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres'
```

## Getting Your Database Connection String

### From Supabase Dashboard

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Navigate to **Project Settings** > **Database**
4. Scroll to **Connection string** section
5. Select **URI** format
6. Copy the connection string

It will look like:
```
postgresql://postgres:[YOUR-PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres
```

**Important:** Replace `[YOUR-PASSWORD]` with your actual database password.

### Connection String Formats

#### Direct Connection (Port 5432)
```
postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
```

#### Connection Pooler (Port 6543) - Recommended for Production
```
postgresql://postgres.[PROJECT-REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres
```

## Verifying DATABASE_URL is Set

### Check if it's set:
```bash
echo $DATABASE_URL
```

### Test the connection:
```bash
# Using psql
psql $DATABASE_URL -c "SELECT version();"

# Or using the migration script
./scripts/run-migration-010.sh
```

## Using with Different Tools

### With direnv (Auto-load .env)

1. Install direnv:
   ```bash
   brew install direnv  # macOS
   # or
   apt-get install direnv  # Linux
   ```

2. Add to your shell config (`~/.zshrc` or `~/.bashrc`):
   ```bash
   eval "$(direnv hook zsh)"  # or bash
   ```

3. Create `.envrc` in project root:
   ```bash
   dotenv .env
   ```

4. Allow direnv:
   ```bash
   direnv allow
   ```

### With Docker

Add to your `docker-compose.yml`:
```yaml
services:
  api:
    environment:
      - DATABASE_URL=${DATABASE_URL}
    env_file:
      - .env
```

### With Railway/Render/Heroku

Set the environment variable in your platform's dashboard:
- **Railway**: Project Settings > Variables
- **Render**: Environment > Environment Variables
- **Heroku**: Settings > Config Vars

## Security Best Practices

⚠️ **Never commit your `.env` file to version control!**

1. **Add to `.gitignore`**:
   ```
   .env
   .env.local
   .env.*.local
   ```

2. **Use `.env.example`** for documentation (without real credentials)

3. **Use secrets management** in production:
   - AWS Secrets Manager
   - HashiCorp Vault
   - Platform-specific secrets (Railway, Render, etc.)

## Troubleshooting

### "Database connection failed"
- Verify your password is correct
- Check that the connection string format is correct
- Ensure your IP is allowed in Supabase (if using IP restrictions)

### "DATABASE_URL not set"
- Make sure you've exported the variable or loaded the `.env` file
- Check that the variable name is exactly `DATABASE_URL` (case-sensitive)

### "Connection refused"
- Verify the host and port are correct
- Check if you're using the connection pooler (port 6543) vs direct connection (port 5432)
- Ensure your Supabase project is active

## Example .env File

```env
# Database Configuration
DATABASE_URL=postgresql://postgres:your-actual-password@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres

# Supabase Configuration
SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here

# Server Configuration
PORT=8080
SERVICE_NAME=kyb-platform-v4-complete
```

## Next Steps

Once `DATABASE_URL` is set:

1. **Run the migration** (if not already done):
   ```bash
   ./scripts/run-migration-010.sh
   ```

2. **Start the server**:
   ```bash
   go run cmd/railway-server/main.go
   ```

3. **Verify routes are registered**:
   Look for this log message:
   ```
   ✅ New API routes registered:
      - GET /api/v1/merchants/{merchantId}/analytics
      - GET /api/v1/merchants/{merchantId}/website-analysis
      - POST /api/v1/risk/assess
      - GET /api/v1/risk/assess/{assessmentId}
   ```

4. **Test the endpoints** using Postman/Insomnia collections in `tests/api/merchant-details/`

