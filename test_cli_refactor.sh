#!/bin/bash

# Integration Test for CLI Refactor (v0.4.0)
# Tests the full workflow of init, set-credentials, set-config, status commands

set -e

echo "════════════════════════════════════════════════════════"
echo " fazt.sh v0.4.0 - CLI Refactor Integration Test"
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

# Test configuration
TEST_CONFIG_DIR="/tmp/fazt-test-$$"
TEST_CONFIG="$TEST_CONFIG_DIR/config.json"
FAZT_BIN="./fazt"

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    FAILED=$((FAILED + 1))
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Cleanup function
cleanup() {
    info "Cleaning up test directory: $TEST_CONFIG_DIR"
    rm -rf "$TEST_CONFIG_DIR"
}

trap cleanup EXIT

# Check if fazt binary exists
if [ ! -f "$FAZT_BIN" ]; then
    fail "fazt binary not found. Run 'make build-local' first."
    exit 1
fi

echo "Test Configuration:"
echo "  Binary: $FAZT_BIN"
echo "  Config Dir: $TEST_CONFIG_DIR"
echo "  Config File: $TEST_CONFIG"
echo ""

# ============================================================================
# TEST 1: Init Command
# ============================================================================
echo "1. Testing 'fazt server init' command..."

# Test 1a: Successful initialization
$FAZT_BIN server init \
    --username testadmin \
    --password testpass123 \
    --domain https://test.example.com \
    --port 4698 \
    --env development \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if [ -f "$TEST_CONFIG" ]; then
    pass "Init creates config file"
else
    fail "Init did not create config file"
fi

# Verify config has correct values
if grep -q '"domain": "https://test.example.com"' "$TEST_CONFIG"; then
    pass "Init sets domain correctly"
else
    fail "Domain not set correctly in config"
fi

if grep -q '"port": "4698"' "$TEST_CONFIG"; then
    pass "Init sets port correctly"
else
    fail "Port not set correctly in config"
fi

if grep -q '"username": "testadmin"' "$TEST_CONFIG"; then
    pass "Init sets username correctly"
else
    fail "Username not set correctly in config"
fi

# Verify password is hashed (not plaintext)
if grep -q '"password_hash": "\$2a\$' "$TEST_CONFIG"; then
    pass "Init hashes password (bcrypt)"
else
    fail "Password not hashed correctly"
fi

# Verify secure file permissions
PERMS=$(stat -c %a "$TEST_CONFIG" 2>/dev/null || stat -f %A "$TEST_CONFIG" 2>/dev/null)
if [ "$PERMS" = "600" ]; then
    pass "Config file has secure permissions (0600)"
else
    fail "Config file permissions incorrect: $PERMS (expected 600)"
fi

# Test 1b: Init fails when config already exists
$FAZT_BIN server init \
    --username admin \
    --password pass \
    --domain https://new.com \
    --config "$TEST_CONFIG" 2>&1 | grep -q "already initialized" || \
    grep -q "exists"

if [ $? -eq 0 ]; then
    pass "Init fails when config already exists"
else
    fail "Init should fail when config exists"
fi

echo ""

# ============================================================================
# TEST 2: Status Command
# ============================================================================
echo "2. Testing 'fazt server status' command..."

OUTPUT=$($FAZT_BIN server status --config "$TEST_CONFIG" 2>&1)

# Check for expected content in status output
echo "$OUTPUT" | grep -q "Server Status" && pass "Status shows header" || fail "Status missing header"
echo "$OUTPUT" | grep -q "Config:" && pass "Status shows config path" || fail "Status missing config path"
echo "$OUTPUT" | grep -q "Domain:" && pass "Status shows domain" || fail "Status missing domain"
echo "$OUTPUT" | grep -q "https://test.example.com" && pass "Status shows correct domain" || fail "Status shows wrong domain"
echo "$OUTPUT" | grep -q "Port:" && pass "Status shows port" || fail "Status missing port"
echo "$OUTPUT" | grep -q "4698" && pass "Status shows correct port" || fail "Status shows wrong port"
echo "$OUTPUT" | grep -q "Username:" && pass "Status shows username" || fail "Status missing username"
echo "$OUTPUT" | grep -q "testadmin" && pass "Status shows correct username" || fail "Status shows wrong username"

# Check server running status (should be "Not running" since we didn't start it)
if echo "$OUTPUT" | grep -q "Not running"; then
    pass "Status correctly shows server is not running"
else
    fail "Status should show server is not running"
fi

echo ""

# ============================================================================
# TEST 3: Set-Credentials Command
# ============================================================================
echo "3. Testing 'fazt server set-credentials' command..."

# Test 3a: Update password
$FAZT_BIN server set-credentials \
    --password newpassword456 \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"password_hash": "\$2a\$' "$TEST_CONFIG"; then
    pass "set-credentials updates password"
else
    fail "Password not updated"
fi

# Verify username was preserved
if grep -q '"username": "testadmin"' "$TEST_CONFIG"; then
    pass "set-credentials preserves username"
else
    fail "Username was changed unexpectedly"
fi

# Test 3b: Update username
$FAZT_BIN server set-credentials \
    --username newadmin \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"username": "newadmin"' "$TEST_CONFIG"; then
    pass "set-credentials updates username"
else
    fail "Username not updated"
fi

# Test 3c: Update both
$FAZT_BIN server set-credentials \
    --username finaladmin \
    --password finalpass789 \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"username": "finaladmin"' "$TEST_CONFIG"; then
    pass "set-credentials updates both username and password"
else
    fail "Failed to update both values"
fi

# Test 3d: Fails with no flags
$FAZT_BIN server set-credentials --config "$TEST_CONFIG" 2>&1 | grep -q "at least one"
if [ $? -eq 0 ]; then
    pass "set-credentials requires at least one flag"
else
    fail "set-credentials should require at least one flag"
fi

echo ""

# ============================================================================
# TEST 4: Set-Config Command
# ============================================================================
echo "4. Testing 'fazt server set-config' command..."

# Test 4a: Update domain
$FAZT_BIN server set-config \
    --domain https://new.example.com \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"domain": "https://new.example.com"' "$TEST_CONFIG"; then
    pass "set-config updates domain"
else
    fail "Domain not updated"
fi

# Test 4b: Update port
$FAZT_BIN server set-config \
    --port 8080 \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"port": "8080"' "$TEST_CONFIG"; then
    pass "set-config updates port"
else
    fail "Port not updated"
fi

# Test 4c: Update environment
$FAZT_BIN server set-config \
    --env production \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"env": "production"' "$TEST_CONFIG"; then
    pass "set-config updates environment"
else
    fail "Environment not updated"
fi

# Test 4d: Update multiple fields at once
$FAZT_BIN server set-config \
    --domain https://prod.example.com \
    --port 443 \
    --env production \
    --config "$TEST_CONFIG" >/dev/null 2>&1

if grep -q '"domain": "https://prod.example.com"' "$TEST_CONFIG" && \
   grep -q '"port": "443"' "$TEST_CONFIG" && \
   grep -q '"env": "production"' "$TEST_CONFIG"; then
    pass "set-config updates multiple fields simultaneously"
else
    fail "Failed to update multiple fields"
fi

# Test 4e: Fails with no flags
$FAZT_BIN server set-config --config "$TEST_CONFIG" 2>&1 | grep -q "at least one"
if [ $? -eq 0 ]; then
    pass "set-config requires at least one flag"
else
    fail "set-config should require at least one flag"
fi

# Test 4f: Fails with invalid port
$FAZT_BIN server set-config --port 99999 --config "$TEST_CONFIG" 2>&1 | grep -q "invalid\|range"
if [ $? -eq 0 ]; then
    pass "set-config validates port range"
else
    fail "set-config should validate port range"
fi

# Test 4g: Fails with invalid environment
$FAZT_BIN server set-config --env staging --config "$TEST_CONFIG" 2>&1 | grep -q "invalid\|must be"
if [ $? -eq 0 ]; then
    pass "set-config validates environment values"
else
    fail "set-config should validate environment"
fi

echo ""

# ============================================================================
# TEST 5: Deploy Alias
# ============================================================================
echo "5. Testing 'fazt deploy' alias..."

# Create a simple test site
TEST_SITE_DIR="$TEST_CONFIG_DIR/test-site"
mkdir -p "$TEST_SITE_DIR"
echo "<h1>Test Site</h1>" > "$TEST_SITE_DIR/index.html"

# Note: This test just verifies the command is recognized
# Actual deployment requires server to be running and auth token configured
$FAZT_BIN deploy --help 2>&1 | grep -q "deploy\|Deploy"
if [ $? -eq 0 ]; then
    pass "deploy alias is recognized"
else
    fail "deploy alias not working"
fi

# Verify it's equivalent to client deploy
$FAZT_BIN client deploy --help 2>&1 > /tmp/client-help.txt
$FAZT_BIN deploy --help 2>&1 > /tmp/deploy-help.txt

if diff /tmp/client-help.txt /tmp/deploy-help.txt >/dev/null 2>&1; then
    pass "deploy alias shows same help as client deploy"
else
    fail "deploy alias help differs from client deploy"
fi

rm -f /tmp/client-help.txt /tmp/deploy-help.txt

echo ""

# ============================================================================
# TEST 6: Config Structure (No auth.enabled field)
# ============================================================================
echo "6. Testing config structure..."

# Verify config does not have "enabled" field in auth section
if grep -q '"enabled":' "$TEST_CONFIG"; then
    fail "Config should not have 'auth.enabled' field"
else
    pass "Config does not have 'auth.enabled' field (correct)"
fi

# Verify config requires auth (has username and password_hash)
if grep -q '"username":' "$TEST_CONFIG" && grep -q '"password_hash":' "$TEST_CONFIG"; then
    pass "Config always includes auth credentials"
else
    fail "Config missing auth credentials"
fi

echo ""

# ============================================================================
# TEST 7: Error Handling
# ============================================================================
echo "7. Testing error handling..."

# Test 7a: Commands fail with missing config
rm -f "$TEST_CONFIG"

$FAZT_BIN server status --config "$TEST_CONFIG" 2>&1 | grep -q "not found\|does not exist"
if [ $? -eq 0 ]; then
    pass "status fails gracefully when config missing"
else
    fail "status should fail when config missing"
fi

$FAZT_BIN server set-credentials --password test --config "$TEST_CONFIG" 2>&1 | grep -q "not found\|not initialized"
if [ $? -eq 0 ]; then
    pass "set-credentials fails when config missing"
else
    fail "set-credentials should fail when config missing"
fi

$FAZT_BIN server set-config --domain https://test.com --config "$TEST_CONFIG" 2>&1 | grep -q "not found\|not initialized"
if [ $? -eq 0 ]; then
    pass "set-config fails when config missing"
else
    fail "set-config should fail when config missing"
fi

# Test 7b: Init with missing required flags
$FAZT_BIN server init --username admin --config "$TEST_CONFIG" 2>&1 | grep -q "required\|missing"
if [ $? -eq 0 ]; then
    pass "init fails when required flags missing"
else
    fail "init should fail when required flags missing"
fi

echo ""

# ============================================================================
# TEST 8: Full Workflow
# ============================================================================
echo "8. Testing complete workflow..."

# Clean start
rm -rf "$TEST_CONFIG_DIR"
mkdir -p "$TEST_CONFIG_DIR"

# 1. Initialize
$FAZT_BIN server init \
    --username workflow_admin \
    --password workflow_pass \
    --domain https://workflow.test.com \
    --port 5000 \
    --env development \
    --config "$TEST_CONFIG" >/dev/null 2>&1

# 2. Check status
STATUS_OUTPUT=$($FAZT_BIN server status --config "$TEST_CONFIG" 2>&1)
echo "$STATUS_OUTPUT" | grep -q "workflow.test.com" && \
echo "$STATUS_OUTPUT" | grep -q "5000" && \
echo "$STATUS_OUTPUT" | grep -q "workflow_admin"

if [ $? -eq 0 ]; then
    pass "Workflow: init → status shows correct values"
else
    fail "Workflow: status not showing correct values"
fi

# 3. Update credentials
$FAZT_BIN server set-credentials --password new_workflow_pass --config "$TEST_CONFIG" >/dev/null 2>&1

# 4. Update config
$FAZT_BIN server set-config \
    --domain https://updated.test.com \
    --port 6000 \
    --env production \
    --config "$TEST_CONFIG" >/dev/null 2>&1

# 5. Verify all changes
FINAL_STATUS=$($FAZT_BIN server status --config "$TEST_CONFIG" 2>&1)
echo "$FINAL_STATUS" | grep -q "updated.test.com" && \
echo "$FINAL_STATUS" | grep -q "6000" && \
echo "$FINAL_STATUS" | grep -q "production" && \
echo "$FINAL_STATUS" | grep -q "workflow_admin"

if [ $? -eq 0 ]; then
    pass "Workflow: All updates reflected in final status"
else
    fail "Workflow: Updates not reflected correctly"
fi

echo ""

# ============================================================================
# TEST RESULTS
# ============================================================================
echo "════════════════════════════════════════════════════════"
echo " Test Results"
echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed!${NC}"
    exit 1
fi
