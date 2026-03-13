# Plan J: Netflix-style Series Grouping + URL Routing
Created: 2026-03-13
Status: 🟡 In Progress

## Overview
Restructure Velox to group TV series like Netflix — one card per series instead of one card per episode. New URL routes: `/movies/:id` for movies, `/series/:seriesId` for series overview. Backend has `series` table + `SeriesRepo` with List/GetByID/Search ready but unexposed.

## Tech Stack
- Backend: Go stdlib net/http + existing SeriesRepo/SeasonRepo/EpisodeRepo
- Frontend: React 19 + TanStack Query + React Router v7
- Database: SQLite (existing `series`, `seasons`, `episodes` tables)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Backend Series API | ⬜ Pending | 0% |
| 02 | Frontend Types + Hooks | ⬜ Pending | 0% |
| 03 | Routes + SeriesDetailPage | ⬜ Pending | 0% |
| 04 | Update Listing Pages + Components | ⬜ Pending | 0% |

### Phase Summary
- **Phase 01:** Add `seriesRepo` to `SeriesHandler`, expose `GET /api/series`, `GET /api/series/{id}`, `GET /api/series/search?q=`
- **Phase 02:** Add `Series`/`SeriesWithSeasons` TS types, `seriesApi` functions, TanStack Query hooks (`useSeriesList`, `useSeriesDetail`, `useSeriesSearch`)
- **Phase 03:** Create `SeriesDetailPage` + `EpisodeCard` (fix episode.id→media_id bug), simplify `MediaDetailPage` to movie-only, update Router (`/movies/:id`, `/series/:seriesId`, xóa `/media/:id`)
- **Phase 04:** MediaCard type-aware routing (series cards: no progress/favorite, link `/series/:seriesId`), SeriesPage real series data (drop genre filter), HomePage series row, SearchPage dual fetch (movies search + series search, merge results, remove broken client-side type filter), WatchPage back button, LibraryListPage type-aware links

## Key Design Decisions

### Route Migration (no backwards compat)
| Location | Current | New | Phase |
|----------|---------|-----|-------|
| Router.tsx | `/media/:id` → MediaDetailPage | Xóa route | 3d |
| Router.tsx | — | `/movies/:id` → MediaDetailPage | 3d |
| Router.tsx | — | `/series/:seriesId` → SeriesDetailPage | 3d |
| MediaCard.tsx | `/media/${id}` | type-aware `/movies/` or `/series/` | 4a |
| WatchPage.tsx back btn | `/media/${mediaId}` | `navigate(-1)` | 4f |
| LibraryListPage.tsx | `/media/${item.id}` | type-aware routing | 4f |

No redirect from `/media/:id` — self-hosted app, no external deep links.

### Series Card Semantics
- Series cards use `series.id` (NOT `media_id`) — different ID domain
- **No progress bar** on series cards (`showProgress={false}`)
- **No favorite button** on series cards (`showFavorite={false}`)
- Click → opens `/series/:seriesId` overview (NOT auto-play)
- Future: series-level progress/favorites require new DB design (out of scope)

### Series API Response Contract
`GET /api/series` returns:
```json
[{
  "id": 1, "library_id": 1, "title": "Malcolm in the Middle",
  "sort_title": "malcolm in the middle", "tmdb_id": 2004,
  "overview": "...", "status": "Ended", "network": "FOX",
  "first_air_date": "2000-01-09",
  "poster_path": "/...", "backdrop_path": "/...",
  "logo_path": "/...", "thumb_path": "/..."
}]
```
**Not available on series:** genres, rating, media-level progress/favorites.
SeriesPage: year filter (from first_air_date) + title/newest sort only. No genre filter.

### Existing Endpoints (unchanged)
- `GET /api/series/{id}/seasons` — already works
- `GET /api/series/{id}/seasons/{seasonId}/episodes` — already works

## Acceptance Criteria
- [ ] `go build ./...` compiles
- [ ] `npx tsc --noEmit` compiles
- [ ] `GET /api/series` returns array of Series (not episodes)
- [ ] `GET /api/series/{id}` returns single series
- [ ] `GET /api/series/search?q=X` returns matching series
- [ ] SeriesPage shows one card per series (not per episode)
- [ ] HomePage series row shows grouped series
- [ ] Search returns series results
- [ ] Clicking series card → `/series/:seriesId` → shows seasons/episodes
- [ ] Clicking episode → `/watch/:mediaId` → plays correctly
- [ ] No `/media/` links remain in codebase (except API endpoints)
- [ ] WatchPage back button navigates correctly
- [ ] Series cards have NO progress bar or favorite button

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
