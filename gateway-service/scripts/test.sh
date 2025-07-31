#!/bin/bash

# Ice Cream Store Gateway Service - Test Script
# This script runs comprehensive tests against the gateway service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
GATEWAY_URL="http://localhost:8082"
TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

# Helper functions
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    echo -e "${BLUE}🧪 Test $TEST_COUNT: $test_name${NC}"
    
    if result=$(eval "$test_command" 2>&1); then
        if [[ -z "$expected_pattern" ]] || echo "$result" | grep -q "$expected_pattern"; then
            echo -e "   ${GREEN}✅ PASS${NC}"
            PASS_COUNT=$((PASS_COUNT + 1))
            return 0
        else
            echo -e "   ${RED}❌ FAIL - Expected pattern '$expected_pattern' not found${NC}"
            echo -e "   ${YELLOW}Response: $result${NC}"
            FAIL_COUNT=$((FAIL_COUNT + 1))
            return 1
        fi
    else
        echo -e "   ${RED}❌ FAIL - Command failed${NC}"
        echo -e "   ${YELLOW}Error: $result${NC}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi
}

echo -e "${CYAN}🧪 Starting Gateway Service Tests...${NC}"
echo "=================================="

# Test 1: Basic connectivity
run_test "Basic Connectivity" \
    "curl -s -w '%{http_code}' -o /dev/null $GATEWAY_URL/api/health" \
    "200"

# Test 2: Health endpoint response
run_test "Health Endpoint Response" \
    "curl -s $GATEWAY_URL/api/health" \
    "healthy"

# Test 3: Health check via proxy (if auth service is up)
run_test "Auth Service Proxy Health" \
    "curl -s -w '%{http_code}' -o /dev/null $GATEWAY_URL/api/v1/sessions/health" \
    "200"

# Test 4: Orders service proxy health (if orders service is up)
run_test "Orders Service Proxy Health" \
    "curl -s -w '%{http_code}' -o /dev/null $GATEWAY_URL/api/v1/orders/health" \
    "200"

# Test 5: CORS headers
run_test "CORS Headers" \
    "curl -s -I -X OPTIONS $GATEWAY_URL/api/health" \
    "Access-Control-Allow-Origin"

# Test 6: Gateway health endpoint
run_test "Gateway Health Endpoint" \
    "curl -s $GATEWAY_URL/api/health" \
    "status"

# Test 7: Invalid endpoint (should return 404)
run_test "Invalid Endpoint (404)" \
    "curl -s -w '%{http_code}' -o /dev/null $GATEWAY_URL/api/invalid" \
    "404"

# Test 8: Health endpoint format check
run_test "Health Endpoint Format" \
    "curl -s $GATEWAY_URL/api/health" \
    "gateway"

# Test 9: Gateway service container health
if docker ps | grep -q "icecream_gateway"; then
    health_status=$(docker inspect icecream_gateway --format='{{.State.Health.Status}}' 2>/dev/null || echo "no-healthcheck")
    if [[ "$health_status" == "healthy" || "$health_status" == "starting" ]]; then
        echo -e "${BLUE}🧪 Test 9: Container Health Check${NC}"
        echo -e "   ${GREEN}✅ PASS${NC} (Status: $health_status)"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${BLUE}🧪 Test 9: Container Health Check${NC}"
        echo -e "   ${RED}❌ FAIL${NC} (Status: $health_status)"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    TEST_COUNT=$((TEST_COUNT + 1))
else
    echo -e "${YELLOW}⚠️  Skipping container health check - container not found${NC}"
fi

# Test 10: Service discovery test
run_test "Service Discovery" \
    "curl -s $GATEWAY_URL/api/health" \
    "operational"

echo ""
echo "=================================="
echo -e "${CYAN}📊 Test Results Summary${NC}"
echo "=================================="
echo -e "Total Tests: ${BLUE}$TEST_COUNT${NC}"
echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
echo -e "Failed: ${RED}$FAIL_COUNT${NC}"

if [ $FAIL_COUNT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed! Gateway Service is working correctly.${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please check the service configuration.${NC}"
    exit 1
fi 