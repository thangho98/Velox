package fshare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// shortenRetryDelay drops baseRetryDelay to 1ms for the duration of a test.
func shortenRetryDelay(t *testing.T) {
	t.Helper()
	prev := baseRetryDelay
	baseRetryDelay = time.Millisecond
	t.Cleanup(func() { baseRetryDelay = prev })
}

// newTestClient builds a Client pointed at srv with a harmless AppKey.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(Options{
		AppKey:      "test-app-key",
		BaseURL:     srv.URL,
		UserAgent:   "test/1.0",
		HTTPTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestNewClient_DefaultsWhenEmpty(t *testing.T) {
	c, err := NewClient(Options{})
	if err != nil {
		t.Fatalf("NewClient empty: %v", err)
	}
	if c.appKey != DefaultAppKey {
		t.Errorf("appKey = %q; want DefaultAppKey", c.appKey)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q; want %q", c.baseURL, defaultBaseURL)
	}
	if c.userAgent != defaultUserAgent {
		t.Errorf("userAgent = %q; want %q", c.userAgent, defaultUserAgent)
	}
}

func TestLogin_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/login" {
			t.Errorf("path = %q; want /api/user/login", r.URL.Path)
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.AppKey != "test-app-key" {
			t.Errorf("app_key = %q; want test-app-key", req.AppKey)
		}
		writeJSON(w, LoginResponse{
			Code:      200,
			Token:     "tok-xyz",
			SessionID: "sess-abc",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	sess, err := c.Login(context.Background(), "u@e.com", "pw")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if sess.Token != "tok-xyz" || sess.SessionID != "sess-abc" {
		t.Errorf("session = %+v; want token+sid set", sess)
	}
	if sess.Email != "u@e.com" {
		t.Errorf("Email = %q; want from login arg", sess.Email)
	}

	// Credentials should be stashed for auto-relogin.
	c.mu.RLock()
	creds := c.credentials
	c.mu.RUnlock()
	if creds == nil || creds.Password != "pw" {
		t.Errorf("credentials not stashed: %+v", creds)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, LoginResponse{Code: 400, Msg: "Wrong password"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Login(context.Background(), "u@e.com", "bad")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v; want ErrInvalidCredentials", err)
	}
}

func TestLogin_SessionExpiredDuringLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, LoginResponse{Code: 201, Msg: "session expired"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Login(context.Background(), "u@e.com", "pw")
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("err = %v; want ErrSessionExpired", err)
	}
}

func TestCheckSession_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Healthy /api/user/get returns the full UserInfo object (no "code" field).
		writeJSON(w, UserInfo{
			ID:          "123",
			Email:       "u@e.com",
			AccountType: "Vip",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	if err := c.CheckSession(context.Background()); err != nil {
		t.Errorf("CheckSession: %v", err)
	}
	// Session should be updated with fetched account info.
	sess := c.Session()
	if sess.Email != "u@e.com" || sess.AccountType != "Vip" {
		t.Errorf("session not populated: %+v", sess)
	}
}

func TestCheckSession_Expired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, HealthResponse{Code: 201, Msg: "expired"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	err := c.CheckSession(context.Background())
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("err = %v; want ErrSessionExpired", err)
	}
}

func TestCheckSession_NoSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit without session")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.CheckSession(context.Background())
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("err = %v; want ErrNotLoggedIn", err)
	}
}

func TestListFolder_Root(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/fileops/list") {
			t.Errorf("unexpected %s %s", r.Method, r.URL)
		}
		writeJSON(w, []FolderItem{
			{Linkcode: "a", Name: "foo.mkv", Size: "100", Mimetype: "video/x-matroska"},
			{Linkcode: "b", Name: "bar", Size: "0"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	items, err := c.ListFolder(context.Background(), "")
	if err != nil {
		t.Fatalf("ListFolder: %v", err)
	}
	if len(items) != 2 || items[0].Linkcode != "a" {
		t.Errorf("items = %+v", items)
	}
	if !items[0].IsFolder() == false { // items[0] has mimetype → file
		t.Errorf("items[0] IsFolder = true; want file")
	}
	if !items[1].IsFolder() {
		t.Errorf("items[1] IsFolder = false; want folder (no mimetype)")
	}
	if items[0].SizeBytes() != 100 {
		t.Errorf("SizeBytes = %d; want 100", items[0].SizeBytes())
	}
}

func TestListFolder_Nested_Pagination(t *testing.T) {
	var pagesServed int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/fileops/getFolderList" {
			t.Errorf("unexpected %s %s", r.Method, r.URL)
		}
		var req FolderListRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !strings.Contains(req.URL, "/folder/xyz") {
			t.Errorf("URL = %q; want /folder/xyz", req.URL)
		}
		page := int32(atomic.AddInt32(&pagesServed, 1) - 1)
		items := make([]FolderItem, 0, pageSize)
		if page < 2 {
			// Full pages 0 and 1 → triggers another fetch.
			for i := 0; i < pageSize; i++ {
				items = append(items, FolderItem{
					Linkcode: fmt.Sprintf("p%d-i%d", page, i),
					Name:     fmt.Sprintf("file-p%d-%d.mkv", page, i),
				})
			}
		} else {
			// Last page (smaller than pageSize) → caller stops.
			items = []FolderItem{{Linkcode: "last", Name: "last.mkv"}}
		}
		writeJSON(w, items)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	items, err := c.ListFolder(context.Background(), "xyz")
	if err != nil {
		t.Fatalf("ListFolder: %v", err)
	}
	wantCount := pageSize*2 + 1
	if len(items) != wantCount {
		t.Errorf("len = %d; want %d", len(items), wantCount)
	}
	if atomic.LoadInt32(&pagesServed) != 3 {
		t.Errorf("pagesServed = %d; want 3", pagesServed)
	}
}

func TestListFolder_SessionExpired_Relogin(t *testing.T) {
	var loginHits, listHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/login":
			atomic.AddInt32(&loginHits, 1)
			writeJSON(w, LoginResponse{Code: 200, Token: "tok2", SessionID: "sid2"})
		case "/api/fileops/getFolderList":
			hit := atomic.AddInt32(&listHits, 1)
			if hit == 1 {
				writeJSON(w, map[string]any{"code": 201, "msg": "expired"})
				return
			}
			writeJSON(w, []FolderItem{{Linkcode: "x", Name: "x.mkv"}})
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("old-tok", "old-sid")
	c.SetCredentials("u@e.com", "pw")
	items, err := c.ListFolder(context.Background(), "xyz")
	if err != nil {
		t.Fatalf("ListFolder: %v", err)
	}
	if len(items) != 1 || items[0].Linkcode != "x" {
		t.Errorf("items = %+v", items)
	}
	if loginHits != 1 {
		t.Errorf("loginHits = %d; want 1", loginHits)
	}
	if c.Session().Token != "tok2" {
		t.Errorf("session not refreshed: %+v", c.Session())
	}
}

func TestGetDirectLink_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/session/download" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var req DownloadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !strings.Contains(req.URL, "/file/abc") {
			t.Errorf("URL = %q; want /file/abc", req.URL)
		}
		writeJSON(w, DownloadResponse{Location: "https://download.fsxx.fshare.vn/real"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	url, err := c.GetDirectLink(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetDirectLink: %v", err)
	}
	if url != "https://download.fsxx.fshare.vn/real" {
		t.Errorf("url = %q", url)
	}
}

func TestGetDirectLink_FilePassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"msg":"password required"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	_, err := c.GetDirectLink(context.Background(), "locked")
	if !errors.Is(err, ErrFilePassword) {
		t.Errorf("err = %v; want ErrFilePassword", err)
	}
}

func TestGetDirectLink_LinkDead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, DownloadResponse{Code: 404, Msg: "link dead"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	_, err := c.GetDirectLink(context.Background(), "dead")
	if !errors.Is(err, ErrLinkDead) {
		t.Errorf("err = %v; want ErrLinkDead", err)
	}
}

func TestRetry_Backoff_RecoversAfter400s(t *testing.T) {
	shortenRetryDelay(t)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit := atomic.AddInt32(&hits, 1)
		if hit < 3 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		writeJSON(w, DownloadResponse{Location: "https://cdn.example/file"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	url, err := c.GetDirectLink(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetDirectLink: %v", err)
	}
	if url == "" {
		t.Errorf("url empty after retries")
	}
	if atomic.LoadInt32(&hits) != 3 {
		t.Errorf("hits = %d; want 3", hits)
	}
}

func TestRetry_Exhaustion(t *testing.T) {
	shortenRetryDelay(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	_, err := c.GetDirectLink(context.Background(), "abc")
	if !errors.Is(err, ErrRateLimit) {
		t.Errorf("err = %v; want ErrRateLimit", err)
	}
}

func TestCookieJar_SessionIDHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		if !strings.Contains(cookie, "session_id=sid-hdr") {
			t.Errorf("Cookie header missing session_id: %q", cookie)
		}
		writeJSON(w, HealthResponse{Code: 200})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid-hdr")
	if err := c.CheckSession(context.Background()); err != nil {
		t.Errorf("CheckSession: %v", err)
	}
}

func TestSetHeaders_UserAgent(t *testing.T) {
	var uaSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uaSeen = r.Header.Get("User-Agent")
		writeJSON(w, HealthResponse{Code: 200})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.RestoreSession("tok", "sid")
	_ = c.CheckSession(context.Background())
	if uaSeen != "test/1.0" {
		t.Errorf("UA = %q; want test/1.0", uaSeen)
	}
}
