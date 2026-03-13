package omdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL = "https://www.omdbapi.com/"

	// DefaultAPIKey is a built-in OMDb API key for convenience.
	// Admins can override it in Settings → Metadata.
	DefaultAPIKey = "9f3b182f"
)

// Client is an OMDb API client.
type Client struct {
	apiKey string
	http   *http.Client
}

// New creates a new OMDb client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetByIMDbID fetches movie/show details by IMDb ID (e.g., "tt3896198").
func (c *Client) GetByIMDbID(ctx context.Context, imdbID string) (*Response, error) {
	url := fmt.Sprintf("%s?apikey=%s&i=%s", baseURL, c.apiKey, imdbID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("omdb: creating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("omdb: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("omdb: unexpected status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("omdb: decoding response: %w", err)
	}

	if result.Response == "False" {
		return nil, fmt.Errorf("omdb: %s", result.Error)
	}

	return &result, nil
}
