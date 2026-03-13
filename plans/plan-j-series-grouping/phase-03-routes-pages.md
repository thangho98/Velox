# Phase 03: Routes + SeriesDetailPage
Status: ⬜ Pending
Dependencies: Phase 02

## Objective
Create Netflix-style series overview page, extract EpisodeCard, update routes, simplify MediaDetailPage to movie-only.

## Implementation Steps

### 3a. Extract EpisodeCard
1. [ ] Create `webapp/src/components/EpisodeCard.tsx` — move from MediaDetailPage
2. [ ] Bug fix: `episode.id` → `episode.media_id` in play links (episode.id is PK, not media_id)
3. [ ] Keep `formatDuration` inline in EpisodeCard

### 3b. Create SeriesDetailPage
4. [ ] Create `webapp/src/pages/SeriesDetailPage.tsx`
5. [ ] Route param: `useParams<{ seriesId: string }>()`
6. [ ] Fetch series via new `useSeriesDetail(seriesId)` hook (series + seasons)
7. [ ] Fetch episodes via existing `useEpisodes(seriesId, selectedSeasonId)`
8. [ ] Layout: backdrop + poster + title + overview + status/network badges
9. [ ] Season selector tabs (reuse pattern from MediaDetailPage)
10. [ ] Episode list using EpisodeCard
11. [ ] Play button → `/watch/${firstEpisode.media_id}`

### 3c. Simplify MediaDetailPage (movie-only)
12. [ ] Remove `isSeries` branch entirely (seasons/episodes section)
13. [ ] Remove `seriesMedia`, `useSeasons`, `useEpisodes` imports
14. [ ] Remove `displayTitle` logic — just use `media.media.title`
15. [ ] Remove `seriesId`, `selectedSeasonId` state

### 3d. Update Router
16. [ ] Add `/movies/:id` → MediaDetailPage
17. [ ] Add `/series/:seriesId` → SeriesDetailPage
18. [ ] Remove `/media/:id` route entirely
19. [ ] Import SeriesDetailPage

## Files to Create
- `webapp/src/components/EpisodeCard.tsx` — Extracted + bug fix
- `webapp/src/pages/SeriesDetailPage.tsx` — Netflix-style series page

## Files to Modify
- `webapp/src/pages/MediaDetailPage.tsx` — Simplify to movie-only
- `webapp/src/providers/Router.tsx` — New routes

## Test Criteria
- [ ] `npx tsc --noEmit` compiles
- [ ] `/movies/:id` renders movie detail
- [ ] `/series/:seriesId` renders series with seasons/episodes
- [ ] Episode play links go to `/watch/:mediaId` (correct ID)

---
Next Phase: phase-04-listing-pages.md
