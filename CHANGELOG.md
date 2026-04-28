# Changelog

## v0.1.8 [2026-04-28]

5 commits since v0.1.7 (`dbe76cf`, `cee3ef0`, `de43089`, `18a86e4`, `4cacbbc`) — 184 files, ~22k LOC. Highlights: Live TV / IPTV across all platforms, NAS download + delete pipeline, Android TV launcher, first comprehensive backend test pass, design tokens v2.

### Added

#### Live TV / IPTV (`cee3ef0`) — primary feature

**Backend**
- Migrations 040–045 add the IPTV schema:
  - `040_live_tv`: `live_playlists` (M3U source URL + headers), `live_channels` (tvg-id, tvg-logo, group-title, hidden flag, manual EPG URL), `live_programs` (mock EPG store).
  - `041_add_epg_url`: per-channel optional XMLTV URL override.
  - `042_user_preferences_last_channel_id`: last-watched channel per user.
  - `043_live_channels_hidden`: admin can hide channels from listings.
  - `044_live_channel_headers`: per-channel custom HTTP headers (User-Agent, Origin, Referer) for upstream fetches.
  - `045_user_live_channel_data`: per-user recently-watched (`user_id`, `channel_id`, `play_count`, `last_watched_at`), indexed by `(user_id, last_watched_at DESC)`.
- `internal/service/m3uparser/parser.go` — `#EXTM3U` / `#EXTINF` parser with `tvg-id`, `tvg-logo`, `group-title` extraction.
- `internal/service/livetv_service.go` + `internal/repository/livetv_repo.go` — playlist CRUD, channel listing, EPG (mock until XMLTV parser lands), recently-watched listing.
- `internal/handler/livetv_handler.go` + `internal/handler/livetv_proxy.go` — HTTP routes and HMAC-signed upstream HLS proxy.
- New endpoints (all under `/api/livetv`):
  - `GET /groups`, `GET /channels`, `GET /channels/{id}`, `GET /channels/{id}/epg`
  - `GET /channels/recent?limit=N` — authenticated user's recently-watched
  - `POST /channels/{id}/toggle-hidden` (admin only)
  - `GET /livetv/stream/{id}` — accepts `?token=<jwt>` for native players that can't attach `Authorization` headers per HLS segment
- Stream entry bumps `user_live_channel_data` with a 5-minute in-memory cooldown per `(user, channel)` to keep HLS segment re-requests and rapid channel-surfing from inflating `play_count`.

**Webapp**
- Pages: `LiveTvPage` (channel grid + EPG sidebar) and `WatchLivePage` (full-screen HLS player + collapsible channel sidebar).
- Components: `components/livetv/CategoryPills`, `ChannelCard`, `EpgTimeline`, `LiveTVPlayer`.
- State: `hooks/stores/useLiveTV.ts` (Zustand) + React Query keys.
- Settings: new `LiveTvSection.tsx` for playlist add/refresh and channel visibility toggle.
- Navigation: Navbar / Sidebar / `MobileTabBar` add a Live TV entry with a `LivePulseDot` indicator. `Layout` swaps desktop sidebar for `MobileTabBar` on small screens.
- Design tokens (`index.css`): `ink-0/50/100/200/300/400`, `fog-400/500/600/700/800`, `crimson-500/600`, `shadow-card`, `shadow-pop`. Consumed by every new Live TV surface and the refreshed `Navbar` / `Logo` / `Layout`.
- "Recently Watched Channels" row on `HomePage`.

**Android (phone + TV)**
- New layers under `data/` and `domain/livetv` (DTOs, models, `LiveTvRepository{,Impl}`).
- ViewModels: `LiveTvViewModel`, `LiveTvPlayerViewModel`. Both inject `AuthManager` and expose `streamUrl(channelId)` that URL-encodes the JWT into `?token=` (Jellyfin-style — Media3 `DefaultHttpDataSource` can't attach Authorization headers per HLS segment).
- Screens: `LiveTvScreen`, `LiveTvPlayerScreen` (phone), `TvLiveTvScreen` (TV).
- Components: `LivePulseDot`, `LiveTvChannelCard`, `LiveTvEpgTimeline`, `LiveTvMiniPlayer` (now takes a fully-formed `streamUrl: String`, not a `baseUrl`).
- Bottom tab bar gains a Live TV tab with the pulsing dot.
- `HomeViewModel` injects `LiveTvRepository` and collects `getLiveTvRecentChannels(20).firstOrNull()`; both `HomeScreen` and `TvHomeScreen` render the recent-channels carousel and navigate to `livetv_player/{channelId}`.

#### NAS download pipeline + Android TV baseline (`dbe76cf`)

**Backend**
- `internal/service/download.go`: atomic Cloud-to-NAS downloader via `DownloadService` with task deduplication and error propagation (no more silent partials).
- `internal/handler/download.go`: `POST /api/media/{id}/download` and `POST /api/series/{id}/download` brought into the strict `respondJSON` contract.
- `OphimScanner` API now surfaces upstream errors instead of swallowing them.

**Android**
- `TvMainActivity` (LEANBACK_LAUNCHER) + `presentation/tv/` (TvAppNavigation, `TvNavigationDrawer` with auth-state injection, `TvHomeScreen`, `TvMediaDetailScreen`, `TvSeriesDetailScreen`).
- Universal APK keeps phone + TV in one binary (`uses-feature leanback/touchscreen required="false"`).
- Plans: `plans/.../ophim-provider/`, `plans/.../android-tv/`.

**Webapp fixes** (rolled into `dbe76cf`)
- Premature toast rendering on `MediaDetailPage` and `EpisodeCard` resolved.
- "Download to NAS" series-level button visibility now matches the per-episode rule on `SeriesDetailPage`.
- Duplicate API data loads removed from several Compose `LaunchedEffect` hooks on TV screens.

#### Delete-from-NAS + Episode hydration (`de43089`)

**Backend**
- `DELETE /api/media/{id}/download` (admin) — removes the local file and drops the matching `media_files` row. `DownloadService.DeleteDownloadedFile` is path-safe: only files under `outputDir` are touched, anything matching `cloud://` / `fshare://` is skipped.
- `DELETE /api/series/{id}/download` (admin) — iterates episodes and deletes whichever already have a local copy.
- `SeriesService.GetByID` now hydrates `Episode.MediaFiles` via `MediaFileRepo`, so callers can decide download vs. delete per episode.
- `downloadM3U8` switched to `ffmpegbin.FFmpeg()` for the resolved binary path (matches the rest of the codebase, no more bare `"ffmpeg"`).

**Webapp**
- Hooks: `useRemoveDownloadFromNas` and `useRemoveSeriesDownloadFromNas` (`packages/shared/hooks/media/useMetadataOps.ts`).
- `EpisodeCard` / `MediaDetailPage` / `SeriesDetailPage` now flip between "Download to NAS" and "Delete local download" based on `Episode.media_files` hydration. Confirm dialog + `Toast.error(message, detail)` (Toast gains an optional secondary line).
- `ContinueWatchingCard`: when `duration === 0`, show "watched X minutes" instead of "Y minutes remaining".
- `WatchPage`: clamp `remainingTime` to `>= 0` (defensive against drag overshoot).

**Android (phone + TV)**
- `VeloxApi`: `POST` and `DELETE` for `/api/{media,series}/{id}/download`.
- `MediaRepository{,Impl}`: `startDownload` / `deleteDownload` (+ series variants).
- `EpisodeDto.media_files`: lets the UI detect downloaded vs cloud-only.
- `MediaDetailScreen`, `SeriesDetailScreen`, `TvMediaDetailScreen`, `TvSeriesDetailScreen`: Download / Delete buttons with snackbar progress. ViewModels expose `downloadMedia` / `deleteDownload` (+ per-series variants).

#### Backend test coverage (`18a86e4`) — first broad pass
- 29 new `*_test.go` files spanning every layer:
  - **handler**: `fs`, `health`, `respond`, `settings`
  - **model**: `series`
  - **playback**: `playback`
  - **repository**: `media_file`, `pretranscode`, `series`
  - **service**: `activity`, `admin`, `browse`, `download`, `media`, `scheduler`, `settings`, `stream`
  - **storage**: `image`
  - **transcoder**: `encoding`, `hls`
  - **pkg**: `fanart`, `ffmpegbin`, `ffprobe`, `nfo`, `omdb`, `opensubs`, `subprovider`, `thetvdb`, `tmdb`
- Bundled refactor: `SettingsServiceInterface` extracted from `SettingsHandler` so handler tests can mock the settings service. Bundled here (not its own commit) because it was extracted *only* for testability.
- `go test ./...` green.

#### Documentation (`4cacbbc`)
- `AGENTS.md` + `GEMINI.md`: shared agent collaboration guidelines for the repo.
- `CLAUDE.md`: +66 lines of behavioral rules — *think before coding*, *simplicity first*, *surgical changes*, *goal-driven execution*.
- `docs/DESIGN_SYSTEM_V2.md`: ink/fog/crimson token spec (the source-of-truth for the new Live TV palette).
- `plans/260418-0847-shoko-integration/`: 5-phase plan for Shoko + AniList anime metadata integration.
- `.brain/{brain.json,session_log.txt}`: project-memory refresh; stale `handover.md` removed.

### Changed
- App version bumped to `0.1.8` (`versionCode=108`); seeded as **non-mandatory** in `app_versions` — IPTV and NAS-delete are additive APIs and older clients keep working.
- Backend `const version` corrected from a stale `velox v0.1.1` to `velox v0.1.8` (the constant had drifted behind release tags since v0.1.1).
- `LiveTvMiniPlayer` API tightened — takes a fully-formed `streamUrl: String` instead of a `baseUrl`, so callers consistently route through `viewModel.streamUrl(channelId)`.

### Fixed
- Android: Live TV ExoPlayer no longer 401s hitting `/livetv/stream/{id}` — the URL now ships `?token=<urlencoded JWT>` (auth middleware already accepted `token` as a query param before falling back to the JWT header).
- Android: Screen no longer auto-dims during Live TV playback — `PlayerView.keepScreenOn` bound to `!isPaused` in the `update` block, matching the main `VideoPlayer` pattern.
- Android: Lingering `MediaSession` notification when leaving Live TV — `PlaybackManager.releasePlayer()` now explicitly calls `stopService(VeloxPlaybackService)`. Previously the service kept running with a stale player reference.
- Webapp: Premature toast rendering on `MediaDetailPage` / `EpisodeCard` (rolled in via `dbe76cf`).
- Webapp: Duplicate "Download to NAS" series-level button now hidden when no episode is cloud-backed.
- Webapp: Duplicate API loads in several Compose `LaunchedEffect` hooks on TV screens removed (`dbe76cf`).
- Webapp: Tablet portrait (e.g. 1600×2560) no longer crams Vietnamese nav labels into the top bar. Bottom-tab breakpoint moved from `md` to `lg`, so anything below 1024 CSS px now uses `MobileTabBar` (matching the mobile UX) and the top bar only shows logo + bell + search + avatar. Layout's bottom-padding stays at `pb-28` at md (was `pb-16`) to leave room for the bottom nav. Defensive `whitespace-nowrap` added to top-bar links for narrow lg viewports.

### Notes
- Confirmed single universal APK continues to serve phone + TV (`MainActivity` LAUNCHER + `TvMainActivity` LEANBACK_LAUNCHER, `uses-feature leanback/touchscreen required="false"`). No separate TV build.
- **XMLTV parser** for exact EPG schedule data still pending — backend currently returns mock EPG until a cron task is wired up to populate `live_programs`.
- **Optional HLS reverse-proxy** for IPTV web CORS deferred — native apps don't hit CORS, only the webapp would benefit; not worth the NAS CPU cost vs the lightweight-core vision.
- Repo-root junk (`filter_m3u*.py`, `*.m3u`, `screenshot/`, `.scratch/`, `Velox Dashboard.html`) intentionally left unstaged — pending `.gitignore` entries in a follow-up Chore commit.

## v0.1.7 [2026-04-19]
### Added
- Backend: Full cloud storage foundation with provider registry, encrypted credentials, FShare driver, cloud library linking, cloud scanner, provider refresh service, stream URL resolver, and storage-provider admin APIs.
- Backend: Cloud subtitle extraction — `ProbeAndUpdateCloudMetadata` now persists embedded subtitles (was audio tracks only). Cloud files get full ffprobe data (codec, resolution, audio tracks, subtitles) during scan.
- Backend: `POST /api/media/{id}/cloud-probe` admin endpoint for on-demand cloud file probing.
- Backend: `cloud-media-probe` scheduled task (6h interval) — backfills unprobed cloud files with metadata sequentially to avoid FShare rate limiting.
- Backend: AniList metadata integration with OAuth helper flow and backend wiring for anime matching.
- Backend: HLS V2 session filenames and session keys now include `audioTrackId`, with parser round-trip tests for playlist and segment filenames.
- Backend: Metadata fallback — when TMDb movie search fails, automatically tries TV series search. Matches BBC documentaries and similar content that only exist as TV series on TMDb.
- Backend: `ListUnprobedCloud` repository method for querying cloud files missing video codec data.
- Backend: `hdr_resolve.go` — cloud-aware HDR detection using persisted DB fields instead of re-probing with ffprobe.
- Backend: `ffprobe` context-aware variants (`IsHDRLikeCtx`, `NeedsHDRColorMetadataFallbackCtx`, `GetDVProfileCtx`) for proper timeout/cancellation on network-backed files.
- Backend: Migration 035 stores HDR / Dolby Vision metadata for media files, and scanner extraction now persists HDR / DV details from ffprobe during scan.
- Backend: VAAPI + NVENC Vulkan hardware HDR tonemap path added for supported transcode flows.
- Webapp: New Storage settings UI for adding, refreshing, validating, and linking cloud providers to libraries.
- Webapp: "Extract cloud subtitles" button in movie detail ActionMenu (admin only, cloud media only).
- Webapp: AniList Connect link + clipboard paste button in Settings → Metadata.
- Webapp: HLS audio picker now tracks exact track ID and avoids unnecessary HLS re-initialization when the resolved stream URL has not changed.
- Android: Updated cloud playback URL handling with `CloudUrlRefreshInterceptor` for expired CDN links.
- Android: "Extract cloud subtitles" button in movie detail ActionMenu (admin only, cloud media only).
- Android: AniList Connect link + clipboard paste button in Settings → Metadata provider cards.
- Android: Pull-to-refresh on Home, Media Detail, and Series Detail screens using `PullToRefreshBox` and explicit `isRefreshing` state.
- Dockerfile: Added `sqlite3` CLI for production DB inspection.

### Changed
- Backend: Stream sessions now resolve cloud playback URLs lazily at FFmpeg start/restart instead of capturing a single resolved URL when the session is created.
- Backend: Exact audio-track HLS selection is now preserved across playback info, HLS master, and transcoder session keys.
- Backend: `GenerateHLSWithAudio` keeps using the multi-output path even for a single selected track so the chosen stream index is respected.
- Backend: `cloud-media-probe` scheduled task is throttled with a 3-second probe delay and stops the run immediately on rate-limit errors.
- Android: HLS audio switching now rebuilds the playback URL with `at={trackId}` instead of relying only on preferred audio language.
- Android: App version bumped to `0.1.7` (`versionCode=107`) to match the seeded mandatory update record.

### Fixed
- Backend: File verifier no longer marks cloud files (`fshare://...`) as missing — `fileExists()` skips `os.Stat` for cloud paths.
- Backend: Pretranscode worker skips cloud files early instead of spamming ffprobe HDR probe failures.
- Backend: `SubtitleService.ServeContent` and `TranslateSubtitle` now resolve cloud paths to HTTP URLs before calling FFmpeg.
- Backend: Cloud probe audio/subtitle persistence wrapped in transaction to prevent intermediate empty state.
- Backend: `ParseCloudPath("fshare://")` returns `ok=false` for empty native IDs.
- Backend: Scheduled task pagination uses offset-0 strategy to avoid skipping files after successful probes.
- Backend: Centralized cloud path detection through `scanner.IsCloudPath()` across all services.
- Backend: Fixed HLS V2 parser mismatch after `_at{audioTrackId}` was added to filenames. This removes the bogus `MaxHeight=1` parsing path that caused VAAPI `Hardware does not support scaling to size 2x1`.
- Backend: HLS playlist query rewriting now also covers `#EXT-X-MAP`, `#EXT-X-KEY`, and `#EXT-X-I-FRAME-STREAM-INF` URIs.
- Android: Fullscreen rotation fix — unwrap `ContextWrapper` chain to find Activity, use `SENSOR_LANDSCAPE` + force-portrait-then-release for reliable orientation toggle.
- Android: Restored missing Settings screens and `SystemAdminViewModel` integration needed for the expanded admin/settings surface.

### Security
- Backend: Storage provider credentials are encrypted at rest with AES-256-GCM and support external secret override via `VELOX_CLOUD_SECRET`.
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
