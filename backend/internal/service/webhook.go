package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

const webhookTimeout = 5 * time.Second

// WebhookService manages webhook subscriptions and dispatch.
type WebhookService struct {
	repo *repository.WebhookRepo
}

func NewWebhookService(repo *repository.WebhookRepo) *WebhookService {
	return &WebhookService{repo: repo}
}

// Create adds a new webhook.
func (s *WebhookService) Create(ctx context.Context, w *model.Webhook) error {
	return s.repo.Create(ctx, w)
}

// GetByID retrieves a webhook.
func (s *WebhookService) GetByID(ctx context.Context, id int64) (*model.Webhook, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all webhooks.
func (s *WebhookService) List(ctx context.Context) ([]model.Webhook, error) {
	return s.repo.List(ctx)
}

// Update modifies an existing webhook.
func (s *WebhookService) Update(ctx context.Context, w *model.Webhook) error {
	return s.repo.Update(ctx, w)
}

// Delete removes a webhook.
func (s *WebhookService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// Dispatch sends an event payload to all active webhooks matching the event.
// Each webhook is called in its own goroutine with a timeout.
func (s *WebhookService) Dispatch(event string, payload any) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	hooks, err := s.repo.ListByEvent(ctx, event)
	if err != nil {
		log.Printf("webhook dispatch: failed to list hooks for %s: %v", event, err)
		return
	}

	if len(hooks) == 0 {
		return
	}

	body, err := json.Marshal(map[string]any{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	})
	if err != nil {
		log.Printf("webhook dispatch: failed to marshal payload for %s: %v", event, err)
		return
	}

	for _, hook := range hooks {
		go s.send(hook, body)
	}
}

func (s *WebhookService) send(hook model.Webhook, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), webhookTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(body))
	if err != nil {
		log.Printf("webhook: failed to create request for %s: %v", hook.URL, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Velox-Webhook/0.1")

	// Sign with HMAC-SHA256 if secret is set
	if hook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(hook.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Velox-Signature", fmt.Sprintf("sha256=%s", sig))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("webhook: failed to send to %s: %v", hook.URL, err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("webhook: %s returned %d", hook.URL, resp.StatusCode)
	}
}
