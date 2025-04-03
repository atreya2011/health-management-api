#!/bin/bash

# This script demonstrates how to use curl to interact with the BodyRecord API
# It assumes the server is running on localhost:8080

# Set the base URL
BASE_URL="http://localhost:8081"

# Generate a fresh JWT token using the generate_token.go script
echo "Generating a fresh JWT token..."
JWT_TOKEN=$(go run scripts/generate_token.go | grep -v "Generated JWT Token:" | tr -d '\n')

echo "Token generated successfully!"

echo "Testing BodyRecord API..."
echo "=========================="
echo ""

# 1. Create a new body record
echo "1. Creating a new body record..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"date":"2025-04-01","weight_kg":75.5,"body_fat_percentage":16.2}' \
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
