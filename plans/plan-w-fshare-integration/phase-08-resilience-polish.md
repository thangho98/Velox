# Phase 08: Provider Resilience + Polish
Status: ⬜ Pending
Dependencies: Phase 07

## Objective

Harden cloud storage integration cho production: per-provider auto-refresh (works for fshare password-auth + future OAuth providers), rate-limit backoff, admin dashboard health card, captcha fallback, rescan scheduling. Polish edge cases phát hiện qua testing.

## Context

Đây là "cleanup + hardening" phase sau khi core flow (Phase 01-07) hoạt động end-to-end. Target: production-ready.

## Implementation Steps

### 1. Provider Auto-Refresh Service
- [ ] Tạo `backend/internal/service/provider_refresh.go`:
  ```go
  type ProviderRefreshService struct {
      repo      *repository.StorageProviderRepo
      registry  *cloudstorage.Registry
      cryptoKey []byte
      logger    *slog.Logger
  }

  // RefreshExpiring loops over providers nearing expiry (<5min) and dispatches
  // to the right auth flow. Works for password-auth AND OAuth drivers.
  func (s *ProviderRefreshService) RefreshExpiring(ctx context.Context) error {
      providers, err := s.repo.ListExpiringSoon(ctx, time.Now().Add(5*time.Minute))
      if err != nil { return err }

      for _, p := range providers {
          if err := s.refresh(ctx, p); err != nil {
              msg := err.Error()
              p.LastError = &msg
              s.repo.Update(ctx, p)
              s.logger.Warn("provider.refresh_failed", "id", p.ID, "type", p.ProviderType, "err", err)
              continue
          }
      }
      return nil
  }

  func (s *ProviderRefreshService) refresh(ctx context.Context, p *model.StorageProvider) error {
      driver, err := s.registry.Get(p.ProviderType)
      if err != nil { return err }

      creds, err := cloudstorage.DecryptCredentials(s.cryptoKey, p.CredentialsEncrypted)
      if err != nil { return err }

      // Dispatch by auth flow — polymorphic per driver
      var newCreds *cloudstorage.Credentials
      switch d := driver.(type) {
      case cloudstorage.PasswordAuthDriver:
          // fshare: re-login with stored password (no refresh_token grant)
          newCreds, err = d.AuthenticatePassword(ctx, creds.Email, creds.Password)
      case cloudstorage.OAuthDriver:
          // GDrive/OneDrive (future): use refresh_token grant
          newCreds, err = d.RefreshToken(ctx, creds)
      default:
          return fmt.Errorf("driver %s: no refresh mechanism", driver.Type())
      }
      if err != nil { return err }

      // Persist refreshed credentials + update account info
      blob, err := cloudstorage.EncryptCredentials(s.cryptoKey, newCreds)
      if err != nil { return err }

      now := time.Now()
      p.CredentialsEncrypted = blob
      p.TokenExpiresAt = &newCreds.ExpiresAt
      p.LastRefreshAt = &now
      p.LastError = nil

      // Optional: call Provider.GetAccountInfo to refresh account_type/quota
      provider, _ := driver.NewProvider(newCreds)
      if info, aerr := provider.GetAccountInfo(ctx); aerr == nil {
          p.AccountType = &info.AccountType
          p.QuotaBytes = &info.QuotaBytes
          p.UsedBytes = &info.UsedBytes
          if !info.ExpiresAt.IsZero() {
              p.AccountExpiresAt = &info.ExpiresAt
          }
      }

      return s.repo.Update(ctx, p)
  }
  ```
- [ ] Dispatch polymorphic by driver auth flow — fshare uses password re-login, GDrive (future) uses OAuth refresh_token — service code unchanged when new driver added
- [ ] Add to `tasks` table: `provider_refresh`, interval **5min**
- [ ] Provider-agnostic: works với password-auth hôm nay và OAuth drivers sau này

### 2. Rate Limit Backoff (Already Implemented Phase 01)

`requestWithBackoff` đã có trong Phase 01. Phase 08 polish:

- [ ] Add per-session rate limit counter: nếu 1 session hit ErrRateLimit > 3x/hour → mark `last_error = "rate_limit"` → admin dashboard show warning
- [ ] Global rate limit (tất cả sessions): nếu tổng flaky-400 rate > 20/min → throttle API calls (leaky bucket)
- [ ] Metrics: `fshare_flaky_400_total`, `fshare_rate_limited_total` exposed qua `/api/admin/cloud/stats`

**Note:** fshare KHÔNG có 429 hoặc Retry-After headers (verified research). Rate limiting hiện ra như persistent 400s hoặc `code != 200`. Backoff: exponential 2s base, 5 max attempts (Phase 01).

### 3. Admin Dashboard — Provider Health Card
- [ ] Backend endpoint `GET /api/admin/cloud/providers/health` — aggregated health across all providers
- [ ] Reuses per-provider list from Phase 06 `GET /api/admin/cloud/providers`
- [ ] Health states (per-provider):
  - ✅ `healthy` — token valid, last_refresh < 20min ago
  - ⚠️ `expiring_soon` — <5min until estimated expiry
  - ❌ `expired` — past estimated expiry (TokenExpiresAt < now)
  - ❌ `error` — last_error not null
- [ ] Dashboard widget: top-K unhealthy providers với one-click "Refresh all"
- [ ] Webapp card trên admin dashboard:
  ```tsx
  <CloudSessionsCard>
    <SessionRow email="..." provider="fshare" status="healthy" />
    <SessionRow email="..." provider="fshare" status="expiring_soon" />
  </CloudSessionsCard>
  ```
- [ ] "Refresh now" button — trigger manual refresh
- [ ] "Re-authenticate" button — reopen FshareLoginForm với prefilled email

### 4. Silent Rate-Limit Fallback (Cookie Paste)

**Note:** Research xác nhận fshare KHÔNG có captcha logic trong API (OSS repos không handle captcha). Khi trigger rate limit cực đoan hoặc account suspended → API trả silent failures. Cookie paste workaround dùng cho edge case khi email/password login fail liên tục.

- [ ] Scenario: persistent login failures (3+ consecutive `ErrInvalidCredentials` sau reset password)
- [ ] UI option: "Alternative: Paste session cookie from browser"
- [ ] User flow:
  1. Login fshare.vn qua browser (may solve manual challenge)
  2. F12 DevTools → Application → Cookies → copy `session_id` + extract `token` từ any API response
  3. Paste vào Velox form → backend validate bằng `/api/user/get` → save as session (credentials null, token + session_id only)
- [ ] Backend: `POST /api/admin/cloud/fshare/sessions/from-cookie` — body `{email, token, session_id}` → validate `/api/user/get` → save
- [ ] **Limitation:** Sessions from cookie paste KHÔNG auto-refresh (no credentials to re-login). User phải re-paste mỗi 25-30min. Document trong tooltip.
- [ ] Cookie paste sessions: `credentials_encrypted` NULL → auto-refresh service skip row.

### 5. Rescan Scheduling
- [ ] Add scheduled task: `cloud_library_rescan`
- [ ] Default interval: 24h (admin configurable)
- [ ] Triggers scan for each cloud library (any provider): walks tree, detects new files, marks missing files
- [ ] Websocket event `scan.scheduled_start` + `scan.scheduled_complete`
- [ ] Admin UI: "Last scan" timestamp trên library card

### 6. Error Observability
- [ ] Structured logging tagged by provider:
  ```go
  logger.Info("provider.auth", "type", "fshare", "email", email, "duration_ms", dur)
  logger.Warn("provider.rate_limited", "type", "fshare", "retry_after_s", 10)
  logger.Error("provider.refresh_failed", "id", id, "type", "fshare", "err", err)
  ```
- [ ] Admin dashboard: "Recent cloud errors" widget — last 10 errors aggregated across providers
- [ ] Metrics counters (lightweight atomic, không Prometheus):
  - `provider_api_calls_total{type, endpoint, status}`
  - `provider_url_refreshes_total{type}`
  - `provider_auth_refreshes_total{type, outcome}`
- [ ] Expose via `GET /api/admin/cloud/stats`

## Acceptance Criteria

- [ ] Auto-refresh service runs on schedule — verified qua admin dashboard "last_refresh_at" updates
- [ ] Rate limit test: simulate 429 storm → client backs off exponentially, eventually succeeds
- [ ] Captcha fallback flow works: paste cookie → session created → list folders OK
- [ ] Rescan scheduled task runs every 24h — new files appear in library
- [ ] Error dashboard shows recent fshare errors với timestamp + session
- [ ] Load test: 5 concurrent users playing fshare media → URL refresh doesn't thrash fshare API (metric: <1 API call per user per hour steady-state)

## Non-functional Targets

- Session refresh latency: < 5s
- Memory overhead per session: < 10KB
- Rate limit backoff max wait: **32s** (5 attempts, base 2s exponential: 2+4+8+16 ≈ 30s cumulative; capped before 6th attempt)
- Admin dashboard load time: < 200ms (cached session status)
- Preemptive refresh tick: 5min (check every 5min for sessions expiring in <5min)
- Target re-login frequency: every 25min per active session (~2.4x/hour per library)

## Gotchas

- **Password storage**: Research confirmed — fshare KHÔNG có refresh_token grant. Email+password persistence required cho auto-refresh. AES-256-GCM mitigate. Document in security README.
- **Scheduler conflict**: If auto-refresh + user manual refresh cùng lúc → mutex per session row (Go mutex map `map[int64]*sync.Mutex`).
- **Rate limit per-IP**: Fshare likely rate limit per source IP — Velox server = single IP cho tất cả users. Monitor `fshare_flaky_400_total` metric.
- **Cookie paste sessions**: Không auto-refresh. User phải re-paste mỗi ~25min. UI gợi ý switch sang email+password khi có thể.
- **Relogin storm**: Khi app restart, nhiều sessions cùng expired → tất cả auto-refresh đồng thời = rate limit spike. Mitigate: stagger refresh với random jitter (0-60s) khi startup.
- **Account_type validation**: Check `AccountType == "Vip"` mỗi refresh — nếu account hết VIP → `last_error = "not_vip"`, admin dashboard alert.

## Release Checklist

- [ ] CHANGELOG.md entry: "Add fshare cloud library support (opt-in via VELOX_CLOUD_SOURCES_ENABLED)"
- [ ] `.env.example` updated với:
  - `VELOX_CLOUD_SOURCES_ENABLED=false`
  - `VELOX_CLOUD_SECRET=<generate via: openssl rand -hex 32>`
  - `VELOX_FSHARE_APP_KEY=<user-registered via fshare dev portal>`
  - `VELOX_FSHARE_USER_AGENT=okhttp/3.6.0` (optional, default)
- [ ] `docs/cloud-sources.md` user guide:
  - Prerequisites (VIP account)
  - Setup steps (enable flag, generate secret, restart)
  - Troubleshooting (captcha, rate limit)
  - Limitations (no transcode, Android only)
- [ ] Backend audit checklist: fshare error paths all wrap context (reuse Plan Q patterns)
- [ ] `make test` + `make lint` clean
- [ ] Smoke test on Synology prod: enable flag, add library, scan, play → verify end-to-end
- [ ] Rollback plan documented: disable flag → all fshare endpoints return 503, local libraries unaffected

## Out of Scope (Future Plans)

- **Plan W+1**: Google Drive driver (reuse cloudstorage.Provider interface — ~5 days implementation)
- **Plan W+2**: OneDrive driver (OAuth + MSAL)
- **Plan W+3**: Thumbnail generation cho cloud files (partial download + ffmpeg)
- **Plan W+4**: Web browser playback cho cloud media (HLS proxy — big effort)
- Offline download / sync cho specific files user starred
- Chromecast support cho cloud media (requires app server proxy vs direct CDN trade-off)
- Cross-provider file migration (e.g., fshare → gdrive copy)
