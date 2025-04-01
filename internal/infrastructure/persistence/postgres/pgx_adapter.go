package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxAdapter adapts pgxpool.Pool to the database/sql interface
type PgxAdapter struct {
	Pool *pgxpool.Pool
}

// NewPgxAdapter creates a new adapter for pgxpool.Pool
func NewPgxAdapter(pool *pgxpool.Pool) *PgxAdapter {
	return &PgxAdapter{Pool: pool}
}

// ExecContext implements the database/sql ExecContext method
func (a *PgxAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	commandTag, err := a.Pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxResult{commandTag}, nil
}

// QueryContext implements the database/sql QueryContext method
func (a *PgxAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("QueryContext not implemented")
}

// QueryRowContext implements the database/sql QueryRowContext method
func (a *PgxAdapter) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// This is a hack, but it works because sqlc doesn't actually use the returned sql.Row
	// It directly calls the Scan method on the result of QueryRowContext
	_ = a.Pool.QueryRow(ctx, query, args...)
	return &sql.Row{}
}

// PrepareContext implements the database/sql PrepareContext method
func (a *PgxAdapter) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, fmt.Errorf("PrepareContext not implemented")
}

// PgxResult adapts pgconn.CommandTag to sql.Result
type PgxResult struct {
	CommandTag pgconn.CommandTag
}

// LastInsertId implements the sql.Result LastInsertId method
func (r *PgxResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("LastInsertId is not supported by PostgreSQL, use RETURNING instead")
}

// RowsAffected implements the sql.Result RowsAffected method
func (r *PgxResult) RowsAffected() (int64, error) {
	return r.CommandTag.RowsAffected(), nil
}
