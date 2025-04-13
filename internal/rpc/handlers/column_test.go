package handlers

import (
	"context"
	// "log/slog" // No longer needed
	// "os" // No longer needed
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool" // Add import for pgxpool.Pool
)

// setupTestColumns creates test columns in the database for testing using the global pool
func setupTestColumns(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID, uuid.UUID, string, string) {
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)
	futureDate := now.Add(24 * time.Hour)

	column1ID := uuid.New()
	category1 := "health"
	// Use the passed-in pool directly
	err := testutil.CreateTestColumn(ctx, pool, column1ID, "Test Column 1", "Test content 1",
		pgtype.Text{String: category1, Valid: true},
		[]string{"diet", "exercise"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	column2ID := uuid.New()
	category2 := "nutrition"
	// Use the passed-in pool directly
	err = testutil.CreateTestColumn(ctx, pool, column2ID, "Test Column 2", "Test content 2",
		pgtype.Text{String: category2, Valid: true},
		[]string{"diet", "food"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	column3ID := uuid.New()
	// Use the passed-in pool directly
	err = testutil.CreateTestColumn(ctx, pool, column3ID, "Unpublished Column", "This should not appear",
		pgtype.Text{String: "health", Valid: true},
		[]string{"diet"},
		pgtype.Timestamptz{Time: futureDate, Valid: true})
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	return column1ID, column2ID, column3ID, category1, category2
}

// setupHandler is removed, handler created directly in tests

func TestColumnHandler_ListPublishedColumns(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// DB Setup/Teardown handled by TestMain

	// Create handler using global resources
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background() // Use background context

	setupTestColumns(t, ctx, testPool)   // Use global testPool

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
}

func TestColumnHandler_GetColumn(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// DB Setup/Teardown handled by TestMain

	// Create handler using global resources
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background() // Use background context

	column1ID, _, column3ID, _, _ := setupTestColumns(t, ctx, testPool) // Use global testPool

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

func TestColumnHandler_ListColumnsByCategory(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// DB Setup/Teardown handled by TestMain

	// Create handler using global resources
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background() // Use background context

	_, _, _, category1, _ := setupTestColumns(t, ctx, testPool) // Use global testPool

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
}

func TestColumnHandler_ListColumnsByTag(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// DB Setup/Teardown handled by TestMain

	// Create handler using global resources
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background() // Use background context

	setupTestColumns(t, ctx, testPool)   // Use global testPool

	// Test ListColumnsByTag
	tag := "diet"
	tagReq := connect.NewRequest(&v1.ListColumnsByTagRequest{
		Tag: tag,
		Pagination: &v1.PageRequest{
			PageSize:   10,
			PageNumber: 1,
		},
	})

	tagResp, err := handler.ListColumnsByTag(ctx, tagReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// We should have 2 published columns with the "diet" tag
	if tagResp.Msg.Pagination.TotalItems != 2 {
		t.Errorf("Expected count 2, got %d", tagResp.Msg.Pagination.TotalItems)
	}

	if len(tagResp.Msg.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(tagResp.Msg.Columns))
	}

	// Test with a tag that doesn't exist
	nonExistentTag := "nonexistent"
	nonExistentTagReq := connect.NewRequest(&v1.ListColumnsByTagRequest{
		Tag: nonExistentTag,
		Pagination: &v1.PageRequest{
			PageSize:   10,
			PageNumber: 1,
		},
	})

	nonExistentTagResp, err := handler.ListColumnsByTag(ctx, nonExistentTagReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if nonExistentTagResp.Msg.Pagination.TotalItems != 0 {
		t.Errorf("Expected count 0, got %d", nonExistentTagResp.Msg.Pagination.TotalItems)
	}

	if len(nonExistentTagResp.Msg.Columns) != 0 {
		t.Errorf("Expected 0 columns, got %d", len(nonExistentTagResp.Msg.Columns))
	}
}
