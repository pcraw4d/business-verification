# Railway Deployment Instructions

## 🚀 Quick Deploy to Railway

### Step 1: Create Railway Project
1. Go to [railway.app](https://railway.app)
2. Click "New Project"
3. Choose "Deploy from GitHub repo"
4. Select this repository

### Step 2: Configure Environment Variables
1. In your Railway project dashboard
2. Go to "Variables" tab
3. Copy variables from `.env.railway.template`
4. Fill in your actual values:
   - `JWT_SECRET`: Generate a random 32-character string
   - `ENCRYPTION_KEY`: Generate a random 32-character string
   - `SUPABASE_URL`: Your Supabase project URL
   - `SUPABASE_ANON_KEY`: Your Supabase anon key
   - `SUPABASE_SERVICE_ROLE_KEY`: Your Supabase service role key

### Step 3: Add PostgreSQL Database
1. In Railway dashboard, click "New"
2. Select "Database" → "PostgreSQL"
3. Railway will automatically set `DATABASE_URL`

### Step 4: Deploy
1. Railway will automatically detect the Dockerfile.beta
2. Click "Deploy" to start the build
3. Wait for deployment to complete

### Step 5: Get Your URLs
1. Go to "Settings" tab
2. Copy your deployment URL
3. Share with beta testers

## 🔧 Manual Deployment Commands

If you prefer using Railway CLI:

```bash
# Login to Railway
railway login

# Link to project
railway link

# Deploy
railway up

# Get deployment URL
railway domain
```

## 📊 Monitoring

- **Logs**: View in Railway dashboard
- **Metrics**: Built-in monitoring
- **Health Checks**: Automatic health monitoring

## 🔒 Security Notes

- All secrets are encrypted in Railway
- Environment variables are secure
- SSL is automatically enabled
- Health checks ensure uptime

## 🐛 Troubleshooting

### Build Failures
- Check Dockerfile.beta syntax
- Verify all files are present
- Check Railway logs

### Runtime Errors
- Verify environment variables
- Check database connection
- Review application logs

### Performance Issues
- Monitor resource usage
- Scale up if needed
- Optimize Docker image
