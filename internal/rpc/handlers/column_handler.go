package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	postgres "github.com/atreya2011/health-management-api/internal/db"
	db "github.com/atreya2011/health-management-api/internal/db/gen"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ColumnHandler implements the column service RPCs
type ColumnHandler struct {
	repo *postgres.PgColumnRepository // Use concrete repository type
	log  *slog.Logger
}

// NewColumnHandler creates a new column handler
func NewColumnHandler(repo *postgres.PgColumnRepository, log *slog.Logger) *ColumnHandler { // Use concrete repository type
	return &ColumnHandler{
		repo: repo,
		log:  log,
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
	// Ensure page is at least 1 (from service)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	// Apply max page size (from service)
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset (from service)
	offset := (pageNumber - 1) * pageSize

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching published columns", "page", pageNumber, "pageSize", pageSize)
	columns, err := h.repo.FindPublished(ctx, pageSize, offset) // Changed from columnApp.ListPublishedColumns
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch published columns", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch published columns"))
	}

	// Get total count (from service)
	total, err := h.repo.CountPublished(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count published columns", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count published columns"))
	}

	// Convert persistence models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns)) // columns is now []db.Column
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column) // Pass db.Column
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

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching column", "columnID", columnID)
	column, err := h.repo.FindByID(ctx, columnID) // Changed from columnApp.GetColumn
	if err != nil {
		if errors.Is(err, postgres.ErrColumnNotFound) { // Use postgres error
			h.log.WarnContext(ctx, "Column not found", "columnID", columnID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("column not found"))
		}
		h.log.ErrorContext(ctx, "Failed to fetch column", "columnID", columnID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch column"))
	}

	// Check if the column is published using the PublishedAt field
	// A column is published if PublishedAt is not NULL and is in the past.
	isPublished := column.PublishedAt.Valid && column.PublishedAt.Time.Before(time.Now())
	if !isPublished {
		h.log.WarnContext(ctx, "Attempted to access unpublished column", "id", columnID)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("column not found")) // Treat unpublished as not found
	}

	// Convert persistence model to protobuf message
	protoColumn := toProtoColumn(column) // column is now db.Column

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
	// Ensure page is at least 1 (from service)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	// Apply max page size (from service)
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset (from service)
	offset := (pageNumber - 1) * pageSize

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching columns by category", "category", req.Msg.Category, "page", pageNumber, "pageSize", pageSize)
	columns, err := h.repo.FindByCategory(ctx, req.Msg.Category, pageSize, offset) // Changed from columnApp.ListColumnsByCategory
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch columns by category", "category", req.Msg.Category, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch columns by category"))
	}

	// Get total count (from service)
	total, err := h.repo.CountByCategory(ctx, req.Msg.Category)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count columns by category", "category", req.Msg.Category, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count columns by category"))
	}

	// Convert persistence models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns)) // columns is now []db.Column
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column) // Pass db.Column
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
	// Ensure page is at least 1 (from service)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	// Apply max page size (from service)
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset (from service)
	offset := (pageNumber - 1) * pageSize

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching columns by tag", "tag", req.Msg.Tag, "page", pageNumber, "pageSize", pageSize)
	columns, err := h.repo.FindByTag(ctx, req.Msg.Tag, pageSize, offset) // Changed from columnApp.ListColumnsByTag
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch columns by tag", "tag", req.Msg.Tag, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch columns by tag"))
	}

	// Get total count (from service)
	total, err := h.repo.CountByTag(ctx, req.Msg.Tag)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count columns by tag", "tag", req.Msg.Tag, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count columns by tag"))
	}

	// Convert persistence models to protobuf messages
	protoColumns := make([]*v1.Column, len(columns)) // columns is now []db.Column
	for i, column := range columns {
		protoColumns[i] = toProtoColumn(column) // Pass db.Column
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

// toProtoColumn converts a db.Column (sqlc generated) to a v1.Column
func toProtoColumn(column db.Column) *v1.Column { // Accept db.Column
	protoColumn := &v1.Column{
		Id:        column.ID.String(),
		Title:     column.Title,
		Content:   column.Content,
		Tags:      column.Tags, // Assuming Tags is []string in db.Column
		CreatedAt: timestamppb.New(column.CreatedAt),
		UpdatedAt: timestamppb.New(column.UpdatedAt),
	}

	// Handle pgtype.Text for Category
	if column.Category.Valid {
		protoColumn.Category = &wrapperspb.StringValue{Value: column.Category.String}
	}

	// Handle pgtype.Timestamp for PublishedAt
	if column.PublishedAt.Valid {
		protoColumn.PublishedAt = timestamppb.New(column.PublishedAt.Time)
	}

	return protoColumn
}

// Helper function to check if a column is published (needed for GetColumn)
// This replaces the removed IsPublished method
func isColumnPublished(column db.Column) bool {
	return column.PublishedAt.Valid && column.PublishedAt.Time.Before(time.Now())
}
