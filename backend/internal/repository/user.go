package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// UserRepo handles user database operations
type UserRepo struct {
	db DBTX
}

// NewUserRepo creates a new user repository
func NewUserRepo(db DBTX) *UserRepo {
	return &UserRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction
func (r *UserRepo) WithTx(tx *sql.Tx) *UserRepo {
	return &UserRepo{db: tx}
}

// Create inserts a new user
func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users
		(username, display_name, password_hash, is_admin, avatar_path)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	isAdmin := 0
	if u.IsAdmin {
		isAdmin = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		u.Username, u.DisplayName, u.PasswordHash, isAdmin, u.AvatarPath)

	return row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	var isAdmin int
	err := r.db.QueryRowContext(ctx, `SELECT id, username, display_name, password_hash,
		is_admin, avatar_path, created_at, updated_at
		FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash,
			&isAdmin, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin == 1
	return &u, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	var isAdmin int
	err := r.db.QueryRowContext(ctx, `SELECT id, username, display_name, password_hash,
		is_admin, avatar_path, created_at, updated_at
		FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash,
			&isAdmin, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin == 1
	return &u, nil
}

// List retrieves all users
func (r *UserRepo) List(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, username, display_name, password_hash,
		is_admin, avatar_path, created_at, updated_at
		FROM users ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		var isAdmin int
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash,
			&isAdmin, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		u.IsAdmin = isAdmin == 1
		users = append(users, &u)
	}
	return users, rows.Err()
}

// Update updates a user (excluding password - use UpdatePassword for that)
func (r *UserRepo) Update(ctx context.Context, u *model.User) error {
	isAdmin := 0
	if u.IsAdmin {
		isAdmin = 1
	}

	_, err := r.db.ExecContext(ctx, `UPDATE users SET
		username = ?, display_name = ?, is_admin = ?, avatar_path = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		u.Username, u.DisplayName, isAdmin, u.AvatarPath, u.ID)
	return err
}

// UpdatePassword updates only the password hash
func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, hash string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		hash, userID)
	return err
}

// Delete removes a user
func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// Count returns the total number of users
func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// CountAdmins returns the number of admin users
func (r *UserRepo) CountAdmins(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&count)
	return count, err
}

// GetLibraryIDs returns the library IDs a user has access to
func (r *UserRepo) GetLibraryIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT library_id FROM user_library_access WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("listing library access: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning library id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SetLibraryAccess sets which libraries a user has access to (replaces all existing)
func (r *UserRepo) SetLibraryAccess(ctx context.Context, userID int64, libraryIDs []int64) error {
	// Use transaction if we're not already in one
	tx, ok := r.db.(*sql.Tx)
	if !ok {
		// Need to get the underlying *sql.DB to start a transaction
		// This is a bit awkward - in practice, caller should use WithTx
		return fmt.Errorf("SetLibraryAccess requires a transaction - use WithTx")
	}

	// Delete existing access
	if _, err := tx.ExecContext(ctx,
		"DELETE FROM user_library_access WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("deleting existing access: %w", err)
	}

	// Insert new access
	for _, libID := range libraryIDs {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO user_library_access (user_id, library_id) VALUES (?, ?)",
			userID, libID); err != nil {
			return fmt.Errorf("inserting library access: %w", err)
		}
	}
	return nil
}

// GrantAllLibraries gives a user access to all existing libraries
func (r *UserRepo) GrantAllLibraries(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_library_access (user_id, library_id)
		SELECT ?, id FROM libraries
		ON CONFLICT (user_id, library_id) DO NOTHING`,
		userID)
	return err
}

// UserPreferencesRepo handles user preferences
type UserPreferencesRepo struct {
	db DBTX
}

// NewUserPreferencesRepo creates a new preferences repository
func NewUserPreferencesRepo(db DBTX) *UserPreferencesRepo {
	return &UserPreferencesRepo{db: db}
}

// Get retrieves user preferences
func (r *UserPreferencesRepo) Get(ctx context.Context, userID int64) (*model.UserPreferences, error) {
	var p model.UserPreferences
	err := r.db.QueryRowContext(ctx, `SELECT user_id, subtitle_language, audio_language,
		max_streaming_quality, theme
		FROM user_preferences WHERE user_id = ?`, userID).
		Scan(&p.UserID, &p.SubtitleLanguage, &p.AudioLanguage,
			&p.MaxStreamingQuality, &p.Theme)
	if err == sql.ErrNoRows {
		// Return defaults
		return &model.UserPreferences{
			UserID:              userID,
			SubtitleLanguage:    "",
			AudioLanguage:       "",
			MaxStreamingQuality: "auto",
			Theme:               "dark",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update updates user preferences (upsert)
func (r *UserPreferencesRepo) Update(ctx context.Context, p *model.UserPreferences) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_preferences (user_id, subtitle_language, audio_language, max_streaming_quality, theme)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			subtitle_language = excluded.subtitle_language,
			audio_language = excluded.audio_language,
			max_streaming_quality = excluded.max_streaming_quality,
			theme = excluded.theme`,
		p.UserID, p.SubtitleLanguage, p.AudioLanguage, p.MaxStreamingQuality, p.Theme)
	return err
}
