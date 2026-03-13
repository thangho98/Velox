package tvmaze

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "https://api.tvmaze.com"

// Client for TVmaze API (free, no API key required).
// Rate limit: 20 calls per 10 seconds.
type Client struct {
	httpClient *http.Client
	limiter    chan struct{}
}

// New creates a new TVmaze client with built-in rate limiting.
func New() *Client {
	limiter := make(chan struct{}, 20)
	for range 20 {
		limiter <- struct{}{}
	}

	// Refill tokens: 20 per 10 seconds = 1 every 500ms
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case limiter <- struct{}{}:
			default:
			}
		}
	}()

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: limiter,
	}
}

// do executes a rate-limited GET request.
func (c *Client) do(ctx context.Context, path string, v any) error {
	select {
	case <-c.limiter:
	case <-ctx.Done():
		return ctx.Err()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("tvmaze: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tvmaze: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("tvmaze: rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tvmaze: API error %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// SearchShows searches for TV shows by name.
func (c *Client) SearchShows(ctx context.Context, query string) ([]SearchResult, error) {
	var results []SearchResult
	if err := c.do(ctx, "/search/shows?q="+url.QueryEscape(query), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetShow retrieves a show by TVmaze ID.
func (c *Client) GetShow(ctx context.Context, id int) (*Show, error) {
	var show Show
	if err := c.do(ctx, fmt.Sprintf("/shows/%d", id), &show); err != nil {
		return nil, err
	}
	return &show, nil
}

// GetShowFull retrieves a show with embedded episodes and seasons.
func (c *Client) GetShowFull(ctx context.Context, id int) (*Show, error) {
	var show Show
	if err := c.do(ctx, fmt.Sprintf("/shows/%d?embed[]=episodes&embed[]=seasons", id), &show); err != nil {
		return nil, err
	}
	return &show, nil
}

// LookupByTVDB looks up a show by TheTVDB ID.
func (c *Client) LookupByTVDB(ctx context.Context, tvdbID int) (*Show, error) {
	var show Show
	if err := c.do(ctx, "/lookup/shows?thetvdb="+strconv.Itoa(tvdbID), &show); err != nil {
		return nil, err
	}
	return &show, nil
}

// LookupByIMDb looks up a show by IMDb ID (e.g. "tt0944947").
func (c *Client) LookupByIMDb(ctx context.Context, imdbID string) (*Show, error) {
	var show Show
	if err := c.do(ctx, "/lookup/shows?imdb="+url.QueryEscape(imdbID), &show); err != nil {
		return nil, err
	}
	return &show, nil
}

// GetEpisodes retrieves all episodes for a show.
func (c *Client) GetEpisodes(ctx context.Context, showID int) ([]Episode, error) {
	var episodes []Episode
	if err := c.do(ctx, fmt.Sprintf("/shows/%d/episodes", showID), &episodes); err != nil {
		return nil, err
	}
	return episodes, nil
}

// GetEpisodeByNumber retrieves a specific episode by season and episode number.
func (c *Client) GetEpisodeByNumber(ctx context.Context, showID, season, episode int) (*Episode, error) {
	var ep Episode
	path := fmt.Sprintf("/shows/%d/episodebynumber?season=%d&number=%d", showID, season, episode)
	if err := c.do(ctx, path, &ep); err != nil {
		return nil, err
	}
	return &ep, nil
}

// GetSeasons retrieves all seasons for a show.
func (c *Client) GetSeasons(ctx context.Context, showID int) ([]Season, error) {
	var seasons []Season
	if err := c.do(ctx, fmt.Sprintf("/shows/%d/seasons", showID), &seasons); err != nil {
		return nil, err
	}
	return seasons, nil
}
