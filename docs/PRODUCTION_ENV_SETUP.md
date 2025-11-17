# Production Environment Setup Guide

## Overview

This guide explains how to configure environment variables for production deployment of the KYB Platform with restored risk management functionality.

## Required Environment Variables

### 1. SUPABASE_URL
**Required**: Yes  
**Description**: Supabase project URL  
**Format**: `https://your-project.supabase.co`  
**Example**:
```bash
export SUPABASE_URL="https://abcdefghijklmnop.supabase.co"
```

**How to find**: Supabase Dashboard → Project Settings → API → Project URL

---

### 2. SUPABASE_ANON_KEY
**Required**: Yes  
**Description**: Supabase anonymous/public API key  
**Format**: JWT token string  
**Example**:
```bash
export SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**How to find**: Supabase Dashboard → Project Settings → API → anon/public key

**Security Note**: This is a public key, but still keep it secure. It's used for client-side operations.

---

## Optional Environment Variables

### 3. DATABASE_URL
**Required**: No (but recommended for production)  
**Description**: Direct PostgreSQL connection string for threshold persistence  
**Format**: `postgres://user:password@host:port/database`  
**Example**:
```bash
export DATABASE_URL="postgres://postgres:password@db.example.com:5432/kyb_platform"
```

**When to use**:
- Production deployments requiring threshold persistence
- When you need direct database access for risk thresholds
- For better performance with connection pooling

**How to find**:
- **Supabase**: Dashboard → Project Settings → Database → Connection string
- **Custom PostgreSQL**: Provided by your database administrator

**Note**: If not set, the system will use in-memory storage (data lost on restart)

---

### 4. REDIS_URL
**Required**: No  
**Description**: Redis connection string for caching  
**Format**: `redis://host:port` or `redis://:password@host:port`  
**Example**:
```bash
export REDIS_URL="redis://redis.example.com:6379"
# Or with password:
export REDIS_URL="redis://:mypassword@redis.example.com:6379"
```

**When to use**:
- High-traffic deployments
- When classification caching is needed
- For improved performance

**Benefits**:
- Faster classification responses
- Reduced database load
- Better scalability

---

### 5. PORT
**Required**: No  
**Description**: Server port number  
**Default**: `8080`  
**Example**:
```bash
export PORT="3000"
```

**When to change**:
- Port conflicts
- Platform-specific requirements (e.g., Railway, Heroku)
- Multiple services on same host

---

### 6. SERVICE_NAME
**Required**: No  
**Description**: Service identifier for logging and monitoring  
**Default**: `kyb-platform`  
**Example**:
```bash
export SERVICE_NAME="kyb-platform-prod"
```

**When to use**:
- Multiple environments (dev, staging, prod)
- Service identification in logs
- Monitoring and alerting

---

## Environment Setup Methods

### Method 1: Export in Shell (Development/Testing)

```bash
# Required
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your-anon-key"

# Optional but recommended
export DATABASE_URL="postgres://user:pass@host:port/db"
export REDIS_URL="redis://host:port"
export PORT="8080"
export SERVICE_NAME="kyb-platform"
```

### Method 2: .env File (Local Development)

Create a `.env` file in the project root:

```bash
# .env
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
DATABASE_URL=postgres://user:pass@host:port/db
REDIS_URL=redis://host:port
PORT=8080
SERVICE_NAME=kyb-platform
```

**Note**: `.env` files are gitignored. Never commit secrets to version control.

### Method 3: Platform Environment Variables (Production)

#### Railway
1. Go to your project → Settings → Variables
2. Add each variable with its value
3. Variables are automatically available to your application

#### Heroku
```bash
heroku config:set SUPABASE_URL="https://your-project.supabase.co"
heroku config:set SUPABASE_ANON_KEY="your-anon-key"
heroku config:set DATABASE_URL="postgres://..."
```

#### Docker
```bash
docker run -e SUPABASE_URL="..." -e SUPABASE_ANON_KEY="..." your-image
```

Or use `docker-compose.yml`:
```yaml
services:
  api:
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}
      - DATABASE_URL=${DATABASE_URL}
```

---

## Validation

### Validate Configuration

Use the validation script to check your configuration:

```bash
./scripts/validate_config.sh
```

This will:
- ✅ Check all required variables are set
- ✅ Validate URL formats
- ✅ Check port numbers
- ✅ Report any issues

### Test Database Connection

If `DATABASE_URL` is set, test the connection:

```bash
./scripts/test_database_connection.sh
```

### Verify Database Schema

Ensure the database schema is set up correctly:

```bash
./scripts/verify_database_schema.sh
```

---

## Production Checklist

Before deploying to production, ensure:

- [ ] All required variables are set
- [ ] `DATABASE_URL` is configured (for persistence)
- [ ] Database schema is migrated (`risk_thresholds` table exists)
- [ ] Configuration is validated (`./scripts/validate_config.sh`)
- [ ] Database connection is tested (`./scripts/test_database_connection.sh`)
- [ ] Environment variables are set in your deployment platform
- [ ] Secrets are stored securely (not in code)
- [ ] `.env` files are gitignored

---

## Security Best Practices

1. **Never commit secrets**: Use `.gitignore` for `.env` files
2. **Use platform secrets**: Use your platform's secret management (Railway, Heroku, etc.)
3. **Rotate keys regularly**: Update API keys periodically
4. **Limit access**: Only give access to those who need it
5. **Monitor usage**: Watch for unusual API usage patterns

---

## Troubleshooting

### "DATABASE_URL not set" Warning

**Symptom**: Server starts but shows warning about database  
**Solution**: Set `DATABASE_URL` or accept in-memory mode (data lost on restart)

### "Database ping failed" Error

**Symptom**: Database connection fails  
**Solutions**:
1. Verify `DATABASE_URL` format is correct
2. Check database is accessible from your server
3. Verify credentials are correct
4. Check firewall/network rules

### "Table does not exist" Error

**Symptom**: Database connected but `risk_thresholds` table missing  
**Solution**: Run migration:
```bash
psql $DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql
```

### Configuration Validation Fails

**Symptom**: `validate_config.sh` reports errors  
**Solutions**:
1. Check all required variables are set
2. Verify URL formats are correct
3. Check port numbers are valid (1-65535)

---

## Example Production Setup

### Complete Example

```bash
# Required
export SUPABASE_URL="https://abcdefghijklmnop.supabase.co"
export SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Recommended for production
export DATABASE_URL="postgres://postgres:secure_password@db.railway.app:5432/railway"
export REDIS_URL="redis://redis.railway.app:6379"

# Optional
export PORT="8080"
export SERVICE_NAME="kyb-platform-prod"

# Validate
./scripts/validate_config.sh

# Test database
./scripts/test_database_connection.sh

# Verify schema
./scripts/verify_database_schema.sh
```

---

## Next Steps

After configuring environment variables:

1. ✅ Validate configuration: `./scripts/validate_config.sh`
2. ✅ Test database connection: `./scripts/test_database_connection.sh`
3. ✅ Verify database schema: `./scripts/verify_database_schema.sh`
4. ✅ Review deployment checklist: `docs/DEPLOYMENT_CHECKLIST.md`
5. ✅ Deploy to staging first
6. ✅ Monitor and verify in production

---

**Last Updated**: November 15, 2025  
**Version**: 1.0

