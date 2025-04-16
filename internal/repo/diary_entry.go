package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDiaryEntryNotFound is returned when a diary entry is not found
var ErrDiaryEntryNotFound = errors.New("diary entry not found")

// DiaryEntryRepository provides database operations for DiaryEntry
type DiaryEntryRepository struct {
	q *db.Queries
}

// NewDiaryEntryRepository creates a new PostgreSQL diary entry repository
func NewDiaryEntryRepository(pool *pgxpool.Pool) *DiaryEntryRepository { // Return exported type
	return &DiaryEntryRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new diary entry, accepting the current time.
func (r *DiaryEntryRepository) Create(ctx context.Context, userID uuid.UUID, title *string, content string, entryDate time.Time, now time.Time) (db.DiaryEntry, error) {
	var titleVal pgtype.Text
	if title != nil {
		titleVal = pgtype.Text{String: *title, Valid: true}
	}

	pgDate := pgtype.Date{Time: entryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    userID,
		Title:     titleVal,
		Content:   content,
		EntryDate: pgDate,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dbEntry, err := r.q.CreateDiaryEntry(ctx, params)
	if err != nil {
		return db.DiaryEntry{}, fmt.Errorf("failed to create diary entry: %w", err)
	}

	// Return generated struct directly
	return dbEntry, nil
}

// Update updates an existing diary entry, accepting the current time.
func (r *DiaryEntryRepository) Update(ctx context.Context, id, userID uuid.UUID, title *string, content string, now time.Time) (db.DiaryEntry, error) {
	var titleVal pgtype.Text
	if title != nil {
		titleVal = pgtype.Text{String: *title, Valid: true}
	}

	params := db.UpdateDiaryEntryParams{
		ID:        id,
		Title:     titleVal,
		Content:   content,
		UserID:    userID, // Need UserID to ensure user owns the entry being updated
		UpdatedAt: now,
	}

	dbEntry, err := r.q.UpdateDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DiaryEntry{}, ErrDiaryEntryNotFound // Return zero value and local error
		}
		return db.DiaryEntry{}, fmt.Errorf("failed to update diary entry: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbEntry, nil
}

// FindByID retrieves a diary entry by ID and user ID
func (r *DiaryEntryRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (db.DiaryEntry, error) {
	params := db.GetDiaryEntryByIDParams{
		ID:     id,
		UserID: userID,
	}

	dbEntry, err := r.q.GetDiaryEntryByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DiaryEntry{}, ErrDiaryEntryNotFound // Return zero value and local error
		}
		return db.DiaryEntry{}, fmt.Errorf("failed to find diary entry: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbEntry, nil
}

// FindByUser retrieves paginated diary entries for a user
func (r *DiaryEntryRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.DiaryEntry, error) {
	params := db.ListDiaryEntriesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbEntries, err := r.q.ListDiaryEntriesByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list diary entries: %w", err)
	}

	// Return generated structs directly
	return dbEntries, nil
}

// Delete deletes a diary entry by ID and user ID
func (r *DiaryEntryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteDiaryEntryParams{
		ID:     id,
		UserID: userID,
	}

	// 1. Check if the entry exists and belongs to the user *before* deleting.
	_, err := r.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, ErrDiaryEntryNotFound) {
			// Entry doesn't exist or doesn't belong to the user. Return the specific error.
			return ErrDiaryEntryNotFound
		}
		// Some other error occurred during the check.
		return fmt.Errorf("failed to check diary entry existence before delete: %w", err)
	}

	// 2. Entry exists, proceed with deletion.
	// We assume the sqlc generated DeleteDiaryEntry only returns error.
	err = r.q.DeleteDiaryEntry(ctx, params)
	if err != nil {
		// We don't expect ErrNoRows here anymore because we checked existence first.
		// Any error here is likely a real database issue.
		return fmt.Errorf("failed to execute delete diary entry query: %w", err)
	}

	return nil // Deletion successful
}

// CountByUser returns the total number of diary entries for a user
func (r *DiaryEntryRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountDiaryEntriesByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count diary entries: %w", err)
	}

	return count, nil
}
