package fshare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// pageSize is the default items-per-page for list endpoints.
	// 60 is the observed cap across OSS clients.
	pageSize = 60
)

// ListFolder returns all items inside folderCode. Pass an empty string to list
// the account root. Pagination is handled internally — callers receive a flat slice.
//
// Session expiry (code:201) triggers one auto-relogin if credentials are stashed.
func (c *Client) ListFolder(ctx context.Context, folderCode string) ([]FolderItem, error) {
	var items []FolderItem
	err := c.withSessionRetry(ctx, func() error {
		items = nil // reset on retry
		for pageIndex := 0; ; pageIndex++ {
			page, err := c.listFolderPage(ctx, folderCode, pageIndex)
			if err != nil {
				return err
			}
			items = append(items, page...)
			if len(page) < pageSize {
				return nil
			}
		}
	})
	return items, err
}

// listFolderPage fetches one page of folder items.
// Root listing uses GET /api/fileops/list (no body, query params).
// Nested listing uses POST /api/fileops/getFolderList with token + folder URL.
func (c *Client) listFolderPage(ctx context.Context, folderCode string, pageIndex int) ([]FolderItem, error) {
	session := c.Session()
	if session == nil {
		return nil, ErrNotLoggedIn
	}

	var req *http.Request
	var err error
	if folderCode == "" {
		url := fmt.Sprintf("%s/api/fileops/list?pageIndex=%d&dirOnly=0&limit=%d", c.baseURL, pageIndex, pageSize)
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	} else {
		body, marshalErr := json.Marshal(FolderListRequest{
			Token:     session.Token,
			URL:       "https://www.fshare.vn/folder/" + folderCode,
			DirOnly:   0,
			PageIndex: pageIndex,
			Limit:     pageSize,
		})
		if marshalErr != nil {
			return nil, fmt.Errorf("fshare: marshal folder list: %w", marshalErr)
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/fileops/getFolderList", bytes.NewReader(body))
	}
	if err != nil {
		return nil, fmt.Errorf("fshare: build folder list: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.requestWithBackoff(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Some fshare error payloads come as objects {code, msg} instead of arrays.
	// Sniff the first non-whitespace byte to decide between shapes.
	body, err := readPeek(resp)
	if err != nil {
		return nil, err
	}
	if isJSONObject(body) {
		var envelope struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(body, &envelope); err == nil {
			if envelope.Code == 201 {
				return nil, ErrSessionExpired
			}
			return nil, fmt.Errorf("fshare: folder list code=%d msg=%s", envelope.Code, envelope.Msg)
		}
	}

	var items []FolderItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("fshare: decode folder list: %w", err)
	}
	return items, nil
}
