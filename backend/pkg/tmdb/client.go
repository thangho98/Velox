package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	baseURL = "https://api.themoviedb.org/3"
)

// Client for TMDb API.
// Uses Bearer token auth (TMDb v4 read access token) which works with v3 endpoints.
type Client struct {
	apiKey     string // v4 read access token (not v3 API key)
	httpClient *http.Client
	config     *Configuration
}

// Configuration for image URLs
type Configuration struct {
	Images struct {
		BaseURL       string   `json:"base_url"`
		SecureBaseURL string   `json:"secure_base_url"`
		BackdropSizes []string `json:"backdrop_sizes"`
		LogoSizes     []string `json:"logo_sizes"`
		PosterSizes   []string `json:"poster_sizes"`
		ProfileSizes  []string `json:"profile_sizes"`
		StillSizes    []string `json:"still_sizes"`
	} `json:"images"`
}

// New creates a new TMDb client
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewWithHTTPClient creates a client with custom HTTP client
func NewWithHTTPClient(apiKey string, httpClient *http.Client) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// LoadConfiguration fetches TMDb configuration for image URLs
func (c *Client) LoadConfiguration(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", "/configuration", nil)
	if err != nil {
		return err
	}

	var config Configuration
	if err := c.do(req, &config); err != nil {
		return err
	}

	c.config = &config
	return nil
}

// GetImageURL returns a full image URL
func (c *Client) GetImageURL(path string, size string) string {
	if c.config == nil || path == "" {
		return ""
	}
	return c.config.Images.SecureBaseURL + size + path
}

// SearchMovies searches for movies by title
func (c *Client) SearchMovies(ctx context.Context, query string, year int, page int) (*MovieSearchResults, error) {
	params := url.Values{
		"query": {query},
		"page":  {strconv.Itoa(page)},
	}
	if year > 0 {
		params.Set("year", strconv.Itoa(year))
	}

	req, err := c.newRequest(ctx, "GET", "/search/movie", params)
	if err != nil {
		return nil, err
	}

	var results MovieSearchResults
	if err := c.do(req, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// SearchTV searches for TV shows by title
func (c *Client) SearchTV(ctx context.Context, query string, year int, page int) (*TVSearchResults, error) {
	params := url.Values{
		"query": {query},
		"page":  {strconv.Itoa(page)},
	}
	if year > 0 {
		params.Set("first_air_date_year", strconv.Itoa(year))
	}

	req, err := c.newRequest(ctx, "GET", "/search/tv", params)
	if err != nil {
		return nil, err
	}

	var results TVSearchResults
	if err := c.do(req, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// GetMovieDetails gets full details for a movie
func (c *Client) GetMovieDetails(ctx context.Context, movieID int) (*MovieDetails, error) {
	params := url.Values{
		"append_to_response": {"credits,keywords"},
	}

	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/movie/%d", movieID), params)
	if err != nil {
		return nil, err
	}

	var details MovieDetails
	if err := c.do(req, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetTVDetails gets full details for a TV show
func (c *Client) GetTVDetails(ctx context.Context, tvID int) (*TVDetails, error) {
	params := url.Values{
		"append_to_response": {"credits,keywords,external_ids"},
	}

	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/tv/%d", tvID), params)
	if err != nil {
		return nil, err
	}

	var details TVDetails
	if err := c.do(req, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetTVSeason gets details for a specific season
func (c *Client) GetTVSeason(ctx context.Context, tvID int, seasonNumber int) (*SeasonDetails, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/tv/%d/season/%d", tvID, seasonNumber), nil)
	if err != nil {
		return nil, err
	}

	var details SeasonDetails
	if err := c.do(req, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetTVEpisode gets details for a specific episode
func (c *Client) GetTVEpisode(ctx context.Context, tvID int, seasonNumber int, episodeNumber int) (*EpisodeDetails, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/tv/%d/season/%d/episode/%d", tvID, seasonNumber, episodeNumber), nil)
	if err != nil {
		return nil, err
	}

	var details EpisodeDetails
	if err := c.do(req, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// FindByExternalID finds media by external ID (IMDb, etc.)
func (c *Client) FindByExternalID(ctx context.Context, externalID string, source string) (*FindResults, error) {
	params := url.Values{
		"external_source": {source},
	}

	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/find/%s", externalID), params)
	if err != nil {
		return nil, err
	}

	var results FindResults
	if err := c.do(req, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// newRequest creates a new HTTP request with context
func (c *Client) newRequest(ctx context.Context, method, path string, params url.Values) (*http.Request, error) {
	u := baseURL + path
	if params != nil {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// do executes the request and decodes the response
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tmdb API error: %d - %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
