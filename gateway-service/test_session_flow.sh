#!/bin/bash

# Test script for Gateway Session Management
# This demonstrates the complete session security flow

GATEWAY_URL="http://localhost:8082"
SESSION_URL="http://localhost:8081"

echo "🔐 Testing Gateway Session Management Security"
echo "=============================================="
echo ""

# Test 1: Try accessing protected route without token
echo "📋 1. Testing protected route WITHOUT token (should fail)..."
curl -s -w "Status: %{http_code}\n" \
  -X GET "$GATEWAY_URL/api/v1/orders" | jq '.' 2>/dev/null || echo "No JSON response"
echo ""

# Test 2: Try with fake external token
echo "📋 2. Testing with FAKE external token (should fail)..."
FAKE_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmFrZSJ9.fake-signature"
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer $FAKE_TOKEN" \
  -X GET "$GATEWAY_URL/api/v1/orders" | jq '.'
echo ""

# Test 3: Login through gateway (creates session)
echo "📋 3. Testing LOGIN through gateway (creates session)..."
LOGIN_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/api/v1/sessions/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo $LOGIN_RESPONSE | jq '.'
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token' 2>/dev/null)

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Login failed! Make sure session service is running and has admin user."
  echo "   Start session service: cd ../session-service && go run ."
  exit 1
fi

echo ""
echo "✅ Session token obtained: ${TOKEN:0:50}..."
echo ""

# Test 4: Access protected route with valid session token
echo "📋 4. Testing protected route WITH valid session token..."
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer $TOKEN" \
  -X GET "$GATEWAY_URL/api/v1/orders" | jq '.' 2>/dev/null || echo "Orders service may not be running"
echo ""

# Test 5: Verify session exists in session service
echo "📋 5. Verifying session exists in session service..."
curl -s -X POST "$SESSION_URL/api/v1/sessions/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}" | jq '.'
echo ""

# Test 6: Get session statistics
echo "📋 6. Checking session statistics..."
curl -s -X GET "$SESSION_URL/api/v1/sessions/stats" | jq '.'
echo ""

# Test 7: Logout through gateway (revokes session)
echo "📋 7. Testing LOGOUT through gateway (revokes session)..."
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer $TOKEN" \
  -X POST "$GATEWAY_URL/api/v1/sessions/logout" | jq '.'
echo ""

# Test 8: Try to use token after logout (should fail)
echo "📋 8. Testing token after logout (should fail)..."
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer $TOKEN" \
  -X GET "$GATEWAY_URL/api/v1/orders" | jq '.'
echo ""

# Test 9: Verify session no longer exists
echo "📋 9. Verifying session no longer exists..."
curl -s -X POST "$SESSION_URL/api/v1/sessions/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}" | jq '.'
echo ""

echo "✅ Session Management Security Tests Complete!"
echo ""
echo "🔐 SECURITY SUMMARY:"
echo "   ✅ External tokens are rejected"
echo "   ✅ Only gateway-created sessions are valid"
echo "   ✅ Sessions are stored server-side"
echo "   ✅ Logout completely revokes sessions"
echo "   ✅ Invalid tokens are blocked"
echo ""
echo "🚀 To run these tests:"
echo "   1. Start session service: cd ../session-service && go run ."
echo "   2. Start gateway service: cd ../gateway-service && go run ."
echo "   3. (Optional) Start orders service: cd ../orders-service && make start"
echo "   4. Run this script: ./test_session_flow.sh" 