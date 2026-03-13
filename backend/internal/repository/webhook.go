package repository

import (
	"context"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// WebhookRepo handles webhook database operations
type WebhookRepo struct {
	db DBTX
}

func NewWebhookRepo(db DBTX) *WebhookRepo {
	return &WebhookRepo{db: db}
}

// Create inserts a new webhook.
func (r *WebhookRepo) Create(ctx context.Context, w *model.Webhook) error {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO webhooks (url, events, secret, active)
		 VALUES (?, ?, ?, ?)
		 RETURNING id, created_at, updated_at`,
		w.URL, w.Events, w.Secret, w.Active)
	return row.Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

// GetByID retrieves a webhook by ID.
func (r *WebhookRepo) GetByID(ctx context.Context, id int64) (*model.Webhook, error) {
	var w model.Webhook
	row := r.db.QueryRowContext(ctx,
		`SELECT id, url, events, secret, active, created_at, updated_at
		 FROM webhooks WHERE id = ?`, id)
	if err := row.Scan(&w.ID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, fmt.Errorf("getting webhook %d: %w", id, err)
	}
	return &w, nil
}

// List retrieves all webhooks.
func (r *WebhookRepo) List(ctx context.Context) ([]model.Webhook, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, url, events, secret, active, created_at, updated_at
		 FROM webhooks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks: %w", err)
	}
	defer rows.Close()

	var items []model.Webhook
	for rows.Next() {
		var w model.Webhook
		if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

// Update updates a webhook's URL, events, secret, and active status.
func (r *WebhookRepo) Update(ctx context.Context, w *model.Webhook) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE webhooks SET url = ?, events = ?, secret = ?, active = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		w.URL, w.Events, w.Secret, w.Active, w.ID)
	if err != nil {
		return fmt.Errorf("updating webhook %d: %w", w.ID, err)
	}
	return nil
}

// Delete removes a webhook by ID.
func (r *WebhookRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting webhook %d: %w", id, err)
	}
	return nil
}

// ListByEvent retrieves active webhooks that subscribe to a given event.
// Events are stored as JSON arrays, e.g. ["scan.complete","media.added"].
func (r *WebhookRepo) ListByEvent(ctx context.Context, event string) ([]model.Webhook, error) {
	// Use JSON_EACH to search within the JSON array of events
	rows, err := r.db.QueryContext(ctx,
		`SELECT w.id, w.url, w.events, w.secret, w.active, w.created_at, w.updated_at
		 FROM webhooks w, json_each(w.events) e
		 WHERE w.active = 1 AND (e.value = ? OR e.value = '*')`,
		event)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks for event %s: %w", event, err)
	}
	defer rows.Close()

	var items []model.Webhook
	for rows.Next() {
		var w model.Webhook
		if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		items = append(items, w)
	}
	return items, rows.Err()
}
