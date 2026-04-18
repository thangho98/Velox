package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL = "https://graphql.anilist.co"

	// AniList documents 90 requests/minute by default, while temporary degraded
	// periods may clamp requests further. The client therefore also retries 429s
	// using Retry-After / X-RateLimit-Reset headers.
	rateLimitPerMinute = 90
	defaultPerPage     = 10
)

const mediaFields = `
	id
	idMal
	type
	format
	status
	episodes
	season
	seasonYear
	description(asHtml: false)
	siteUrl
	title {
		romaji
		english
		native
		userPreferred
	}
	coverImage {
		extraLarge
		large
		medium
		color
	}
	bannerImage
	startDate {
		year
		month
		day
	}
	endDate {
		year
		month
		day
	}
	studios(isMain: true) {
		nodes {
			id
			name
			isAnimationStudio
		}
	}
	streamingEpisodes {
		title
		thumbnail
		url
		site
	}
	nextAiringEpisode {
		airingAt
		timeUntilAiring
		episode
	}
`

const searchAnimeQuery = `
query SearchAnime($search: String!, $page: Int!, $perPage: Int!) {
	Page(page: $page, perPage: $perPage) {
		pageInfo {
			total
			perPage
			currentPage
			lastPage
			hasNextPage
		}
		media(search: $search, type: ANIME) {
` + mediaFields + `
		}
	}
}`

const getAnimeByIDQuery = `
query AnimeByID($id: Int!) {
	Media(id: $id, type: ANIME) {
` + mediaFields + `
	}
}`

// Client wraps the AniList GraphQL API.
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
	limiter     chan struct{}
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// GraphQLError is the GraphQL error format AniList returns.
type GraphQLError struct {
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

type searchAnimeEnvelope struct {
	Page MediaPage `json:"Page"`
}

type animeByIDEnvelope struct {
	Media *Media `json:"Media"`
}

// New creates a new AniList client.
// accessToken may be empty because public AniList data is available without auth.
func New(accessToken string) *Client {
	c := &Client{
		accessToken: accessToken,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: make(chan struct{}, rateLimitPerMinute),
	}
	c.startLimiter()
	return c
}

// NewWithHTTPClient creates a client with a custom HTTP client.
func NewWithHTTPClient(accessToken string, httpClient *http.Client) *Client {
	c := &Client{
		accessToken: accessToken,
		baseURL:     baseURL,
		httpClient:  httpClient,
		limiter:     make(chan struct{}, rateLimitPerMinute),
	}
	c.startLimiter()
	return c
}

func (c *Client) startLimiter() {
	for range rateLimitPerMinute {
		c.limiter <- struct{}{}
	}

	go func() {
		ticker := time.NewTicker(time.Minute / time.Duration(rateLimitPerMinute))
		defer ticker.Stop()
		for range ticker.C {
			select {
			case c.limiter <- struct{}{}:
			default:
			}
		}
	}()
}

func (c *Client) wait(ctx context.Context) error {
	select {
	case <-c.limiter:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SearchAnime searches AniList anime by title.
func (c *Client) SearchAnime(ctx context.Context, search string, page, perPage int) (*MediaPage, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return nil, fmt.Errorf("anilist search: search query is required")
	}
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = defaultPerPage
	}

	var envelope searchAnimeEnvelope
	if err := c.doGraphQL(ctx, searchAnimeQuery, map[string]any{
		"search":  search,
		"page":    page,
		"perPage": perPage,
	}, &envelope); err != nil {
		return nil, err
	}

	return &envelope.Page, nil
}

// GetAnimeByID fetches a single anime by AniList ID.
func (c *Client) GetAnimeByID(ctx context.Context, id int) (*Media, error) {
	if id <= 0 {
		return nil, fmt.Errorf("anilist get anime: id must be positive")
	}

	var envelope animeByIDEnvelope
	if err := c.doGraphQL(ctx, getAnimeByIDQuery, map[string]any{"id": id}, &envelope); err != nil {
		return nil, err
	}
	if envelope.Media == nil {
		return nil, fmt.Errorf("anilist get anime: anime %d not found", id)
	}
	return envelope.Media, nil
}

func (c *Client) doGraphQL(ctx context.Context, query string, variables map[string]any, out any) error {
	payload, err := json.Marshal(graphQLRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return fmt.Errorf("anilist marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := c.wait(ctx); err != nil {
			return fmt.Errorf("anilist rate limit wait: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("anilist create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("anilist request failed: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("anilist read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			delay := retryDelay(resp.Header)
			if err := sleepContext(ctx, delay); err != nil {
				return fmt.Errorf("anilist retry wait: %w", err)
			}
			lastErr = fmt.Errorf("anilist API error: %d - %s", resp.StatusCode, strings.TrimSpace(string(body)))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("anilist API error: %d - %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var envelope graphQLResponse
		if err := json.Unmarshal(body, &envelope); err != nil {
			return fmt.Errorf("anilist decode response: %w", err)
		}
		if len(envelope.Errors) > 0 {
			return fmt.Errorf("anilist graphql error: %s", joinGraphQLErrors(envelope.Errors))
		}
		if out == nil || len(envelope.Data) == 0 {
			return nil
		}
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("anilist decode data: %w", err)
		}
		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("anilist request failed after retry")
}

func retryDelay(headers http.Header) time.Duration {
	if retryAfter := strings.TrimSpace(headers.Get("Retry-After")); retryAfter != "" {
		if secs, err := strconv.Atoi(retryAfter); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}

	if resetAt := strings.TrimSpace(headers.Get("X-RateLimit-Reset")); resetAt != "" {
		if unixTS, err := strconv.ParseInt(resetAt, 10, 64); err == nil {
			delay := time.Until(time.Unix(unixTS, 0))
			if delay > 0 {
				return delay
			}
		}
	}

	return time.Minute
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func joinGraphQLErrors(errors []GraphQLError) string {
	parts := make([]string, 0, len(errors))
	for _, gqlErr := range errors {
		msg := strings.TrimSpace(gqlErr.Message)
		if msg == "" {
			continue
		}
		if gqlErr.Status > 0 {
			msg = fmt.Sprintf("%s (status %d)", msg, gqlErr.Status)
		}
		parts = append(parts, msg)
	}
	if len(parts) == 0 {
		return "unknown graphql error"
	}
	return strings.Join(parts, "; ")
}
