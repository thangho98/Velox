package ophim

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
	baseURL      = "https://ophim1.com"
	imageBaseURL = "https://img.ophim.live/uploads/movies/"
)

// Client is a HTTP Client to communicate with OPhim public API
type Client struct {
	httpClient *http.Client
}

// New creates a new OPhim client
func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetImageURL returns a full image URL based on thumb/poster filenames
func (c *Client) GetImageURL(filename string) string {
	if filename == "" {
		return ""
	}
	return imageBaseURL + filename
}

// GetRecentMovies fetches recently updated movies with pagination
func (c *Client) GetRecentMovies(ctx context.Context, page int) (*PageResponse, error) {
	params := url.Values{
		"page": {strconv.Itoa(page)},
	}

	req, err := c.newRequest(ctx, "GET", "/v1/api/danh-sach/phim-moi-cap-nhat", params)
	if err != nil {
		return nil, err
	}

	var results PageResponse
	if err := c.do(req, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// GetMovieDetails gets full details and streaming links for a movie by slug
func (c *Client) GetMovieDetails(ctx context.Context, slug string) (*MovieDetailResponse, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/phim/%s", url.PathEscape(slug)), nil)
	if err != nil {
		return nil, err
	}

	var details MovieDetailResponse
	if err := c.do(req, &details); err != nil {
		return nil, err
	}

	if !details.Status {
		return nil, fmt.Errorf("ophim api error: %s", details.Message)
	}

	return &details, nil
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

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// do executes the request and decodes the response.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ophim api returned status %d - %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
