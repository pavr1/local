#!/bin/bash

# Ice Cream Store Inventory Service - Test Script
set -e

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

BASE_URL="http://localhost:8082"

echo -e "${CYAN}🍦📦 Testing Ice Cream Store Inventory Service${RESET}"
echo -e "${CYAN}=============================================${RESET}"

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    
    echo -e "${YELLOW}Testing: $description${RESET}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
                   -H "Content-Type: application/json" || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
                   -H "Content-Type: application/json" \
                   -d "$data" || echo "000")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ $method $endpoint - HTTP $http_code${RESET}"
        [ -n "$body" ] && echo "   Response: $body"
    else
        echo -e "${RED}❌ $method $endpoint - HTTP $http_code${RESET}"
        [ -n "$body" ] && echo "   Error: $body"
    fi
    echo ""
}

# Test health endpoint
echo -e "${CYAN}=== Health Check ===${RESET}"
test_endpoint "GET" "/api/v1/inventory/health" "Health check"

# Test root endpoint
echo -e "${CYAN}=== Service Info ===${RESET}"
test_endpoint "GET" "/" "Service information"

# Test suppliers endpoints
echo -e "${CYAN}=== Suppliers Endpoints ===${RESET}"
test_endpoint "GET" "/api/v1/suppliers" "List suppliers"

# Test creating a supplier
test_endpoint "POST" "/api/v1/suppliers" "Create supplier" '{
    "supplier_name": "Test Supplier",
    "contact_number": "123-456-7890",
    "email": "test@supplier.com",
    "address": "123 Test St",
    "notes": "Test supplier for API testing"
}'

# Test ingredients endpoints  
echo -e "${CYAN}=== Ingredients Endpoints ===${RESET}"
test_endpoint "GET" "/api/v1/ingredients" "List ingredients"

# Test creating an ingredient
test_endpoint "POST" "/api/v1/ingredients" "Create ingredient" '{
    "name": "Test Ingredient"
}'

# Test other endpoints (should return "Not implemented yet")
echo -e "${CYAN}=== Other Endpoints (Not Implemented) ===${RESET}"
test_endpoint "GET" "/api/v1/existences" "List existences (not implemented)"
test_endpoint "GET" "/api/v1/runout-reports" "List runout reports (not implemented)"
test_endpoint "GET" "/api/v1/recipe-categories" "List recipe categories (not implemented)"
test_endpoint "GET" "/api/v1/recipes" "List recipes (not implemented)"
test_endpoint "GET" "/api/v1/recipe-ingredients" "List recipe ingredients (not implemented)"

echo -e "${CYAN}=== Test Summary ===${RESET}"
echo -e "${GREEN}✅ Basic API testing completed${RESET}"
echo -e "${YELLOW}💡 Some endpoints return 'Not implemented yet' - this is expected${RESET}"
echo -e "${YELLOW}💡 Suppliers and Ingredients should work fully${RESET}" 