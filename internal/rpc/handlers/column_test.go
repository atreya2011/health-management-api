package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestColumns(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID, uuid.UUID, string, string) {
	t.Helper() // Mark as test helper
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)
	futureDate := now.Add(24 * time.Hour)

	column1ID := uuid.New()
	category1 := "health"
	err := testutil.CreateTestColumn(ctx, pool, column1ID, "Test Column 1", "Test content 1",
		pgtype.Text{String: category1, Valid: true},
		[]string{"diet", "exercise"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	require.NoError(t, err, "Failed to create test column 1")

	column2ID := uuid.New()
	category2 := "nutrition"
	err = testutil.CreateTestColumn(ctx, pool, column2ID, "Test Column 2", "Test content 2",
		pgtype.Text{String: category2, Valid: true},
		[]string{"diet", "food"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	require.NoError(t, err, "Failed to create test column 2")

	column3ID := uuid.New()
	err = testutil.CreateTestColumn(ctx, pool, column3ID, "Unpublished Column", "This should not appear",
		pgtype.Text{String: "health", Valid: true},
		[]string{"diet"},
		pgtype.Timestamptz{Time: futureDate, Valid: true})
	require.NoError(t, err, "Failed to create unpublished column")

	return column1ID, column2ID, column3ID, category1, category2
}

func TestColumnHandler_ListPublishedColumns(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	setupTestColumns(t, ctx, testPool)

	testCases := []struct {
		name        string
		req         *v1.ListPublishedColumnsRequest
		expectLen   int
		expectTotal int32
		expectPage  int32
	}{
		{
			name: "Default Pagination",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   2, // Only published columns
			expectTotal: 2,
			expectPage:  1,
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 2,
			expectPage:  1,
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectLen:   1,
			expectTotal: 2,
			expectPage:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListPublishedColumns(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.Columns, tc.expectLen)
			require.NotNil(t, resp.Msg.Pagination)
			assert.Equal(t, tc.expectTotal, resp.Msg.Pagination.TotalItems)
			assert.Equal(t, tc.expectPage, resp.Msg.Pagination.CurrentPage)
		})
	}
}

func TestColumnHandler_GetColumn(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	column1ID, _, column3ID, _, _ := setupTestColumns(t, ctx, testPool)
	nonExistentID := uuid.New()

	testCases := []struct {
		name        string
		req         *v1.GetColumnRequest
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.GetColumnResponse], err error)
	}{
		{
			name: "Success - Get Published Column",
			req:  &v1.GetColumnRequest{Id: column1ID.String()},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.GetColumnResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.Column)
				assert.Equal(t, column1ID.String(), resp.Msg.Column.Id)
				assert.Equal(t, "Test Column 1", resp.Msg.Column.Title)
			},
		},
		{
			name: "Error - Get Unpublished Column",
			req:  &v1.GetColumnRequest{Id: column3ID.String()},
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.GetColumnResponse], err error) {
				require.Error(t, err) // Expecting a 'not found' or similar error
				assert.Nil(t, resp)
			},
		},
		{
			name: "Error - Get Non-existent Column",
			req:  &v1.GetColumnRequest{Id: nonExistentID.String()},
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.GetColumnResponse], err error) {
				require.Error(t, err) // Expecting a 'not found' or similar error
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.GetColumn(ctx, req)
			tc.verify(t, resp, err)
		})
	}
}

func TestColumnHandler_ListColumnsByCategory(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	_, _, _, category1, category2 := setupTestColumns(t, ctx, testPool)

	testCases := []struct {
		name        string
		req         *v1.ListColumnsByCategoryRequest
		expectLen   int
		expectTotal int32
		expectCat   string
	}{
		{
			name: "List by Category 1",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   category1,
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 1,
			expectCat:   category1,
		},
		{
			name: "List by Category 2",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   category2,
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 1,
			expectCat:   category2,
		},
		{
			name: "List by Non-existent Category",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   "nonexistent",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   0,
			expectTotal: 0,
			expectCat:   "", // Not applicable
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListColumnsByCategory(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.Columns, tc.expectLen)
			require.NotNil(t, resp.Msg.Pagination)
			assert.Equal(t, tc.expectTotal, resp.Msg.Pagination.TotalItems)
			if tc.expectLen > 0 {
				require.NotNil(t, resp.Msg.Columns[0].Category)
				assert.Equal(t, tc.expectCat, resp.Msg.Columns[0].Category.Value)
			}
		})
	}
}

func TestColumnHandler_ListColumnsByTag(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	setupTestColumns(t, ctx, testPool)

	testCases := []struct {
		name        string
		req         *v1.ListColumnsByTagRequest
		expectLen   int
		expectTotal int32
	}{
		{
			name: "List by Tag 'diet'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "diet",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   2, // Both published columns have 'diet'
			expectTotal: 2,
		},
		{
			name: "List by Tag 'exercise'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "exercise",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 1,
		},
		{
			name: "List by Tag 'food'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "food",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 1,
		},
		{
			name: "List by Non-existent Tag",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "nonexistent",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   0,
			expectTotal: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListColumnsByTag(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.Columns, tc.expectLen)
			require.NotNil(t, resp.Msg.Pagination)
			assert.Equal(t, tc.expectTotal, resp.Msg.Pagination.TotalItems)
		})
	}
}
