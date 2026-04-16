# Phase 05: Stream URL Resolver (Provider-based)
Status: ⬜ Pending
Dependencies: Phase 03

## Objective

Extend `POST /api/stream/{id}/url` endpoint — khi media belongs to cloud library, resolve via `cloudstorage.Provider.GetDownloadURL()` and return direct CDN URL. Works cho fshare hôm nay + any driver registered trong Registry.

## Context

- Current endpoint: [backend/internal/handler/stream.go](backend/internal/handler/stream.go) — generate api_key → return `/api/stream/{id}?api_key=...`
- Playback flow hiện tại (local):
  1. Android `POST /api/stream/{id}/url` → `{url, api_key, expires_in: 7200}`
  2. Android ExoPlayer GET URL → Velox backend `http.ServeFile`
- Target flow (cloud):
  1. Android `POST /api/stream/{id}/url` → `{url: "https://cdn.provider.com/...", expires_in: 300, direct_cdn: true, provider_type: "fshare"}`
  2. Android ExoPlayer GET URL → provider CDN direct (Velox backend không involved)

**⚠️ URL TTL reality (verified từ OSS research):** fshare direct URL short-lived — minutes, có thể single-use. KHÔNG cache server-side. `expires_in` conservative = 300 seconds (5 min). Android refresh strategy reactive (Phase 07). GDrive direct URLs (future) have different TTL characteristics — Provider interface leaves TTL to driver.

## Response Contract

- [ ] Extend response struct:
  ```go
  type StreamURLResponse struct {
      URL          string    `json:"url"`
      ProviderType string    `json:"provider_type,omitempty"` // "fshare"|"google_drive"|... empty = local
      APIKey       string    `json:"api_key,omitempty"`       // only for local
      ExpiresAt    time.Time `json:"expires_at"`
      ExpiresIn    int       `json:"expires_in"`              // seconds
      DirectCDN    bool      `json:"direct_cdn"`              // true = no proxy (cloud)
  }
  ```
- [ ] Android app xem `provider_type` + `direct_cdn` để enable URL refresh interceptor khi cần

## Implementation Steps

### 1. Refactor `stream.go` Handler
- [ ] File: [backend/internal/handler/stream.go](backend/internal/handler/stream.go)
- [ ] Current: mixed logic generate api_key inline → dispatch by library type
- [ ] New structure:
  ```go
  func (h *StreamHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
      id := parseID(r)
      media, err := h.mediaService.GetByID(ctx, id)
      if err != nil { http.Error(...); return }

      library, err := h.libraryRepo.GetByID(ctx, media.LibraryID)
      if err != nil { http.Error(...); return }

      // Local library → existing api_key flow
      if library.StorageProviderID == nil {
          h.serveLocalURL(w, r, media) // existing logic extracted
          return
      }

      // Cloud library → resolve via Provider
      h.serveCloudURL(w, r, media, library)
  }

  func (h *StreamHandler) serveCloudURL(w http.ResponseWriter, r *http.Request, media *model.MediaFile, library *model.Library) {
      sp, err := h.providerRepo.GetByID(ctx, *library.StorageProviderID)
      if err != nil { http.Error(w, "provider not found", 404); return }

      driver, err := h.registry.Get(sp.ProviderType)
      if err != nil { http.Error(w, err.Error(), 500); return }

      creds, err := cloudstorage.DecryptCredentials(h.cryptoKey, sp.CredentialsEncrypted)
      if err != nil { http.Error(w, "decrypt creds", 500); return }

      provider, err := driver.NewProvider(creds)
      if err != nil { http.Error(w, err.Error(), 500); return }

      // Parse provider-scheme path: "fshare://XYZ" → ("fshare", "XYZ")
      _, fileID, ok := parseCloudPath(media.Path)
      if !ok { http.Error(w, "invalid cloud path", 500); return }

      downloadURL, err := provider.GetDownloadURL(ctx, fileID)
      if err != nil {
          switch {
          case errors.Is(err, cloudstorage.ErrSessionExpired):
              http.Error(w, "cloud session expired, re-authenticate", 401)
          case errors.Is(err, cloudstorage.ErrRateLimit):
              http.Error(w, "cloud provider rate limited", 429)
          case errors.Is(err, cloudstorage.ErrFileNotFound):
              http.Error(w, "file removed from cloud", 404)
          default:
              http.Error(w, err.Error(), 500)
          }
          return
      }

      // NOTE: After GetDownloadURL, the provider's client may have silently
      // re-logged in (withSessionRetry on code:201). Compare creds pre/post
      // and persist new token back to storage_providers row if changed.
      // See Phase 08 provider_refresh service for auto-save helper.

      response := StreamURLResponse{
          URL:          downloadURL,
          ProviderType: sp.ProviderType,
          ExpiresAt:    time.Now().Add(5 * time.Minute), // conservative
          ExpiresIn:    300,
          DirectCDN:    true,
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
