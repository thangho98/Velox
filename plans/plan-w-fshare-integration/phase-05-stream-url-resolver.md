# Phase 05: Stream URL Resolver (Backend)
Status: ⬜ Pending
Dependencies: Phase 03

## Objective

Extend existing `POST /api/stream/{id}/url` endpoint để return fshare direct URL thay vì internal api_key URL khi `media.library.source_type = 'fshare'`. Android ExoPlayer stream trực tiếp từ fshare CDN.

## Context

- Current endpoint: [backend/internal/handler/stream.go](backend/internal/handler/stream.go) — generate api_key → return `/api/stream/{id}?api_key=...`
- Playback flow hiện tại (local):
  1. Android `POST /api/stream/{id}/url` → `{url, api_key, expires_in: 7200}`
  2. Android ExoPlayer GET URL → Velox backend `http.ServeFile`
- Target flow (fshare):
  1. Android `POST /api/stream/{id}/url` → `{url: "https://download.fsxxx.fshare.vn/...", expires_in: 300, direct_cdn: true}`
  2. Android ExoPlayer GET URL → fshare CDN direct (Velox backend không involved)

**⚠️ URL TTL reality (verified từ OSS research):** fshare direct URL short-lived — minutes, có thể single-use. KHÔNG cache server-side. `expires_in` conservative = 300 seconds (5 min) cho Android biết pattern. Android refresh strategy reactive (Phase 07).

## Response Contract

- [ ] Extend response struct:
  ```go
  type StreamURLResponse struct {
      URL          string    `json:"url"`
      SourceType   string    `json:"source_type"` // "local" | "fshare"
      APIKey       string    `json:"api_key,omitempty"` // only for local
      ExpiresAt    time.Time `json:"expires_at"`
      ExpiresIn    int       `json:"expires_in"` // seconds
      DirectCDN    bool      `json:"direct_cdn"` // true = no proxy (fshare)
  }
  ```
- [ ] Android app xem `source_type` để biết cần refresh URL logic khác

## Implementation Steps

### 1. Refactor `stream.go` Handler
- [ ] File: [backend/internal/handler/stream.go](backend/internal/handler/stream.go)
- [ ] Current: mixed logic generate api_key inline → tách ra thành service call
- [ ] New structure:
  ```go
  func (h *StreamHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
      id := parseID(r)
      media, err := h.mediaService.GetByID(ctx, id)
      if err != nil { http.Error(...); return }

      library, err := h.libraryRepo.GetByID(ctx, media.LibraryID)
      if err != nil { http.Error(...); return }

      resolver, err := h.resolverFactory.For(library.SourceType)
      if err != nil { http.Error(...); return }

      direct, err := resolver.ResolveStreamURL(ctx, source.StreamRef{
          SourcePath: media.Path,
          LibraryID:  library.ID,
      })
      if err != nil {
          // Typed error → status code
          switch {
          case errors.Is(err, source.ErrSessionExpired):
              http.Error(w, "cloud session expired, re-authenticate", 401)
          case errors.Is(err, source.ErrRateLimit):
              http.Error(w, "cloud provider rate limited", 429)
          default:
              http.Error(w, err.Error(), 500)
          }
          return
      }

      response := StreamURLResponse{
          URL:        direct.URL,
          SourceType: library.SourceType,
          ExpiresAt:  direct.ExpiresAt,
          ExpiresIn:  int(time.Until(direct.ExpiresAt).Seconds()),
          DirectCDN:  library.SourceType != "local",
      }
      if library.SourceType == "local" {
          response.APIKey = extractAPIKeyFromURL(direct.URL)
      }
      writeJSON(w, response)
  }
  ```

### 2. TTL Strategy
- [ ] **Local:** Based on media duration (existing logic — 2h minimum)
- [ ] **Fshare:** Fixed conservative value `expires_in = 300` (5 min). fshare API không return TTL explicitly — URL có thể single-use hoặc expire trong phút. Client phải treat như fragile.
- [ ] Response includes `direct_cdn: true` để Android biết enable reactive refresh logic (403 interceptor)
- [ ] Android strategy: **KHÔNG proactive refresh** cho fshare (URL single-use không benefit). Reactive only — interceptor refresh on 403/404 (Phase 07).

### 3. Error Handling + Typed Errors
- [ ] Add trong `backend/internal/source/errors.go`:
  ```go
  var (
      ErrSessionExpired     = errors.New("cloud session expired")
      ErrInvalidCredentials = errors.New("cloud credentials invalid")
      ErrRateLimit          = errors.New("cloud provider rate limited")
      ErrFileNotFound       = errors.New("source file not found")
      ErrCaptchaRequired    = errors.New("cloud provider requires captcha")
  )
  ```
- [ ] FshareResolver wrap fshare.Client errors → source errors (decouple layer)

### 4. Bypass api_key Auth cho Direct CDN
- [ ] Current middleware: [backend/internal/middleware/auth.go](backend/internal/middleware/auth.go) — check api_key query param
- [ ] Fshare response không có api_key → Android ExoPlayer call trực tiếp fshare domain, KHÔNG qua Velox middleware
- [ ] **NO code change needed** — middleware chỉ apply trên `/api/stream/...` routes của Velox

### 5. NO URL Caching ⚠️ (Changed from original plan)

**Decision:** KHÔNG cache `location` URL. Research xác nhận URL có thể single-use / expire trong phút. Caching = stale URLs → 403 errors.

- Mỗi `POST /api/stream/{id}/url` call → fresh fshare API call → new URL
- Cost: 1 fshare API call per playback start. Với library scan 1x/day và playback ~10 request/user/day → không trigger rate limit.
- Server mem footprint: 0 (không store URL)
- If future observation shows URL valid >60s → reconsider minimal cache (defer to Phase 08 polish)

## Acceptance Criteria

- [ ] Unit test `GetStreamURL` handler với mocked resolver factory:
  - [ ] Local library → returns response với `api_key` + `direct_cdn: false`
  - [ ] Fshare library → returns response với CDN URL + `direct_cdn: true`, NO api_key
  - [ ] Resolver returns ErrSessionExpired → HTTP 401
  - [ ] Resolver returns ErrRateLimit → HTTP 429
- [ ] Integration test: real fshare library (dev env) → `POST /api/stream/{id}/url` → curl trả URL → verify fshare CDN returns file bytes
- [ ] Zero regression trên local stream: existing test suite pass
- [ ] Postman/curl manual: `POST /api/stream/{fshare_id}/url` với JWT → response format đúng

## Gotchas

- **expires_at trong JSON**: time.Time marshals to RFC3339 — Android phải parse ISO 8601 (already supported).
- **Direct CDN URL có thể HTTP (not HTTPS)**: fshare CDN historically HTTPS. Verify khi smoke test. Nếu HTTP → Android cần cleartext permission (dev option hoặc manifest).
- **Session auto-refresh trong resolver:** FshareResolver từ Phase 03 handle code:201 retry. Handler chỉ thấy error cuối cùng. Nếu session refresh thành công → response OK. Fail → 401.
- **CORS không relevant**: Android client không enforce CORS. Web browser không hỗ trợ fshare streaming phase này.
- **URL encoding**: fshare trả URL có query params (auth tokens). Android phải pass nguyên — không re-encode.
- **fshare rate limit (persistent 400s):** Resolver retry với backoff (Phase 01 `requestWithBackoff`). Nếu exhaust → `ErrRateLimit` → handler trả 429. Client wait + retry.

## Out of Scope

- Thumbnail URL resolution cho fshare (cần phase riêng)
- Subtitle URL resolution
- Stream session tracking (analytics) — local streams có log, fshare direct CDN không track được
- Bandwidth monitoring cho fshare (user check trên fshare.vn dashboard)
- URL caching / deduplication (decided above — not safe given TTL reality)
