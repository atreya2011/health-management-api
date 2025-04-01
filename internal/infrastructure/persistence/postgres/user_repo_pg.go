package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgUserRepository implements the domain.UserRepository interface
type pgUserRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPgUserRepository creates a new PostgreSQL user repository
func NewPgUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	adapter := NewPgxAdapter(pool)
	return &pgUserRepository{
		pool: pool,
		q:    db.New(adapter),
	}
}

// Create creates a new user record
func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
	dbUser, err := r.q.CreateUser(ctx, user.SubjectID)
	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return fmt.Errorf("user with this subject ID already exists: %w", err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// Update the domain user with the generated ID and timestamps
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	
	return nil
}

// FindByID retrieves a user by their internal UUID
func (r *pgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return toDomainUser(dbUser), nil
}

// FindBySubjectID retrieves a user by their JWT Subject claim
func (r *pgUserRepository) FindBySubjectID(ctx context.Context, subjectID string) (*domain.User, error) {
	dbUser, err := r.q.GetUserBySubjectID(ctx, subjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by subject ID: %w", err)
	}
	return toDomainUser(dbUser), nil
}

// toDomainUser converts a db.User to a domain.User
func toDomainUser(dbUser db.User) *domain.User {
	return &domain.User{
		ID:        dbUser.ID,
		SubjectID: dbUser.SubjectID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}
