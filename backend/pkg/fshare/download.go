package fshare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetDirectLink returns a short-lived download URL for the given file linkcode.
// URLs are minutes-scale TTL and possibly single-use — callers MUST NOT cache
// the returned string, only hand it to the HTTP client for immediate download.
//
// Session expiry (code:201) triggers one auto-relogin if credentials are stashed.
func (c *Client) GetDirectLink(ctx context.Context, fileCode string) (string, error) {
	var location string
	err := c.withSessionRetry(ctx, func() error {
		session := c.Session()
		if session == nil {
			return ErrNotLoggedIn
		}

		body, err := json.Marshal(DownloadRequest{
			Token:    session.Token,
			URL:      "https://www.fshare.vn/file/" + fileCode,
			Password: "",
			Zipflag:  0,
		})
		if err != nil {
			return fmt.Errorf("fshare: marshal download: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/session/download", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("fshare: build download: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.requestWithBackoff(ctx, req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// 403 is the canonical "wrong file password" response and is NOT retried
		// by requestWithBackoff. Surface it as a typed error.
		if resp.StatusCode == 403 {
			return ErrFilePassword
		}

		var dr DownloadResponse
		if err := decodeJSONOrError(resp, &dr); err != nil {
			return err
		}

		if dr.Code == 201 {
			return ErrSessionExpired
		}
		if dr.Location == "" {
			if dr.Code != 0 && dr.Code != 200 {
				return fmt.Errorf("%w: code=%d msg=%s", ErrLinkDead, dr.Code, dr.Msg)
			}
			return fmt.Errorf("%w: empty location", ErrLinkDead)
		}

		location = dr.Location
		return nil
	})
	return location, err
}
