# Beta Deployment Guide

**Version:** 1.0.0-beta  
**Last Updated:** January 2025

---

## Overview

This guide provides step-by-step instructions for deploying the KYB Platform beta release to production or staging environments.

---

## Prerequisites

### Infrastructure Requirements

- **Server:** Linux-based (Ubuntu 20.04+ recommended)
- **Database:** PostgreSQL 13+
- **Cache:** Redis 6+ (optional but recommended)
- **Reverse Proxy:** Nginx or similar
- **SSL Certificate:** For HTTPS
- **Domain:** Configured DNS

### Software Requirements

- **Go:** 1.22 or later
- **Node.js:** 18+ and npm
- **Docker:** 20.10+ (optional, for containerized deployment)
- **Git:** For code deployment

### Environment Variables

Required environment variables:

```bash
# Database
DATABASE_URL=postgresql://user:password@host:5432/dbname

# Redis (optional)
REDIS_URL=redis://host:6379

# API Configuration
API_PORT=8080
API_HOST=0.0.0.0

# Frontend Configuration
NEXT_PUBLIC_API_BASE_URL=https://api.yourdomain.com

# Security
JWT_SECRET=your-secret-key
API_KEY=your-api-key

# Environment
ENVIRONMENT=production
```

---

## Deployment Steps

### 1. Prepare Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y postgresql postgresql-contrib redis-server nginx git

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs
```

### 2. Clone Repository

```bash
# Create application directory
sudo mkdir -p /opt/kyb-platform
sudo chown $USER:$USER /opt/kyb-platform

# Clone repository
cd /opt/kyb-platform
git clone https://github.com/pcraw4d/business-verification.git .

# Checkout beta release
git checkout v1.0.0-beta
```

### 3. Set Up Database

```bash
# Create database
sudo -u postgres psql
CREATE DATABASE kyb_platform;
CREATE USER kyb_user WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE kyb_platform TO kyb_user;
\q

# Run migrations
cd /opt/kyb-platform
export DATABASE_URL="postgresql://kyb_user:your-password@localhost:5432/kyb_platform"
# Follow migration guide to run migrations
```

### 4. Build Frontend

```bash
# Navigate to frontend directory
cd /opt/kyb-platform/frontend

# Install dependencies
npm ci

# Build Next.js application
npm run build

# Verify build
ls -la .next/
```

### 5. Build Backend

```bash
# Navigate to project root
cd /opt/kyb-platform

# Install Go dependencies
go mod download

# Build backend
go build -o bin/kyb-server ./cmd/railway-server

# Verify build
./bin/kyb-server --version
```

### 6. Configure Nginx

Create Nginx configuration file:

```nginx
# /etc/nginx/sites-available/kyb-platform

server {
    listen 80;
    server_name yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/ssl/certs/yourdomain.crt;
    ssl_certificate_key /etc/ssl/private/yourdomain.key;

    # Frontend
    location / {
        root /opt/kyb-platform/frontend/.next/static;
        try_files $uri $uri/ /index.html;
    }

    # API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:8080/health;
    }
}
```

Enable configuration:

```bash
sudo ln -s /etc/nginx/sites-available/kyb-platform /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 7. Create Systemd Service

Create service file:

```ini
# /etc/systemd/system/kyb-platform.service

[Unit]
Description=KYB Platform API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=kyb
WorkingDirectory=/opt/kyb-platform
Environment="DATABASE_URL=postgresql://kyb_user:password@localhost:5432/kyb_platform"
Environment="REDIS_URL=redis://localhost:6379"
Environment="API_PORT=8080"
Environment="ENVIRONMENT=production"
ExecStart=/opt/kyb-platform/bin/kyb-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable kyb-platform
sudo systemctl start kyb-platform
sudo systemctl status kyb-platform
```

### 8. Set Up Logging

```bash
# Create log directory
sudo mkdir -p /var/log/kyb-platform
sudo chown kyb:kyb /var/log/kyb-platform

# Configure log rotation
sudo nano /etc/logrotate.d/kyb-platform
```

Log rotation configuration:

```
/var/log/kyb-platform/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 kyb kyb
    sharedscripts
    postrotate
        systemctl reload kyb-platform > /dev/null 2>&1 || true
    endscript
}
```

### 9. Set Up Monitoring

```bash
# Install monitoring tools (optional)
# Prometheus, Grafana, or similar

# Set up health check monitoring
# Configure alerts for service downtime
```

### 10. Verify Deployment

```bash
# Check service status
sudo systemctl status kyb-platform

# Check logs
sudo journalctl -u kyb-platform -f

# Test API endpoint
curl http://localhost:8080/health

# Test frontend
curl https://yourdomain.com

# Check database connection
psql -U kyb_user -d kyb_platform -c "SELECT 1;"
```

---

## Docker Deployment (Alternative)

### Build Docker Images

```bash
# Build backend image
docker build -t kyb-platform-api:beta -f cmd/railway-server/Dockerfile .

# Build frontend image
docker build -t kyb-platform-frontend:beta -f cmd/frontend-service/Dockerfile .
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: kyb_platform
      POSTGRES_USER: kyb_user
      POSTGRES_PASSWORD: your-password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    volumes:
      - redis_data:/data

  api:
    image: kyb-platform-api:beta
    environment:
      DATABASE_URL: postgresql://kyb_user:your-password@postgres:5432/kyb_platform
      REDIS_URL: redis://redis:6379
      API_PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis

  frontend:
    image: kyb-platform-frontend:beta
    environment:
      NEXT_PUBLIC_API_BASE_URL: http://api:8080
    ports:
      - "3000:3000"
    depends_on:
      - api

volumes:
  postgres_data:
  redis_data:
```

Deploy:

```bash
docker-compose up -d
```

---

## Rollback Procedure

### Manual Rollback

1. **Stop current service**
   ```bash
   sudo systemctl stop kyb-platform
   ```

2. **Checkout previous version**
   ```bash
   cd /opt/kyb-platform
   git checkout v0.9.0  # Previous version
   ```

3. **Rebuild and restart**
   ```bash
   go build -o bin/kyb-server ./cmd/railway-server
   sudo systemctl start kyb-platform
   ```

### Database Rollback

```bash
# Restore database from backup
pg_restore -U kyb_user -d kyb_platform backup.dump
```

---

## Health Checks

### API Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-XX",
  "version": "1.0.0-beta"
}
```

### Frontend Health Check

```bash
curl https://yourdomain.com
```

Should return HTML page.

---

## Troubleshooting

### Service Won't Start

1. Check logs: `sudo journalctl -u kyb-platform -n 50`
2. Verify environment variables
3. Check database connection
4. Verify port availability

### Database Connection Issues

1. Verify PostgreSQL is running: `sudo systemctl status postgresql`
2. Check connection string
3. Verify user permissions
4. Check firewall rules

### Frontend Not Loading

1. Verify Next.js build completed
2. Check Nginx configuration
3. Verify SSL certificates
4. Check file permissions

---

## Security Checklist

- [ ] SSL certificates configured
- [ ] Environment variables secured
- [ ] Database credentials protected
- [ ] API keys rotated
- [ ] Firewall rules configured
- [ ] Regular backups scheduled
- [ ] Monitoring and alerts set up
- [ ] Log rotation configured

---

## Post-Deployment

1. **Monitor logs** for errors
2. **Test all endpoints** manually
3. **Verify database** connections
4. **Check performance** metrics
5. **Set up alerts** for critical issues

---

## Support

For deployment issues:
- Check logs: `sudo journalctl -u kyb-platform`
- Review documentation
- Contact DevOps team

---

**Deployment Complete!** 🚀

