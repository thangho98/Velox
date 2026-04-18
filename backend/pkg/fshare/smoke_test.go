//go:build smoke

// Smoke test against the real fshare.vn API.
// Run with: go test -tags smoke ./pkg/fshare/ -v -run TestSmoke
//
// Requires env vars:
//
//	VELOX_FSHARE_SMOKE_EMAIL     (required)
//	VELOX_FSHARE_SMOKE_PASSWORD  (required)
//	VELOX_FSHARE_SMOKE_FOLDER    (optional) — linkcode of a nested folder to list
//	VELOX_FSHARE_SMOKE_FILE      (optional) — linkcode of a file to resolve download URL
//	VELOX_FSHARE_APP_KEY         (optional) — overrides DefaultAppKey
//
// This test hits the live fshare API. Use a VIP account. The test does NOT
// download the file, only asserts that the returned URL is well-formed.

package fshare

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSmoke_LoginListDownload(t *testing.T) {
	email := os.Getenv("VELOX_FSHARE_SMOKE_EMAIL")
	password := os.Getenv("VELOX_FSHARE_SMOKE_PASSWORD")
	if email == "" || password == "" {
		t.Skip("set VELOX_FSHARE_SMOKE_EMAIL + VELOX_FSHARE_SMOKE_PASSWORD to run")
	}

	c, err := NewClient(Options{
		AppKey:      os.Getenv("VELOX_FSHARE_APP_KEY"),  // "" → DefaultAppKey
		UserAgent:   os.Getenv("VELOX_FSHARE_SMOKE_UA"), // "" → default
		HTTPTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Log("→ Login")
	sess, err := c.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	t.Logf("  token=%s… session_id=%s…", truncate(sess.Token, 8), truncate(sess.SessionID, 8))

	t.Log("→ GetUserInfo (VIP check + session health)")
	info, err := c.GetUserInfo(ctx)
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	t.Logf("  email=%s account_type=%s level=%s webspace=%s", info.Email, info.AccountType, info.Level, info.Webspace)
	if info.AccountType != "Vip" {
		t.Fatalf("account_type = %q; VIP required for /session/download", info.AccountType)
	}

	t.Log("→ ListFolder(root)")
	rootItems, err := c.ListFolder(ctx, "")
	if err != nil {
		t.Fatalf("ListFolder root: %v", err)
	}
	t.Logf("  %d items at root", len(rootItems))
	if len(rootItems) == 0 {
		t.Log("  (empty root — consider placing test files)")
	}
	for i, item := range rootItems {
		if i >= 3 {
			t.Logf("  … +%d more", len(rootItems)-3)
			break
		}
		t.Logf("  [%d] %s  size=%d  folder=%v  linkcode=%s", i, item.Name, item.SizeBytes(), item.IsFolder(), item.Linkcode)
	}

	if folderCode := os.Getenv("VELOX_FSHARE_SMOKE_FOLDER"); folderCode != "" {
		t.Logf("→ ListFolder(%s)", folderCode)
		items, err := c.ListFolder(ctx, folderCode)
		if err != nil {
			t.Fatalf("ListFolder nested: %v", err)
		}
		t.Logf("  %d items", len(items))
	}

	if fileCode := os.Getenv("VELOX_FSHARE_SMOKE_FILE"); fileCode != "" {
		t.Logf("→ GetDirectLink(%s)", fileCode)
		url, err := c.GetDirectLink(ctx, fileCode)
		if err != nil {
			t.Fatalf("GetDirectLink: %v", err)
		}
		if !strings.HasPrefix(url, "http") {
			t.Errorf("url = %q; want http(s) prefix", url)
		}
		t.Logf("  url=%s…", truncate(url, 60))

		// HEAD the URL — don't download body.
		req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Logf("  HEAD failed (ok if server requires GET): %v", err)
		} else {
			resp.Body.Close()
			t.Logf("  HEAD status=%d content-length=%d", resp.StatusCode, resp.ContentLength)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
