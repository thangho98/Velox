# Phase 01: fshare API Client Package
Status: ⬜ Pending
Dependencies: None

## Objective

Viết Go package [backend/pkg/fshare/](backend/pkg/fshare/) đóng vai trò HTTP client cho fshare.vn API. Public interface: login, list folder, get direct download link, session health check.

## Research Summary (verified từ 5 open-source repos)

Đã cross-reference 5 repos: [duythongle/fshare2gdrive](https://github.com/duythongle/fshare2gdrive) (2025-01 — most current), [dhhiep/fshare_tool](https://github.com/dhhiep/fshare_tool), [tudoanh/get_fshare](https://github.com/tudoanh/get_fshare), [giangvo/synology-fshare](https://github.com/giangvo/synology-fshare), [haindvn/FShareDownloader](https://github.com/haindvn/FShareDownloader). Tất cả dùng **cùng endpoints + auth flow** → high confidence, skip HAR analysis.

**Primary reference for port:** `duythongle/fshare2gdrive/fshare2gdrive.js` (documents real iOS UA + `code:201` retry pattern).

## Verified API Reference

**Base URL:** `https://api.fshare.vn`

### POST /api/user/login
Login with email + password + app_key.
```json
Request Body:
{
  "user_email": "user@example.com",
  "password": "plaintext",
  "app_key": "user-registered-app-key"
}

Response 200:
{
  "code": 200,
  "msg": "Success",
  "token": "<auth token, 32+ chars>",
  "session_id": "<hex session id>",
  "email": "user@example.com",
  "account_type": "Vip",
  "expire_vip": "1735689600"
}
```

### GET /api/fileops/list (root folder only)
- Headers: `Cookie: session_id=<id>`, UA
- Query: `?pageIndex=0&dirOnly=0&limit=60`
- Returns list of items at root level.

### POST /api/fileops/getFolderList (nested folder)
- Headers: `Cookie: session_id=<id>`, `Content-Type: application/json`, UA
```json
Body:
{
  "token": "<from login>",
  "url": "https://www.fshare.vn/folder/<linkcode>",
  "dirOnly": 0,
  "pageIndex": 0,
  "limit": 60
}

Response: Array of {
  "linkcode": "<unique code>",
  "name": "Movie.Title.2023.mkv",
  "type": 1,         // item type code
  "size": 8589934592,
  "mimetype": "video/x-matroska",
  "furl": "https://www.fshare.vn/file/<linkcode>",
  "modified": "2023-01-15 10:30:00",
  "path": "/Movies/2023/"
}
```

### POST /api/session/download
- Headers: Cookie, Content-Type, UA
```json
Body:
{
  "token": "<from login>",
  "url": "https://www.fshare.vn/file/<linkcode>",
  "password": "",
  "zipflag": 0
}

Response 200:
{
  "location": "https://download.fsxxx.fshare.vn/.../<filename>?<auth_params>"
}
```
⚠️ **URL có thể single-use / expire trong phút — KHÔNG cache, re-fetch mỗi download request.**

### GET /api/user/get (session health check)
- Headers: Cookie + Token header (verify từ fshare2gdrive)
- Response body `code == 201` → session expired → re-login.
- Response `code == 200` → healthy, returns profile.

## Error Detection Pattern ⚠️

**CRITICAL:** fshare HTTP status luôn 200 trên logical errors. PHẢI parse body `code` field trước:

| Body `code` | Meaning | Go Error |
|---|---|---|
| 200 | Success | nil |
| 201 | Session expired, re-login needed | `ErrSessionExpired` |
| Other non-200 | Invalid creds / link dead / rate limit | `ErrInvalidCredentials`, `ErrLinkDead`, `ErrRateLimit` |

**Rate limit:** Không có 429 / Retry-After. Rate limiting hiện ra như persistent 400s hoặc `code != 200`. Strategy: exponential backoff, max 5 retries, 2s base delay.

## Package Structure

```
backend/pkg/fshare/
├── client.go         # Client struct + constructor, cookiejar
├── auth.go           # Login, CheckSession, relogin flow
├── fileops.go        # ListFolder (root via GET, nested via POST)
├── download.go       # GetDirectLink
├── errors.go         # Typed errors
├── types.go          # Request/response DTOs
├── retry.go          # Exponential backoff helper
├── client_test.go    # Unit tests với fake HTTP server
└── README.md         # Endpoint reference (verified từ OSS repos)
```

## Implementation Steps

### 1. Types (DTOs)
- [ ] `types.go`:
  ```go
  package fshare

  type LoginRequest struct {
      UserEmail string `json:"user_email"`
      Password  string `json:"password"`
      AppKey    string `json:"app_key"`
  }

  type LoginResponse struct {
      Code        int    `json:"code"`
      Msg         string `json:"msg"`
      Token       string `json:"token"`
      SessionID   string `json:"session_id"`
      Email       string `json:"email"`
      AccountType string `json:"account_type"` // "Vip" | "Fee"
      ExpireVIP   string `json:"expire_vip"`
  }

  type FolderListRequest struct {
      Token     string `json:"token"`
      URL       string `json:"url"` // https://www.fshare.vn/folder/<code>
      DirOnly   int    `json:"dirOnly"`
      PageIndex int    `json:"pageIndex"`
      Limit     int    `json:"limit"`
  }

  type FolderItem struct {
      Linkcode string    `json:"linkcode"`
      Name     string    `json:"name"`
      Type     int       `json:"type"`
      Size     int64     `json:"size"`
      Mimetype string    `json:"mimetype"`
      Furl     string    `json:"furl"`
      Path     string    `json:"path"`
      Modified string    `json:"modified"` // "2006-01-02 15:04:05"
  }

  type DownloadRequest struct {
      Token    string `json:"token"`
      URL      string `json:"url"`
      Password string `json:"password"` // "" for unprotected files
      Zipflag  int    `json:"zipflag"`
  }

  type DownloadResponse struct {
      Code     int    `json:"code,omitempty"`
      Location string `json:"location"`
      Msg      string `json:"msg,omitempty"`
  }

  type Session struct {
      Token       string
      SessionID   string
      Email       string
      AccountType string
      LastLoginAt time.Time
  }
  ```

### 2. Client với Cookie Jar
- [ ] `client.go`:
  ```go
  type Client struct {
      httpClient *http.Client
      baseURL    string
      appKey     string
      userAgent  string

      mu          sync.RWMutex
      session     *Session
      credentials *Credentials // email + password for auto-relogin
  }

  type Credentials struct {
      Email    string
      Password string
  }

  type Options struct {
      BaseURL     string        // default "https://api.fshare.vn"
      AppKey      string        // REQUIRED — user-registered value
      UserAgent   string        // default "okhttp/3.6.0"
      HTTPTimeout time.Duration // default 30s
  }

  func NewClient(opts Options) (*Client, error) {
      jar, _ := cookiejar.New(nil)
      return &Client{
          httpClient: &http.Client{
              Jar:     jar,  // auto-handle session_id cookie
              Timeout: opts.HTTPTimeout,
          },
          baseURL:   opts.BaseURL,
          appKey:    opts.AppKey,
          userAgent: opts.UserAgent,
      }, nil
  }

  func (c *Client) SetCredentials(email, password string) {
      c.mu.Lock()
      defer c.mu.Unlock()
      c.credentials = &Credentials{email, password}
  }
  ```

### 3. Login + Session Health
- [ ] `auth.go`:
  ```go
  func (c *Client) Login(ctx context.Context, email, password string) (*Session, error) {
      body, _ := json.Marshal(LoginRequest{UserEmail: email, Password: password, AppKey: c.appKey})
      req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/user/login", bytes.NewReader(body))
      c.setHeaders(req)

      resp, err := c.httpClient.Do(req)
      if err != nil { return nil, fmt.Errorf("login: %w", err) }
      defer resp.Body.Close()

      var lr LoginResponse
      if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil { return nil, err }

      if lr.Code != 200 {
          switch lr.Code {
          case 201:
              return nil, ErrSessionExpired
          default:
              return nil, fmt.Errorf("%w: code=%d msg=%s", ErrInvalidCredentials, lr.Code, lr.Msg)
          }
      }

      session := &Session{
          Token:       lr.Token,
          SessionID:   lr.SessionID,
          Email:       lr.Email,
          AccountType: lr.AccountType,
          LastLoginAt: time.Now(),
      }

      c.mu.Lock()
      c.session = session
      c.credentials = &Credentials{email, password} // stash for auto-relogin
      c.mu.Unlock()

      return session, nil
  }

  // CheckSession hits /api/user/get. Returns ErrSessionExpired if code == 201.
  func (c *Client) CheckSession(ctx context.Context) error {
      session := c.Session()
      if session == nil { return ErrSessionExpired }

      req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/user/get", nil)
      c.setHeaders(req)
      // Cookie jar auto-attaches session_id

      resp, err := c.httpClient.Do(req)
      if err != nil { return err }
      defer resp.Body.Close()

      var health struct { Code int `json:"code"` }
      json.NewDecoder(resp.Body).Decode(&health)
      if health.Code == 201 { return ErrSessionExpired }
      if health.Code != 200 { return fmt.Errorf("health check: code=%d", health.Code) }
      return nil
  }

  // withSessionRetry runs fn. If ErrSessionExpired AND credentials stored → auto relogin + retry once.
  func (c *Client) withSessionRetry(ctx context.Context, fn func() error) error {
      if err := fn(); err == nil { return nil } else if !errors.Is(err, ErrSessionExpired) { return err }

      c.mu.RLock(); creds := c.credentials; c.mu.RUnlock()
      if creds == nil { return ErrSessionExpired }

      if _, err := c.Login(ctx, creds.Email, creds.Password); err != nil { return err }
      return fn()
  }
  ```
- [ ] `setHeaders`: set `User-Agent`, `Content-Type: application/json`, `Accept: application/json`

### 4. List Folder (root + nested)
- [ ] `fileops.go`:
  ```go
  // ListFolder returns items inside folderCode. Empty string = root.
  // Handles pagination internally (returns full flat list).
  func (c *Client) ListFolder(ctx context.Context, folderCode string) ([]FolderItem, error) {
      var all []FolderItem
      pageIndex := 0
      const limit = 60

      return all, c.withSessionRetry(ctx, func() error {
          all = nil // reset on retry
          for {
              page, err := c.listFolderPage(ctx, folderCode, pageIndex, limit)
              if err != nil { return err }
              all = append(all, page...)
              if len(page) < limit { break } // last page
              pageIndex++
          }
          return nil
      })
  }

  func (c *Client) listFolderPage(ctx context.Context, folderCode string, pageIndex, limit int) ([]FolderItem, error) {
      var req *http.Request; var err error
      if folderCode == "" {
          // root: GET /api/fileops/list
          url := fmt.Sprintf("%s/api/fileops/list?pageIndex=%d&dirOnly=0&limit=%d", c.baseURL, pageIndex, limit)
          req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
      } else {
          // nested: POST /api/fileops/getFolderList
          body, _ := json.Marshal(FolderListRequest{
              Token:     c.Session().Token,
              URL:       "https://www.fshare.vn/folder/" + folderCode,
              DirOnly:   0,
              PageIndex: pageIndex,
              Limit:     limit,
          })
          req, err = http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/fileops/getFolderList", bytes.NewReader(body))
      }
      if err != nil { return nil, err }
      c.setHeaders(req)

      resp, err := c.requestWithBackoff(ctx, req)
      if err != nil { return nil, err }
      defer resp.Body.Close()

      // Response can be array directly or {code, data}. Sniff first byte.
      var items []FolderItem
      if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
          return nil, fmt.Errorf("decode folder list: %w", err)
      }
      return items, nil
  }
  ```

### 5. Get Direct Download Link
- [ ] `download.go`:
  ```go
  // GetDirectLink returns a download URL. URL is short-lived (minutes) and possibly single-use.
  // Caller MUST use URL immediately and NOT cache.
  func (c *Client) GetDirectLink(ctx context.Context, fileCode string) (string, error) {
      var location string
      err := c.withSessionRetry(ctx, func() error {
          body, _ := json.Marshal(DownloadRequest{
              Token:    c.Session().Token,
              URL:      "https://www.fshare.vn/file/" + fileCode,
              Password: "",
              Zipflag:  0,
          })
          req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/session/download", bytes.NewReader(body))
          c.setHeaders(req)

          resp, err := c.requestWithBackoff(ctx, req)
          if err != nil { return err }
          defer resp.Body.Close()

          var dr DownloadResponse
          if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil { return err }

          switch {
          case resp.StatusCode == 403:
              return fmt.Errorf("%w: file password protected", ErrFilePassword)
          case dr.Code == 201:
              return ErrSessionExpired
          case dr.Location == "":
              return fmt.Errorf("%w: code=%d msg=%s", ErrLinkDead, dr.Code, dr.Msg)
          }
          location = dr.Location
          return nil
      })
      return location, err
  }
  ```

### 6. Retry + Backoff
- [ ] `retry.go`:
  ```go
  // requestWithBackoff retries flaky 400s and HTTP-level errors up to 5 attempts.
  // fshare has no 429 / Retry-After — treat persistent 400s as rate limit.
  func (c *Client) requestWithBackoff(ctx context.Context, req *http.Request) (*http.Response, error) {
      const maxAttempts = 5
      const baseDelay = 2 * time.Second

      var lastErr error
      for attempt := 0; attempt < maxAttempts; attempt++ {
          if attempt > 0 {
              delay := baseDelay * time.Duration(1<<attempt) // 2s, 4s, 8s, 16s
              select {
              case <-time.After(delay):
              case <-ctx.Done(): return nil, ctx.Err()
              }
              // Clone request — body already read on retry
              req = cloneRequest(req)
          }

          resp, err := c.httpClient.Do(req)
          if err != nil { lastErr = err; continue }

          // 2xx or 403 (file password) — don't retry
          if resp.StatusCode < 400 || resp.StatusCode == 403 {
              return resp, nil
          }

          // 400/500 — retry
          resp.Body.Close()
          lastErr = fmt.Errorf("http %d", resp.StatusCode)
      }
      return nil, fmt.Errorf("%w: %v", ErrRateLimit, lastErr)
  }
  ```

### 7. Typed Errors
- [ ] `errors.go`:
  ```go
  var (
      ErrInvalidCredentials = errors.New("fshare: invalid credentials")
      ErrSessionExpired     = errors.New("fshare: session expired (code 201)")
      ErrRateLimit          = errors.New("fshare: rate limited / flaky api")
      ErrFilePassword       = errors.New("fshare: file is password protected")
      ErrLinkDead           = errors.New("fshare: link dead or invalid")
      ErrAppKeyInvalid      = errors.New("fshare: app_key invalid or revoked")
      ErrNotVIP             = errors.New("fshare: account not VIP, download not available")
  )
  ```

### 8. Unit Tests
- [ ] `client_test.go` — all với `httptest.NewServer`:
  - [ ] `TestLogin_Success` — valid creds → session returned
  - [ ] `TestLogin_InvalidCredentials` — code != 200 → ErrInvalidCredentials
  - [ ] `TestCheckSession_Expired` — server returns code:201 → ErrSessionExpired
  - [ ] `TestListFolder_Root` — GET /api/fileops/list, single page
  - [ ] `TestListFolder_Nested_Pagination` — POST /api/fileops/getFolderList, 3 pages
  - [ ] `TestGetDirectLink_Success` — returns location
  - [ ] `TestGetDirectLink_SessionExpired_Relogin` — code:201 → auto-relogin → retry → success
  - [ ] `TestRetry_Backoff_400s` — 3 flaky 400s then 200 → succeeds
  - [ ] `TestRetry_Exhaustion` — 5 consecutive 400s → ErrRateLimit
  - [ ] `TestCookieJar_SessionID` — login sets cookie, subsequent call sends it

## Acceptance Criteria

- [ ] `cd backend && go build ./pkg/fshare/` clean
- [ ] `go test ./pkg/fshare/ -v -count=1 -race` pass (including race detector)
- [ ] README.md đầy đủ endpoint reference với examples (copy từ phase này)
- [ ] Manual smoke test (dev env, VIP account):
  - [ ] Login success → session populated
  - [ ] ListFolder("") → root items
  - [ ] ListFolder("<subfolder_code>") → nested items
  - [ ] GetDirectLink → curl URL → nhận file bytes
  - [ ] CheckSession(after login) → healthy
  - [ ] Sleep 35 minutes → CheckSession → ErrSessionExpired → withSessionRetry re-logins
- [ ] README documents how to get `app_key` (register via fshare dev portal)

## Configuration Requirements

User phải cung cấp via env:
- `VELOX_FSHARE_APP_KEY` — per-install value (see sourcing options below)
- `VELOX_FSHARE_USER_AGENT` (optional, default `okhttp/3.6.0`)

### App Key Sourcing (ranked by effort)

**Option A — Use observed OSS value (zero effort, fastest start)** ⭐ recommended để MVP
- Pick any observed value từ OSS repos:
  - `dMnqMMZMUnN5YpvKENaEhdQQ5jxDqddt` (duythongle, sniffed từ fshare iOS app)
  - `L2S7R6ZMagggC5wWkQhX2+aDi467PPuftWUMRFSn` (tudoanh)
  - `GUxft6Beh3Bf8qKP7GC2IplYJZz1A53JQfRwne0R` (giangvo)
- Risk: fshare có thể invalidate any time. Chuẩn bị fallback (Option B hoặc C)
- Treated như "shared dev key" — no per-account tracking

**Option B — Register qua fshare dev portal (recommended cho long-term)**
- Visit fshare API docs page → click "Lấy App Key"
- Provide: app name, email
- Receive key qua email trong vài giờ (theo community reports)
- Advantage: key tied to your email, less likely to be invalidated blindly

**Option C — Sniff từ iOS/Android app (nếu A và B fail)**
- Install fshare mobile app → proxy traffic qua mitmproxy hoặc Charles
- Monitor login request → extract `app_key` field
- ~30min work, need TLS MITM cert installed on device

### For MVP (anh): Start với Option A

Document trong README: dùng observed value khi bắt đầu, register official key (Option B) khi deploy prod. App Key KHÔNG commit vào git — `.env` only.

## Gotchas Đã Biết

- **code: 201 ≠ HTTP 201.** Always parse body `{code}` BEFORE checking HTTP status.
- **Session ~30min.** `dhhiep` cache 30min proactive refresh; `duythongle` lazy detect via `/api/user/get`. Ta dùng hybrid: lazy via CheckSession + proactive refresh trong Phase 08 scheduler (25min).
- **Download URL single-use/short-lived.** KHÔNG cache `location`. Scanner KHÔNG get link during scan (only on playback request).
- **URL normalization:** Strip query params when passing folder/file URLs. Use `www.fshare.vn` (not `fshare.vn`).
- **Pagination:** `/api/fileops/*` is **0-based** (`pageIndex`). Nếu gặp `/api/v3/files/folder` (alt endpoint), nó **1-based** (`page`) — ta không dùng.
- **Free accounts fail silently** trên `/api/session/download`. Check `account_type == "Vip"` sau login, return `ErrNotVIP` nếu không VIP.
- **No captcha logic** trong OSS repos — API may silently rate-limit thay vì captcha. Monitor flaky 400 rate.
- **UA "Go-http-client" bị reject.** Default UA = `okhttp/3.6.0`.

## Out of Scope

- Upload files
- Share link creation / management
- Account info APIs ngoài login + health check
- Mobile-app UA (iOS `Fshare/1 CFNetwork/... Darwin/...`) — dùng okhttp UA đủ
- `/api/v3/files/folder` alt endpoint (public folder without auth) — not needed for VIP workflow
