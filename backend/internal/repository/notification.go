package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/thawng/velox/internal/model"
)

// NotificationRepo handles notifications database operations
type NotificationRepo struct {
	db DBTX
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo(db DBTX) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Create inserts a new notification into the database
func (r *NotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, type, title, message, data, read, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		n.UserID, n.Type, n.Title, n.Message, n.Data, boolToInt(n.Read), n.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	n.ID = id
	return nil
}

// GetByID retrieves a single notification by ID
func (r *NotificationRepo) GetByID(ctx context.Context, id int64) (*model.Notification, error) {
	var n model.Notification
	var dataStr string
	var readInt int
	var readAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, type, title, message, data, read, created_at, read_at
		 FROM notifications WHERE id = ?`, id).
		Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &dataStr, &readInt, &n.CreatedAt, &readAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting notification: %w", err)
	}

	n.Read = readInt == 1
	n.Data = json.RawMessage(dataStr)
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}

	return &n, nil
}

// GetByUser retrieves notifications for a user with optional filtering
func (r *NotificationRepo) GetByUser(ctx context.Context, filter model.NotificationFilter) ([]model.Notification, error) {
	query := `SELECT id, user_id, type, title, message, data, read, created_at, read_at
		      FROM notifications WHERE 1=1`
	args := []any{}

	if filter.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filter.UserID)
	}

	if filter.UnreadOnly {
		query += " AND read = 0"
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	defer rows.Close()

	var items []model.Notification
	for rows.Next() {
		var n model.Notification
		var dataStr string
		var readInt int
		var readAt sql.NullTime

		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
			&dataStr, &readInt, &n.CreatedAt, &readAt); err != nil {
			return nil, fmt.Errorf("scanning notification: %w", err)
		}

		n.Read = readInt == 1
		n.Data = json.RawMessage(dataStr)
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}

		items = append(items, n)
	}

	return items, rows.Err()
}

// MarkAsRead marks specific notifications as read for a user
func (r *NotificationRepo) MarkAsRead(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	// args order must match query: read_at=?, id IN (ids...), user_id=?
	args := make([]any, 0, len(ids)+2)
	args = append(args, time.Now())
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, userID)

	query := fmt.Sprintf(
		`UPDATE notifications SET read = 1, read_at = ?
		 WHERE id IN (%s) AND user_id = ?`,
		strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("marking notifications as read: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = 1, read_at = ?
		 WHERE user_id = ? AND read = 0`,
		time.Now(), userID)
	if err != nil {
		return fmt.Errorf("marking all notifications as read: %w", err)
	}
	return nil
}

// Delete removes specific notifications for a user
func (r *NotificationRepo) Delete(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	// args order must match query: id IN (ids...), user_id=?
	args := make([]any, 0, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, userID)

	query := fmt.Sprintf(
		`DELETE FROM notifications WHERE id IN (%s) AND user_id = ?`,
		strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting notifications: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteOld removes notifications older than a specific time (for cleanup)
func (r *NotificationRepo) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM notifications WHERE created_at < ? AND read = 1`,
		before)
	if err != nil {
		return 0, fmt.Errorf("deleting old notifications: %w", err)
	}

	return result.RowsAffected()
}

// CountUnread returns the number of unread notifications for a user
func (r *NotificationRepo) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0`,
		userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
