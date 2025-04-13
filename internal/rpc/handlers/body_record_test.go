package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestBodyRecordHandler_CreateBodyRecord(t *testing.T) {
	date := time.Now().UTC().Truncate(24 * time.Hour)
	dateStr := date.Format("2006-01-02")

	testCases := []struct {
		name        string
		req         *v1.CreateBodyRecordRequest
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.CreateBodyRecordResponse], err error)
	}{
		{
			name: "Success",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: 75.5},
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateBodyRecordResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.BodyRecord)
				assert.Equal(t, testUserID.String(), resp.Msg.BodyRecord.UserId)
				assert.Equal(t, dateStr, resp.Msg.BodyRecord.Date)
				require.NotNil(t, resp.Msg.BodyRecord.WeightKg)
				assert.Equal(t, 75.5, resp.Msg.BodyRecord.WeightKg.Value)
			},
		},
		{
			name: "Error - Invalid Weight",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: -10.0}, // Negative weight
			},
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateBodyRecordResponse], err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
			},
		},
		{
			name: "Success - Only Weight",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: 76.0},
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateBodyRecordResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg.BodyRecord)
				assert.Equal(t, 76.0, resp.Msg.BodyRecord.WeightKg.Value)
				assert.Nil(t, resp.Msg.BodyRecord.BodyFatPercentage) // Ensure other fields are nil if not provided
			},
		},
		{
			name: "Success - With Body Fat",
			req: &v1.CreateBodyRecordRequest{
				Date:              dateStr,
				WeightKg:          &wrapperspb.DoubleValue{Value: 75.0},
				BodyFatPercentage: &wrapperspb.DoubleValue{Value: 15.5},
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateBodyRecordResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg.BodyRecord)
				assert.Equal(t, 75.0, resp.Msg.BodyRecord.WeightKg.Value)
				require.NotNil(t, resp.Msg.BodyRecord.BodyFatPercentage)
				assert.Equal(t, 15.5, resp.Msg.BodyRecord.BodyFatPercentage.Value)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := testutil.NewBodyRecordRepository(testPool)
			handler := NewBodyRecordHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			req := connect.NewRequest(tc.req)
			resp, err := handler.CreateBodyRecord(testCtx, req)

			tc.verify(t, resp, err)
		})
	}
}

func TestBodyRecordHandler_ListBodyRecords(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	weight1 := 75.5
	weight2 := 76.0
	bodyFat := 15.5
	_, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	require.NoError(t, err, "Failed to create test body record 1")
	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	require.NoError(t, err, "Failed to create test body record 2")

	testCases := []struct {
		name       string
		req        *v1.ListBodyRecordsRequest
		expectLen  int
		expectTotal int32
		expectPage int32
	}{
		{
			name: "Default Pagination",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:  2,
			expectTotal: 2,
			expectPage: 1,
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectLen:  1,
			expectTotal: 2,
			expectPage: 1,
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectLen:  1,
			expectTotal: 2,
			expectPage: 2,
		},
		{
			name: "Pagination - Page 3 Size 1 (Empty)",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 3},
			},
			expectLen:  0,
			expectTotal: 2,
			expectPage: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListBodyRecords(testCtx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.BodyRecords, tc.expectLen)
			require.NotNil(t, resp.Msg.Pagination)
			assert.Equal(t, tc.expectTotal, resp.Msg.Pagination.TotalItems)
			assert.Equal(t, tc.expectPage, resp.Msg.Pagination.CurrentPage)
		})
	}
}

func TestBodyRecordHandler_GetBodyRecordsByDateRange(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)
	weight1, weight2, weight3 := 75.5, 76.0, 77.0
	bodyFat := 15.5
	_, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	require.NoError(t, err)
	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	require.NoError(t, err)
	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, lastWeek, &weight3, nil)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		expectLen int
	}{
		{
			name:      "Range including today and yesterday",
			startDate: today.Add(-3 * 24 * time.Hour),
			endDate:   today,
			expectLen: 2,
		},
		{
			name:      "Range including only today",
			startDate: today,
			endDate:   today,
			expectLen: 1,
		},
		{
			name:      "Range including all three",
			startDate: lastWeek,
			endDate:   today,
			expectLen: 3,
		},
		{
			name:      "Range with no records",
			startDate: today.Add(24 * time.Hour),
			endDate:   today.Add(48 * time.Hour),
			expectLen: 0,
		},
		{
			name:      "Range including only last week",
			startDate: lastWeek,
			endDate:   lastWeek,
			expectLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
				StartDate: tc.startDate.Format("2006-01-02"),
				EndDate:   tc.endDate.Format("2006-01-02"),
			})
			resp, err := handler.GetBodyRecordsByDateRange(testCtx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.BodyRecords, tc.expectLen)
		})
	}
}
