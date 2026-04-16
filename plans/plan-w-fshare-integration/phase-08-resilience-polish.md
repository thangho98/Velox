# Phase 08: Session Resilience + Polish
Status: ⬜ Pending
Dependencies: Phase 07

## Objective

Harden fshare integration cho production: session auto-refresh, rate-limit backoff, admin dashboard health card, captcha fallback, rescan scheduling. Polish edge cases phát hiện qua testing.

## Context

Đây là "cleanup + hardening" phase sau khi core flow (Phase 01-07) hoạt động end-to-end. Target: production-ready.

## Implementation Steps

### 1. Session Auto-Refresh Service
- [ ] Tạo `backend/internal/service/cloud_session_service.go`:
  ```go
  type CloudSessionService struct {
      repo          *repository.CloudSessionRepo
      clientFactory func() *fshare.Client
      crypto        []byte
      logger        *slog.Logger
  }

  // RefreshExpiring loops over sessions nearing expiry (<5min) and re-logins.
  // fshare session TTL ~30min — preemptively refresh mỗi 25min tick.
  func (s *CloudSessionService) RefreshExpiring(ctx context.Context) error {
      sessions, err := s.repo.ListExpiringSoon(ctx, time.Now().Add(5*time.Minute))
      if err != nil { return err }

      for _, sess := range sessions {
          if err := s.refresh(ctx, sess); err != nil {
              msg := err.Error()
              sess.LastError = &msg
              s.repo.Update(ctx, sess)
              s.logger.Warn("fshare.session_refresh_failed", "id", sess.ID, "err", err)
              continue // don't fail batch
          }
      }
      return nil
  }

  // refresh decrypts credentials, logs in, persists new token + session_id.
  func (s *CloudSessionService) refresh(ctx context.Context, sess *model.CloudSession) error {
      credsJSON, _ := crypto.Decrypt(s.crypto, sess.CredentialsEncrypted)
      var creds model.CredentialsPayload; json.Unmarshal(credsJSON, &creds)

      client := s.clientFactory()
      loginSess, err := client.Login(ctx, creds.Email, creds.Password)
      if err != nil { return err }

      tokEnc, _ := crypto.Encrypt(s.crypto, []byte(loginSess.Token))
      sidEnc, _ := crypto.Encrypt(s.crypto, []byte(loginSess.SessionID))
      now := time.Now()
      expiry := now.Add(25 * time.Minute) // conservative buffer before 30min TTL
      sess.TokenEncrypted = tokEnc
      sess.SessionIDEncrypted = sidEnc
      sess.TokenExpiresAt = &expiry
      sess.LastRefreshAt = &now
      sess.AccountType = &loginSess.AccountType
      sess.LastError = nil
      return s.repo.Update(ctx, sess)
  }
  ```
- [ ] Password persistence đã decided trong Phase 02 — stored in `credentials_encrypted` BLOB.
- [ ] Add to `tasks` table: `cloud_session_refresh`, interval **5min** (check expiring_soon = <5min), re-login nếu cần.
- [ ] Adjusted from original draft: TTL 30min, refresh tick 5min, re-login at 25min mark.

### 2. Rate Limit Backoff (Already Implemented Phase 01)

`requestWithBackoff` đã có trong Phase 01. Phase 08 polish:

- [ ] Add per-session rate limit counter: nếu 1 session hit ErrRateLimit > 3x/hour → mark `last_error = "rate_limit"` → admin dashboard show warning
- [ ] Global rate limit (tất cả sessions): nếu tổng flaky-400 rate > 20/min → throttle API calls (leaky bucket)
- [ ] Metrics: `fshare_flaky_400_total`, `fshare_rate_limited_total` exposed qua `/api/admin/cloud/stats`

**Note:** fshare KHÔNG có 429 hoặc Retry-After headers (verified research). Rate limiting hiện ra như persistent 400s hoặc `code != 200`. Backoff: exponential 2s base, 5 max attempts (Phase 01).

### 3. Admin Dashboard — Cloud Health Card
- [ ] Backend endpoint `GET /api/admin/cloud/sessions` — list all sessions với health status
- [ ] Health states (adjusted to 30min TTL reality):
  - ✅ `healthy` — token valid, last_refresh < 20min ago
  - ⚠️ `expiring_soon` — <5min until estimated expiry
  - ❌ `expired` — past estimated expiry (TokenExpiresAt < now)
  - ❌ `error` — last_error not null
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
- [ ] Add scheduled task: `fshare_library_rescan`
- [ ] Default interval: 24h (admin configurable)
- [ ] Triggers scan for each fshare library: walks tree, detects new files, marks missing files
- [ ] Websocket event `scan.scheduled_start` + `scan.scheduled_complete`
- [ ] Admin UI: "Last scan" timestamp trên library card

### 6. Error Observability
- [ ] Structured logging cho fshare operations:
  ```go
  logger.Info("fshare.login", "email", email, "duration_ms", dur)
  logger.Warn("fshare.rate_limited", "retry_after_s", 10)
  logger.Error("fshare.session_refresh_failed", "session_id", id, "err", err)
  ```
- [ ] Admin dashboard: "Recent cloud errors" widget — last 10 errors từ logs
- [ ] Metrics counters (lightweight — atomic counters, không Prometheus):
  - `fshare_api_calls_total{endpoint, status}`
  - `fshare_url_refreshes_total`
  - `fshare_session_refreshes_total{outcome}`
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

- **Plan W+1**: Google Drive support (reuse SourceResolver abstraction)
- **Plan W+2**: Thumbnail generation cho cloud files (partial download + ffmpeg)
- **Plan W+3**: Web browser playback cho fshare (HLS proxy — big effort)
- **Plan W+4**: Cross-region fshare load balancing (nếu fshare expose multi-CDN)
- Offline download / sync cho specific files user starred
- Chromecast support cho cloud media (requires app server proxy vs direct CDN trade-off)
