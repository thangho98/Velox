# Phase 04: Catalog Shell — Home / Browse / Detail / Search
Status: ⬜ Pending
Dependencies: Phase 03

## Objective
Ship a usable browsing shell for phone/tablet before deep player work finishes, with real data from backend and a navigation structure compatible with later TV focus rules.

## Scope
- Home
- Browse/libraries
- Media detail
- Series detail
- Search
- Favorites entry point placeholder if full feature not ready yet

## Verified Contract Sources
- `packages/shared/hooks/media/useMediaQuery.ts`
- `packages/shared/hooks/media/useGenres.ts`
- `packages/shared/hooks/media/useSeries.ts`
- `packages/shared/hooks/media/useProgress.ts`
- `packages/shared/lib/image.ts`

## Implementation Steps

### 1. Port Models + APIs
- [ ] media list/detail models
- [ ] series/season/episode models
- [ ] search result model
- [ ] browse result model
- [ ] genre/credit models

### 2. Build Repositories
- [ ] `MediaRepository`
- [ ] `SeriesRepository`
- [ ] `SearchRepository`
- [ ] `LibraryRepository`
- [ ] `HomeRepository`

### 3. Build Shell Navigation
- [ ] home route
- [ ] browse route
- [ ] search route
- [ ] media detail route
- [ ] series detail route
- [ ] player route

### 4. Build UI Surfaces
- [ ] Home rows
- [ ] library/folder traversal
- [ ] movie/series cards
- [ ] media detail hero
- [ ] series detail with seasons/episodes
- [ ] search results page

### 5. Build Cross-Screen Reusables
- [ ] poster/backdrop image composables
- [ ] section header
- [ ] content row/grid
- [ ] empty/error/loading states
- [ ] progress badge or progress bar where relevant

## Tasks
1. [ ] Port core media/list/detail/search DTOs from current TS contracts.
2. [ ] Create repositories for:
   - home feed
   - libraries/folders
   - media detail
   - series detail
   - search
3. [ ] Create image URL helpers equivalent to current shared `mediaImage`.
4. [ ] Build app shell nav graph:
   - home
   - browse
   - search
   - detail
   - player route
5. [ ] Implement Home screen sections:
   - continue watching if available
   - recent/popular sections as supported
   - row-based browsing
6. [ ] Implement Browse screen:
   - libraries
   - folder traversal
   - media grids
7. [ ] Implement Media Detail screen:
   - poster/backdrop
   - metadata
   - play CTA
   - related action placeholders
8. [ ] Implement Series Detail screen:
   - seasons/episodes
   - episode play actions
9. [ ] Implement Search screen:
   - debounced query
   - mixed result rendering
   - empty/loading/error states
10. [ ] Add reusable shared UI:
   - section headers
   - media cards
   - loading placeholders
   - error states
11. [ ] Add navigation deep-link support for detail/player routes.
12. [ ] Add UI tests for basic browse/detail/search flows.

## Files / Modules to Create
- [ ] `core:model/.../media/*.kt`
- [ ] `core:model/.../series/*.kt`
- [ ] `core:network/.../MediaApi.kt`
- [ ] `core:network/.../SeriesApi.kt`
- [ ] `core:network/.../SearchApi.kt`
- [ ] `feature:home/...`
- [ ] `feature:browse/...`
- [ ] `feature:detail/...`
- [ ] `feature:search/...`
- [ ] shared UI in `core:designsystem`

## Test Criteria
- [ ] Home loads real backend data.
- [ ] Browse can traverse libraries/folders.
- [ ] Search returns movies and series.
- [ ] Media detail opens correct metadata.
- [ ] Series detail shows seasons and episodes.
- [ ] Detail screen can route into player with correct IDs.

## Done When
- [ ] Native app is already useful for browsing before advanced playback parity lands.

## Verification
- [ ] User can log in and browse content end-to-end.
- [ ] Search returns real backend results.
- [ ] Media detail can route into player screen with media ID context.
- [ ] No placeholder-only screens remain in Ship 1 critical flow.

## Exit Criteria
- Native app is already useful as a browsing client even before full playback polish.
