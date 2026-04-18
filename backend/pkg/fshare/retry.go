package fshare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	maxRetryAttempts = 5
)

// baseRetryDelay is the starting delay for exponential backoff.
// Exposed as var (not const) so tests can shorten it via testOverrideRetryDelay.
var baseRetryDelay = 2 * time.Second

// requestWithBackoff retries flaky 400/5xx responses with exponential backoff.
// fshare has no 429/Retry-After — rate limiting surfaces as persistent 400s.
//
// Success criteria: response status 2xx or 403 (file-password, not transient).
// Retry delays: 2s, 4s, 8s, 16s (cumulative ~30s before final attempt).
func (c *Client) requestWithBackoff(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Snapshot the body so we can rebuild the request on retry.
	var bodyBytes []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("fshare: read request body: %w", err)
		}
		_ = req.Body.Close()
		bodyBytes = b
	}

	var lastErr error
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if attempt > 0 {
			delay := baseRetryDelay << (attempt - 1) // 2s, 4s, 8s, 16s
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		attemptReq := req.Clone(ctx)
		if bodyBytes != nil {
			attemptReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := c.httpClient.Do(attemptReq)
		if err != nil {
			lastErr = err
			continue
		}

		// 2xx or 403 (file-password) — return immediately.
		if resp.StatusCode < 400 || resp.StatusCode == 403 {
			return resp, nil
		}

		// 4xx (other) / 5xx — retry.
		_ = resp.Body.Close()
		lastErr = fmt.Errorf("http %d", resp.StatusCode)
	}
	return nil, fmt.Errorf("%w: %v", ErrRateLimit, lastErr)
}
