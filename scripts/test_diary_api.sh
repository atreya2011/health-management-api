#!/bin/bash

# This script demonstrates how to use curl to interact with the DiaryService API
# It assumes the server is running on localhost:8081

# Set the base URL
BASE_URL="http://localhost:8081"

# Generate a fresh JWT token using the generate_token.go script
echo "Generating a fresh JWT token..."
JWT_TOKEN=$(go run scripts/generate_token.go | grep -v "Generated JWT Token:" | tr -d '\n')

echo "Token generated successfully!"

echo "Testing DiaryService API..."
echo "==========================="
echo ""

# 1. Create a new diary entry
echo "1. Creating a new diary entry..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"title":"My First Diary Entry","content":"This is the content of my first diary entry.","entry_date":"2025-04-01"}' \
  "$BASE_URL/healthapp.v1.DiaryService/CreateDiaryEntry" | jq .
echo ""

# 2. Create another diary entry (for testing list and delete)
echo "2. Creating another diary entry..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"title":"My Second Diary Entry","content":"This is the content of my second diary entry.","entry_date":"2025-04-02"}' \
  "$BASE_URL/healthapp.v1.DiaryService/CreateDiaryEntry" | jq .
echo ""

# Store the ID of the second entry for later use
ENTRY_ID=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"title":"My Second Diary Entry","content":"This is the content of my second diary entry.","entry_date":"2025-04-02"}' \
  "$BASE_URL/healthapp.v1.DiaryService/CreateDiaryEntry" | jq -r '.diaryEntry.id')

# 3. List diary entries (paginated)
echo "3. Listing diary entries (paginated)..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"pagination":{"page_size":5,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.DiaryService/ListDiaryEntries" | jq .
echo ""

# 4. Get a specific diary entry
echo "4. Getting a specific diary entry..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$ENTRY_ID\"}" \
  "$BASE_URL/healthapp.v1.DiaryService/GetDiaryEntry" | jq .
echo ""

# 5. Update a diary entry
echo "5. Updating a diary entry..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$ENTRY_ID\",\"title\":\"Updated Diary Entry\",\"content\":\"This content has been updated.\"}" \
  "$BASE_URL/healthapp.v1.DiaryService/UpdateDiaryEntry" | jq .
echo ""

# 6. Delete a diary entry
echo "6. Deleting a diary entry..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$ENTRY_ID\"}" \
  "$BASE_URL/healthapp.v1.DiaryService/DeleteDiaryEntry" | jq .
echo ""

# 7. Verify the entry was deleted by trying to get it
echo "7. Verifying the entry was deleted..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$ENTRY_ID\"}" \
  "$BASE_URL/healthapp.v1.DiaryService/GetDiaryEntry" | jq .
echo ""

echo "API testing complete!"
