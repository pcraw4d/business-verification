#!/usr/bin/env python3
"""
Script to fix logger calls throughout the codebase to use structured logging format.
"""

import os
import re
import glob

def fix_logger_calls_in_file(filepath):
    """Fix logger calls in a single file."""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # Pattern to match logger calls with multiple arguments
        # Matches: logger.Info("message", "key1", value1, "key2", value2)
        pattern = r'(\w+\.logger\.(Info|Warn|Error|Debug|Fatal))\s*\(\s*"([^"]*)"\s*,\s*((?:"[^"]*"\s*,\s*[^,)]+\s*,\s*)*"[^"]*"\s*,\s*[^,)]+)\s*\)'
        
        def replace_logger_call(match):
            logger_call = match.group(1)
            message = match.group(3)
            args = match.group(4)
            
            # Parse the key-value pairs
            pairs = []
            # Split by comma, but be careful about strings with commas
            parts = []
            current_part = ""
            in_string = False
            paren_count = 0
            
            for char in args:
                if char == '"' and (not current_part or current_part[-1] != '\\'):
                    in_string = not in_string
                elif char == '(' and not in_string:
                    paren_count += 1
                elif char == ')' and not in_string:
                    paren_count -= 1
                elif char == ',' and not in_string and paren_count == 0:
                    parts.append(current_part.strip())
                    current_part = ""
                    continue
                current_part += char
            
            if current_part.strip():
                parts.append(current_part.strip())
            
            # Group into key-value pairs
            kv_pairs = []
            for i in range(0, len(parts), 2):
                if i + 1 < len(parts):
                    key = parts[i].strip().strip('"')
                    value = parts[i + 1].strip()
                    kv_pairs.append(f'"{key}": {value}')
            
            if kv_pairs:
                fields_map = "map[string]interface{}{\n" + ",\n".join(f"\t\t{kv}" for kv in kv_pairs) + ",\n\t}"
                return f'{logger_call}("{message}", {fields_map})'
            else:
                return f'{logger_call}("{message}", map[string]interface{{}})'
        
        # Apply the replacement
        content = re.sub(pattern, replace_logger_call, content)
        
        # Also fix simple logger calls with just a message
        simple_pattern = r'(\w+\.logger\.(Info|Warn|Error|Debug|Fatal))\s*\(\s*"([^"]*)"\s*\)'
        content = re.sub(simple_pattern, r'\1("\3", map[string]interface{})', content)
        
        # Fix logger calls with error parameters
        error_pattern = r'(\w+\.logger\.(Info|Warn|Error|Debug|Fatal))\s*\(\s*"([^"]*)"\s*,\s*"error"\s*,\s*([^)]+)\s*\)'
        content = re.sub(error_pattern, r'\1("\3", map[string]interface{}{"error": \4})', content)
        
        if content != original_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"Fixed logger calls in: {filepath}")
            return True
        
        return False
        
    except Exception as e:
        print(f"Error processing {filepath}: {e}")
        return False

def main():
    """Main function to fix logger calls in all Go files."""
    # Find all Go files in internal directory
    go_files = glob.glob("internal/**/*.go", recursive=True)
    
    fixed_count = 0
    for filepath in go_files:
        if fix_logger_calls_in_file(filepath):
            fixed_count += 1
    
    print(f"Fixed logger calls in {fixed_count} files")

if __name__ == "__main__":
    main()
