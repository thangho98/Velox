package repository

import (
	"context"
	"fmt"
)

// AdminRepo provides admin-level database operations
type AdminRepo struct {
	db DBTX
}

func NewAdminRepo(db DBTX) *AdminRepo {
	return &AdminRepo{db: db}
}

// CountTable returns the row count for a table.
// tableName must be one of: media, series, users, libraries
func (r *AdminRepo) CountTable(ctx context.Context, tableName string) (int, error) {
	// Validate table name to prevent SQL injection
	switch tableName {
	case "media", "series", "users", "libraries":
	default:
		return 0, fmt.Errorf("invalid table name: %s", tableName)
	}

	var count int
	row := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName))
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("counting %s: %w", tableName, err)
	}
	return count, nil
}
