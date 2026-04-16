# Plan W: Fshare Cloud Integration

Created: 2026-04-16
Status: ⬜ Pending
Research: ✅ Done (verified API shape from 5 OSS repos — saved 2-4h HAR analysis)

## Overview

Tích hợp **fshare.vn cloud storage** vào Velox như một **library source type mới**. User point Velox đến folder fshare → Velox scan metadata → Android app stream trực tiếp từ fshare CDN (không qua backend).

Mục tiêu: Cho phép host library media >10TB trên fshare VIP mà không cần NAS local.

## Constraints (đã chốt với user)

| Constraint | Value | Impact |
|---|---|---|
| Fshare account | **VIP** (full speed ~100MB/s) | Direct CDN stream khả thi |
| Library size | **>10TB** | Bắt buộc catalog-only (không mirror local) |
| Playback target | **Android direct play ONLY** | Skip FFmpeg transcode/HLS |
| Integration type | **New library source type** | Đổi `libraries` schema + source resolver abstraction |

## Simplifications đạt được

- ❌ **KHÔNG** cần FFmpeg transcode / HLS pipeline cho fshare media
- ❌ **KHÔNG** cần ffprobe remote file (filename + TMDb đủ metadata)
- ❌ **KHÔNG** cần proxy stream qua backend (Android stream thẳng fshare CDN)
- ❌ **KHÔNG** cần thumbnail / trickplay cho fshare phase đầu
- ✅ Backend chỉ đóng vai trò **catalog + URL resolver**

## Architecture

```
┌──────────────────┐                              ┌─────────────┐
│  Velox Backend   │                              │   fshare    │
│  (Go + SQLite)   │ ─── API login/list/getLink ──► │    .vn    │
│                  │                              │  (CDN+API)  │
│  ┌────────────┐  │                              └──────┬──────┘
│  │ fshare pkg │  │                                     │
│  │ (Phase 1)  │  │                                     │
│  └────────────┘  │                                     │
│  ┌────────────┐  │                                     │
│  │ scanner/   │  │                                     │
│  │ fshare     │  │                                     │
│  │ (Phase 4)  │  │                                     │
│  └────────────┘  │                                     │
│  ┌────────────┐  │                                     │
│  │ source     │  │                                     │
│  │ resolver   │  │                                     │
│  │ (Phase 3)  │  │                                     │
│  └────────────┘  │                                     │
│  ┌────────────┐  │                                     │
│  │ media_files│  │                                     │
│  │ (SQLite)   │  │                                     │
│  └────────────┘  │                                     │
└────────┬─────────┘                                     │
         │                                               │
         │ POST /api/stream/{id}/url                     │
         │ → returns { direct_url, expires_in }          │
         ▼                                               │
┌──────────────────┐                                     │
│   Android App    │  ───── HTTP Range GET ──────────────┘
│   (ExoPlayer)    │        (direct from fshare CDN)
└──────────────────┘
```

## Path Scheme

Reuse `media_files.path` column. New URI prefix:

| Source | Path Format | Example |
|---|---|---|
| Local | Absolute filesystem path | `/mnt/data/movies/Avatar.2009.mkv` |
| Fshare | `fshare://{linkcode}` | `fshare://abc123xyz` |

Service layer parses prefix → route to correct resolver.

## fshare API Reference (verified from OSS research)

**Base URL:** `https://api.fshare.vn`

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/user/login` | POST | Login with `{user_email, password, app_key}` → returns `{token, session_id}` |
| `/api/user/get` | GET | Session health check. Body `code == 201` = expired |
| `/api/fileops/list` | GET | Root folder listing `?pageIndex=0&limit=60` |
| `/api/fileops/getFolderList` | POST | Nested folder listing `{token, url, pageIndex, limit}` |
| `/api/session/download` | POST | Get direct URL `{token, url, password, zipflag}` → `{location}` |

**Key invariants:**
- HTTP status always 200 on logical errors. Parse body `code` field first.
- `code == 201` = session expired → re-login needed
- Token (body param) + session_id (cookie) BOTH required for authed calls
- Session TTL ~30min (not documented, empirical from OSS usage)
- Download URL short-lived (minutes, may be single-use) — **never cache server-side**
- App Key: user-registered per account, passed as config (`VELOX_FSHARE_APP_KEY`)
- User-Agent: REQUIRED. Default `okhttp/3.6.0` (default Go UA gets rejected)
- No 429 / Retry-After headers — rate limit shows as persistent 400s
- No captcha logic in any OSS client — API silently fails on suspicious activity

## Database Schema Changes

### Migration 036 — extend `libraries` table

```sql
ALTER TABLE libraries ADD COLUMN source_type TEXT NOT NULL DEFAULT 'local';
-- 'local' | 'fshare'
ALTER TABLE libraries ADD COLUMN source_credentials_id INTEGER REFERENCES cloud_sessions(id);
ALTER TABLE libraries ADD COLUMN source_root_id TEXT; -- fshare folder code (root of this library)
```

### Migration 037 — `cloud_sessions` table

```sql
CREATE TABLE cloud_sessions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL, -- 'fshare'
    account_email TEXT NOT NULL,
    token_encrypted BLOB NOT NULL, -- AES-GCM(token) using VELOX_CLOUD_SECRET
    token_expires_at DATETIME NOT NULL,
    last_refresh_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider, account_email)
);
```

## Phases

| Phase | Name | Status | Tasks | Dep |
|-------|------|--------|-------|-----|
| 01 | [fshare API Client Package](phase-01-fshare-api-client.md) | ⬜ | 8 | — |
| 02 | [Database Schema + Encryption](phase-02-database-schema.md) | ⬜ | 5 | 01 |
| 03 | [Source Resolver Abstraction](phase-03-source-abstraction.md) | ⬜ | 6 | 02 |
| 04 | [Fshare Scanner + Metadata Matching](phase-04-fshare-scanner.md) | ⬜ | 7 | 03 |
| 05 | [Stream URL Resolver (Backend)](phase-05-stream-url-resolver.md) | ⬜ | 4 | 03 |
| 06 | [Admin UI — Add Fshare Library (Webapp)](phase-06-admin-ui-webapp.md) | ⬜ | 7 | 05 |
| 07 | [Android Integration — Reactive URL Refresh + Badge](phase-07-android-integration.md) | ⬜ | 4 | 05 |
| 08 | [Session Resilience + Polish](phase-08-resilience-polish.md) | ⬜ | 6 | 07 |

**Total:** 47 tasks across 8 phases

## Backend Changes Summary

| File | Change Type | Phase |
|---|---|---|
| `backend/pkg/fshare/` | NEW package | 01 |
| `backend/pkg/crypto/aesgcm.go` | NEW — encryption helper | 02 |
| `backend/internal/database/migrate/036_libraries_source_type.go` | NEW migration | 02 |
| `backend/internal/database/migrate/037_cloud_sessions.go` | NEW migration | 02 |
| `backend/internal/model/library.go` | Extend — add `SourceType`, `SourceRootID` | 02 |
| `backend/internal/model/cloud_session.go` | NEW model | 02 |
| `backend/internal/repository/library.go` | Extend — source fields | 02 |
| `backend/internal/repository/cloud_session.go` | NEW repo | 02 |
| `backend/internal/source/resolver.go` | NEW — `SourceResolver` interface | 03 |
| `backend/internal/source/local_resolver.go` | NEW — wrap existing logic | 03 |
| `backend/internal/source/fshare_resolver.go` | NEW — fshare implementation | 03 |
| `backend/internal/scanner/fshare_walker.go` | NEW — cloud scanner | 04 |
| `backend/internal/scanner/pipeline.go` | Modify — dispatch by source_type | 04 |
| `backend/internal/handler/stream.go` | Modify — source-aware URL gen | 05 |
| `backend/internal/handler/library.go` | Modify — source_type field in API | 06 |
| `backend/internal/handler/cloud_auth.go` | NEW — fshare login test endpoint | 06 |

## Frontend Changes Summary

| File | Change Type | Phase |
|---|---|---|
| `webapp/src/pages/admin/LibrariesPage.tsx` | Modify — source_type picker | 06 |
| `webapp/src/components/admin/FshareLoginForm.tsx` | NEW | 06 |
| `webapp/src/components/admin/FshareFolderPicker.tsx` | NEW — tree browser | 06 |
| `webapp/src/api/cloud.ts` | NEW — fshare API client | 06 |
| `webapp/src/components/MediaCard.tsx` | Modify — add ☁️ badge cho fshare source | 06 |

## Android Changes Summary

| File | Change Type | Phase |
|---|---|---|
| `android/.../data/repository/PlaybackRepository.kt` | Modify — handle 403 URL refresh | 07 |
| `android/.../presentation/ui/components/MediaCard.kt` | Modify — ☁️ badge | 07 |
| `android/.../utils/UrlRefreshInterceptor.kt` | NEW — OkHttp interceptor for ExoPlayer | 07 |

## Critical Gotchas

1. ~~**fshare API không public**~~ **→ RESOLVED:** 5 OSS repos dùng cùng endpoints + auth flow. API reference verified (see above).
2. **Session TTL ~30min** — fshare không document. Auto-refresh scheduler every 5min, preemptive re-login at 25min mark. Client also has lazy `withSessionRetry` (Phase 01) on `code == 201`.
3. **Download URL short-lived / single-use** — Research xác nhận. Backend KHÔNG cache; Android reactive-refresh only (interceptor on 403/404). NO proactive URL rotation.
4. **Error layer inversion** — HTTP always 200 on errors. PHẢI parse body `code` field BEFORE reading HTTP status. `code == 201` = session expired → re-login.
5. **Password persistence required** — fshare không có refresh_token grant. Auto-refresh cần email+password decrypt → re-login. Mitigate: AES-256-GCM, unique per-install key (`VELOX_CLOUD_SECRET`).
6. **Credentials at rest** — Full account access. PHẢI encrypt. Key rotation documented as manual process (cutover: disable flag → re-login all sessions with new key → enable flag).
7. **TMDb rate limit khi first scan 10TB** — 10TB ~= 2500 files (trung bình 4GB/file). TMDb cho 40 req/10s. Cần batch + throttle (reuse existing logic).
8. **IsHDR / DVProfile unknown cho fshare files** — ExoPlayer tự detect runtime. Column `is_hdr = 0`, `dv_profile = 0`. Safe nhờ migration 035 sentinel pattern (IsHDR always runtime-probed, never DB-read for decisions).
9. **No 429 / no captcha in API** — Rate limit = persistent 400s. Captcha (if triggered) = silent failure. Cookie-paste fallback trong Phase 08 cho edge case login loops.
10. **File move/rename bên fshare** — `linkcode` stable, không ảnh hưởng. File DELETE bên fshare → next scan mark `media_files.status = 'missing'`.
11. **User-Agent bắt buộc** — Default `okhttp/3.6.0`. Go default UA gets rejected.
12. **VIP account required** — Free accounts fail silently on `/api/session/download`. Phase 01 checks `account_type == "Vip"` after login, returns `ErrNotVIP`. Admin UI gates library creation.

## Non-Goals (Out of Scope)

- ❌ Web browser playback cho fshare media (CORS + codec hell — không support phase này)
- ❌ Trickplay / thumbnail cho fshare files (downloading 10TB để generate thumbnails không realistic)
- ❌ Subtitle extract từ container (cần download full file → bỏ qua; dùng external subtitle search provider)
- ❌ iOS playback (không có app iOS, user chỉ dùng Android)
- ❌ Multi-cloud providers (Google Drive, Dropbox) — abstract interface mở rộng được nhưng implement sau

## Rollout Strategy

1. **Feature flag:** `VELOX_CLOUD_SOURCES_ENABLED=true` default OFF. Admin UI và API 404 khi flag OFF.
2. **Prerequisites:**
   - Generate `VELOX_CLOUD_SECRET` (`openssl rand -hex 32`)
   - Set `VELOX_FSHARE_APP_KEY` — MVP dùng observed OSS value (e.g., `dMnqMMZMUnN5YpvKENaEhdQQ5jxDqddt`). Register official key qua fshare dev portal cho prod (see Phase 01 sourcing options)
   - fshare account: email + password + VIP status active
3. **First test:** 1 library nhỏ (~50 files) để verify scan + play flow end-to-end
4. **Scale test:** Full library >10TB → monitor scan time, TMDb rate limit, DB size, fshare session refresh cadence
5. **Prod rollout:** Flag ON sau khi scale test pass

## Success Criteria

- [ ] User add được fshare library qua admin UI (email + password login, VIP account)
- [ ] Scan 10TB library < 2h (TMDb throttle dominant bottleneck)
- [ ] Android app stream fshare media direct từ CDN (adb logcat verifies `download.fs*.fshare.vn`; backend bandwidth ≈ 0)
- [ ] URL reactive-refresh khi 403/404 giữa chừng playback (OkHttp interceptor)
- [ ] Session auto re-login every 25min scheduler + lazy on `code: 201` (TTL ~30min)
- [ ] Flaky 400s retry với exponential backoff (max 5 attempts)
- [ ] 0 regression trên local libraries (backwards compat)
- [ ] Tất cả fshare credentials (email + password + token + session_id) encrypted at rest với AES-256-GCM
