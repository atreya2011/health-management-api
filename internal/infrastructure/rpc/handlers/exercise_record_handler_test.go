package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	// "github.com/atreya2011/health-management-api/internal/application" // Removed
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestExerciseRecordHandler_CreateExerciseRecord(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository
	repo := testutil.NewExerciseRecordRepository(testDB.Pool)
	// service := application.NewExerciseRecordService(repo, logger) // Removed

	// Create the handler with the real repository
	handler := NewExerciseRecordHandler(repo, logger) // Changed from service

	// Test data
	recordedAt := time.Now().UTC()
	exerciseName := "Running"
	durationMinutes := int32(30)
	caloriesBurned := int32(250)

	// Create a request
	req := connect.NewRequest(&v1.CreateExerciseRecordRequest{
		ExerciseName:    exerciseName,
		DurationMinutes: &wrapperspb.Int32Value{Value: durationMinutes},
		CaloriesBurned:  &wrapperspb.Int32Value{Value: caloriesBurned},
		RecordedAt:      timestamppb.New(recordedAt),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateExerciseRecord(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil || resp.Msg.ExerciseRecord == nil {
		t.Fatal("Expected response with exercise record, got nil")
	}

	if resp.Msg.ExerciseRecord.UserId != userID.String() {
		t.Errorf("Expected UserID %v, got %v", userID.String(), resp.Msg.ExerciseRecord.UserId)
	}

	if resp.Msg.ExerciseRecord.ExerciseName != exerciseName {
		t.Errorf("Expected ExerciseName %v, got %v", exerciseName, resp.Msg.ExerciseRecord.ExerciseName)
	}

	if resp.Msg.ExerciseRecord.DurationMinutes == nil || resp.Msg.ExerciseRecord.DurationMinutes.Value != durationMinutes {
		t.Errorf("Expected DurationMinutes %v, got %v", durationMinutes, resp.Msg.ExerciseRecord.DurationMinutes)
	}

	if resp.Msg.ExerciseRecord.CaloriesBurned == nil || resp.Msg.ExerciseRecord.CaloriesBurned.Value != caloriesBurned {
		t.Errorf("Expected CaloriesBurned %v, got %v", caloriesBurned, resp.Msg.ExerciseRecord.CaloriesBurned)
	}
}

func TestExerciseRecordHandler_CreateExerciseRecord_Error(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository
	repo := testutil.NewExerciseRecordRepository(testDB.Pool)
	// service := application.NewExerciseRecordService(repo, logger) // Removed

	// Create the handler with the real repository
	handler := NewExerciseRecordHandler(repo, logger) // Changed from service

	// Test data - invalid duration to trigger validation error
	recordedAt := time.Now().UTC()
	exerciseName := "Running"
	durationMinutes := int32(-10) // Negative duration should fail validation

	// Create a request
	req := connect.NewRequest(&v1.CreateExerciseRecordRequest{
		ExerciseName:    exerciseName,
		DurationMinutes: &wrapperspb.Int32Value{Value: durationMinutes},
		RecordedAt:      timestamppb.New(recordedAt),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateExerciseRecord(testCtx, req)

	// Check for errors
	if err == nil {
		t.Error("Expected error for invalid duration, got nil")
	}

	// Verify the response
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
}

func TestExerciseRecordHandler_ListExerciseRecords(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test records
	today := time.Now().UTC()
	yesterday := today.Add(-24 * time.Hour)

	duration1 := int32(30)
	duration2 := int32(45)
	calories1 := int32(250)
	calories2 := int32(350)

	_, err = testutil.CreateTestExerciseRecord(ctx, testDB.Queries, userID, "Running", &duration1, &calories1, today) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test exercise record: %v", err)
	}

	_, err = testutil.CreateTestExerciseRecord(ctx, testDB.Queries, userID, "Weight Training", &duration2, &calories2, yesterday) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test exercise record: %v", err)
	}

	// Create a real repository
	repo := testutil.NewExerciseRecordRepository(testDB.Pool)
	// service := application.NewExerciseRecordService(repo, logger) // Removed

	// Create the handler with the real repository
	handler := NewExerciseRecordHandler(repo, logger) // Changed from service

	// Test parameters
	pageSize := int32(10)
	pageNumber := int32(1)

	// Create a request
	req := connect.NewRequest(&v1.ListExerciseRecordsRequest{
		Pagination: &v1.PageRequest{
			PageSize:   pageSize,
			PageNumber: pageNumber,
		},
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.ListExerciseRecords(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	expectedCount := 2
	if len(resp.Msg.ExerciseRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(resp.Msg.ExerciseRecords))
	}

	if resp.Msg.Pagination == nil {
		t.Fatal("Expected pagination in response, got nil")
	}

	if resp.Msg.Pagination.TotalItems != int32(expectedCount) {
		t.Errorf("Expected total items %d, got %d", expectedCount, resp.Msg.Pagination.TotalItems)
	}

	if resp.Msg.Pagination.CurrentPage != pageNumber {
		t.Errorf("Expected current page %d, got %d", pageNumber, resp.Msg.Pagination.CurrentPage)
	}
}

func TestExerciseRecordHandler_DeleteExerciseRecord(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test exercise record
	recordedAt := time.Now().UTC()
	duration := int32(30)
	calories := int32(250)

	record, err := testutil.CreateTestExerciseRecord(ctx, testDB.Queries, userID, "Record to Delete", &duration, &calories, recordedAt) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test exercise record: %v", err)
	}

	// Create a real repository
	repo := testutil.NewExerciseRecordRepository(testDB.Pool)
	// service := application.NewExerciseRecordService(repo, logger) // Removed

	// Create the handler with the real repository
	handler := NewExerciseRecordHandler(repo, logger) // Changed from service

	// Create a request
	req := connect.NewRequest(&v1.DeleteExerciseRecordRequest{
		Id: record.ID.String(),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.DeleteExerciseRecord(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	if !resp.Msg.Success {
		t.Errorf("Expected success to be true, got false")
	}

	// Verify the record was deleted by listing records
	listReq := connect.NewRequest(&v1.ListExerciseRecordsRequest{
		Pagination: &v1.PageRequest{
			PageSize:   10,
			PageNumber: 1,
		},
	})
	
	listResp, err := handler.ListExerciseRecords(testCtx, listReq)
	if err != nil {
		t.Errorf("Expected no error when listing records, got %v", err)
	}
	
	if len(listResp.Msg.ExerciseRecords) != 0 {
		t.Errorf("Expected 0 records after deletion, got %d", len(listResp.Msg.ExerciseRecords))
	}
}
