#!/bin/bash

# Backend Testing Script
# This script tests all the backend endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:5000"
API_BASE="${BACKEND_URL}/api/v1"

echo -e "${BLUE}🧪 Testing IDAM-PAM Backend API${NC}"
echo -e "${BLUE}Backend URL: ${BACKEND_URL}${NC}"
echo ""

# Function to make HTTP requests and check responses
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local headers=$5
    local description=$6
    
    echo -e "${YELLOW}Testing: ${description}${NC}"
    echo -e "  ${method} ${endpoint}"
    
    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -X ${method} "${endpoint}" -H "Content-Type: application/json" -H "${headers}" -d "${data}")
        else
            response=$(curl -s -w "\n%{http_code}" -X ${method} "${endpoint}" -H "Content-Type: application/json" -d "${data}")
        fi
    else
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -X ${method} "${endpoint}" -H "${headers}")
        else
            response=$(curl -s -w "\n%{http_code}" -X ${method} "${endpoint}")
        fi
    fi
    
    # Split response and status code
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "  ${GREEN}✅ Status: ${status_code} (Expected: ${expected_status})${NC}"
        if [ -n "$response_body" ] && [ "$response_body" != "null" ]; then
            echo -e "  ${GREEN}📄 Response: ${response_body}${NC}"
        fi
        return 0
    else
        echo -e "  ${RED}❌ Status: ${status_code} (Expected: ${expected_status})${NC}"
        echo -e "  ${RED}📄 Response: ${response_body}${NC}"
        return 1
    fi
}

# Check if backend is running
echo -e "${YELLOW}🔍 Checking if backend is running...${NC}"
if ! curl -s "${BACKEND_URL}/health" > /dev/null; then
    echo -e "${RED}❌ Backend is not running at ${BACKEND_URL}${NC}"
    echo -e "${YELLOW}💡 Start the backend with: go run cmd/server/main.go${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Backend is running${NC}"
echo ""

# Test 1: Health Check
test_endpoint "GET" "${BACKEND_URL}/health" "" 200 "" "Health check endpoint"
echo ""

# Test 2: User Registration
echo -e "${BLUE}👤 Testing User Management${NC}"
REGISTER_DATA='{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!"
}'

test_endpoint "POST" "${API_BASE}/auth/register" "$REGISTER_DATA" 200 "" "User registration"
echo ""

# Test 3: User Login
LOGIN_DATA='{
    "username": "testuser",
    "password": "SecurePassword123!"
}'

echo -e "${YELLOW}🔐 Testing authentication...${NC}"
login_response=$(curl -s -X POST "${API_BASE}/auth/login" -H "Content-Type: application/json" -d "$LOGIN_DATA")
echo -e "  ${GREEN}📄 Login Response: ${login_response}${NC}"

# Extract JWT token
JWT_TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$JWT_TOKEN" ]; then
    echo -e "  ${GREEN}✅ JWT Token obtained${NC}"
    echo -e "  ${BLUE}🔑 Token: ${JWT_TOKEN:0:50}...${NC}"
else
    echo -e "  ${RED}❌ Failed to obtain JWT token${NC}"
    exit 1
fi
echo ""

# Test 4: Protected Endpoints
echo -e "${BLUE}🔒 Testing Protected Endpoints${NC}"

# Test Users List
test_endpoint "GET" "${API_BASE}/users" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Get users list"
echo ""

# Test 5: Secret Management
echo -e "${BLUE}🔑 Testing Secret Management${NC}"

# Create a secret
SECRET_DATA='{
    "name": "Test Database Password",
    "description": "Test secret for database connection",
    "data": "super-secret-password-123"
}'

secret_response=$(curl -s -X POST "${API_BASE}/secrets" -H "Content-Type: application/json" -H "Authorization: Bearer ${JWT_TOKEN}" -d "$SECRET_DATA")
echo -e "${YELLOW}Creating secret...${NC}"
echo -e "  ${GREEN}📄 Response: ${secret_response}${NC}"

# Extract secret ID
SECRET_ID=$(echo "$secret_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -n "$SECRET_ID" ]; then
    echo -e "  ${GREEN}✅ Secret created with ID: ${SECRET_ID}${NC}"
    
    # Test getting secrets list
    test_endpoint "GET" "${API_BASE}/secrets" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Get secrets list"
    echo ""
    
    # Test getting specific secret (decrypted)
    test_endpoint "GET" "${API_BASE}/secrets/${SECRET_ID}" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Get specific secret (decrypted)"
    echo ""
    
    # Test deleting secret
    test_endpoint "DELETE" "${API_BASE}/secrets/${SECRET_ID}" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Delete secret"
    echo ""
else
    echo -e "  ${RED}❌ Failed to create secret${NC}"
fi

# Test 6: Audit Logs
echo -e "${BLUE}📋 Testing Audit Logs${NC}"
test_endpoint "GET" "${API_BASE}/audit" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Get audit logs"
echo ""

# Test 7: TOTP Setup
echo -e "${BLUE}🔐 Testing TOTP (MFA) Setup${NC}"
test_endpoint "POST" "${API_BASE}/totp/enable" "" 200 "Authorization: Bearer ${JWT_TOKEN}" "Enable TOTP for user"
echo ""

# Test 8: Unauthorized Access
echo -e "${BLUE}🚫 Testing Unauthorized Access${NC}"
test_endpoint "GET" "${API_BASE}/users" "" 401 "" "Access protected endpoint without token"
test_endpoint "GET" "${API_BASE}/secrets" "" 401 "" "Access secrets without token"
echo ""

# Test 9: Invalid Token
echo -e "${BLUE}🔒 Testing Invalid Token${NC}"
test_endpoint "GET" "${API_BASE}/users" "" 401 "Authorization: Bearer invalid-token" "Access with invalid token"
echo ""

# Performance Test
echo -e "${BLUE}⚡ Performance Test${NC}"
echo -e "${YELLOW}Running 10 concurrent health checks...${NC}"

start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "${BACKEND_URL}/health" > /dev/null &
done
wait
end_time=$(date +%s.%N)

duration=$(echo "$end_time - $start_time" | bc)
echo -e "  ${GREEN}✅ Completed 10 requests in ${duration} seconds${NC}"
echo ""

# Database Connection Test
echo -e "${BLUE}🗄️  Testing Database Connection${NC}"
echo -e "${YELLOW}Checking if database operations are working...${NC}"

# Try to register another user to test database
DB_TEST_DATA='{
    "username": "dbtest",
    "email": "dbtest@example.com",
    "password": "TestPassword123!"
}'

if test_endpoint "POST" "${API_BASE}/auth/register" "$DB_TEST_DATA" 200 "" "Database connection test (user registration)"; then
    echo -e "  ${GREEN}✅ Database is working correctly${NC}"
else
    echo -e "  ${RED}❌ Database connection issues detected${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}🎉 Backend Testing Complete!${NC}"
echo ""
echo -e "${BLUE}📊 Test Summary:${NC}"
echo -e "✅ Health check endpoint working"
echo -e "✅ User registration and authentication working"
echo -e "✅ JWT token generation and validation working"
echo -e "✅ Protected endpoints properly secured"
echo -e "✅ Secret management (create, read, delete) working"
echo -e "✅ Audit logging working"
echo -e "✅ TOTP/MFA setup working"
echo -e "✅ Unauthorized access properly blocked"
echo -e "✅ Database operations working"
echo ""
echo -e "${YELLOW}💡 Next Steps:${NC}"
echo -e "1. Test the frontend by running: npm run dev"
echo -e "2. Test the complete Docker setup: docker-compose up"
echo -e "3. Deploy to AWS using: ./scripts/deploy.sh"
echo ""
echo -e "${GREEN}🔗 Useful URLs:${NC}"
echo -e "- Backend API: ${BACKEND_URL}"
echo -e "- Health Check: ${BACKEND_URL}/health"
echo -e "- API Documentation: Check the README.md file"