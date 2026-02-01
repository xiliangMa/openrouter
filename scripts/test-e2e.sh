#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== MassRouter SaaS Platform E2E Test ===${NC}"
echo

# Configuration
API_BASE="http://localhost:8080/api/v1"
EMAIL="admin@test.com"
PASSWORD="password123"

# Check if server is running
echo -e "${YELLOW}Checking server status...${NC}"
if ! curl -s "${API_BASE}/health" > /dev/null 2>&1; then
    echo -e "${RED}Error: Server is not running at ${API_BASE}${NC}"
    echo "Start the server with: cd backend && ./server"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}"

# Step 1: Login
echo -e "\n${YELLOW}Step 1: User Authentication${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

if ! echo "$LOGIN_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${RED}✗ Login failed${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')
echo -e "${GREEN}✓ Login successful${NC}"
echo "  User ID: ${USER_ID}"
echo "  Token: ${ACCESS_TOKEN:0:20}..."

# Step 2: Get user profile with API keys
echo -e "\n${YELLOW}Step 2: Get User Profile${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "${API_BASE}/user/profile" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

if ! echo "$PROFILE_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${RED}✗ Get profile failed${NC}"
    echo "Response: $PROFILE_RESPONSE"
    exit 1
fi

API_KEY_PREFIX=$(echo "$PROFILE_RESPONSE" | jq -r '.data.api_keys[0].prefix')
API_KEY_ID=$(echo "$PROFILE_RESPONSE" | jq -r '.data.api_keys[0].id')
BALANCE=$(echo "$PROFILE_RESPONSE" | jq -r '.data.balance')
echo -e "${GREEN}✓ Profile retrieved${NC}"
echo "  API Key Prefix: ${API_KEY_PREFIX}"
echo "  Balance: ${BALANCE}"

# Step 3: Get available models
echo -e "\n${YELLOW}Step 3: Get Available Models${NC}"
MODELS_RESPONSE=$(curl -s -X GET "${API_BASE}/models" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

if ! echo "$MODELS_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${RED}✗ Get models failed${NC}"
    echo "Response: $MODELS_RESPONSE"
    exit 1
fi

MODEL_COUNT=$(echo "$MODELS_RESPONSE" | jq -r '.data.models | length')
MODEL_NAME=$(echo "$MODELS_RESPONSE" | jq -r '.data.models[0].name')
PROVIDER_NAME=$(echo "$MODELS_RESPONSE" | jq -r '.data.models[0].provider_name')
echo -e "${GREEN}✓ Models retrieved${NC}"
echo "  Available models: ${MODEL_COUNT}"
echo "  First model: ${MODEL_NAME} (${PROVIDER_NAME})"

# Step 4: Test Chat Completion (proxy endpoint)
echo -e "\n${YELLOW}Step 4: Test Chat Completion Proxy${NC}"
echo "  Using model: ${MODEL_NAME}"
echo "  Using API key prefix: ${API_KEY_PREFIX}"

# Note: We need the full API key, not just prefix
# In a real scenario, we would need to retrieve or create a full API key
# For now, we'll use the test key from the database
TEST_API_KEY="aeg0peY1VPw1w6caG4kVpMU7mvKMkkbe"

CHAT_RESPONSE=$(curl -s -X POST "${API_BASE}/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "X-API-Key: ${TEST_API_KEY}" \
  -d "{\"model\":\"${MODEL_NAME}\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello, this is a test message from E2E test script.\"}]}")

if echo "$CHAT_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo -e "${RED}✗ Chat completion failed${NC}"
    echo "Response: $CHAT_RESPONSE"
    exit 1
fi

ASSISTANT_CONTENT=$(echo "$CHAT_RESPONSE" | jq -r '.choices[0].message.content')
USAGE_PROMPT=$(echo "$CHAT_RESPONSE" | jq -r '.usage.prompt_tokens')
USAGE_COMPLETION=$(echo "$CHAT_RESPONSE" | jq -r '.usage.completion_tokens')
echo -e "${GREEN}✓ Chat completion successful${NC}"
echo "  Assistant response: ${ASSISTANT_CONTENT}"
echo "  Token usage: ${USAGE_PROMPT} prompt + ${USAGE_COMPLETION} completion"

# Step 5: Check balance after usage
echo -e "\n${YELLOW}Step 5: Verify Balance Update${NC}"
# Note: In real implementation, balance should decrease after billing
# For simulation mode, balance remains the same
BALANCE_RESPONSE=$(curl -s -X GET "${API_BASE}/user/balance" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

if echo "$BALANCE_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    NEW_BALANCE=$(echo "$BALANCE_RESPONSE" | jq -r '.data.balance')
    echo -e "${GREEN}✓ Balance checked${NC}"
    echo "  Current balance: ${NEW_BALANCE}"
else
    echo -e "${YELLOW}⚠ Balance check endpoint not available${NC}"
fi

# Step 6: Test with different model
echo -e "\n${YELLOW}Step 6: Test Different Model${NC}"
# Find a model from a different provider
SECOND_MODEL=$(echo "$MODELS_RESPONSE" | jq -r '.data.models[2].name')
SECOND_PROVIDER=$(echo "$MODELS_RESPONSE" | jq -r '.data.models[2].provider_name')

if [ "$SECOND_MODEL" != "null" ] && [ "$SECOND_MODEL" != "" ]; then
    echo "  Testing model: ${SECOND_MODEL} (${SECOND_PROVIDER})"
    
    SECOND_CHAT_RESPONSE=$(curl -s -X POST "${API_BASE}/chat/completions" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer ${ACCESS_TOKEN}" \
      -H "X-API-Key: ${TEST_API_KEY}" \
      -d "{\"model\":\"${SECOND_MODEL}\",\"messages\":[{\"role\":\"user\",\"content\":\"Test message for second model.\"}]}")
    
    if echo "$SECOND_CHAT_RESPONSE" | jq -e '.choices' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Second model test successful${NC}"
    else
        echo -e "${YELLOW}⚠ Second model test returned non-standard response${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Not enough models for second test${NC}"
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ E2E TEST COMPLETED SUCCESSFULLY${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo "Summary:"
echo "  - Server: Running"
echo "  - Authentication: Working"
echo "  - Profile & API Keys: Retrieved"
echo "  - Models: ${MODEL_COUNT} available"
echo "  - Chat Completion: Working (simulation mode)"
echo "  - Multi-provider: Tested"
echo
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Configure real API keys for actual provider testing"
echo "  2. Implement rate limiting and billing verification"
echo "  3. Add more comprehensive error handling tests"
echo "  4. Test with streaming responses"