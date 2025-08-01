#!/bin/bash

# This script demonstrates how to use curl to interact with the Column API
# It assumes the server is running on localhost:8081

# Check for required dependencies
if ! command -v jq &>/dev/null; then
  echo -e "${RED}Error: jq is required but not installed. Please install jq to run this script.${NC}"
  echo "On Ubuntu/Debian: sudo apt-get install jq"
  echo "On macOS: brew install jq"
  echo "On CentOS/RHEL: sudo yum install jq"
  exit 1
fi

# Set the base URL
BASE_URL="http://localhost:8081"

# Seed mock columns using the seed command
echo "Seeding mock columns using the seed command..."
go run main.go seed --mock

echo "Mock columns seeded successfully!"

echo "Testing Column API..."
echo "====================="
echo ""

# 1. List published columns
echo "1. Listing published columns..."
PUBLISHED_COLUMNS_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"pagination":{"page_size":5,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.ColumnService/ListPublishedColumns")

echo "$PUBLISHED_COLUMNS_RESPONSE" | jq .
echo ""

# Extract the first column ID from the response
COLUMN_ID=$(echo "$PUBLISHED_COLUMNS_RESPONSE" | jq -r '.columns[0].id // empty')

if [ -z "$COLUMN_ID" ]; then
  echo "Error: No columns found in the response. Cannot proceed with GetColumn test."
  exit 1
fi

# 2. List columns by category
echo "2. Listing columns by category 'health'..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"category":"health","pagination":{"page_size":10,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.ColumnService/ListColumnsByCategory" | jq .
echo ""

# 3. List columns by tag
echo "3. Listing columns by tag 'health'..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"tag":"health","pagination":{"page_size":10,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.ColumnService/ListColumnsByTag" | jq .
echo ""

# 4. Get a specific column by ID using the dynamically obtained ID
echo "4. Getting column by ID: $COLUMN_ID..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$COLUMN_ID\"}" \
  "$BASE_URL/healthapp.v1.ColumnService/GetColumn" | jq .
echo ""

echo "API testing complete!"
