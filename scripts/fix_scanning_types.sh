#!/bin/bash

# Script to fix type conversion issues in scanning.go
# This script will convert string severity values to Severity enum type in switch statements

echo "Fixing scanning.go type conversion issues..."

# Fix container issues switch statement
echo "Fixing container issues switch statement..."
sed -i '' 's/switch issue\.Severity {/switch Severity(issue.Severity) {/g' internal/security/scanning.go

# Fix secrets switch statement
echo "Fixing secrets switch statement..."
sed -i '' 's/switch secret\.Severity {/switch Severity(secret.Severity) {/g' internal/security/scanning.go

# Fix compliance issues switch statement
echo "Fixing compliance issues switch statement..."
sed -i '' 's/switch issue\.Severity {/switch Severity(issue.Severity) {/g' internal/security/scanning.go

echo "Scanning type conversion fixes completed!"
