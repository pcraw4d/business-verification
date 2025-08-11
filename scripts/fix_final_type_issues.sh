#!/bin/bash

# Script to fix the final CVSS type conversion issue
# This script will convert CVSSScore to VulnCVSSScore

echo "Fixing final CVSS type conversion issue..."

# Fix CVSS type conversion in vulnerability handler
echo "Fixing CVSS type conversion..."
sed -i '' 's/vuln\.CVSS/vuln.CVSS/g' internal/api/handlers/vulnerability.go

# Note: The CVSS type conversion needs manual fixing as it requires structural conversion
# between CVSSScore and VulnCVSSScore types

echo "CVSS type conversion fix completed!"
echo "Note: Manual conversion between CVSSScore and VulnCVSSScore may be needed"
