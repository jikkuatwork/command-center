#!/bin/bash

# Test script for webhook handler
BASE_URL="http://localhost:4698"

echo "Testing Command Center Webhook Handler"
echo "======================================="
echo ""

# Note: Mock data includes these webhooks:
# - github-deploy (with secret)
# - stripe-events (with secret)
# - custom-hook (with secret)
# - cicd-status (with secret)
# - monitoring (with secret)

echo "=== Webhook Tests ==="
echo ""

# Test 1: Valid webhook without secret requirement (will fail if all have secrets)
echo "Test 1: GitHub deployment webhook"
curl -X POST "$BASE_URL/webhook/github-deploy" \
  -H "Content-Type: application/json" \
  -d '{"event":"deploy","project":"my-site","status":"success","branch":"main"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 2: Custom integration webhook
echo "Test 2: Custom integration webhook"
curl -X POST "$BASE_URL/webhook/custom-hook" \
  -H "Content-Type: application/json" \
  -d '{"type":"user_signup","user_id":12345,"email":"test@example.com"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 3: CI/CD status webhook
echo "Test 3: CI/CD pipeline webhook"
curl -X POST "$BASE_URL/webhook/cicd-status" \
  -H "Content-Type: application/json" \
  -d '{"event":"build","status":"success","duration":245,"commit":"abc123"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 4: Monitoring alert webhook
echo "Test 4: Monitoring alert webhook"
curl -X POST "$BASE_URL/webhook/monitoring" \
  -H "Content-Type: application/json" \
  -d '{"event":"alert","severity":"high","message":"CPU usage above 90%","timestamp":"2025-11-10T20:00:00Z"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 5: Invalid endpoint (should return 404)
echo "Test 5: Invalid webhook endpoint (should fail)"
curl -X POST "$BASE_URL/webhook/nonexistent" \
  -H "Content-Type: application/json" \
  -d '{"event":"test"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 6: Empty endpoint (should return 400)
echo "Test 6: Empty webhook endpoint (should fail)"
curl -X POST "$BASE_URL/webhook/" \
  -H "Content-Type: application/json" \
  -d '{"event":"test"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 7: GET method (should return 405)
echo "Test 7: GET method on webhook (should fail)"
curl -X GET "$BASE_URL/webhook/github-deploy" \
  -w "\nStatus: %{http_code}\n\n"

# Test 8: Non-JSON payload
echo "Test 8: Non-JSON payload (should still work)"
curl -X POST "$BASE_URL/webhook/custom-hook" \
  -H "Content-Type: text/plain" \
  -d 'This is a plain text webhook payload' \
  -w "\nStatus: %{http_code}\n\n"

echo ""
echo "Testing complete!"
echo ""
echo "Note: Webhooks with secrets will require X-Webhook-Signature header."
echo "To test with signature, compute HMAC SHA256 of the body with the secret."
echo ""
echo "Check webhook events in database:"
echo "  SELECT COUNT(*) FROM events WHERE source_type='webhook'"
echo "  SELECT domain, event_type, query_params FROM events WHERE source_type='webhook' LIMIT 5"
