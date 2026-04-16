# Changelog

## [2026-04-16]
### Added
- Android App: Implement zero-restart dynamic Locale switching (i18n). Switching languages directly overrides Compose's `ConfigurationContext` while keeping `Activity` inheritance intact via an overridden `ContextWrapper`.
- Android App: AlertDialog notification advising users when an application restart is required for deep-seated locale changes.

### Changed
- Webapp & Android App: Complete all outstanding UI text bindings to the multi-language string definitions.
- Android: Decoupled Android `UserProfileViewModel` app language saving logic to persist globally inside `AuthManager` DataStore for instant updates.

### Fixed
- Android: Syntactical escaping errors across all Compose XML String Resource mappings.

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
