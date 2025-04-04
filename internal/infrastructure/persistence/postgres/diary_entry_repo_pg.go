package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// pgDiaryEntryRepository implements the domain.DiaryEntryRepository interface
type pgDiaryEntryRepository struct {
	q *db.Queries
}

// NewPgDiaryEntryRepository creates a new PostgreSQL diary entry repository
func NewPgDiaryEntryRepository(pool *pgxpool.Pool) domain.DiaryEntryRepository {
	return &pgDiaryEntryRepository{
		q: db.New(pool),
	}
}

// Create creates a new diary entry
func (r *pgDiaryEntryRepository) Create(ctx context.Context, entry *domain.DiaryEntry) (*domain.DiaryEntry, error) {
	var titleVal pgtype.Text
	if entry.Title != nil {
		titleVal = pgtype.Text{String: *entry.Title, Valid: true}
	}

	pgDate := pgtype.Date{Time: entry.EntryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    entry.UserID,
		Title:     titleVal,
		Content:   entry.Content,
		EntryDate: pgDate,
	}

	dbEntry, err := r.q.CreateDiaryEntry(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create diary entry: %w", err)
	}

	return toDomainDiaryEntry(dbEntry), nil
}

// Update updates an existing diary entry
func (r *pgDiaryEntryRepository) Update(ctx context.Context, entry *domain.DiaryEntry) (*domain.DiaryEntry, error) {
	var titleVal pgtype.Text
	if entry.Title != nil {
		titleVal = pgtype.Text{String: *entry.Title, Valid: true}
	}

	params := db.UpdateDiaryEntryParams{
		ID:      entry.ID,
		Title:   titleVal,
		Content: entry.Content,
		UserID:  entry.UserID,
	}

	dbEntry, err := r.q.UpdateDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDiaryEntryNotFound
		}
		return nil, errors.Wrap(err, "failed to update diary entry")
	}

	return toDomainDiaryEntry(dbEntry), nil
}

// FindByID retrieves a diary entry by ID and user ID
func (r *pgDiaryEntryRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*domain.DiaryEntry, error) {
	params := db.GetDiaryEntryByIDParams{
		ID:     id,
		UserID: userID,
	}

	dbEntry, err := r.q.GetDiaryEntryByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDiaryEntryNotFound
		}
		return nil, errors.Wrap(err, "failed to find diary entry")
	}

	return toDomainDiaryEntry(dbEntry), nil
}

// FindByUser retrieves paginated diary entries for a user
func (r *pgDiaryEntryRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.DiaryEntry, error) {
	params := db.ListDiaryEntriesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbEntries, err := r.q.ListDiaryEntriesByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list diary entries: %w", err)
	}

	entries := make([]*domain.DiaryEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		entries[i] = toDomainDiaryEntry(dbEntry)
	}

	return entries, nil
}

// Delete deletes a diary entry by ID and user ID
func (r *pgDiaryEntryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteDiaryEntryParams{
		ID:     id,
		UserID: userID,
	}

	err := r.q.DeleteDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return errors.Wrap(err, "failed to delete diary entry")
	}

	return nil
}

// CountByUser returns the total number of diary entries for a user
func (r *pgDiaryEntryRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountDiaryEntriesByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count diary entries: %w", err)
	}

	return count, nil
}

// toDomainDiaryEntry converts a db.DiaryEntry (pgx-based) to a domain.DiaryEntry
func toDomainDiaryEntry(dbEntry db.DiaryEntry) *domain.DiaryEntry {
	var title *string
	if dbEntry.Title.Valid {
		title = &dbEntry.Title.String
	}

	var entryDateVal time.Time
	if dbEntry.EntryDate.Valid {
		entryDateVal = dbEntry.EntryDate.Time
	}

	return &domain.DiaryEntry{
		ID:        dbEntry.ID,
		UserID:    dbEntry.UserID,
		Title:     title,
		Content:   dbEntry.Content,
		EntryDate: entryDateVal,
		CreatedAt: dbEntry.CreatedAt,
		UpdatedAt: dbEntry.UpdatedAt,
	}
}
