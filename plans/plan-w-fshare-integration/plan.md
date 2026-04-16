# Plan W: Cloud Storage Integration (Fshare + extensible)

Created: 2026-04-16
Status: ⬜ Pending (Phase 01 DONE, live smoke test PASS)
Research: ✅ API verified from 5 OSS repos + live fshare VIP account

## Overview

Tích hợp **cloud storage providers** vào Velox làm library source type mới. Architecture interface-based để mở rộng cho Google Drive, OneDrive, Dropbox sau này mà không cần đổi scanner/stream handler.

**Separation of concerns:**
- **Storage Provider** = "how to connect to this cloud" — 1 provider = 1 cloud account (credentials, health, session refresh). Setup 1 lần.
- **Library** = "what to scan from that provider" — N libraries có thể share 1 provider (e.g., /Movies, /TV Shows, /Anime cùng 1 fshare account).

**Use case MVP:** Fshare VIP + Android direct play (no transcode, >10TB library).

## Constraints (đã chốt với user)

| Constraint | Value | Impact |
|---|---|---|
| Fshare account | VIP (full speed ~100MB/s) | Direct CDN stream khả thi |
| Library size | >10TB | Catalog-only, không mirror local |
| Playback target | Android direct play ONLY | Skip FFmpeg transcode/HLS |
| Integration type | New library source type with provider abstraction | Extensible to GDrive/OneDrive |

## Architecture (3 layers)

```
┌─────────────────────────────────────────────────────────────┐
│                         Admin UI                            │
│  Settings → Storage Providers     │  Libraries → Add        │
│  (manage accounts)                 │  (link folder URL)      │
└────────────┬────────────────────┴────────────┬───────────────┘
             │                                 │
┌────────────▼─────────────────────────────────▼───────────────┐
│              internal/cloudstorage                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Registry   │  │   Provider   │  │  Credentials │        │
│  │   (drivers)  │  │  interface   │  │  (AES-GCM)   │        │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘        │
│         │                 │                                  │
│  ┌──────▼─────────────────▼───────┐                          │
│  │  drivers/fshare               │ ← implements Provider     │
│  │  drivers/googledrive (future) │ ← implements Provider     │
│  │  drivers/onedrive (future)    │ ← implements Provider     │
│  └──────┬─────────────────────────┘                          │
└─────────┼────────────────────────────────────────────────────┘
          │
┌─────────▼──────────┐
│   pkg/fshare       │ ← raw API client (Phase 01 done)
│   (portable lib)   │
└────────────────────┘
```

## Interfaces

```go
// cloudstorage/provider.go

type Item struct {
    ID       string    // provider-native ID (fshare linkcode, gdrive fileId)
    Name     string
    IsFolder bool
    Size     int64
    Mimetype string
    Modified time.Time
    URL      string    // canonical URL
}

type AccountInfo struct {
    Email       string
    DisplayName string
    AccountType string    // "Vip" | "Free" | "Paid"
    QuotaBytes  int64
    UsedBytes   int64
    ExpiresAt   time.Time
}

// Provider operates on an authenticated cloud session.
type Provider interface {
    Type() string
    ParseFolderURL(url string) (folderID string, err error)
    ParseFileURL(url string) (fileID string, err error)
    ListFolder(ctx context.Context, folderID string) ([]Item, error)
    GetDownloadURL(ctx context.Context, fileID string) (string, error)
    GetAccountInfo(ctx context.Context) (*AccountInfo, error)
    CheckHealth(ctx context.Context) error
}

// Driver is the factory + auth flow for a provider type.
type Driver interface {
    Type() string
    DisplayName() string
    AuthFlow() AuthFlow
    NewProvider(creds *Credentials) (Provider, error)
}

// Password-auth providers (fshare, basic FTP-like).
type PasswordAuthDriver interface {
    Driver
    AuthenticatePassword(ctx context.Context, email, password string) (*Credentials, error)
}

// OAuth 2.0 providers (GDrive, OneDrive, Dropbox).
type OAuthDriver interface {
    Driver
    AuthURL(state, redirectURI string) string
    ExchangeCode(ctx context.Context, code, redirectURI string) (*Credentials, error)
    RefreshToken(ctx context.Context, creds *Credentials) (*Credentials, error)
}
```

## Simplifications đạt được

- ❌ KHÔNG cần FFmpeg transcode / HLS pipeline cho cloud media
- ❌ KHÔNG cần ffprobe remote file (filename + TMDb đủ metadata)
- ❌ KHÔNG cần proxy stream qua backend (Android stream thẳng CDN)
- ❌ KHÔNG cần thumbnail / trickplay cho cloud files phase MVP
- ✅ Backend chỉ đóng vai trò catalog + URL resolver

## Data Model

```
┌──────────────────────┐         ┌──────────────────────┐
│ storage_providers    │    ┌────┤ libraries            │
├──────────────────────┤    │    ├──────────────────────┤
│ id                   │◄───┘    │ id                   │
│ user_id              │         │ name                 │
│ provider_type        │         │ path                 │ (local: filesystem path)
│ display_name         │         │ storage_provider_id  │ (nullable FK → providers)
│ account_email        │         │ source_url           │ (cloud: folder URL)
│ credentials_encrypted│         │ ...                  │
│ account_type         │         └──────────────────────┘
│ quota/used/expires   │
│ last_refresh/error   │         Library source type:
└──────────────────────┘         - storage_provider_id IS NULL → local (use path)
                                 - storage_provider_id IS NOT NULL → cloud (use source_url + provider)
```

### Migration 036 — `storage_providers` table

```sql
CREATE TABLE storage_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_type TEXT NOT NULL,          -- "fshare" | "google_drive" | ...
    display_name TEXT NOT NULL,           -- user-defined label
    account_email TEXT NOT NULL,
    credentials_encrypted BLOB NOT NULL,  -- AES-GCM(CredentialsPayload JSON)
    account_type TEXT,                    -- "Vip" | "Free" | "Paid"
    quota_bytes INTEGER,
    used_bytes INTEGER,
    account_expires_at DATETIME,          -- provider's VIP/subscription expiry
    token_expires_at DATETIME,            -- access token expiry (local bookkeeping)
    last_refresh_at DATETIME,
    last_check_at DATETIME,
    last_error TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider_type, account_email)
);
CREATE INDEX idx_storage_providers_user ON storage_providers(user_id);
CREATE INDEX idx_storage_providers_type ON storage_providers(provider_type);
```

### Migration 037 — extend `libraries`

```sql
ALTER TABLE libraries ADD COLUMN storage_provider_id INTEGER REFERENCES storage_providers(id) ON DELETE SET NULL;
ALTER TABLE libraries ADD COLUMN source_url TEXT;
CREATE INDEX idx_libraries_provider ON libraries(storage_provider_id);
```

**Existing libraries:** `storage_provider_id = NULL` → local filesystem (use existing `path` column). Zero migration side-effects.

## Phases

| Phase | Name | Status | Tasks | Dep |
|-------|------|--------|-------|-----|
| 01 | [fshare API Client Package](phase-01-fshare-api-client.md) | ✅ DONE | 8 | — |
| 02 | [Database Schema + Encryption](phase-02-database-schema.md) | ⬜ | 6 | 01 |
| 03 | [CloudStorage Abstraction + Fshare Driver](phase-03-source-abstraction.md) | ⬜ | 8 | 02 |
| 04 | [Scanner using Provider Interface](phase-04-fshare-scanner.md) | ⬜ | 7 | 03 |
| 05 | [Stream URL Resolver via Provider](phase-05-stream-url-resolver.md) | ⬜ | 5 | 03 |
| 06 | [Admin UI — Providers + Libraries (2 screens)](phase-06-admin-ui-webapp.md) | ⬜ | 9 | 05 |
| 07 | [Android Integration](phase-07-android-integration.md) | ⬜ | 4 | 05 |
| 08 | [Resilience + Polish](phase-08-resilience-polish.md) | ⬜ | 6 | 07 |

**Total:** 53 tasks across 8 phases (Phase 01 complete).

## Backend Changes Summary

| File | Change | Phase |
|---|---|---|
| `backend/pkg/fshare/` | ✅ Done — raw API client | 01 |
| `backend/pkg/crypto/aesgcm.go` | NEW — AES-GCM helper | 02 |
| `backend/internal/database/migrate/036_storage_providers.go` | NEW | 02 |
| `backend/internal/database/migrate/037_libraries_provider_link.go` | NEW | 02 |
| `backend/internal/model/storage_provider.go` | NEW | 02 |
| `backend/internal/model/library.go` | Extend — provider_id + source_url | 02 |
| `backend/internal/repository/storage_provider.go` | NEW | 02 |
| `backend/internal/repository/library.go` | Extend | 02 |
| `backend/internal/cloudstorage/provider.go` | NEW — interfaces | 03 |
| `backend/internal/cloudstorage/registry.go` | NEW — driver registry | 03 |
| `backend/internal/cloudstorage/credentials.go` | NEW — encrypt/decrypt | 03 |
| `backend/internal/cloudstorage/errors.go` | NEW — typed errors | 03 |
| `backend/internal/cloudstorage/drivers/fshare/` | NEW — fshare driver | 03 |
| `backend/internal/scanner/cloud_walker.go` | NEW — Provider-based scanner | 04 |
| `backend/internal/scanner/pipeline.go` | Modify — dispatch by provider_id | 04 |
| `backend/internal/handler/stream.go` | Modify — Provider.GetDownloadURL | 05 |
| `backend/internal/handler/storage_provider.go` | NEW — CRUD + auth | 06 |
| `backend/internal/handler/library.go` | Modify — provider_id + source_url | 06 |
| `backend/internal/service/provider_refresh.go` | NEW — scheduled refresh | 08 |

## Frontend Changes Summary

| File | Change | Phase |
|---|---|---|
| `webapp/src/pages/admin/StorageProvidersPage.tsx` | NEW — providers list + add | 06 |
| `webapp/src/pages/admin/LibrariesPage.tsx` | Extend — source picker + URL input | 06 |
| `webapp/src/components/admin/AddProviderDialog.tsx` | NEW — dynamic form per driver | 06 |
| `webapp/src/components/admin/ProviderHealthBadge.tsx` | NEW | 06 |
| `webapp/src/components/admin/StorageProviderPicker.tsx` | NEW — dropdown | 06 |
| `webapp/src/api/cloudstorage.ts` | NEW — API client | 06 |

## Android Changes Summary

| File | Change | Phase |
|---|---|---|
| `android/.../data/repository/PlaybackRepository.kt` | Modify — handle cloud source URL | 07 |
| `android/.../utils/CloudUrlRefreshInterceptor.kt` | NEW — OkHttp 403/404 refresh | 07 |
| `android/.../presentation/ui/components/MediaCard.kt` | Modify — cloud source badge | 07 |

## Critical Gotchas (carry forward từ research)

1. ✅ **fshare API verified** — endpoints stable, see [Phase 01 README](../../backend/pkg/fshare/README.md)
2. **Session TTL ~30min** — scheduler refresh per provider every 25min
3. **Download URL short-lived / single-use** — NO caching, Android reactive-refresh on 403
4. **Error layer inversion** — HTTP always 200 on errors, parse body `code` first
5. **Password persistence required** — fshare has no refresh_token grant; GDrive/OneDrive use refresh tokens (future)
6. **Credentials at rest** — AES-256-GCM với `VELOX_CLOUD_SECRET` (32-byte hex)
7. **TMDb rate limit khi scan 10TB** — reuse existing throttle
8. **IsHDR/DVProfile** unknown for cloud files — ExoPlayer runtime detection (safe under migration 035 sentinel)
9. **User-Agent required** (fshare) — iOS UA in `pkg/fshare` default
10. **Account type check** — `Provider.GetAccountInfo().AccountType` validated before library creation
11. **URL parsing** — each driver handles its own format (fshare `/folder/XXX`, gdrive `/drive/folders/XXX`)

## Non-Goals (Out of Scope)

- ❌ Web browser playback cho cloud media (CORS + codec hell)
- ❌ Trickplay / thumbnail cho cloud files (phase MVP)
- ❌ Subtitle extract từ container
- ❌ iOS playback
- ❌ Google Drive / OneDrive drivers (architecture supports, implementation phase sau)
- ❌ OAuth web callback flow (MVP chỉ password-auth cho fshare)

## Rollout Strategy

1. **Feature flag:** `VELOX_CLOUD_SOURCES_ENABLED=false` default OFF
2. **Prerequisites:**
   - Generate `VELOX_CLOUD_SECRET` (`openssl rand -hex 32`)
   - `VELOX_FSHARE_APP_KEY` optional — `pkg/fshare.DefaultAppKey` works (verified 2026-04)
   - fshare VIP account active
3. **First test:** Small test library (~50 files) qua admin UI
4. **Scale test:** 10TB library → monitor scan time, TMDb rate limit, provider refresh cadence
5. **Prod rollout:** Flag ON sau scale test pass

## Success Criteria

- [x] Phase 01: `pkg/fshare` smoke test end-to-end (Login + VIP + ListFolder root + nested)
- [ ] Admin adds Fshare provider qua Settings → shows health ✅
- [ ] Admin adds Library với provider + folder URL → scan triggers
- [ ] Scan 10TB < 2h (TMDb throttle dominant)
- [ ] Android streams fshare media direct từ CDN (backend bandwidth ≈ 0)
- [ ] URL reactive-refresh khi 403/404 (OkHttp interceptor)
- [ ] Provider auto-refresh every 25min (scheduler)
- [ ] 0 regression trên local libraries
- [ ] Tất cả credentials AES-256-GCM encrypted at rest
- [ ] Architecture supports adding new drivers without touching scanner/stream handler
