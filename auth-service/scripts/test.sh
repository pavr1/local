#!/bin/bash

# Ice Cream Store Auth Service Test Script
# This script tests the authentication service endpoints

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

BASE_URL="http://localhost:8081"
API_BASE="$BASE_URL/api/v1"

echo "🍦🔐 Testing Ice Cream Store Auth Service..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

# Test results
TESTS_RUN=0
TESTS_PASSED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_code="${3:-200}"
    
    echo -e "${CYAN}🧪 Testing: $test_name${RESET}"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Run the command and capture output and exit code
    set +e
    local output
    local http_code
    
    if [[ "$command" == *"curl"* ]]; then
        # Run curl and capture both response body and HTTP code
        local temp_file="/tmp/auth_test_$$"
        http_code=$(curl -s -w '%{http_code}' -o "$temp_file" $(echo "$command" | sed 's/curl -s -w.*$//' | sed "s/curl -s/curl -s/"))
        output=$(cat "$temp_file" 2>/dev/null || echo "")
        rm -f "$temp_file"
    else
        output=$(eval "$command" 2>&1)
        http_code="N/A"
    fi
    
    local exit_code=$?
    set -e
    
    # Check if test passed
    if [[ "$http_code" == "$expected_code" ]] || [[ "$expected_code" == "0" && "$exit_code" == "0" ]]; then
        echo -e "   ${GREEN}✅ PASSED${RESET}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "   ${RED}❌ FAILED${RESET}"
        echo "   Expected HTTP code: $expected_code"
        echo "   Got HTTP code: $http_code"
        echo "   Exit code: $exit_code"
    fi
}

# Test 1: Health check
run_test "Health Check" \
    "curl -s -X GET '$API_BASE/auth/health'" \
    "200"

# Test 2: Login with admin user (should work if database is properly initialized)
run_test "Admin Login" \
    "curl -s -X POST '$API_BASE/auth/login' -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'" \
    "200"

# Test 3: Login with invalid credentials
run_test "Invalid Login" \
    "curl -s -X POST '$API_BASE/auth/login' -H 'Content-Type: application/json' -d '{\"username\":\"invalid\",\"password\":\"invalid\"}'" \
    "401"

# Test 4: Access protected endpoint without token
run_test "Protected Endpoint Without Token" \
    "curl -s -X GET '$API_BASE/auth/profile'" \
    "401"

# Test 5: Validate token endpoint without token
run_test "Validate Token Without Token" \
    "curl -s -X GET '$API_BASE/auth/validate'" \
    "401"

# Test 6: Try to get a valid token and use it
echo -e "${CYAN}🧪 Testing: Full Auth Flow${RESET}"
TESTS_RUN=$((TESTS_RUN + 1))

# Login and extract token
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin123"}' 2>/dev/null || echo "")

if [[ -n "$LOGIN_RESPONSE" ]] && echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -n "$TOKEN" ]]; then
        # Test using the token
        PROFILE_CODE=$(curl -s -w '%{http_code}' -o /dev/null \
            -H "Authorization: Bearer $TOKEN" \
            "$API_BASE/auth/profile" 2>/dev/null || echo "")
        
        if [[ "$PROFILE_CODE" == "200" ]]; then
            echo -e "   ${GREEN}✅ PASSED${RESET}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            echo "   Successfully obtained and used JWT token"
        else
            echo -e "   ${RED}❌ FAILED${RESET}"
            echo "   Could not use JWT token for protected endpoint"
            echo "   Profile endpoint returned: $PROFILE_CODE"
        fi
    else
        echo -e "   ${RED}❌ FAILED${RESET}"
        echo "   Could not extract token from login response"
    fi
else
    echo -e "   ${RED}❌ FAILED${RESET}"
    echo "   Login failed or invalid response format"
    echo "   Response: $LOGIN_RESPONSE"
fi

# Test 7: Container health check
run_test "Container Health Check" \
    "docker inspect icecream_auth --format='{{.State.Health.Status}}' | grep -q healthy" \
    "0"

echo ""
echo "=========================================="
echo -e "${CYAN}🏁 Test Summary${RESET}"
echo "=========================================="
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $((TESTS_RUN - TESTS_PASSED))"

if [[ $TESTS_PASSED -eq $TESTS_RUN ]]; then
    echo -e "${GREEN}✅ All tests passed!${RESET}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please check the auth service configuration.${RESET}"
    exit 1
fi 