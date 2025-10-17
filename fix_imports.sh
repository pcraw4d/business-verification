#!/bin/bash

# Fix external module imports to use local modules
echo "Fixing external module imports..."

# Replace external imports with local ones
find . -name "*.go" -type f -exec sed -i '' 's|github.com/pcraw4d/business-verification/internal/|kyb-platform/internal/|g' {} \;
find . -name "*.go" -type f -exec sed -i '' 's|github.com/pcraw4d/business-verification/test|kyb-platform/test|g' {} \;

echo "Import fixes completed."
