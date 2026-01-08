#!/bin/bash
# Basic integration test script
# Tests key endpoints through the Gateway

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
ACCESS_TOKEN=""

echo "=== Integration Tests ==="
echo ""

# Test health checks
echo "1. Testing health checks..."
curl -f -s "$GATEWAY_URL/health" > /dev/null && echo "✓ Gateway health check passed" || echo "✗ Gateway health check failed"

# Test authentication (if credentials are provided)
if [ -n "$TEST_EMAIL" ] && [ -n "$TEST_PASSWORD" ]; then
    echo ""
    echo "2. Testing authentication..."
    RESPONSE=$(curl -s -X POST "$GATEWAY_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    
    ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$ACCESS_TOKEN" ]; then
        echo "✓ Authentication successful"
        echo "  Token: ${ACCESS_TOKEN:0:20}..."
    else
        echo "✗ Authentication failed"
        echo "  Response: $RESPONSE"
    fi
fi

# Test protected endpoints (if token is available)
if [ -n "$ACCESS_TOKEN" ]; then
    echo ""
    echo "3. Testing protected endpoints..."
    
    # Test inventory endpoints
    curl -f -s -X GET "$GATEWAY_URL/api/brands" \
        -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null && \
        echo "✓ Inventory Service (brands) accessible" || \
        echo "✗ Inventory Service (brands) failed"
    
    # Test geography endpoints
    curl -f -s -X GET "$GATEWAY_URL/api/countries" \
        -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null && \
        echo "✓ Geography Service (countries) accessible" || \
        echo "✗ Geography Service (countries) failed"
    
    # Test navigation endpoints
    curl -f -s -X GET "$GATEWAY_URL/api/menus" \
        -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null && \
        echo "✓ Navigation Service (menus) accessible" || \
        echo "✗ Navigation Service (menus) failed"
    
    # Test help desk endpoints
    curl -f -s -X GET "$GATEWAY_URL/api/help-desk" \
        -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null && \
        echo "✓ Help Desk Service (tickets) accessible" || \
        echo "✗ Help Desk Service (tickets) failed"
fi

echo ""
echo "=== Integration Tests Complete ==="
echo ""
echo "Note: For comprehensive testing, see extra/integration_test_guide.md"

