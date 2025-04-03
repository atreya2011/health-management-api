#!/bin/bash

# This script demonstrates how to use curl to interact with the ExerciseRecordService API
# It assumes the server is running on localhost:8081

# Set the base URL
BASE_URL="http://localhost:8081"

# Generate a fresh JWT token using the generate_token.go script
echo "Generating a fresh JWT token..."
JWT_TOKEN=$(go run scripts/generate_token.go | grep -v "Generated JWT Token:" | tr -d '\n')

echo "Token generated successfully!"

echo "Testing ExerciseRecordService API..."
echo "==================================="
echo ""

# 1. Create a new exercise record
echo "1. Creating a new exercise record..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"exercise_name":"Running","duration_minutes":30,"calories_burned":250,"recorded_at":"2025-04-01T08:00:00Z"}' \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/CreateExerciseRecord" | jq .
echo ""

# 2. Create another exercise record (for testing list and delete)
echo "2. Creating another exercise record..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"exercise_name":"Weight Training","duration_minutes":45,"calories_burned":180,"recorded_at":"2025-04-02T17:30:00Z"}' \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/CreateExerciseRecord" | jq .
echo ""

# Store the ID of the second record for later use
RECORD_ID=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"exercise_name":"Yoga","duration_minutes":60,"calories_burned":150,"recorded_at":"2025-04-03T06:00:00Z"}' \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/CreateExerciseRecord" | jq -r '.exerciseRecord.id')

# 3. List exercise records (paginated)
echo "3. Listing exercise records (paginated)..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"pagination":{"page_size":5,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/ListExerciseRecords" | jq .
echo ""

# 4. Delete an exercise record
echo "4. Deleting an exercise record..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d "{\"id\":\"$RECORD_ID\"}" \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/DeleteExerciseRecord" | jq .
echo ""

# 5. Verify deletion by listing records again
echo "5. Verifying deletion by listing records again..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"pagination":{"page_size":5,"page_number":1}}' \
  "$BASE_URL/healthapp.v1.ExerciseRecordService/ListExerciseRecords" | jq .
echo ""

echo "API testing complete!"
