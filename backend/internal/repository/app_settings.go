package repository

import (
	"context"
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

// AppSettingsRepo handles CRUD for the app_settings table.
type AppSettingsRepo struct {
	db DBTX
}

// NewAppSettingsRepo creates a new settings repository.
func NewAppSettingsRepo(db DBTX) *AppSettingsRepo {
	return &AppSettingsRepo{db: db}
}

// Get returns a setting value by key. Returns empty string if not found.
func (r *AppSettingsRepo) Get(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// Set upserts a setting key-value pair.
func (r *AppSettingsRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		key, value)
	return err
}

// GetMulti returns a map of key-value pairs for the given keys.
func (r *AppSettingsRepo) GetMulti(ctx context.Context, keys ...string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, k := range keys {
		result[k] = ""
	}

	if len(keys) == 0 {
		return result, nil
	}

	// Build query with placeholders
	query := `SELECT key, value FROM app_settings WHERE key IN (`
	args := make([]any, len(keys))
	for i, k := range keys {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = k
	}
	query += ")"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.AppSetting
		if err := rows.Scan(&s.Key, &s.Value); err != nil {
			return nil, err
		}
		result[s.Key] = s.Value
	}
	return result, rows.Err()
}
