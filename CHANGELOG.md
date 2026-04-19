# Changelog

## [2026-04-19]
### Added
- Backend: HLS V2 session filenames and keys now include `audioTrackId`, with parser round-trip tests for playlist and segment filenames.
- Android: Pull-to-refresh on Home, Media Detail, and Series Detail screens using `PullToRefreshBox` and explicit `isRefreshing` state.

### Changed
- Backend: Stream sessions now resolve cloud playback URLs lazily at FFmpeg start/restart instead of capturing a single resolved URL when the session is created.
- Backend: Exact audio-track HLS selection is now preserved across playback info, HLS master, and transcoder session keys.
- Backend: `GenerateHLSWithAudio` keeps using the multi-output path even for a single selected track so the chosen stream index is respected.
- Backend: `cloud-media-probe` scheduled task is throttled with a 3-second probe delay and stops the run immediately on rate-limit errors.
- Webapp: HLS audio picker now highlights by selected track ID and avoids unnecessary HLS re-initialization when the resolved URL has not changed.
- Android: HLS audio switching now rebuilds the playback URL with `at={trackId}` instead of relying only on preferred audio language.

### Fixed
- Backend: Fixed HLS V2 parser mismatch after `_at{audioTrackId}` was added to filenames. This removes the bogus `MaxHeight=1` parsing path that caused VAAPI `Hardware does not support scaling to size 2x1`.
- Backend: HLS playlist query rewriting now also covers `#EXT-X-MAP`, `#EXT-X-KEY`, and `#EXT-X-I-FRAME-STREAM-INF` URIs.

## v0.1.7 [2026-04-18]
### Added
- Backend: Cloud subtitle extraction — `ProbeAndUpdateCloudMetadata` now persists embedded subtitles (was audio tracks only). Cloud files get full ffprobe data (codec, resolution, audio tracks, subtitles) during scan.
- Backend: `POST /api/media/{id}/cloud-probe` admin endpoint for on-demand cloud file probing.
- Backend: `cloud-media-probe` scheduled task (6h interval) — backfills unprobed cloud files with metadata sequentially to avoid FShare rate limiting.
- Backend: Metadata fallback — when TMDb movie search fails, automatically tries TV series search. Matches BBC documentaries and similar content that only exist as TV series on TMDb.
- Backend: `ListUnprobedCloud` repository method for querying cloud files missing video codec data.
- Backend: `hdr_resolve.go` — cloud-aware HDR detection using persisted DB fields instead of re-probing with ffprobe.
- Backend: `ffprobe` context-aware variants (`IsHDRLikeCtx`, `NeedsHDRColorMetadataFallbackCtx`, `GetDVProfileCtx`) for proper timeout/cancellation on network-backed files.
- Webapp: "Extract cloud subtitles" button in movie detail ActionMenu (admin only, cloud media only).
- Webapp: AniList Connect link + clipboard paste button in Settings → Metadata.
- Android: "Extract cloud subtitles" button in movie detail ActionMenu (admin only, cloud media only).
- Android: AniList Connect link + clipboard paste button in Settings → Metadata provider cards.
- Dockerfile: Added `sqlite3` CLI for production DB inspection.

### Fixed
- Backend: File verifier no longer marks cloud files (`fshare://...`) as missing — `fileExists()` skips `os.Stat` for cloud paths.
- Backend: Pretranscode worker skips cloud files early instead of spamming ffprobe HDR probe failures.
- Backend: `SubtitleService.ServeContent` and `TranslateSubtitle` now resolve cloud paths to HTTP URLs before calling FFmpeg.
- Backend: Cloud probe audio/subtitle persistence wrapped in transaction to prevent intermediate empty state.
- Backend: `ParseCloudPath("fshare://")` returns `ok=false` for empty native IDs.
- Backend: Scheduled task pagination uses offset-0 strategy to avoid skipping files after successful probes.
- Backend: Centralized cloud path detection through `scanner.IsCloudPath()` across all services.
- Android: Fullscreen rotation fix — unwrap `ContextWrapper` chain to find Activity, use `SENSOR_LANDSCAPE` + force-portrait-then-release for reliable orientation toggle.

### Security
- Cloud resolver injected into `SubtitleService` — embedded subtitle extraction from cloud files no longer leaks raw `fshare://` paths to FFmpeg.

## [2026-04-17]
### Added
- Backend: Plan W — Cloud Storage Integration. New `internal/cloudstorage` package introduces a driver-based abstraction (`Provider`, `Driver`, `PasswordAuthDriver`, `OAuthDriver`) so future Google Drive / OneDrive support plugs in without touching scanner or stream handler.
- Backend: Fshare driver (`internal/cloudstorage/drivers/fshare`) wrapping `pkg/fshare` with URL parsing (`/folder/<linkcode>`, `/file/<linkcode>`), typed error mapping, and account-info → AccountInfo translation.
- Backend: Migration 036 `storage_providers` table (1 row = 1 cloud account, shared across libraries) with AES-256-GCM encrypted credentials.
- Backend: Migration 037 extends `libraries` with `storage_provider_id` + `source_url` columns (null = local filesystem, zero regression).
- Backend: `pkg/crypto` AES-256-GCM helper for at-rest encryption with `LoadOrGenerateKey` — self-generates `{DataDir}/cloud_secret.key` on first boot (0600 perms). `VELOX_CLOUD_SECRET` env overrides for users who prefer external secret management. Cloud feature is always on (no flag).
- Backend: `CloudWalker` scanner (provider-agnostic BFS) writing `media_files.path = {provider_type}://{native_id}`.
- Backend: Stream URL handler (`/api/stream/{id}/url`) dispatches by library — cloud libraries return direct CDN URL + `provider_type` + `direct_cdn:true`; local libraries keep the existing `api_key` flow.
- Backend: Storage provider admin API — `GET /api/admin/cloud/drivers`, `GET|POST /api/admin/cloud/providers`, `DELETE|POST-refresh|POST-validate-url` on `{id}`.
- Backend: `ProviderRefreshService` scheduled every 5 min (fshare TTL ~30 min). Polymorphic by auth flow: password-auth re-logs in, OAuth refreshes token.
- Webapp: `Settings → Storage` new section to add/refresh/delete cloud storage providers (fshare today, future drivers from driver registry).
- Webapp: `Settings → Libraries` source picker — local filesystem vs cloud storage with folder URL paste.
- Android: Updated `StreamUrlResponse` DTO with `provider_type` + `direct_cdn`. New `CloudUrlRefreshInterceptor` reactively refreshes expired fshare CDN URLs on 403/404 (max 2 retries per media).

### Security
- Stored fshare passwords encrypted at rest with AES-256-GCM. Key is auto-generated on first boot into `{DataDir}/cloud_secret.key` (0600 perms) unless `VELOX_CLOUD_SECRET` is explicitly set.

## [2026-04-16]
### Added
- Android App: Implement zero-restart dynamic Locale switching (i18n). Switching languages directly overrides Compose's `ConfigurationContext` while keeping `Activity` inheritance intact via an overridden `ContextWrapper`.
- Android App: AlertDialog notification advising users when an application restart is required for deep-seated locale changes.

### Changed
- Webapp & Android App: Complete all outstanding UI text bindings to the multi-language string definitions.
- Android: Decoupled Android `UserProfileViewModel` app language saving logic to persist globally inside `AuthManager` DataStore for instant updates.
- Android: Added baseline `detekt.yml` static analysis rules and `.editorconfig` formatting rules with ktlint overrides for relaxed Compose UI constraints.

### Fixed
- Android: Syntactical escaping errors across all Compose XML String Resource mappings.

## [2026-04-15]
### Added
- Backend: Responsive image system with typed TMDb path wrappers (`Poster`/`Backdrop`/`Still`/`Logo`/`Profile`) routed through `/api/images/tmdb/{type}/{path}` with per-type size buckets.
- Backend: `ImageResource` struct carrying `srcset`, aspect ratio, width/height and blurhash placeholder across media/series/playback/common payloads.
- Backend: Blurhash service with scan + force-scan backfill integration. New migrations `033_self_describing_image_paths` + `034_image_metadata` table.
- Webapp: `ResponsiveImage` component using `<picture srcset>` with `react-blurhash` LQIP placeholder for smoother image loading.
- Shared: `ImageResource` TypeScript type replacing legacy string image fields across all shared contracts.

### Changed
- Backend: `SearchSeries` N+1 query batched to single `GetBatch`; Still kind fixed in NextUp attach; ProfilePath wired for Person and User avatar.
- Webapp: Removed legacy string image fields — all consumers now read from `ImageResource`.
- Docs: Split `development-rules.md` into per-platform files (`development-rules-backend.md`, `development-rules-webapp.md`, `development-rules-mobile.md`) + shared conventions index. `CLAUDE.md` updated to reference the new layout.

### Fixed
- Backend: Indexed LLM subtitle response validation — enforce expected count, reject duplicate and negative indexes. Prevents silent cue corruption where extra items would overwrite cues of the next batch via `cues[b.start+j].Text = t`, duplicate indexes collapsed the list, and negative indexes were swallowed during reconstruction.

## [2026-04-14]
### Changed
- Android App: Massive Clean Architecture refactor. Componentized `SettingsScreen.kt` (3600+ lines) into 17 isolated section files.
- Android App: Decoupled `SettingsViewModel` from direct network calls by generating `SettingsRepository` with 92 abstract interface methods.
- Android App: Extracted logic from `SettingsViewModel` God Object into logical partitioned extension files (`SettingsViewModel_Admin.kt`, `SettingsViewModel_System.kt`).
- Android App: Full Unidirectional Data Flow decoupling for `VideoPlayer`. Converted parameters bounding it to `PlayerViewModel` into a dedicated `PlayerActions` intent interface, making the component stateless and reusable.

### Fixed
- Android App: Kotlin multi-pass verification to repair split-induced KSP parsing syntax errors across screens and repositories.

## [2026-04-13]
### Added
- Backend: Scheduled tasks API and `app_versions` database seed migration.
- Android App: Implement `AppUpdater` with `DownloadManager` for automatic in-app Android updates using `REQUEST_INSTALL_PACKAGES`.
- Webapp/Android App: Added scheduled tasks editor and status dashboard.
- Backend: Background Subtitle Auto-Translate worker that scans database and local library continuously.
- Webapp: Added Auto-Translate card inside Subtitles configuration page to toggle and select the target output languages.

### Changed
- Android App: AppVersionCard now fetches updates dynamically using `VeloxApi` and `AuthManager` tokens instead of a hardcoded HTTP connection.
- Backend: `AppVersion` Github tag parsing implementation updated to accurately calculate version codes (e.g. `v0.1.5` -> `105`).
- Webapp: Unified all HTML `<select>` elements to use a customized `Select.tsx` component matching the overall Design System dark theme.
- Backend: Enhanced AI subtitle translation system prompt to automatically deduce relative pronouns across continuous subtitle cues to maintain dialog consistency.
- Backend: Increased `MaxTokens` mapping limit up to 8192 to prevent 'no text content returned' bugs caused by truncated responses from Reasoning models (e.g. Anthropic/MiniMax `<thinking>` layers).

### Fixed
- Android App: Fixed duplicate `@Composable` annotation on `AppVersionCard` blocking Settings compilation.
- Backend/Transcoder: HDR & Dolby Vision color correctness — replaced the `zscale` tonemap chain with `tonemapx` (Jellyfin SIMD) across all encode paths (transcoder, pretranscode worker, trickplay generator). DV Profile 5 uses `IPT-PQ-C2`, not BT.2020, so the old chain produced a green/magenta shift.
- Backend/Transcoder: Force software decode for ALL HDR hwaccel paths (VideoToolbox / QSV / AMF were stripping color metadata). Scale filter now correctly ordered AFTER tonemap for HDR content.
- Backend/Transcoder: Added `NeedsServerTonemap` decision flag for DV content lacking standard color tags — browsers always tonemap on server; Android ExoPlayer direct-plays DV natively.

## [2026-04-11]
### Added
- Backend: AI subtitle translation package (`pkg/translate/ai.go`) - supports OpenAI, Gemini, and Anthropic compatible APIs.
- Backend: `aisubtest` CLI tool for testing AI subtitle translation.
- Backend: `ffprobe_hdr_test.go` for FFprobe HDR metadata testing.

### Changed
- Backend: Stream URL TTL now calculated from media file duration instead of fixed 2-hour value.
- Android App: Disable native text track rendering in ExoPlayer to prevent double subtitle rendering.
- Android App: Add release signing config with keystore for signed APK builds.

## [2026-04-10]
### Added
- Android App: Implement Skip Intro / Credits Admin Dashboard in Settings Screen.
- Android App: API integration for MarkerStats (`/api/admin/markers/stats`) and Backfill (`/api/admin/markers/backfill`).
- `.brain` folder for saving project knowledge context via AWF save-brain.

### Changed
- Android App: Refactored grid-based screens (`MoviesScreen`, `SeriesScreen`, `SearchScreen`, `BrowseScreen`, `FavoritesScreen`) to utilize responsive `GridCells.Adaptive` and aspect ratio-based flexible sizing for fluid scaling across different device screens.
- Android App: Standardized horizontal scroll components (`HomeScreen` carousels) to use fixed dp dimensions to preserve the off-screen "peek" affordance.
- Android App: Moved subjective "Auto-Skip Intro / Credits / Sponsor" toggles to the Preferences settings tab section.
- Webapp: Added boundary threshold logic for the Skip components to prevent UI flickering.
- Android App: Corrected Settings layout routing for the Admin Preferences group.

### Fixed
- Android App: Fixed gray background artifact on the application's launcher icon (`ic_launcher`) and animated splash screen.
- Webapp: Admin user management API hooks routing `/api/users` instead of `/api/admin/users`.
- Android App: Material 3 LinearProgressIndicator component was rendering an ugly stop indicator dot. Set `drawStopIndicator = {}` globally.
