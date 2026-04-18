// Package fshare provides a Go client for the fshare.vn API.
//
// The package targets the mobile/desktop API endpoint (api.fshare.vn) and
// handles the hybrid token+cookie auth, the code:201 session-expiry pattern,
// and fshare's lack of 429/Retry-After headers.
//
// See README.md for the endpoint reference compiled from cross-referencing
// five open-source fshare clients.
package fshare

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.fshare.vn"
	defaultTimeout = 30 * time.Second

	// defaultUserAgent is the only UA observed to work against api.fshare.vn
	// (2026-04). "okhttp/3.6.0" and browser UAs now return code:37 Invalid UA.
	defaultUserAgent = "Fshare/1 CFNetwork/1209 Darwin/20.2.0"

	// DefaultAppKey is active as of 2026-04 (verified via live smoke test).
	// The newer iOS-sniffed key "dMnqMMZMUnN5YpvKENaEhdQQ5jxDqddt" was revoked.
	// Fallback keys if this one gets revoked:
	//   "GUxft6Beh3Bf8qKP7GC2IplYJZz1A53JQfRwne0R"  (giangvo/synology-fshare)
	//   "dMnqMMZMUnN5YpvKENaEhdQQ5jxDqddt"          (iOS, rejected 2026-04)
	DefaultAppKey = "L2S7R6ZMagggC5wWkQhX2+aDi467PPuftWUMRFSn"
)

// Options configures a Client. All fields are optional; sensible defaults apply.
type Options struct {
	// AppKey overrides DefaultAppKey. Set via VELOX_FSHARE_APP_KEY env if the
	// default gets revoked.
	AppKey string

	// BaseURL overrides the default https://api.fshare.vn (for tests).
	BaseURL string

	// UserAgent overrides the default okhttp/3.6.0. Fshare rejects the
	// default Go-http-client UA, so a non-default value is required.
	UserAgent string

	// HTTPTimeout applies to the underlying http.Client. Default 30s.
	HTTPTimeout time.Duration
}

// Client is a fshare.vn API client. Use NewClient to construct.
// Client maintains session state internally; a single Client belongs to one
// fshare account.
type Client struct {
	httpClient *http.Client
	baseURL    string
	appKey     string
	userAgent  string

	mu          sync.RWMutex
	session     *Session
	credentials *Credentials
}

// NewClient constructs a Client. All Options fields are optional.
func NewClient(opts Options) (*Client, error) {
	appKey := opts.AppKey
	if appKey == "" {
		appKey = DefaultAppKey
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	timeout := opts.HTTPTimeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("fshare: cookiejar: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: timeout,
		},
		baseURL:   baseURL,
		appKey:    appKey,
		userAgent: userAgent,
	}, nil
}

// SetCredentials stashes the email+password so withSessionRetry can auto-relogin
// when the server returns code:201 mid-call. Called by the resolver layer after
// decrypting credentials from the cloud_sessions row.
func (c *Client) SetCredentials(email, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.credentials = &Credentials{Email: email, Password: password}
}

// Session returns a copy of the active session, or nil if not logged in.
func (c *Client) Session() *Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.session == nil {
		return nil
	}
	s := *c.session
	return &s
}

// RestoreSession injects token + session_id from persistent storage so the
// client can skip the initial Login call. The caller is still responsible for
// calling SetCredentials so auto-relogin works if the restored session is stale.
func (c *Client) RestoreSession(token, sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = &Session{
		Token:     token,
		SessionID: sessionID,
	}
	// Cookie jar doesn't know about session_id unless we inject it via a real
	// response. Since authenticated endpoints require it, we send session_id
	// explicitly via Cookie header in setHeaders instead of relying on the jar
	// for restored sessions. See setHeaders.
}

// setHeaders applies the standard headers for an authenticated request.
// If a session is active, Cookie: session_id=... is added explicitly to work
// alongside the cookiejar (which only captures cookies from prior responses).
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.mu.RLock()
	session := c.session
	c.mu.RUnlock()
	if session != nil && session.SessionID != "" {
		// Explicit Cookie header so RestoreSession works even when jar is empty.
		req.Header.Set("Cookie", "session_id="+session.SessionID)
	}
}
