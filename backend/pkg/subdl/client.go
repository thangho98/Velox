package subdl

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/pkg/subprovider"
)

const (
	searchURL   = "https://api.subdl.com/api/v1/subtitles"
	downloadURL = "https://dl.subdl.com/subtitle"
	userAgent   = "Velox v0.1.0"
)

var rateLimit = subprovider.NewThrottle(time.Second)

// Client is a Subdl.com subtitle API client.
type Client struct {
	apiKey string
	http   *http.Client
}

// SearchParams configures a subtitle search query.
type SearchParams struct {
	FilmName      string
	FileName      string
	ImdbID        string
	TmdbID        int
	Year          int
	Language      string // comma-separated: "EN,FR"
	SeasonNumber  int
	EpisodeNumber int
	Type          string // "movie" or "tv"
}

// New creates a new Subdl client.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search queries Subdl for matching subtitles.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	q := url.Values{}
	q.Set("api_key", c.apiKey)
	q.Set("subs_per_page", "30")
	q.Set("hi", "1")

	if params.ImdbID != "" {
		q.Set("imdb_id", params.ImdbID)
	}
	if params.TmdbID > 0 {
		q.Set("tmdb_id", strconv.Itoa(params.TmdbID))
	}
	if params.FilmName != "" {
		q.Set("film_name", params.FilmName)
	}
	if params.FileName != "" {
		q.Set("file_name", params.FileName)
	}
	if params.Language != "" {
		q.Set("languages", strings.ToUpper(params.Language))
	}
	if params.Year > 0 {
		q.Set("year", strconv.Itoa(params.Year))
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}
	if params.SeasonNumber > 0 {
		q.Set("season_number", strconv.Itoa(params.SeasonNumber))
	}
	if params.EpisodeNumber > 0 {
		q.Set("episode_number", strconv.Itoa(params.EpisodeNumber))
	}

	fullURL := searchURL + "?" + q.Encode()
	rateLimit.Wait()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("subdl search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("subdl search failed (%d): %s", resp.StatusCode, string(b))
	}

	var apiResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding subdl response: %w", err)
	}

	if !apiResp.Status {
		errMsg := apiResp.Error
		if errMsg == "" {
			errMsg = apiResp.Message
		}
		return nil, fmt.Errorf("subdl search error: %s", errMsg)
	}

	results := make([]subprovider.Result, 0, len(apiResp.Subtitles))
	for _, sub := range apiResp.Subtitles {
		format := strings.ToLower(sub.Format)
		if format == "" {
			format = "srt"
		}

		results = append(results, subprovider.Result{
			Provider:        "subdl",
			ExternalID:      sub.URL, // download path e.g. "3197651-3213944"
			Title:           sub.ReleaseName,
			Language:        strings.ToLower(sub.Language),
			Format:          format,
			Downloads:       0, // not provided by API
			HearingImpaired: sub.HI,
		})
	}
	return results, nil
}

// Download fetches a subtitle zip and extracts the first subtitle file.
func (c *Client) Download(ctx context.Context, subtitlePath string) ([]byte, string, error) {
	cleaned := strings.TrimPrefix(subtitlePath, "/subtitle/")
	cleaned = strings.TrimSuffix(cleaned, ".zip")
	u := downloadURL + "/" + cleaned + ".zip"

	rateLimit.Wait()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("subdl download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("subdl download failed (%d): %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading download: %w", err)
	}

	data, filename, err := extractDownloadedSubtitlePayload(body)
	if err != nil {
		return nil, "", err
	}
	return data, filename, nil
}

func extractDownloadedSubtitlePayload(body []byte) ([]byte, string, error) {
	if looksLikeHTMLDocument(body) {
		return nil, "", fmt.Errorf("subdl returned HTML instead of a subtitle file")
	}

	// Subdl returns a .zip file — extract the first subtitle file
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		if looksLikeSubtitleText(body) {
			return body, "subtitle.srt", nil
		}
		return nil, "", fmt.Errorf("subdl returned an unexpected payload")
	}

	for _, f := range zipReader.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".srt" || ext == ".vtt" || ext == ".ass" || ext == ".sub" || ext == ".ssa" {
			rc, err := f.Open()
			if err != nil {
				return nil, "", fmt.Errorf("opening zip entry: %w", err)
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, "", fmt.Errorf("reading zip entry: %w", err)
			}
			return data, f.Name, nil
		}
	}

	return nil, "", fmt.Errorf("no subtitle file found in archive")
}

func looksLikeHTMLDocument(body []byte) bool {
	snippet := strings.ToLower(strings.TrimSpace(string(body)))
	if len(snippet) > 2048 {
		snippet = snippet[:2048]
	}
	return strings.HasPrefix(snippet, "<!doctype html") ||
		strings.HasPrefix(snippet, "<html") ||
		(strings.Contains(snippet, "<head") && strings.Contains(snippet, "<body"))
}

func looksLikeSubtitleText(body []byte) bool {
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return false
	}
	lower := strings.ToLower(snippet)
	if strings.HasPrefix(lower, "webvtt") {
		return true
	}
	return strings.Contains(snippet, "-->") &&
		(strings.Contains(snippet, "00:") || strings.Contains(snippet, "0:"))
}

// API response types

type searchResponse struct {
	Status    bool            `json:"status"`
	Error     string          `json:"error,omitempty"`
	Message   string          `json:"message,omitempty"`
	Results   []searchResult  `json:"results,omitempty"`
	Subtitles []subtitleEntry `json:"subtitles,omitempty"`
}

type searchResult struct {
	ImdbID string `json:"imdb_id"`
	TmdbID int    `json:"tmdb_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	SdID   int    `json:"sd_id"`
	Year   int    `json:"year"`
}

type subtitleEntry struct {
	ReleaseName string `json:"release_name"`
	Language    string `json:"language"`
	URL         string `json:"url"`
	Format      string `json:"format,omitempty"`
	HI          bool   `json:"hi"`
	Author      string `json:"author,omitempty"`
}
