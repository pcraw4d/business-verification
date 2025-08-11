#!/bin/bash

# Script to fix remaining type conflicts in the security package
# This script will replace all undefined type references with the correct shared types

echo "Fixing security package type conflicts..."

# Fix dashboard.go
echo "Fixing dashboard.go..."
sed -i '' 's/SecurityEventType/EventType/g' internal/security/dashboard.go
sed -i '' 's/SecuritySeverity/Severity/g' internal/security/dashboard.go

# Fix audit_logging.go
echo "Fixing audit_logging.go..."
sed -i '' 's/AuditEventType/EventType/g' internal/security/audit_logging.go
sed -i '' 's/AuditEventCategory/EventCategory/g' internal/security/audit_logging.go
sed -i '' 's/AuditEventSeverity/Severity/g' internal/security/audit_logging.go

# Fix vulnerability_management.go
echo "Fixing vulnerability_management.go..."
sed -i '' 's/SecuritySeverity/Severity/g' internal/security/vulnerability_management.go

# Fix logger calls in access_control.go
echo "Fixing logger calls in access_control.go..."
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/security/access_control.go

# Fix logger calls in audit_logging.go
echo "Fixing logger calls in audit_logging.go..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/security/audit_logging.go

# Fix logger calls in vulnerability_management.go
echo "Fixing logger calls in vulnerability_management.go..."
sed -i '' 's/logger\.Info(ctx, /logger.Info(/g' internal/security/vulnerability_management.go

# Fix logger calls in dashboard.go
echo "Fixing logger calls in dashboard.go..."
sed -i '' 's/logger\.Error(ctx, /logger.Error(/g' internal/security/dashboard.go

echo "Type conflict fixes completed!"
