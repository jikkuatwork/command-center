#!/bin/bash

# Command Center v0.2.0 - Authentication Flow Test Script
# Tests the complete authentication and authorization flow

set -e

echo "════════════════════════════════════════════════════════"
echo " Command Center v0.2.0 - Authentication Flow Test"
echo "════════════════════════════════════════════════════════"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Cleanup function
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        info "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
    rm -f /tmp/cc-test-*.db*
    rm -f /tmp/cc-test-config.json
}

trap cleanup EXIT

echo "1. Testing credential setup..."
rm -f /tmp/cc-test-config.json
./cc-server --username testuser --password testpass123 --config /tmp/cc-test-config.json --db /tmp/cc-test-data.db >/dev/null 2>&1
if [ -f /tmp/cc-test-config.json ]; then
    pass "Credential setup creates config file"
    if grep -q "testuser" /tmp/cc-test-config.json; then
        pass "Username saved correctly"
    else
        fail "Username not found in config"
    fi
    if grep -q "password_hash" /tmp/cc-test-config.json; then
        pass "Password hash saved"
    else
        fail "Password hash not found"
    fi
else
    fail "Config file not created"
fi

echo ""
echo "2. Starting test server..."
./cc-server --config /tmp/cc-test-config.json --db /tmp/cc-test-data.db --port 14698 > /dev/null 2>&1 &
SERVER_PID=$!
sleep 3

if ps -p $SERVER_PID > /dev/null; then
    pass "Server started successfully"
else
    fail "Server failed to start"
    exit 1
fi

echo ""
echo "3. Testing authentication..."

# Test login with invalid credentials
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:14698/api/login \
    -H "Content-Type: application/json" \
    -d '{"username":"wrong","password":"wrong"}')
if [ "$HTTP_CODE" = "401" ]; then
    pass "Invalid credentials rejected (401)"
else
    fail "Invalid credentials should return 401, got $HTTP_CODE"
fi

# Test login with valid credentials
RESPONSE=$(curl -s -c /tmp/cookies.txt -X POST http://localhost:14698/api/login \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"testpass123"}')
if echo "$RESPONSE" | grep -q "success"; then
    pass "Valid credentials accepted"
    if [ -f /tmp/cookies.txt ]; then
        pass "Session cookie set"
    else
        fail "Session cookie not set"
    fi
else
    fail "Login failed with valid credentials"
fi

echo ""
echo "4. Testing protected endpoints..."

# Test dashboard without auth
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:14698/)
if [ "$HTTP_CODE" = "303" ] || [ "$HTTP_CODE" = "307" ] || [ "$HTTP_CODE" = "302" ]; then
    pass "Dashboard redirects without auth"
else
    fail "Dashboard should redirect without auth, got $HTTP_CODE"
fi

# Test dashboard with auth cookie
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -b /tmp/cookies.txt http://localhost:14698/)
if [ "$HTTP_CODE" = "200" ]; then
    pass "Dashboard accessible with valid session"
else
    fail "Dashboard should be accessible with session, got $HTTP_CODE"
fi

echo ""
echo "5. Testing public endpoints..."

# Test tracking endpoint (should be public)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:14698/track \
    -H "Content-Type: application/json" \
    -d '{"h":"test.com","p":"/","e":"pageview"}')
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    pass "Tracking endpoint public (no auth required)"
else
    fail "Tracking endpoint should be public, got $HTTP_CODE"
fi

# Test pixel endpoint (should be public)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:14698/pixel.gif)
if [ "$HTTP_CODE" = "200" ]; then
    pass "Pixel endpoint public (no auth required)"
else
    fail "Pixel endpoint should be public, got $HTTP_CODE"
fi

# Test health endpoint
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:14698/health)
if [ "$HTTP_CODE" = "200" ]; then
    pass "Health check endpoint working"
else
    fail "Health check failed, got $HTTP_CODE"
fi

echo ""
echo "6. Testing logout..."

# Logout
RESPONSE=$(curl -s -b /tmp/cookies.txt -X POST http://localhost:14698/api/logout)
if echo "$RESPONSE" | grep -q "success"; then
    pass "Logout successful"
else
    fail "Logout failed"
fi

# Try accessing dashboard after logout
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -b /tmp/cookies.txt http://localhost:14698/)
if [ "$HTTP_CODE" = "303" ] || [ "$HTTP_CODE" = "307" ] || [ "$HTTP_CODE" = "302" ]; then
    pass "Dashboard redirects after logout"
else
    fail "Dashboard should redirect after logout, got $HTTP_CODE"
fi

echo ""
echo "════════════════════════════════════════════════════════"
echo " Test Results"
echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
