# Phase 04: Update Listing Pages + Components
Status: ⬜ Pending
Dependencies: Phase 03

## Objective
All pages show grouped series (one card per series, not per episode). Links use correct `/movies/` and `/series/` routes.

## Implementation Steps

### 4a. MediaCard — type-aware routing + series card semantics
1. [ ] Add `seriesId?: number` prop to MediaCardProps
2. [ ] Link: `type === 'series' ? /series/${seriesId} : /movies/${id}`
3. [ ] When `type === 'series'`: force `showProgress={false}`, `showFavorite={false}`
4. [ ] Skip `useProgress(id)` call when type is series (series.id is NOT a media_id)

### 4b. MediaRow — pass seriesId
3. [ ] Pass `seriesId={item.series_id}` to MediaCard for non-UserData items

### 4c. SeriesPage — use real series data
4. [ ] Replace `useMediaList({ type: 'episode' })` → `useSeriesList({ limit: 100 })`
5. [ ] Pass `seriesId={s.id}` to MediaCard
6. [ ] Year from `s.first_air_date`, remove genre filter (not on series table)

### 4d. HomePage — series row uses real series
7. [ ] Replace episode fetch → `useSeriesList({ limit: 20 })`
8. [ ] Map `Series[]` to MediaRow-compatible shape with `series_id`

### 4e. SearchPage — merge movies + series results
9. [ ] Fetch movies: `useMediaSearch(query)` (existing, returns media items)
10. [ ] Fetch series: `useSeriesSearch(query)` (new hook from Phase 02)
11. [ ] Merge results: movies (media_type=movie) + series (from series search) into one list
12. [ ] Type filter: "Movies" → show only media, "Series" → show only series results
13. [ ] Series cards: pass `type='series'`, `seriesId={s.id}`, link to `/series/:seriesId`
14. [ ] Remove genre filter for series results (series table has no genres)
15. [ ] Remove client-side full-list fetch (`useMediaList({ limit: 500 })`) — use server-side search

### 4f. Other link updates
11. [ ] WatchPage back button — use `navigate(-1)` instead of `/media/${mediaId}`
12. [ ] LibraryListPage — type-aware routing for media items

## Files to Modify
- `webapp/src/components/MediaCard.tsx` — Type-aware routing
- `webapp/src/components/MediaRow.tsx` — Pass seriesId through
- `webapp/src/pages/SeriesPage.tsx` — Rewrite with real series data
- `webapp/src/pages/HomePage.tsx` — Series row uses real series
- `webapp/src/pages/SearchPage.tsx` — Series in search
- `webapp/src/pages/WatchPage.tsx` — Fix back button nav
- `webapp/src/pages/LibraryListPage.tsx` — Type-aware links
- `webapp/src/types/api.ts` — Add series_id to MediaListItem

## Test Criteria
- [ ] `npx tsc --noEmit` compiles
- [ ] SeriesPage shows one card per series (not per episode)
- [ ] HomePage series row shows grouped series
- [ ] Search returns series results
- [ ] All links route to correct `/movies/` or `/series/` URLs
- [ ] WatchPage back button works correctly
