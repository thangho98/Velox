package fanart

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL = "https://webservice.fanart.tv/v3"
)

// Client for fanart.tv API v3.
// Uses simple API key authentication via query parameter.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a new fanart.tv client.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do executes an API request with the API key.
func (c *Client) do(ctx context.Context, path string, v any) error {
	u := baseURL + path + "?api_key=" + c.apiKey

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("fanart: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fanart: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fanart: API error %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// GetMovieImages retrieves fanart images for a movie by TMDb ID.
func (c *Client) GetMovieImages(ctx context.Context, tmdbID int) (*MovieImages, error) {
	var result MovieImages
	if err := c.do(ctx, fmt.Sprintf("/movies/%d", tmdbID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetShowImages retrieves fanart images for a TV show by TVDB ID.
func (c *Client) GetShowImages(ctx context.Context, tvdbID int) (*ShowImages, error) {
	var result ShowImages
	if err := c.do(ctx, fmt.Sprintf("/tv/%d", tvdbID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
