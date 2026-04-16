# Changelog

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
