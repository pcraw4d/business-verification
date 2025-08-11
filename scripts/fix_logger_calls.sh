#!/bin/bash

# Script to fix remaining logger interface issues in the security package
# This script will fix all logger method calls that incorrectly pass context as first parameter

echo "Fixing logger interface issues..."

# Fix audit_logging.go logger calls
echo "Fixing audit_logging.go logger calls..."
sed -i '' 's/logger\.Error(context\.Background(), /logger.Error(/g' internal/security/audit_logging.go
sed -i '' 's/logger\.Info(context\.Background(), /logger.Info(/g' internal/security/audit_logging.go
sed -i '' 's/logger\.Debug(context\.Background(), /logger.Debug(/g' internal/security/audit_logging.go

# Fix vulnerability_management.go logger calls
echo "Fixing vulnerability_management.go logger calls..."
sed -i '' 's/logger\.Info(context\.Background(), /logger.Info(/g' internal/security/vulnerability_management.go
sed -i '' 's/logger\.Error(context\.Background(), /logger.Error(/g' internal/security/vulnerability_management.go

# Fix dashboard.go logger calls
echo "Fixing dashboard.go logger calls..."
sed -i '' 's/logger\.Error(context\.Background(), /logger.Error(/g' internal/security/dashboard.go
sed -i '' 's/logger\.Info(context\.Background(), /logger.Info(/g' internal/security/dashboard.go

# Fix access_control.go logger calls
echo "Fixing access_control.go logger calls..."
sed -i '' 's/logger\.Info(context\.Background(), /logger.Info(/g' internal/security/access_control.go
sed -i '' 's/logger\.Error(context\.Background(), /logger.Error(/g' internal/security/access_control.go

# Fix monitoring.go logger calls
echo "Fixing monitoring.go logger calls..."
sed -i '' 's/logger\.Info(context\.Background(), /logger.Info(/g' internal/security/monitoring.go
sed -i '' 's/logger\.Error(context\.Background(), /logger.Error(/g' internal/security/monitoring.go

echo "Logger interface fixes completed!"
