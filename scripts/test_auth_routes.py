#!/usr/bin/env python3
"""
Test script for authentication routes
Tests POST /api/v1/auth/register and POST /api/v1/auth/login
"""

import requests
import json
import sys
from datetime import datetime

API_BASE_URL = "https://api-gateway-service-production-21fd.up.railway.app"

def test_auth_register():
    """Test POST /api/v1/auth/register"""
    print("\n" + "="*60)
    print("Testing POST /api/v1/auth/register")
    print("="*60)
    
    # Test 1: Valid registration data
    print("\n1. Testing with valid registration data...")
    valid_data = {
        "email": f"test_{datetime.now().strftime('%Y%m%d%H%M%S')}@example.com",
        "password": "TestPassword123!",
        "username": "testuser",
        "first_name": "Test",
        "last_name": "User"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/register",
            json=valid_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code in [200, 201]:
            print("   ✅ Valid registration: PASSED")
        else:
            print(f"   ⚠️ Valid registration: Status {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test 2: Missing required fields
    print("\n2. Testing with missing required fields...")
    invalid_data = {
        "email": "test@example.com"
        # Missing password
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/register",
            json=invalid_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 400:
            print("   ✅ Missing fields validation: PASSED")
        else:
            print(f"   ⚠️ Missing fields validation: Expected 400, got {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test 3: Invalid email format
    print("\n3. Testing with invalid email format...")
    invalid_email_data = {
        "email": "not-an-email",
        "password": "TestPassword123!"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/register",
            json=invalid_email_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 400:
            print("   ✅ Invalid email validation: PASSED")
        else:
            print(f"   ⚠️ Invalid email validation: Expected 400, got {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

def test_auth_login():
    """Test POST /api/v1/auth/login"""
    print("\n" + "="*60)
    print("Testing POST /api/v1/auth/login")
    print("="*60)
    
    # Test 1: Valid credentials (if user exists)
    print("\n1. Testing with valid credentials...")
    valid_data = {
        "email": "test@example.com",
        "password": "TestPassword123!"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/login",
            json=valid_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 200:
            data = response.json()
            if "token" in data:
                print("   ✅ Valid login: PASSED (token received)")
            else:
                print("   ⚠️ Valid login: No token in response")
        elif response.status_code == 401:
            print("   ⚠️ Valid login: 401 (user may not exist, expected for test)")
        else:
            print(f"   ⚠️ Valid login: Status {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test 2: Invalid credentials
    print("\n2. Testing with invalid credentials...")
    invalid_data = {
        "email": "nonexistent@example.com",
        "password": "WrongPassword123!"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/login",
            json=invalid_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 401:
            print("   ✅ Invalid credentials: PASSED (401 Unauthorized)")
        else:
            print(f"   ⚠️ Invalid credentials: Expected 401, got {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test 3: Missing required fields
    print("\n3. Testing with missing required fields...")
    missing_fields_data = {
        "email": "test@example.com"
        # Missing password
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/login",
            json=missing_fields_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 400:
            print("   ✅ Missing fields validation: PASSED")
        else:
            print(f"   ⚠️ Missing fields validation: Expected 400, got {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test 4: Invalid email format
    print("\n4. Testing with invalid email format...")
    invalid_email_data = {
        "email": "not-an-email",
        "password": "TestPassword123!"
    }
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/v1/auth/login",
            json=invalid_email_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        print(f"   Status Code: {response.status_code}")
        print(f"   Response: {response.text[:200]}")
        
        if response.status_code == 400:
            print("   ✅ Invalid email validation: PASSED")
        else:
            print(f"   ⚠️ Invalid email validation: Expected 400, got {response.status_code}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

def main():
    print("="*60)
    print("Authentication Routes Test Suite")
    print("="*60)
    print(f"API Base URL: {API_BASE_URL}")
    print(f"Test Time: {datetime.now().isoformat()}")
    
    test_auth_register()
    test_auth_login()
    
    print("\n" + "="*60)
    print("Test Suite Complete")
    print("="*60)

if __name__ == "__main__":
    main()

