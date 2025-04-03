package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgDiaryEntryRepository implements the domain.DiaryEntryRepository interface
type pgDiaryEntryRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPgDiaryEntryRepository creates a new PostgreSQL diary entry repository
func NewPgDiaryEntryRepository(pool *pgxpool.Pool) domain.DiaryEntryRepository {
	adapter := NewPgxAdapter(pool)
	return &pgDiaryEntryRepository{
		pool: pool,
		q:    db.New(adapter),
	}
}

// Create creates a new diary entry
func (r *pgDiaryEntryRepository) Create(ctx context.Context, entry *domain.DiaryEntry) (*domain.DiaryEntry, error) {
	// Convert *string to sql.NullString
	var title sql.NullString
	if entry.Title != nil {
		title = sql.NullString{
			String: *entry.Title,
			Valid:  true,
		}
	}

	params := db.CreateDiaryEntryParams{
		UserID:    entry.UserID,
		Title:     title,
		Content:   entry.Content,
		EntryDate: entry.EntryDate,
	}

	dbEntry, err := r.q.CreateDiaryEntry(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create diary entry: %w", err)
	}

	return toDomainDiaryEntry(dbEntry), nil
}

// Update updates an existing diary entry
func (r *pgDiaryEntryRepository) Update(ctx context.Context, entry *domain.DiaryEntry) (*domain.DiaryEntry, error) {
	// Convert *string to sql.NullString
	var title sql.NullString
	if entry.Title != nil {
		title = sql.NullString{
			String: *entry.Title,
			Valid:  true,
		}
	}

	params := db.UpdateDiaryEntryParams{
		ID:      entry.ID,
		Title:   title,
		Content: entry.Content,
		UserID:  entry.UserID,
	}

	dbEntry, err := r.q.UpdateDiaryEntry(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDiaryEntryNotFound
		}
		return nil, fmt.Errorf("failed to update diary entry: %w", err)
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
		if err == sql.ErrNoRows {
			return nil, domain.ErrDiaryEntryNotFound
		}
		return nil, fmt.Errorf("failed to find diary entry: %w", err)
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
		if err == sql.ErrNoRows {
			return domain.ErrDiaryEntryNotFound
		}
		return fmt.Errorf("failed to delete diary entry: %w", err)
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

// toDomainDiaryEntry converts a db.DiaryEntry to a domain.DiaryEntry
func toDomainDiaryEntry(dbEntry db.DiaryEntry) *domain.DiaryEntry {
	// Convert sql.NullString to *string
	var title *string
	if dbEntry.Title.Valid {
		t := dbEntry.Title.String
		title = &t
	}

	return &domain.DiaryEntry{
		ID:        dbEntry.ID,
		UserID:    dbEntry.UserID,
		Title:     title,
		Content:   dbEntry.Content,
		EntryDate: dbEntry.EntryDate,
		CreatedAt: dbEntry.CreatedAt,
		UpdatedAt: dbEntry.UpdatedAt,
	}
}
