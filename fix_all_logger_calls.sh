#!/bin/bash

# Script to fix all logger calls in the codebase

echo "Fixing logger calls throughout the codebase..."

# Function to fix logger calls in a file
fix_logger_calls() {
    local file="$1"
    echo "Processing: $file"
    
    # Fix simple logger calls with just a message
    sed -i '' 's/\([a-zA-Z_][a-zA-Z0-9_]*\)\.logger\.\(Info\|Warn\|Error\|Debug\|Fatal\)("\([^"]*\)")/\1.logger.\2("\3", map[string]interface{}{})/g' "$file"
    
    # Fix logger calls with error parameter
    sed -i '' 's/\([a-zA-Z_][a-zA-Z0-9_]*\)\.logger\.\(Info\|Warn\|Error\|Debug\|Fatal\)("\([^"]*\)", "error", \([^)]*\))/\1.logger.\2("\3", map[string]interface{}{"error": \4})/g' "$file"
    
    # Fix logger calls with two parameters
    sed -i '' 's/\([a-zA-Z_][a-zA-Z0-9_]*\)\.logger\.\(Info\|Warn\|Error\|Debug\|Fatal\)("\([^"]*\)", "\([^"]*\)", \([^)]*\))/\1.logger.\2("\3", map[string]interface{}{"\4": \5})/g' "$file"
    
    # Fix logger calls with three parameters
    sed -i '' 's/\([a-zA-Z_][a-zA-Z0-9_]*\)\.logger\.\(Info\|Warn\|Error\|Debug\|Fatal\)("\([^"]*\)", "\([^"]*\)", \([^,]*\), "\([^"]*\)", \([^)]*\))/\1.logger.\2("\3", map[string]interface{}{"\4": \5, "\6": \7})/g' "$file"
    
    # Fix logger calls with four parameters
    sed -i '' 's/\([a-zA-Z_][a-zA-Z0-9_]*\)\.logger\.\(Info\|Warn\|Error\|Debug\|Fatal\)("\([^"]*\)", "\([^"]*\)", \([^,]*\), "\([^"]*\)", \([^,]*\), "\([^"]*\)", \([^)]*\))/\1.logger.\2("\3", map[string]interface{}{"\4": \5, "\6": \7, "\8": \9})/g' "$file"
}

# Find all Go files in internal directory and fix them
find internal/ -name "*.go" -type f | while read -r file; do
    fix_logger_calls "$file"
done

echo "Logger call fixes completed!"
