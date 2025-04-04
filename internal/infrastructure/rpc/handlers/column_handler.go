package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/domain"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ColumnHandler implements the column service
type ColumnHandler struct {
	columnApp application.ColumnService
	log       *slog.Logger
}

// NewColumnHandler creates a new column handler
func NewColumnHandler(columnApp application.ColumnService, log *slog.Logger) *ColumnHandler {
	return &ColumnHandler{
		columnApp: columnApp,
		log:       log,
	}
}

// ListPublishedColumns lists published columns
func (h *ColumnHandler) ListPublishedColumns(ctx context.Context, req *connect.Request[v1.ListPublishedColumnsRequest]) (*connect.Response[v1.ListPublishedColumnsResponse], error) {
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
	h.log.InfoContext(ctx, "Listing published columns", "pageSize", pageSize, "pageNumber", pageNumber)
	columns, total, err := h.columnApp.ListPublishedColumns(ctx, pageNumber, pageSize)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to list published columns", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list published columns"))
	}

	// Convert domain models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns))
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column)
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListPublishedColumnsResponse{
		Columns: protoColumns,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// GetColumn retrieves a specific column by ID
func (h *ColumnHandler) GetColumn(ctx context.Context, req *connect.Request[v1.GetColumnRequest]) (*connect.Response[v1.GetColumnResponse], error) {
	// Parse column ID
	columnID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid column ID", "columnID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid column ID: %w", err))
	}

	// Call application service
	h.log.InfoContext(ctx, "Getting column", "columnID", columnID)
	column, err := h.columnApp.GetColumn(ctx, columnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			h.log.WarnContext(ctx, "Column not found", "columnID", columnID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("column not found"))
		}
		h.log.ErrorContext(ctx, "Failed to get column", "columnID", columnID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get column"))
	}

	// Convert domain model to protobuf message
	protoColumn := toProtoColumn(column)

	// Create response
	res := connect.NewResponse(&v1.GetColumnResponse{
		Column: protoColumn,
	})

	return res, nil
}

// ListColumnsByCategory lists columns by category
func (h *ColumnHandler) ListColumnsByCategory(ctx context.Context, req *connect.Request[v1.ListColumnsByCategoryRequest]) (*connect.Response[v1.ListColumnsByCategoryResponse], error) {
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
	h.log.InfoContext(ctx, "Listing columns by category", "category", req.Msg.Category, "pageSize", pageSize, "pageNumber", pageNumber)
	columns, total, err := h.columnApp.ListColumnsByCategory(ctx, req.Msg.Category, pageNumber, pageSize)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to list columns by category", "category", req.Msg.Category, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list columns by category"))
	}

	// Convert domain models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns))
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column)
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListColumnsByCategoryResponse{
		Columns: protoColumns,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// ListColumnsByTag lists columns by tag
func (h *ColumnHandler) ListColumnsByTag(ctx context.Context, req *connect.Request[v1.ListColumnsByTagRequest]) (*connect.Response[v1.ListColumnsByTagResponse], error) {
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
	h.log.InfoContext(ctx, "Listing columns by tag", "tag", req.Msg.Tag, "pageSize", pageSize, "pageNumber", pageNumber)
	columns, total, err := h.columnApp.ListColumnsByTag(ctx, req.Msg.Tag, pageNumber, pageSize)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to list columns by tag", "tag", req.Msg.Tag, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list columns by tag"))
	}

	// Convert domain models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns))
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column)
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListColumnsByTagResponse{
		Columns: protoColumns,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// toProtoColumn converts a domain.Column to a v1.Column
func toProtoColumn(column *domain.Column) *v1.Column {
	protoColumn := &v1.Column{
		Id:        column.ID.String(),
		Title:     column.Title,
		Content:   column.Content,
		Tags:      column.Tags,
		CreatedAt: timestamppb.New(column.CreatedAt),
		UpdatedAt: timestamppb.New(column.UpdatedAt),
	}

	// Handle optional fields
	if column.Category != nil {
		protoColumn.Category = &wrapperspb.StringValue{Value: *column.Category}
	}

	if column.PublishedAt != nil {
		protoColumn.PublishedAt = timestamppb.New(*column.PublishedAt)
	}

	return protoColumn
}
