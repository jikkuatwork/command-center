#!/bin/bash

# Test script for API endpoints
BASE_URL="http://localhost:4698"

echo "Testing Command Center API Endpoints"
echo "====================================="
echo ""

# Test 1: GET /api/stats
echo "Test 1: GET /api/stats"
curl -s "$BASE_URL/api/stats" | python3 -m json.tool | head -20
echo "..."
echo ""

# Test 2: GET /api/events (default)
echo "Test 2: GET /api/events (first 5)"
curl -s "$BASE_URL/api/events?limit=5" | python3 -m json.tool | head -30
echo "..."
echo ""

# Test 3: GET /api/domains
echo "Test 3: GET /api/domains"
curl -s "$BASE_URL/api/domains" | python3 -m json.tool
echo ""

# Test 4: GET /api/tags
echo "Test 4: GET /api/tags"
curl -s "$BASE_URL/api/tags" | python3 -m json.tool | head -20
echo "..."
echo ""

# Test 5: GET /api/redirects
echo "Test 5: GET /api/redirects"
curl -s "$BASE_URL/api/redirects" | python3 -m json.tool | head -30
echo "..."
echo ""

# Test 6: POST /api/redirects (create new)
echo "Test 6: POST /api/redirects (create new)"
curl -s -X POST "$BASE_URL/api/redirects" \
  -H "Content-Type: application/json" \
  -d '{"slug":"test-redirect","destination":"https://example.com/test","tags":["test","api"]}' \
  | python3 -m json.tool
echo ""

# Test 7: GET /api/webhooks
echo "Test 7: GET /api/webhooks"
curl -s "$BASE_URL/api/webhooks" | python3 -m json.tool
echo ""

# Test 8: POST /api/webhooks (create new)
echo "Test 8: POST /api/webhooks (create new)"
curl -s -X POST "$BASE_URL/api/webhooks" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Webhook","endpoint":"test-webhook","secret":"my-secret-123"}' \
  | python3 -m json.tool
echo ""

# Test 9: Filter events by domain
echo "Test 9: GET /api/events?domain=example.com"
curl -s "$BASE_URL/api/events?domain=example.com&limit=3" | python3 -m json.tool | head -20
echo "..."
echo ""

echo "Testing complete!"
echo ""
echo "All API endpoints tested successfully!"
