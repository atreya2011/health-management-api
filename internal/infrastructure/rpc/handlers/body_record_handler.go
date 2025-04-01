package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// BodyRecordHandler implements the body record service
type BodyRecordHandler struct {
	bodyRecordApp application.BodyRecordService
	log           *slog.Logger
}

// NewBodyRecordHandler creates a new body record handler
func NewBodyRecordHandler(bodyRecordApp application.BodyRecordService, log *slog.Logger) *BodyRecordHandler {
	return &BodyRecordHandler{
		bodyRecordApp: bodyRecordApp,
		log:           log,
	}
}

// CreateBodyRecord creates or updates a body record for a specific date
func (h *BodyRecordHandler) CreateBodyRecord(ctx context.Context, req *connect.Request[v1.CreateBodyRecordRequest]) (*connect.Response[v1.CreateBodyRecordResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Msg.Date)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid date format", "date", req.Msg.Date, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid date format: %w", err))
	}

	// Convert protobuf wrappers to Go pointers
	var weight *float64
	var bodyFat *float64

	if req.Msg.WeightKg != nil {
		w := req.Msg.WeightKg.Value
		weight = &w
	}

	if req.Msg.BodyFatPercentage != nil {
		bf := req.Msg.BodyFatPercentage.Value
		bodyFat = &bf
	}

	// Call application service
	h.log.InfoContext(ctx, "Creating body record", "userID", userID, "date", date)
	record, err := h.bodyRecordApp.CreateOrUpdateBodyRecord(ctx, userID, date, weight, bodyFat)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to create body record", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create body record"))
	}

	// Convert domain model to protobuf message
	protoRecord := toProtoBodyRecord(record)

	// Create response
	res := connect.NewResponse(&v1.CreateBodyRecordResponse{
		BodyRecord: protoRecord,
	})

	return res, nil
}

// ListBodyRecords lists body records for the authenticated user
func (h *BodyRecordHandler) ListBodyRecords(ctx context.Context, req *connect.Request[v1.ListBodyRecordsRequest]) (*connect.Response[v1.ListBodyRecordsResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Get pagination parameters
	pageSize := 20  // Default page size
	pageNumber := 1 // Default page number

	if req.Msg.Pagination != nil {
		if req.Msg.Pagination.PageSize > 0 {
			pageSize = int(req.Msg.Pagination.PageSize)
		}
		if req.Msg.Pagination.PageNumber > 0 {
			pageNumber = int(req.Msg.Pagination.PageNumber)
		}
	}

	// Call application service
	h.log.InfoContext(ctx, "Listing body records", "userID", userID, "pageSize", pageSize, "pageNumber", pageNumber)
	records, total, err := h.bodyRecordApp.GetBodyRecordsForUser(ctx, userID, pageNumber, pageSize)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to list body records", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list body records"))
	}

	// Convert domain models to protobuf messages
	protoRecords := make([]*v1.BodyRecord, len(records))
	for i, record := range records {
		protoRecords[i] = toProtoBodyRecord(record)
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListBodyRecordsResponse{
		BodyRecords: protoRecords,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// GetBodyRecordsByDateRange retrieves body records for a specific date range
func (h *BodyRecordHandler) GetBodyRecordsByDateRange(ctx context.Context, req *connect.Request[v1.GetBodyRecordsByDateRangeRequest]) (*connect.Response[v1.GetBodyRecordsByDateRangeResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.Msg.StartDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid start date format", "startDate", req.Msg.StartDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start date format: %w", err))
	}

	endDate, err := time.Parse("2006-01-02", req.Msg.EndDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid end date format", "endDate", req.Msg.EndDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end date format: %w", err))
	}

	// Call application service
	h.log.InfoContext(ctx, "Getting body records by date range", "userID", userID, "startDate", startDate, "endDate", endDate)
	records, err := h.bodyRecordApp.GetBodyRecordsForUserDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to get body records by date range", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get body records by date range"))
	}

	// Convert domain models to protobuf messages
	protoRecords := make([]*v1.BodyRecord, len(records))
	for i, record := range records {
		protoRecords[i] = toProtoBodyRecord(record)
	}

	// Create response
	res := connect.NewResponse(&v1.GetBodyRecordsByDateRangeResponse{
		BodyRecords: protoRecords,
	})

	return res, nil
}

// toProtoBodyRecord converts a domain.BodyRecord to a v1.BodyRecord
func toProtoBodyRecord(record *domain.BodyRecord) *v1.BodyRecord {
	protoRecord := &v1.BodyRecord{
		Id:        record.ID.String(),
		UserId:    record.UserID.String(),
		Date:      record.Date.Format("2006-01-02"),
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}

	// Handle optional fields
	if record.WeightKg != nil {
		protoRecord.WeightKg = &wrapperspb.DoubleValue{Value: *record.WeightKg}
	}

	if record.BodyFatPercentage != nil {
		protoRecord.BodyFatPercentage = &wrapperspb.DoubleValue{Value: *record.BodyFatPercentage}
	}

	return protoRecord
}
