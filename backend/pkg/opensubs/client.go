package opensubs

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

	"github.com/thawng/velox/pkg/subprovider"
)

const (
	baseURL   = "https://api.opensubtitles.com/api/v1"
	userAgent = "Velox v0.1.0"
)

// Client is an OpenSubtitles.com REST v3 client.
type Client struct {
	apiKey   string
	username string
	password string

	mu       sync.Mutex
	token    string
	tokenExp time.Time

	http *http.Client
}

// SearchParams configures a subtitle search query.
type SearchParams struct {
	ImdbID   string // "tt1234567" or "1234567"
	TmdbID   int
	Query    string // fallback: film title
	Language string // "en", "vi"
	Year     int
}

// New creates a new OpenSubtitles client.
func New(apiKey, username, password string) *Client {
	return &Client{
		apiKey:   apiKey,
		username: username,
		password: password,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Configured returns true if credentials are present.
func (c *Client) Configured() bool {
	return c.apiKey != "" && c.username != "" && c.password != ""
}

// login authenticates and caches the JWT token.
func (c *Client) login(ctx context.Context) error {
	body, _ := json.Marshal(map[string]string{
		"username": c.username,
		"password": c.password,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding login response: %w", err)
	}

	c.token = result.Token
	c.tokenExp = time.Now().Add(24 * time.Hour)
	return nil
}

// ensureToken checks expiry and re-authenticates if needed.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}
	return c.login(ctx)
}

// Search queries OpenSubtitles for matching subtitles.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("opensubtitles auth: %w", err)
	}

	q := url.Values{}
	if params.ImdbID != "" {
		q.Set("imdb_id", params.ImdbID)
	}
	if params.TmdbID > 0 {
		q.Set("tmdb_id", strconv.Itoa(params.TmdbID))
	}
	if params.Query != "" {
		q.Set("query", params.Query)
	}
	if params.Language != "" {
		q.Set("languages", params.Language)
	}
	if params.Year > 0 {
		q.Set("year", strconv.Itoa(params.Year))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/subtitles?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed (%d): %s", resp.StatusCode, string(b))
	}

	var apiResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}

	results := make([]subprovider.Result, 0, len(apiResp.Data))
	for _, item := range apiResp.Data {
		if len(item.Attributes.Files) == 0 {
			continue
		}
		file := item.Attributes.Files[0]

		results = append(results, subprovider.Result{
			Provider:        "opensubtitles",
			ExternalID:      strconv.Itoa(file.FileID),
			Title:           item.Attributes.Release,
			Language:        item.Attributes.Language,
			Format:          item.Attributes.Format,
			Downloads:       item.Attributes.DownloadCount,
			Rating:          item.Attributes.Ratings,
			Forced:          item.Attributes.ForeignPartsOnly,
			HearingImpaired: item.Attributes.HearingImpaired,
			AITranslated:    item.Attributes.AITranslated,
		})
	}
	return results, nil
}

// Download fetches a subtitle file by file_id. Returns the raw bytes and suggested filename.
func (c *Client) Download(ctx context.Context, fileID string) ([]byte, string, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, "", fmt.Errorf("opensubtitles auth: %w", err)
	}

	fid, err := strconv.Atoi(fileID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file_id: %w", err)
	}

	body, _ := json.Marshal(map[string]int{"file_id": fid})
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/download", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(b))
	}

	var dlResp struct {
		Link     string `json:"link"`
		FileName string `json:"file_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dlResp); err != nil {
		return nil, "", fmt.Errorf("decoding download response: %w", err)
	}

	// Fetch actual subtitle file
	fileReq, err := http.NewRequestWithContext(ctx, "GET", dlResp.Link, nil)
	if err != nil {
		return nil, "", err
	}
	fileResp, err := c.http.Do(fileReq)
	if err != nil {
		return nil, "", fmt.Errorf("fetching subtitle file: %w", err)
	}
	defer fileResp.Body.Close()

	data, err := io.ReadAll(fileResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading subtitle file: %w", err)
	}

	return data, dlResp.FileName, nil
}

// API response types

type searchResponse struct {
	Data []searchItem `json:"data"`
}

type searchItem struct {
	Attributes searchAttributes `json:"attributes"`
}

type searchAttributes struct {
	Release          string       `json:"release"`
	Language         string       `json:"language"`
	Format           string       `json:"format"`
	DownloadCount    int          `json:"download_count"`
	Ratings          float64      `json:"ratings"`
	HearingImpaired  bool         `json:"hearing_impaired"`
	ForeignPartsOnly bool         `json:"foreign_parts_only"`
	AITranslated     bool         `json:"ai_translated"`
	Files            []searchFile `json:"files"`
}

type searchFile struct {
	FileID   int    `json:"file_id"`
	FileName string `json:"file_name"`
}
