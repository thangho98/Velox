package repository

import (
	"context"
	"database/sql"
	"errors"
)

// ErrNotFound is returned when an update/delete affects zero rows.
var ErrNotFound = errors.New("not found")

// DBTX is the common interface between *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
