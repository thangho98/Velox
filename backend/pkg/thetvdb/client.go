package thetvdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	baseURL = "https://api4.thetvdb.com/v4"

	// DefaultAPIKey is a built-in TheTVDB v4 API key.
	// Users can override this with their own key via Settings → Metadata.
	DefaultAPIKey = "34fd1f37-fd18-49d3-9601-93d9ebf3e038"
)

// Client for TheTVDB API v4.
// Uses API-key-based login to obtain a JWT, then Bearer auth for all requests.
type Client struct {
	apiKey     string
	httpClient *http.Client

	mu    sync.RWMutex
	token string
}

// New creates a new TheTVDB client.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// login obtains a JWT token from TheTVDB.
func (c *Client) login(ctx context.Context) error {
	body, err := json.Marshal(map[string]string{"apikey": c.apiKey})
	if err != nil {
		return fmt.Errorf("thetvdb: marshal login: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("thetvdb: login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("thetvdb: login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("thetvdb: login error %d: %s", resp.StatusCode, string(respBody))
	}

	var result apiResponse[LoginResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("thetvdb: decode login: %w", err)
	}

	c.mu.Lock()
	c.token = result.Data.Token
	c.mu.Unlock()

	return nil
}

// ensureToken logs in if no token is set.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.RLock()
	hasToken := c.token != ""
	c.mu.RUnlock()

	if hasToken {
		return nil
	}
	return c.login(ctx)
}

// do executes an authenticated request. Retries once on 401 (expired token).
func (c *Client) do(ctx context.Context, method, path string, params url.Values, v any) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}

	for attempt := 0; attempt < 2; attempt++ {
		u := baseURL + path
		if params != nil {
			u += "?" + params.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, method, u, nil)
		if err != nil {
			return fmt.Errorf("thetvdb: create request: %w", err)
		}

		c.mu.RLock()
		token := c.token
		c.mu.RUnlock()

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("thetvdb: request failed: %w", err)
		}
		defer resp.Body.Close()

		// Retry on 401 with fresh token
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			resp.Body.Close()
			if err := c.login(ctx); err != nil {
				return err
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("thetvdb: API error %d: %s", resp.StatusCode, string(body))
		}

		if v != nil {
			return json.NewDecoder(resp.Body).Decode(v)
		}
		return nil
	}

	return fmt.Errorf("thetvdb: request failed after retry")
}

// Search searches for series or movies by title.
// searchType: "series", "movie", or "" for all types.
func (c *Client) Search(ctx context.Context, query string, searchType string, year int) ([]SearchResult, error) {
	params := url.Values{"query": {query}}
	if searchType != "" {
		params.Set("type", searchType)
	}
	if year > 0 {
		params.Set("year", strconv.Itoa(year))
	}

	var result apiResponse[[]SearchResult]
	if err := c.do(ctx, http.MethodGet, "/search", params, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetSeries retrieves a series by TVDB ID.
func (c *Client) GetSeries(ctx context.Context, id int) (*SeriesBase, error) {
	var result apiResponse[SeriesBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/series/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeriesExtended retrieves extended series details (seasons, artworks, etc).
func (c *Client) GetSeriesExtended(ctx context.Context, id int, short bool) (*SeriesBase, error) {
	params := url.Values{}
	if short {
		params.Set("short", "true")
	}

	var result apiResponse[SeriesBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/series/%d/extended", id), params, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeriesEpisodes retrieves episodes for a series using the given season type.
// seasonType: "official" (default), "dvd", "absolute", etc.
func (c *Client) GetSeriesEpisodes(ctx context.Context, seriesID int, seasonType string, seasonNumber int, page int) (*SeriesEpisodesResponse, error) {
	if seasonType == "" {
		seasonType = "official"
	}
	params := url.Values{}
	if seasonNumber > 0 {
		params.Set("season", strconv.Itoa(seasonNumber))
	}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var result apiResponse[SeriesEpisodesResponse]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/series/%d/episodes/%s", seriesID, seasonType), params, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeason retrieves a season by ID.
func (c *Client) GetSeason(ctx context.Context, id int) (*SeasonBase, error) {
	var result apiResponse[SeasonBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/seasons/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeasonExtended retrieves a season with its episodes.
func (c *Client) GetSeasonExtended(ctx context.Context, id int) (*SeasonExtended, error) {
	var result apiResponse[SeasonExtended]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/seasons/%d/extended", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetEpisode retrieves an episode by ID.
func (c *Client) GetEpisode(ctx context.Context, id int) (*EpisodeBase, error) {
	var result apiResponse[EpisodeBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/episodes/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetEpisodeExtended retrieves extended episode details.
func (c *Client) GetEpisodeExtended(ctx context.Context, id int) (*EpisodeExtended, error) {
	var result apiResponse[EpisodeExtended]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/episodes/%d/extended", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetMovie retrieves a movie by TVDB ID.
func (c *Client) GetMovie(ctx context.Context, id int) (*MovieBase, error) {
	var result apiResponse[MovieBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/movies/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetMovieExtended retrieves extended movie details.
func (c *Client) GetMovieExtended(ctx context.Context, id int) (*MovieExtended, error) {
	params := url.Values{"short": {"true"}}

	var result apiResponse[MovieExtended]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/movies/%d/extended", id), params, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeriesTranslation retrieves a translation for a series.
func (c *Client) GetSeriesTranslation(ctx context.Context, id int, lang string) (*TranslationRecord, error) {
	var result apiResponse[TranslationRecord]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/series/%d/translations/%s", id, lang), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSeriesArtworks retrieves artworks for a series with optional language/type filter.
func (c *Client) GetSeriesArtworks(ctx context.Context, id int, lang string, artType int) ([]ArtworkRecord, error) {
	params := url.Values{}
	if lang != "" {
		params.Set("lang", lang)
	}
	if artType > 0 {
		params.Set("type", strconv.Itoa(artType))
	}

	var result apiResponse[SeriesBase]
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/series/%d/artworks", id), params, &result); err != nil {
		return nil, err
	}
	return result.Data.Artworks, nil
}
