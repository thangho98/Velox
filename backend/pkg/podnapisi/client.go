package podnapisi

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
	baseURL   = "https://www.podnapisi.net"
	userAgent = "Velox v0.1.0"
)

var rateLimit = subprovider.NewThrottle(time.Second)

// Client is a Podnapisi subtitle API client. No authentication required.
type Client struct {
	http *http.Client
}

// SearchParams configures a subtitle search query.
type SearchParams struct {
	Keywords string // movie/episode title
	Year     int
	Season   int
	Episode  int
	Language string // Podnapisi language code
}

// New creates a new Podnapisi client.
func New() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// languageCode maps ISO 639-1 codes to Podnapisi language codes.
var languageCode = map[string]string{
	"en": "2", "vi": "52", "fr": "8", "de": "5",
	"es": "28", "pt": "23", "it": "9", "nl": "13",
	"pl": "19", "ru": "22", "ja": "11", "ko": "4",
	"zh": "17", "ar": "26", "tr": "30", "sv": "25",
	"da": "24", "fi": "31", "no": "3", "cs": "7",
	"hu": "15", "ro": "21", "hr": "38", "sr": "36",
	"bg": "33", "el": "16", "he": "20", "th": "53",
	"id": "54", "ms": "55",
}

// Search queries Podnapisi for matching subtitles.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	q := url.Values{}
	if params.Keywords != "" {
		q.Set("keywords", params.Keywords)
	}
	if params.Year > 0 {
		q.Set("year", strconv.Itoa(params.Year))
	}
	if params.Season > 0 {
		q.Set("seasons", strconv.Itoa(params.Season))
	}
	if params.Episode > 0 {
		q.Set("episodes", strconv.Itoa(params.Episode))
	}
	if params.Language != "" {
		if code, ok := languageCode[params.Language]; ok {
			q.Set("language", code)
		}
	}

	u := baseURL + "/subtitles/search/old?sXML=1&" + q.Encode()
	rateLimit.Wait()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/xml")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("podnapisi search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("podnapisi search failed (%d): %s", resp.StatusCode, string(b))
	}

	// Podnapisi XML response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return parseXMLResponse(body)
}

// SearchJSON uses the JSON search endpoint (alternative).
func (c *Client) SearchJSON(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	q := url.Values{}
	if params.Keywords != "" {
		q.Set("keywords", params.Keywords)
	}
	if params.Year > 0 {
		q.Set("year", strconv.Itoa(params.Year))
	}
	if params.Language != "" {
		if code, ok := languageCode[params.Language]; ok {
			q.Set("language", code)
		}
	}

	u := baseURL + "/subtitles/search/old?sJ=1&" + q.Encode()
	rateLimit.Wait()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("podnapisi search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("podnapisi search failed (%d): %s", resp.StatusCode, string(b))
	}

	var apiResp jsonResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding podnapisi response: %w", err)
	}

	results := make([]subprovider.Result, 0, len(apiResp.Data))
	for _, item := range apiResp.Data {
		results = append(results, subprovider.Result{
			Provider:   "podnapisi",
			ExternalID: item.PID,
			Title:      item.Release,
			Language:   reverseLangCode(item.LanguageID),
			Format:     "srt",
			Downloads:  item.Downloads,
			Rating:     item.Rating,
		})
	}
	return results, nil
}

// Download fetches and extracts a subtitle file by PID. Returns raw bytes and filename.
func (c *Client) Download(ctx context.Context, pid string) ([]byte, string, error) {
	u := baseURL + "/subtitles/" + pid + "/download"
	rateLimit.Wait()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("podnapisi download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("podnapisi download failed (%d): %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading download: %w", err)
	}

	// Podnapisi returns a .zip file — extract the first subtitle file
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		// Not a zip — return as-is (sometimes direct file)
		return body, "subtitle.srt", nil
	}

	for _, f := range zipReader.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".srt" || ext == ".vtt" || ext == ".ass" || ext == ".sub" {
			rc, err := f.Open()
			if err != nil {
				return nil, "", fmt.Errorf("opening zip entry: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, "", fmt.Errorf("reading zip entry: %w", err)
			}
			return data, f.Name, nil
		}
	}

	return nil, "", fmt.Errorf("no subtitle file found in archive")
}

// reverseLangCode maps Podnapisi language ID back to ISO 639-1.
func reverseLangCode(langID string) string {
	for iso, pid := range languageCode {
		if pid == langID {
			return iso
		}
	}
	return langID
}

// JSON response types

type jsonResponse struct {
	Data []jsonItem `json:"data"`
}

type jsonItem struct {
	PID        string  `json:"pid"`
	Release    string  `json:"release"`
	LanguageID string  `json:"languageId"`
	Downloads  int     `json:"downloads"`
	Rating     float64 `json:"rating"`
}
