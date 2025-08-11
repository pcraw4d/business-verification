#!/bin/bash

# Script to fix remaining issues in API handlers
# This script will fix type references and logger calls

echo "Fixing API handlers..."

# Fix vulnerability handler
echo "Fixing vulnerability handler..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/api/handlers/vulnerability.go
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/api/handlers/vulnerability.go
sed -i '' 's/logger\.Warn(ctx, /logger.Warn(/g' internal/api/handlers/vulnerability.go

# Fix access control handler
echo "Fixing access control handler..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/api/handlers/access_control.go
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/api/handlers/access_control.go
sed -i '' 's/logger\.Warn(ctx, /logger.Warn(/g' internal/api/handlers/access_control.go

# Fix security monitoring handler
echo "Fixing security monitoring handler..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/api/handlers/security_monitoring.go
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/api/handlers/security_monitoring.go
sed -i '' 's/logger\.Warn(ctx, /logger.Warn(/g' internal/api/handlers/security_monitoring.go

# Fix audit logging handler
echo "Fixing audit logging handler..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/api/handlers/audit_logging.go
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/api/handlers/audit_logging.go
sed -i '' 's/logger\.Warn(ctx, /logger.Warn(/g' internal/api/handlers/audit_logging.go

echo "API handlers fixes completed!"
