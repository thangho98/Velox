package fshare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Login authenticates against /api/user/login and populates the client's session.
// On success it also stashes credentials so withSessionRetry can auto-relogin.
//
// The login response returns only {code, msg, token, session_id}. Email and
// AccountType in the Session stay empty until a separate GetUserInfo call.
func (c *Client) Login(ctx context.Context, email, password string) (*Session, error) {
	body, err := json.Marshal(LoginRequest{
		UserEmail: email,
		Password:  password,
		AppKey:    c.appKey,
	})
	if err != nil {
		return nil, fmt.Errorf("fshare: marshal login: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/user/login", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("fshare: build login: %w", err)
	}
	c.setHeaders(req)

	// Login itself doesn't retry — credential errors shouldn't be re-hammered.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fshare: login request: %w", err)
	}
	defer resp.Body.Close()

	var lr LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("fshare: decode login: %w", err)
	}

	if lr.Code != 200 {
		switch lr.Code {
		case 201:
			return nil, ErrSessionExpired
		case 37:
			return nil, fmt.Errorf("%w: code=37 msg=%s", ErrAppKeyInvalid, lr.Msg)
		default:
			return nil, fmt.Errorf("%w: code=%d msg=%s", ErrInvalidCredentials, lr.Code, lr.Msg)
		}
	}

	session := &Session{
		Token:       lr.Token,
		SessionID:   lr.SessionID,
		Email:       email,
		LastLoginAt: time.Now(),
	}

	c.mu.Lock()
	c.session = session
	c.credentials = &Credentials{Email: email, Password: password}
	c.mu.Unlock()

	return session, nil
}

// GetUserInfo fetches account details from /api/user/get. Returns the account
// type (VIP status) and other profile fields. Also works as a session health
// check — returns ErrSessionExpired when the server signals code:201.
//
// Side effect: updates the stored Session's Email and AccountType.
func (c *Client) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	var info *UserInfo
	err := c.withSessionRetry(ctx, func() error {
		if c.Session() == nil {
			return ErrNotLoggedIn
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/user/get", nil)
		if err != nil {
			return fmt.Errorf("fshare: build user/get: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("fshare: user/get request: %w", err)
		}
		defer resp.Body.Close()

		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("fshare: read user/get: %w", err)
		}

		// Error envelopes come as {code, msg}; healthy responses are the full
		// UserInfo object. Sniff which shape we got.
		if isJSONObjectWithCode(raw) {
			var hr HealthResponse
			if err := json.Unmarshal(raw, &hr); err == nil {
				if hr.Code == 201 {
					return ErrSessionExpired
				}
				if hr.Code != 0 && hr.Code != 200 {
					return fmt.Errorf("fshare: user/get code=%d msg=%s", hr.Code, hr.Msg)
				}
			}
		}

		var ui UserInfo
		if err := json.Unmarshal(raw, &ui); err != nil {
			return fmt.Errorf("fshare: decode user/get: %w", err)
		}
		info = &ui

		// Update stored session with the authoritative account info.
		c.mu.Lock()
		if c.session != nil {
			c.session.Email = ui.Email
			c.session.AccountType = ui.AccountType
		}
		c.mu.Unlock()
		return nil
	})
	return info, err
}

// CheckSession validates the current session by calling GetUserInfo.
// Returns ErrNotLoggedIn if no session, ErrSessionExpired on code:201,
// or nil on success. A successful call also refreshes stored Email/AccountType.
func (c *Client) CheckSession(ctx context.Context) error {
	if c.Session() == nil {
		return ErrNotLoggedIn
	}
	_, err := c.GetUserInfo(ctx)
	return err
}

// withSessionRetry runs fn. If fn returns ErrSessionExpired and stored
// credentials are available, the client auto-relogins and retries fn once.
// Otherwise the original error is returned.
func (c *Client) withSessionRetry(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil || !errors.Is(err, ErrSessionExpired) {
		return err
	}

	c.mu.RLock()
	creds := c.credentials
	c.mu.RUnlock()
	if creds == nil {
		return err
	}

	if _, reloginErr := c.Login(ctx, creds.Email, creds.Password); reloginErr != nil {
		return reloginErr
	}
	return fn()
}

// decodeJSONOrError reads and JSON-decodes a response body into dst.
// Exposes the raw bytes in the returned error for easier debugging.
func decodeJSONOrError(resp *http.Response, dst any) error {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fshare: read body: %w", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		snippet := string(b)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return fmt.Errorf("fshare: decode json: %w (body=%s)", err, snippet)
	}
	return nil
}

// readPeek slurps the response body so the caller can inspect the first byte
// to decide between a JSON array (success) and a JSON object (error envelope).
func readPeek(resp *http.Response) ([]byte, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fshare: read body: %w", err)
	}
	return b, nil
}

// isJSONObject reports whether body starts with '{' (after leading whitespace).
// Used to distinguish fshare error envelopes from the expected array responses.
func isJSONObject(body []byte) bool {
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return true
		default:
			return false
		}
	}
	return false
}

// isJSONObjectWithCode reports whether body is a JSON object that includes a
// top-level "code" integer field (fshare error envelope shape) vs a healthy
// UserInfo object (which doesn't contain "code").
func isJSONObjectWithCode(body []byte) bool {
	if !isJSONObject(body) {
		return false
	}
	var probe struct {
		Code *int `json:"code"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	return probe.Code != nil
}
