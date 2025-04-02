#!/bin/bash

# This script demonstrates how to use curl to interact with the BodyRecord API
# It assumes the server is running on localhost:8080

# Set the base URL
BASE_URL="http://localhost:8081"

# Set the JWT token (replace with a valid token)
# For testing, you can use a token with subject claim "test-subject-id"
# which matches the mock data user
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXN1YmplY3QtaWQiLCJuYW1lIjoiVGVzdCBVc2VyIiwiaWF0IjoxNTE2MjM5MDIyfQ.fhc3wykrAnRpcKApKhXiahxaOe8PSHatad31NuIZ0Zg"

echo "Testing BodyRecord API..."
echo "=========================="
echo ""

# 1. Create a new body record
echo "1. Creating a new body record..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"date":"2025-04-01","weight_kg":{"value":75.5},"body_fat_percentage":{"value":16.2}}' \
  "$BASE_URL/healthapp.v1.BodyRecordService/CreateBodyRecord" | jq .
echo ""

# 2. List body records (paginated)
echo "2. Listing body records (paginated)..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"pagination":{"page_size":5,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.BodyRecordService/ListBodyRecords" | jq .
echo ""

# 3. Get body records by date range
echo "3. Getting body records by date range..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"start_date":"2025-03-25","end_date":"2025-04-01"}' \
  "$BASE_URL/healthapp.v1.BodyRecordService/GetBodyRecordsByDateRange" | jq .
echo ""

echo "API testing complete!"
