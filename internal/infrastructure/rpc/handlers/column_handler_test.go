package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/application"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestColumnHandler_ListPublishedColumns(t *testing.T) {
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ctx := context.Background()

	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)
	futureDate := now.Add(24 * time.Hour)

	column1ID := uuid.New()
	category1 := "health"
	err := testutil.CreateTestColumn(ctx, testDB.Pool, column1ID, "Test Column 1", "Test content 1",
		pgtype.Text{String: category1, Valid: true},
		[]string{"diet", "exercise"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	column2ID := uuid.New()
	category2 := "nutrition"
	err = testutil.CreateTestColumn(ctx, testDB.Pool, column2ID, "Test Column 2", "Test content 2",
		pgtype.Text{String: category2, Valid: true},
		[]string{"diet", "food"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	column3ID := uuid.New()
	err = testutil.CreateTestColumn(ctx, testDB.Pool, column3ID, "Unpublished Column", "This should not appear",
		pgtype.Text{String: "health", Valid: true},
		[]string{"diet"},
		pgtype.Timestamptz{Time: futureDate, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	repo := testutil.NewColumnRepository(testDB.Pool)
	service := application.NewColumnService(repo, logger)
	handler := NewColumnHandler(service, logger)

	// Test ListPublishedColumns
	req := connect.NewRequest(&v1.ListPublishedColumnsRequest{
		Pagination: &v1.PageRequest{
			PageSize:   10,
			PageNumber: 1,
		},
	})

	resp, err := handler.ListPublishedColumns(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp.Msg.Pagination.TotalItems != 2 {
		t.Errorf("Expected count 2, got %d", resp.Msg.Pagination.TotalItems)
	}

	if len(resp.Msg.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(resp.Msg.Columns))
	}

	// Test GetColumn with a published column
	getReq := connect.NewRequest(&v1.GetColumnRequest{
		Id: column1ID.String(),
	})

	getResp, err := handler.GetColumn(ctx, getReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if getResp.Msg.Column == nil {
		t.Fatal("Expected column, got nil")
	}

	if getResp.Msg.Column.Id != column1ID.String() {
		t.Errorf("Expected ID %v, got %v", column1ID.String(), getResp.Msg.Column.Id)
	}

	if getResp.Msg.Column.Title != "Test Column 1" {
		t.Errorf("Expected title 'Test Column 1', got '%s'", getResp.Msg.Column.Title)
	}

	// Test GetColumn with an unpublished column (should return not found)
	getUnpublishedReq := connect.NewRequest(&v1.GetColumnRequest{
		Id: column3ID.String(),
	})

	_, err = handler.GetColumn(ctx, getUnpublishedReq)
	if err == nil {
		t.Error("Expected error for unpublished column, got nil")
	}

	// Test ListColumnsByCategory
	categoryReq := connect.NewRequest(&v1.ListColumnsByCategoryRequest{
		Category: category1,
		Pagination: &v1.PageRequest{
			PageSize:   10,
			PageNumber: 1,
		},
	})

	categoryResp, err := handler.ListColumnsByCategory(ctx, categoryReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if categoryResp.Msg.Pagination.TotalItems != 1 {
		t.Errorf("Expected count 1, got %d", categoryResp.Msg.Pagination.TotalItems)
	}

	if len(categoryResp.Msg.Columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(categoryResp.Msg.Columns))
	}

	if categoryResp.Msg.Columns[0].Category == nil || categoryResp.Msg.Columns[0].Category.Value != category1 {
		t.Errorf("Expected category '%s', got '%v'", category1,
			categoryResp.Msg.Columns[0].Category)
	}

	// Test GetColumn with non-existent ID
	nonExistentID := uuid.New()
	nonExistentReq := connect.NewRequest(&v1.GetColumnRequest{
		Id: nonExistentID.String(),
	})

	_, err = handler.GetColumn(ctx, nonExistentReq)
	if err == nil {
		t.Error("Expected error for non-existent column, got nil")
	}
}
