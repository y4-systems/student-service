#!/bin/bash

# Rate Limiter Manual Test Script
# This script simulates multiple login attempts to test rate limiting

echo "=================================================="
echo "   Rate Limiter Manual Test Script"
echo "=================================================="
echo ""

# Configuration
BASE_URL="${1:-http://localhost:5001}"
LOGIN_ENDPOINT="${BASE_URL}/auth/login"
NUM_REQUESTS=10

echo "🎯 Target: $LOGIN_ENDPOINT"
echo "📊 Number of requests: $NUM_REQUESTS"
echo "⏱️  Rate limit: 5 requests per minute per IP"
echo ""

# Test credentials (will fail authentication, but that's ok - we're testing rate limiting)
TEST_EMAIL="test@example.com"
TEST_PASSWORD="password123"

echo "Starting rapid login attempts..."
echo ""

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0
OTHER_COUNT=0

for i in $(seq 1 $NUM_REQUESTS); do
    echo -n "Request $i: "
    
    # Make the request and capture status code
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$LOGIN_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
        2>/dev/null)
    
    # Extract status code (last line)
    STATUS_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | head -n -1)
    
    # Check the status code
    if [ "$STATUS_CODE" = "429" ]; then
        echo "❌ RATE LIMITED (429 Too Many Requests)"
        RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
    elif [ "$STATUS_CODE" = "401" ]; then
        echo "✅ REQUEST PROCESSED (401 Unauthorized - expected)"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    elif [ "$STATUS_CODE" = "400" ]; then
        echo "✅ REQUEST PROCESSED (400 Bad Request)"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo "⚠️  UNEXPECTED ($STATUS_CODE)"
        OTHER_COUNT=$((OTHER_COUNT + 1))
    fi
    
    # Small delay between requests
    sleep 0.1
done

echo ""
echo "=================================================="
echo "   Test Results"
echo "=================================================="
echo "✅ Processed requests: $SUCCESS_COUNT"
echo "❌ Rate limited: $RATE_LIMITED_COUNT"
echo "⚠️  Other responses: $OTHER_COUNT"
echo ""

# Verify rate limiting is working
if [ $RATE_LIMITED_COUNT -gt 0 ]; then
    echo "✅ SUCCESS: Rate limiting is working!"
    echo "   Expected behavior: First request allowed, subsequent rapid requests blocked."
else
    echo "⚠️  WARNING: No rate limiting detected!"
    echo "   This might indicate:"
    echo "   1. Service is not running at $BASE_URL"
    echo "   2. Rate limit configuration is too permissive"
    echo "   3. Rate limiter is not properly initialized"
fi

echo ""
echo "=================================================="
echo "Testing rate limit recovery..."
echo "=================================================="
echo ""
echo "⏳ Waiting 13 seconds for rate limit to reset..."

for i in {13..1}; do
    echo -n "$i... "
    sleep 1
done
echo ""
echo ""

echo -n "Request after wait: "
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$LOGIN_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
    2>/dev/null)

STATUS_CODE=$(echo "$RESPONSE" | tail -n 1)

if [ "$STATUS_CODE" = "429" ]; then
    echo "❌ STILL RATE LIMITED (unexpected)"
elif [ "$STATUS_CODE" = "401" ] || [ "$STATUS_CODE" = "400" ]; then
    echo "✅ REQUEST PROCESSED (rate limit recovered)"
else
    echo "⚠️  UNEXPECTED ($STATUS_CODE)"
fi

echo ""
echo "=================================================="
echo "   Test Complete"
echo "=================================================="
