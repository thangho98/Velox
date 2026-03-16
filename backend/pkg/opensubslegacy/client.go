// Package opensubslegacy provides a client for the OpenSubtitles legacy REST API.
// This API does not require an API key or account — only a User-Agent string.
// It supports search by IMDB ID, hash, or text query.
// Downloads are gzip-compressed SRT files.
package opensubslegacy

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/thawng/velox/pkg/subprovider"
)

const (
	baseURL   = "https://rest.opensubtitles.org/search"
	userAgent = "VLSub 0.10.2"
)

// Client for the OpenSubtitles legacy REST API.
type Client struct {
	http *http.Client
}

// New creates a new legacy OpenSubtitles client.
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchParams holds search criteria.
type SearchParams struct {
	ImdbID        string // e.g. "tt0108778" (without "tt" prefix also works)
	SeasonNumber  int
	EpisodeNumber int
	Language      string // ISO 639-2/B code: "eng", "vie", "fre"
	Query         string // text query (fallback if no IMDB ID)
}

// Search finds subtitles via the legacy REST API.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	url := c.buildSearchURL(params)
	if url == "" {
		return nil, fmt.Errorf("opensubslegacy: need imdb_id or query")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opensubslegacy search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("opensubslegacy: status %d: %s", resp.StatusCode, string(body))
	}

	var items []apiResult
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("opensubslegacy: decode: %w", err)
	}

	results := make([]subprovider.Result, 0, len(items))
	for _, item := range items {
		lang := iso639bToISO639_1(item.ISO639)
		if lang == "" {
			lang = item.ISO639
		}

		downloads := 0
		if item.SubDownloadsCnt != "" {
			fmt.Sscanf(item.SubDownloadsCnt, "%d", &downloads)
		}

		ext := strings.TrimPrefix(filepath.Ext(item.SubFileName), ".")
		if ext == "" {
			ext = "srt"
		}

		results = append(results, subprovider.Result{
			Provider:        "opensubtitles",
			ExternalID:      item.IDSubtitleFile,
			Title:           item.SubFileName,
			Language:        lang,
			Format:          ext,
			Downloads:       downloads,
			HearingImpaired: item.SubHearingImpaired == "1",
		})
	}

	return results, nil
}

// Download fetches a subtitle file by its file ID (from ExternalID).
// The legacy API serves gzip-compressed files.
func (c *Client) Download(ctx context.Context, externalID string) ([]byte, string, error) {
	url := fmt.Sprintf("https://dl.opensubtitles.org/en/download/file/%s", externalID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("opensubslegacy download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("opensubslegacy download: status %d", resp.StatusCode)
	}

	// Response is gzip-compressed
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" || resp.Header.Get("Content-Type") == "application/x-gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("opensubslegacy: gzip decode: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("opensubslegacy: read body: %w", err)
	}

	// Try gzip decode if data starts with gzip magic bytes
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		gz, err := gzip.NewReader(strings.NewReader(string(data)))
		if err == nil {
			decoded, err := io.ReadAll(gz)
			gz.Close()
			if err == nil {
				data = decoded
			}
		}
	}

	filename := fmt.Sprintf("opensubtitles_%s.srt", externalID)
	return data, filename, nil
}

func (c *Client) buildSearchURL(p SearchParams) string {
	var parts []string

	if p.EpisodeNumber > 0 {
		parts = append(parts, fmt.Sprintf("episode-%d", p.EpisodeNumber))
	}
	if p.ImdbID != "" {
		imdb := strings.TrimPrefix(p.ImdbID, "tt")
		parts = append(parts, fmt.Sprintf("imdbid-%s", imdb))
	}
	if p.Query != "" && p.ImdbID == "" {
		parts = append(parts, fmt.Sprintf("query-%s", strings.ReplaceAll(p.Query, " ", "+")))
	}
	if p.SeasonNumber > 0 {
		parts = append(parts, fmt.Sprintf("season-%d", p.SeasonNumber))
	}
	if p.Language != "" {
		lang := iso639_1ToISO639b(p.Language)
		parts = append(parts, fmt.Sprintf("sublanguageid-%s", lang))
	}

	if len(parts) == 0 {
		return ""
	}

	return baseURL + "/" + strings.Join(parts, "/")
}

// apiResult represents a single result from the legacy API.
type apiResult struct {
	IDSubtitleFile     string `json:"IDSubtitleFile"`
	SubFileName        string `json:"SubFileName"`
	SubDownloadsCnt    string `json:"SubDownloadsCnt"`
	SubDownloadLink    string `json:"SubDownloadLink"`
	ISO639             string `json:"ISO639"`
	SubHearingImpaired string `json:"SubHearingImpaired"`
	SubFormat          string `json:"SubFormat"`
	LanguageName       string `json:"LanguageName"`
}

// iso639_1ToISO639b converts 2-letter to 3-letter language codes (common ones).
func iso639_1ToISO639b(code string) string {
	m := map[string]string{
		"en": "eng", "vi": "vie", "fr": "fre", "de": "ger", "es": "spa",
		"pt": "por", "it": "ita", "nl": "dut", "sv": "swe", "no": "nor",
		"da": "dan", "fi": "fin", "ja": "jpn", "ko": "kor", "zh": "chi",
		"ar": "ara", "th": "tha", "pl": "pol", "ru": "rus", "tr": "tur",
		"cs": "cze", "hu": "hun", "ro": "rum", "el": "gre", "he": "heb",
		"id": "ind", "ms": "may", "hi": "hin", "bg": "bul", "hr": "hrv",
		"sk": "slo", "sl": "slv", "uk": "ukr", "sr": "scc",
	}
	if v, ok := m[code]; ok {
		return v
	}
	return code
}

// iso639bToISO639_1 converts 3-letter to 2-letter language codes.
func iso639bToISO639_1(code string) string {
	m := map[string]string{
		"eng": "en", "vie": "vi", "fre": "fr", "ger": "de", "spa": "es",
		"por": "pt", "ita": "it", "dut": "nl", "swe": "sv", "nor": "no",
		"dan": "da", "fin": "fi", "jpn": "ja", "kor": "ko", "chi": "zh",
		"ara": "ar", "tha": "th", "pol": "pl", "rus": "ru", "tur": "tr",
		"cze": "cs", "hun": "hu", "rum": "ro", "gre": "el", "heb": "he",
		"ind": "id", "may": "ms", "hin": "hi", "bul": "bg", "hrv": "hr",
		"slo": "sk", "slv": "sl", "ukr": "uk", "scc": "sr",
	}
	if v, ok := m[code]; ok {
		return v
	}
	return code
}
