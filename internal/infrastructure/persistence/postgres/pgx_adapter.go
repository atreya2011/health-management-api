package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Import the postgres driver
	"github.com/jackc/pgx/v5"
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
	// Create a connection string from the pool config
	connStr := a.Pool.Config().ConnString()
	
	// Open a standard database/sql connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()
	
	// Use the standard database/sql Query method
	return db.QueryContext(ctx, query, args...)
}

// CustomRow is a custom implementation of sql.Row
type CustomRow struct {
	row pgx.Row
}

// Scan implements the sql.Scanner interface
func (r *CustomRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// QueryRowContext implements the database/sql QueryRowContext method
func (a *PgxAdapter) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Instead of using the pgx adapter, let's use direct SQL for now
	// This is a temporary workaround until we can properly implement the adapter
	
	// Create a connection string from the pool config
	connStr := a.Pool.Config().ConnString()
	
	// Open a standard database/sql connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err) // In a real implementation, we would handle this error properly
	}
	
	// Use the standard database/sql QueryRow method
	return db.QueryRowContext(ctx, query, args...)
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
