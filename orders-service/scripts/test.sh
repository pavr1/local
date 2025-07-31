#!/bin/bash

# Ice Cream Store Orders Service Test Script
# This script runs comprehensive API tests for the orders service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8083"
API_BASE="$BASE_URL/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

echo -e "${CYAN}🍦📦 Testing Ice Cream Store Orders Service${RESET}"
echo "============================================="
echo ""

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "Testing $test_name... "
    
    if response=$(eval "$test_command" 2>/dev/null); then
        if [[ -z "$expected_pattern" ]] || echo "$response" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✅ PASS${RESET}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}❌ FAIL${RESET} (unexpected response)"
            echo "   Expected pattern: $expected_pattern"
            echo "   Got: $response"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${RESET} (request failed)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Helper function to get auth token (requires auth service)
get_auth_token() {
    if ! curl -f http://localhost:8081/api/v1/sessions/health > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Auth service not available - skipping authenticated tests${RESET}"
        return 1
    fi
    
    local response=$(curl -s -X POST http://localhost:8081/api/v1/sessions/login \
        -H 'Content-Type: application/json' \
        -d '{"username":"admin","password":"admin123"}' 2>/dev/null)
    
    if echo "$response" | grep -q '"token"'; then
        echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
        return 0
    else
        echo -e "${YELLOW}⚠️  Could not get auth token - skipping authenticated tests${RESET}"
        return 1
    fi
}

echo -e "${CYAN}📊 Basic Connectivity Tests${RESET}"
echo "----------------------------"

# Test 1: Health Check
run_test "Health Check" \
    "curl -f $API_BASE/orders/health" \
    '"status":"healthy"'

# Test 2: Root endpoint
run_test "Root Endpoint" \
    "curl -f $BASE_URL/" \
    "ice-cream-orders-service"

echo ""
echo -e "${CYAN}🔐 Authentication Tests${RESET}"
echo "------------------------"

# Get auth token for authenticated tests
if AUTH_TOKEN=$(get_auth_token); then
    AUTH_HEADER="Authorization: Bearer $AUTH_TOKEN"
    
    # Test 3: Get orders (authenticated)
    run_test "Get Orders List (Authenticated)" \
        "curl -f -H '$AUTH_HEADER' $API_BASE/orders" \
        '"orders"'
    
    # Test 4: Create order (authenticated)
    run_test "Create Order (Authenticated)" \
        "curl -f -X POST -H 'Content-Type: application/json' -H '$AUTH_HEADER' $API_BASE/orders -d '{\"customer_id\":1,\"order_type\":\"dine_in\",\"ordered_recipes\":[{\"recipe_id\":1,\"quantity\":2}]}'" \
        '"order_id"'
    
    # Test 5: Get order by ID (authenticated)
    run_test "Get Order by ID (Authenticated)" \
        "curl -f -H '$AUTH_HEADER' $API_BASE/orders/1" \
        '"order_id"'
    
else
    echo -e "${YELLOW}⚠️  Skipping authenticated tests - auth service not available${RESET}"
    TOTAL_TESTS=$((TOTAL_TESTS + 3))
fi

echo ""
echo -e "${CYAN}🚫 Error Handling Tests${RESET}"
echo "------------------------"

# Test 6: Invalid endpoint
run_test "Invalid Endpoint (404)" \
    "curl -s -o /dev/null -w '%{http_code}' $API_BASE/invalid" \
    "404"

# Test 7: Unauthorized access
run_test "Unauthorized Access (401)" \
    "curl -s -o /dev/null -w '%{http_code}' $API_BASE/orders" \
    "401"

# Test 8: Invalid order ID
run_test "Invalid Order ID (404)" \
    "curl -s -o /dev/null -w '%{http_code}' -H '$AUTH_HEADER' $API_BASE/orders/99999" \
    "404"

echo ""
echo -e "${CYAN}📈 Performance Tests${RESET}"
echo "---------------------"

# Test 9: Response time check
run_test "Response Time < 1s" \
    "time_ms=\$(curl -o /dev/null -s -w '%{time_total}' $API_BASE/orders/health | awk '{print \$1*1000}'); [ \${time_ms%.*} -lt 1000 ] && echo 'fast'" \
    "fast"

echo ""
echo "============================================="
echo -e "${CYAN}📊 Test Results Summary${RESET}"
echo "============================================="
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! ($TESTS_PASSED/$TOTAL_TESTS)${RESET}"
    echo ""
    echo -e "${CYAN}📝 Orders Service is fully functional!${RESET}"
    echo ""
    echo -e "${YELLOW}🔗 Available Endpoints:${RESET}"
    echo "   Health:      GET  $API_BASE/orders/health"
    echo "   List Orders: GET  $API_BASE/orders (requires auth)"
    echo "   Get Order:   GET  $API_BASE/orders/{id} (requires auth)"
    echo "   Create Order: POST $API_BASE/orders (requires auth)"
    echo ""
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED ($TESTS_FAILED/$TOTAL_TESTS failed)${RESET}"
    echo ""
    echo -e "${YELLOW}🔍 Troubleshooting:${RESET}"
    echo "   • Check if orders service is running: docker-compose ps"
    echo "   • Check service logs: ./scripts/logs.sh"
    echo "   • Verify database connection: make health"
    echo "   • Ensure auth service is running for authenticated tests"
    echo ""
    exit 1
fi 