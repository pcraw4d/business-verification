#!/usr/bin/env python3
import re
import sys

def fix_logger_calls(content):
    # Fix logger calls with multiple parameters
    patterns = [
        # Pattern for logger calls with 3 parameters: msg, key1, value1
        (r'(\w+)\.logger\.(\w+)\(("([^"]*)"),\s*"([^"]*)",\s*([^)]+)\)', 
         r'\1.logger.\2(\3, map[string]interface{}{\5: \6})'),
        
        # Pattern for logger calls with 5 parameters: msg, key1, value1, key2, value2
        (r'(\w+)\.logger\.(\w+)\(("([^"]*)"),\s*"([^"]*)",\s*([^,]+),\s*"([^"]*)",\s*([^)]+)\)',
         r'\1.logger.\2(\3, map[string]interface{}{\5: \6, \7: \8})'),
        
        # Pattern for logger calls with 7 parameters: msg, key1, value1, key2, value2, key3, value3
        (r'(\w+)\.logger\.(\w+)\(("([^"]*)"),\s*"([^"]*)",\s*([^,]+),\s*"([^"]*)",\s*([^,]+),\s*"([^"]*)",\s*([^)]+)\)',
         r'\1.logger.\2(\3, map[string]interface{}{\5: \6, \7: \8, \9: \10})'),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content)
    
    return content

def main():
    files_to_fix = [
        'internal/risk/automated_alerts.go',
        'internal/security/access_control.go',
        'internal/security/audit_logging.go',
        'internal/security/dashboard.go'
    ]
    
    for file_path in files_to_fix:
        try:
            with open(file_path, 'r') as f:
                content = f.read()
            
            fixed_content = fix_logger_calls(content)
            
            with open(file_path, 'w') as f:
                f.write(fixed_content)
            
            print(f"Fixed logger calls in {file_path}")
        except FileNotFoundError:
            print(f"File not found: {file_path}")
        except Exception as e:
            print(f"Error processing {file_path}: {e}")

if __name__ == "__main__":
    main()
