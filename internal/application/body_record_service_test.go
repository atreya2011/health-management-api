package application

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/atreya2011/health-management-api/internal/testutil"
)

func TestBodyRecordService_CreateOrUpdateBodyRecord(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository
	repo := testutil.NewBodyRecordRepository(testDB.Pool)

	// Create the service with the real repository
	service := NewBodyRecordService(repo, logger)

	// Test data
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5

	// Call the method being tested
	result, err := service.CreateOrUpdateBodyRecord(ctx, userID, date, &weight, nil)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, result.UserID)
	}
	if !result.Date.Equal(date) {
		t.Errorf("Expected Date %v, got %v", date, result.Date)
	}
	if *result.WeightKg != weight {
		t.Errorf("Expected WeightKg %v, got %v", weight, *result.WeightKg)
	}
}

func TestBodyRecordService_CreateOrUpdateBodyRecord_Error(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository
	repo := testutil.NewBodyRecordRepository(testDB.Pool)

	// Create the service with the real repository
	service := NewBodyRecordService(repo, logger)

	// Test data - invalid weight to trigger validation error
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := -10.0 // Negative weight should fail validation

	// Call the method being tested
	result, err := service.CreateOrUpdateBodyRecord(ctx, userID, date, &weight, nil)

	// Check for errors
	if err == nil {
		t.Error("Expected error for invalid weight, got nil")
	}

	// Verify the result
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestBodyRecordService_GetBodyRecordsForUser(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	weight1 := 75.5
	weight2 := 76.0
	bodyFat := 15.5

	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Create a real repository
	repo := testutil.NewBodyRecordRepository(testDB.Pool)

	// Create the service with the real repository
	service := NewBodyRecordService(repo, logger)

	// Test parameters
	page := 1
	pageSize := 10

	// Call the method being tested
	records, total, err := service.GetBodyRecordsForUser(ctx, userID, page, pageSize)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the results
	expectedCount := 2
	if len(records) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(records))
	}

	if total != int64(expectedCount) {
		t.Errorf("Expected total %d, got %d", expectedCount, total)
	}
}

func TestBodyRecordService_GetBodyRecordsForUserDateRange(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)

	weight1 := 75.5
	weight2 := 76.0
	weight3 := 77.0
	bodyFat := 15.5

	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, lastWeek, &weight3, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Create a real repository
	repo := testutil.NewBodyRecordRepository(testDB.Pool)

	// Create the service with the real repository
	service := NewBodyRecordService(repo, logger)

	// Test date range (last 3 days)
	startDate := today.Add(-3 * 24 * time.Hour)
	endDate := today

	// Call the method being tested
	records, err := service.GetBodyRecordsForUserDateRange(ctx, userID, startDate, endDate)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the results - should only include today and yesterday, not last week
	expectedCount := 2
	if len(records) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(records))
	}
}
