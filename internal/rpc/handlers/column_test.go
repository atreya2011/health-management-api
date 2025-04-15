package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/repo"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
)

func setupTestColumns(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (db.Column, db.Column, db.Column, string, string) {
	t.Helper()
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)
	futureDate := now.Add(24 * time.Hour)

	category1 := "health"
	column1, err := testutil.CreateTestColumn(ctx, pool, uuid.New(), "Test Column 1", "Test content 1",
		pgtype.Text{String: category1, Valid: true},
		[]string{"diet", "exercise"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	require.NoError(t, err, "Failed to create test column 1")

	category2 := "nutrition"
	column2, err := testutil.CreateTestColumn(ctx, pool, uuid.New(), "Test Column 2", "Test content 2",
		pgtype.Text{String: category2, Valid: true},
		[]string{"diet", "food"},
		pgtype.Timestamptz{Time: publishedAt, Valid: true})
	require.NoError(t, err, "Failed to create test column 2")

	column3, err := testutil.CreateTestColumn(ctx, pool, uuid.New(), "Unpublished Column", "This should not appear",
		pgtype.Text{String: "health", Valid: true},
		[]string{"diet"},
		pgtype.Timestamptz{Time: futureDate, Valid: true})
	require.NoError(t, err, "Failed to create unpublished column")

	return column1, column2, column3, category1, category2
}

func TestListPublishedColumns(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	col1, col2, _, _, _ := setupTestColumns(t, ctx, testPool)

	protoCol1 := ToProtoColumn(col1)
	protoCol2 := ToProtoColumn(col2)

	testCases := []struct {
		name         string
		req          *v1.ListPublishedColumnsRequest
		expectError  bool
		expectedResp *v1.ListPublishedColumnsResponse
	}{
		{
			name: "Default Pagination",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListPublishedColumnsResponse{
				Columns: []*v1.Column{protoCol1, protoCol2},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListPublishedColumnsResponse{
				Columns: []*v1.Column{protoCol1},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListPublishedColumnsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectError: false,
			expectedResp: &v1.ListPublishedColumnsResponse{
				Columns: []*v1.Column{protoCol2},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 2,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListPublishedColumns(ctx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.Column{}, "created_at", "updated_at"),
					cmpopts.SortSlices(func(a, b *v1.Column) bool {
						return a.Id < b.Id
					}),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("ListPublishedColumns response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetColumn(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	col1, _, col3, _, _ := setupTestColumns(t, ctx, testPool)
	nonExistentID := uuid.New()

	protoCol1 := ToProtoColumn(col1)

	testCases := []struct {
		name         string
		req          *v1.GetColumnRequest
		expectError  bool
		expectedResp *v1.GetColumnResponse
	}{
		{
			name:        "Success - Get Published Column",
			req:         &v1.GetColumnRequest{Id: col1.ID.String()},
			expectError: false,
			expectedResp: &v1.GetColumnResponse{
				Column: protoCol1,
			},
		},
		{
			name:         "Error - Get Unpublished Column",
			req:          &v1.GetColumnRequest{Id: col3.ID.String()},
			expectError:  true,
			expectedResp: nil,
		},
		{
			name:         "Error - Get Non-existent Column",
			req:          &v1.GetColumnRequest{Id: nonExistentID.String()},
			expectError:  true,
			expectedResp: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.GetColumn(ctx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.Column{}, "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("GetColumn response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListColumnsByCategory(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	col1, col2, _, category1, category2 := setupTestColumns(t, ctx, testPool)

	protoCol1 := ToProtoColumn(col1)
	protoCol2 := ToProtoColumn(col2)

	testCases := []struct {
		name         string
		req          *v1.ListColumnsByCategoryRequest
		expectError  bool
		expectedResp *v1.ListColumnsByCategoryResponse
	}{
		{
			name: "List by Category 1",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   category1,
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByCategoryResponse{
				Columns: []*v1.Column{protoCol1},
				Pagination: &v1.PageResponse{
					TotalItems:  1,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "List by Category 2",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   category2,
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByCategoryResponse{
				Columns: []*v1.Column{protoCol2},
				Pagination: &v1.PageResponse{
					TotalItems:  1,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "List by Non-existent Category",
			req: &v1.ListColumnsByCategoryRequest{
				Category:   "nonexistent",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByCategoryResponse{
				Columns: []*v1.Column{},
				Pagination: &v1.PageResponse{
					TotalItems:  0,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListColumnsByCategory(ctx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.Column{}, "created_at", "updated_at"),
					cmpopts.SortSlices(func(a, b *v1.Column) bool {
						return a.Id < b.Id
					}),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("ListColumnsByCategory response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListColumnsByTag(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewColumnRepository(testPool)
	handler := NewColumnHandler(repo, testLogger)
	ctx := context.Background()
	col1, col2, _, _, _ := setupTestColumns(t, ctx, testPool)

	protoCol1 := ToProtoColumn(col1)
	protoCol2 := ToProtoColumn(col2)

	testCases := []struct {
		name         string
		req          *v1.ListColumnsByTagRequest
		expectError  bool
		expectedResp *v1.ListColumnsByTagResponse
	}{
		{
			name: "List by Tag 'diet'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "diet",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByTagResponse{
				Columns: []*v1.Column{protoCol1, protoCol2},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "List by Tag 'exercise'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "exercise",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByTagResponse{
				Columns: []*v1.Column{protoCol1},
				Pagination: &v1.PageResponse{
					TotalItems:  1,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "List by Tag 'food'",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "food",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByTagResponse{
				Columns: []*v1.Column{protoCol2},
				Pagination: &v1.PageResponse{
					TotalItems:  1,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "List by Non-existent Tag",
			req: &v1.ListColumnsByTagRequest{
				Tag:        "nonexistent",
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListColumnsByTagResponse{
				Columns: []*v1.Column{},
				Pagination: &v1.PageResponse{
					TotalItems:  0,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListColumnsByTag(ctx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.Column{}, "created_at", "updated_at"),
					cmpopts.SortSlices(func(a, b *v1.Column) bool {
						return a.Id < b.Id
					}),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("ListColumnsByTag response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
