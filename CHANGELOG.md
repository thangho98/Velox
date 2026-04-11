# Changelog

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
