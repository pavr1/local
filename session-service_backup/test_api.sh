#!/bin/bash

# Test script for Session Management API
# Run this after starting the session service

BASE_URL="http://localhost:8081/api/v1/sessions"

echo "🧪 Testing Session Management API"
echo "================================="

# Test 1: Health Check
echo "📋 1. Testing Health Check..."
curl -s -X GET "$BASE_URL/health" | jq '.'
echo ""

# Test 2: Create Session
echo "📋 2. Testing Session Creation..."
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-123",
    "username": "test_user",
    "role_name": "admin",
    "permissions": ["read", "write", "admin"],
    "remember_me": false
  }')

echo $SESSION_RESPONSE | jq '.'
TOKEN=$(echo $SESSION_RESPONSE | jq -r '.token')
SESSION_ID=$(echo $SESSION_RESPONSE | jq -r '.session_id')
echo ""

# Test 3: Validate Session
echo "📋 3. Testing Session Validation..."
curl -s -X POST "$BASE_URL/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}" | jq '.'
echo ""

# Test 4: Session Stats
echo "📋 4. Testing Session Statistics..."
curl -s -X GET "$BASE_URL/stats" | jq '.'
echo ""

# Test 5: Logout (Revoke Session)
echo "📋 5. Testing Logout..."
curl -s -X POST "$BASE_URL/logout" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}" | jq '.'
echo ""

# Test 6: Validate Session (Should be invalid now)
echo "📋 6. Testing Session Validation After Logout..."
curl -s -X POST "$BASE_URL/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}" | jq '.'
echo ""

echo "✅ API Tests Complete!"
echo ""
echo "🚀 To test with the actual service:"
echo "   1. Start the session service: go run ."
echo "   2. Run this script: ./test_api.sh" 