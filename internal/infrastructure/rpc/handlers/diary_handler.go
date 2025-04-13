package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	// "github.com/atreya2011/health-management-api/internal/application" // Removed
	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres" // Added
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// DiaryHandler implements the diary service RPCs
type DiaryHandler struct {
	repo *postgres.PgDiaryEntryRepository // Use concrete repository type
	log  *slog.Logger
}

// NewDiaryHandler creates a new diary handler
func NewDiaryHandler(repo *postgres.PgDiaryEntryRepository, log *slog.Logger) *DiaryHandler { // Use concrete repository type
	return &DiaryHandler{
		repo: repo,
		log:  log,
	}
}

// CreateDiaryEntry creates a new diary entry
func (h *DiaryHandler) CreateDiaryEntry(ctx context.Context, req *connect.Request[v1.CreateDiaryEntryRequest]) (*connect.Response[v1.CreateDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse date
	entryDate, err := time.Parse("2006-01-02", req.Msg.EntryDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid date format", "entryDate", req.Msg.EntryDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid date format: %w", err))
	}

	// Convert protobuf wrapper to Go pointer
	var title *string
	if req.Msg.Title != nil {
		t := req.Msg.Title.Value
		title = &t
	}

	// Construct persistence object (was domain object)
	entry := &postgres.DiaryEntry{ // Use postgres.DiaryEntry
		ID:        uuid.New(), // Generate ID here
		UserID:    userID,
		Title:     title,
		Content:   req.Msg.Content,
		EntryDate: entryDate,
		CreatedAt: time.Now(), // Set timestamp here
		UpdatedAt: time.Now(), // Set timestamp here
	}

	// Validate the entry (moved from service)
	if err := entry.Validate(); err != nil {
		h.log.WarnContext(ctx, "Validation failed for diary entry", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid diary entry data: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Creating diary entry", "userID", userID, "entryDate", entryDate)
	savedEntry, err := h.repo.Create(ctx, entry) // Changed from diaryApp.CreateDiaryEntry
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to create diary entry", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create diary entry"))
	}

	// Convert persistence model to protobuf message
	protoEntry := toProtoDiaryEntry(savedEntry) // Use savedEntry (now *postgres.DiaryEntry)

	// Create response
	res := connect.NewResponse(&v1.CreateDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// UpdateDiaryEntry updates an existing diary entry
func (h *DiaryHandler) UpdateDiaryEntry(ctx context.Context, req *connect.Request[v1.UpdateDiaryEntryRequest]) (*connect.Response[v1.UpdateDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Convert protobuf wrapper to Go pointer
	var title *string
	if req.Msg.Title != nil {
		t := req.Msg.Title.Value
		title = &t
	}

	// First, check if the entry exists and belongs to the user (moved from service)
	existingEntry, err := h.repo.FindByID(ctx, entryID, userID) // existingEntry is now *postgres.DiaryEntry
	if err != nil {
		if errors.Is(err, postgres.ErrDiaryEntryNotFound) { // Use postgres error
			h.log.WarnContext(ctx, "Diary entry not found for update", "id", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, postgres.ErrDiaryEntryNotFound) // Return specific error
		}
		h.log.ErrorContext(ctx, "Failed to fetch diary entry for update", "id", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch diary entry for update"))
	}

	// Update the entry fields (moved from service)
	existingEntry.Title = title
	existingEntry.Content = req.Msg.Content
	existingEntry.UpdatedAt = time.Now() // Set timestamp here

	// Validate the updated entry (moved from service)
	if err := existingEntry.Validate(); err != nil {
		h.log.WarnContext(ctx, "Validation failed for updated diary entry", "id", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid diary entry data: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Updating diary entry", "entryID", entryID, "userID", userID)
	updatedEntry, err := h.repo.Update(ctx, existingEntry) // Changed from diaryApp.UpdateDiaryEntry
	if err != nil {
		// Note: Update might also return ErrDiaryEntryNotFound if ID changed or deleted concurrently
		if errors.Is(err, postgres.ErrDiaryEntryNotFound) { // Use postgres error
			h.log.WarnContext(ctx, "Diary entry not found during update", "entryID", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		}
		h.log.ErrorContext(ctx, "Failed to update diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update diary entry"))
	}

	// Convert persistence model to protobuf message
	protoEntry := toProtoDiaryEntry(updatedEntry) // Use updatedEntry (now *postgres.DiaryEntry)

	// Create response
	res := connect.NewResponse(&v1.UpdateDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// ListDiaryEntries lists diary entries for the authenticated user
func (h *DiaryHandler) ListDiaryEntries(ctx context.Context, req *connect.Request[v1.ListDiaryEntriesRequest]) (*connect.Response[v1.ListDiaryEntriesResponse], error) {
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
	h.log.InfoContext(ctx, "Fetching diary entries for user", "userID", userID, "page", pageNumber, "pageSize", pageSize)
	entries, err := h.repo.FindByUser(ctx, userID, pageSize, offset) // Changed from diaryApp.ListDiaryEntries
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch diary entries", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch diary entries"))
	}

	// Get total count (from service)
	total, err := h.repo.CountByUser(ctx, userID)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count diary entries", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count diary entries"))
	}

	// Convert persistence models to protobuf messages
	protoEntries := make([]*v1.DiaryEntry, len(entries)) // entries is now []*postgres.DiaryEntry
	for i, entry := range entries {
		protoEntries[i] = toProtoDiaryEntry(entry) // Pass *postgres.DiaryEntry
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListDiaryEntriesResponse{
		DiaryEntries: protoEntries,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// GetDiaryEntry retrieves a specific diary entry by ID
func (h *DiaryHandler) GetDiaryEntry(ctx context.Context, req *connect.Request[v1.GetDiaryEntryRequest]) (*connect.Response[v1.GetDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching diary entry", "entryID", entryID, "userID", userID)
	entry, err := h.repo.FindByID(ctx, entryID, userID) // Changed from diaryApp.GetDiaryEntry
	if err != nil {
		if errors.Is(err, postgres.ErrDiaryEntryNotFound) { // Use postgres error
			h.log.WarnContext(ctx, "Diary entry not found", "entryID", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		}
		h.log.ErrorContext(ctx, "Failed to fetch diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch diary entry"))
	}

	// Convert persistence model to protobuf message
	protoEntry := toProtoDiaryEntry(entry) // entry is now *postgres.DiaryEntry

	// Create response
	res := connect.NewResponse(&v1.GetDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// DeleteDiaryEntry deletes a diary entry
func (h *DiaryHandler) DeleteDiaryEntry(ctx context.Context, req *connect.Request[v1.DeleteDiaryEntryRequest]) (*connect.Response[v1.DeleteDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Deleting diary entry", "entryID", entryID, "userID", userID)
	err = h.repo.Delete(ctx, entryID, userID) // Changed from diaryApp.DeleteDiaryEntry
	if err != nil {
		// Note: Delete doesn't return ErrDiaryEntryNotFound in the repo implementation currently
		// if errors.Is(err, postgres.ErrDiaryEntryNotFound) {
		// 	h.log.WarnContext(ctx, "Diary entry not found for deletion", "entryID", entryID, "userID", userID)
		// 	return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		// }
		h.log.ErrorContext(ctx, "Failed to delete diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete diary entry"))
	}

	// Create response
	res := connect.NewResponse(&v1.DeleteDiaryEntryResponse{
		Success: true,
	})

	return res, nil
}

// toProtoDiaryEntry converts a postgres.DiaryEntry to a v1.DiaryEntry
func toProtoDiaryEntry(entry *postgres.DiaryEntry) *v1.DiaryEntry { // Accept *postgres.DiaryEntry
	protoEntry := &v1.DiaryEntry{
		Id:        entry.ID.String(),
		UserId:    entry.UserID.String(),
		Content:   entry.Content,
		EntryDate: entry.EntryDate.Format("2006-01-02"),
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}

	// Handle optional fields
	if entry.Title != nil {
		protoEntry.Title = &wrapperspb.StringValue{Value: *entry.Title}
	}

	return protoEntry
}
