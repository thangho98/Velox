package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thawng/velox/internal/model"
)

// RefreshTokenRepo handles refresh token database operations
type RefreshTokenRepo struct {
	db DBTX
}

// NewRefreshTokenRepo creates a new refresh token repository
func NewRefreshTokenRepo(db DBTX) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Create inserts a new refresh token
func (r *RefreshTokenRepo) Create(ctx context.Context, userID int64, tokenHash string, deviceName, ipAddress string, expiresAt time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, device_name, ip_address, expires_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id`,
		userID, tokenHash, deviceName, ipAddress, expiresAt).Scan(&id)
	return id, err
}

// GetByTokenHash retrieves a refresh token by its hash
func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, device_name, ip_address, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = ?`, tokenHash).
		Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.DeviceName, &rt.IPAddress, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Delete removes a refresh token by ID
func (r *RefreshTokenRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE id = ?", id)
	return err
}

// DeleteByUserID removes all refresh tokens for a user
func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = ?", userID)
	return err
}

// DeleteExpired removes all expired refresh tokens
func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE expires_at < datetime('now')")
	return err
}

// SessionRepo handles session database operations
type SessionRepo struct {
	db DBTX
}

// NewSessionRepo creates a new session repository
func NewSessionRepo(db DBTX) *SessionRepo {
	return &SessionRepo{db: db}
}

// Create inserts a new session
func (r *SessionRepo) Create(ctx context.Context, userID int64, refreshTokenID *int64, deviceName, ipAddress, userAgent string, expiresAt time.Time) (int64, error) {
	var id int64
	var err error

	if refreshTokenID != nil {
		err = r.db.QueryRowContext(ctx,
			`INSERT INTO sessions (user_id, refresh_token_id, device_name, ip_address, user_agent, expires_at)
			VALUES (?, ?, ?, ?, ?, ?)
			RETURNING id`,
			userID, *refreshTokenID, deviceName, ipAddress, userAgent, expiresAt).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx,
			`INSERT INTO sessions (user_id, device_name, ip_address, user_agent, expires_at)
			VALUES (?, ?, ?, ?, ?)
			RETURNING id`,
			userID, deviceName, ipAddress, userAgent, expiresAt).Scan(&id)
	}
	return id, err
}

// GetByID retrieves a session by ID
func (r *SessionRepo) GetByID(ctx context.Context, id int64) (*model.Session, error) {
	var s model.Session
	var lastActive sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, refresh_token_id, device_name, ip_address, user_agent, expires_at, last_active_at, created_at
		FROM sessions WHERE id = ?`, id).
		Scan(&s.ID, &s.UserID, &s.RefreshTokenID, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.ExpiresAt, &lastActive, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	if lastActive.Valid {
		s.LastActiveAt = lastActive.Time
	}
	return &s, nil
}

// ListByUserID retrieves all sessions for a user
func (r *SessionRepo) ListByUserID(ctx context.Context, userID int64) ([]model.Session, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, refresh_token_id, device_name, ip_address, user_agent, expires_at, last_active_at, created_at
		FROM sessions WHERE user_id = ? ORDER BY last_active_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var lastActive sql.NullTime
		if err := rows.Scan(&s.ID, &s.UserID, &s.RefreshTokenID, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.ExpiresAt, &lastActive, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		if lastActive.Valid {
			s.LastActiveAt = lastActive.Time
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// Delete removes a session by ID
func (r *SessionRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	return err
}

// DeleteByUserID removes all sessions for a user
func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// DeleteByRefreshTokenID removes a session by refresh token ID
func (r *SessionRepo) DeleteByRefreshTokenID(ctx context.Context, rtID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE refresh_token_id = ?", rtID)
	return err
}

// DeleteExpired removes all expired sessions
func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < datetime('now')")
	return err
}

// UpdateLastActive updates the last_active_at timestamp
func (r *SessionRepo) UpdateLastActive(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sessions SET last_active_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// UpdateLastActiveByUserID updates last_active_at for all sessions of a user
func (r *SessionRepo) UpdateLastActiveByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sessions SET last_active_at = CURRENT_TIMESTAMP WHERE user_id = ?", userID)
	return err
}
