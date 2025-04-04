package testutil

import (
	"context"
	"fmt"

	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db" // Added db import
	"github.com/google/uuid"
)

// CreateTestUser creates a test user in the database using sqlc
// It now takes Queries directly as TestDatabase might not be needed everywhere.
func CreateTestUser(ctx context.Context, queries *db.Queries) (uuid.UUID, error) {
	subjectID := fmt.Sprintf("test|%s", uuid.New().String())
	user, err := queries.CreateUser(ctx, subjectID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not create test user: %w", err)
	}
	return user.ID, nil
}
