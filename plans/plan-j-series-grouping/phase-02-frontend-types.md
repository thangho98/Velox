# Phase 02: Frontend Types + Hooks
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Add TypeScript types for Series entity and TanStack Query hooks to consume new backend endpoints.

## Implementation Steps
1. [ ] Add `Series` interface to api.ts (mirrors Go model.Series)
2. [ ] Add `SeriesWithSeasons` interface
3. [ ] Add `SeriesListParams` type
4. [ ] Add `series_id?: number` to `MediaListItem` for routing support
5. [ ] Add series API functions: `seriesApi.list()`, `seriesApi.get()`, `seriesApi.search()`
6. [ ] Add query keys: `seriesKeys.list()`, `seriesKeys.detail()`
7. [ ] Add hooks: `useSeriesList(params)`, `useSeriesDetail(id)`, `useSeriesSearch(q)`

## Files to Modify
- `webapp/src/types/api.ts` — Add Series types
- `webapp/src/hooks/stores/useMedia.ts` — Add API + hooks

## Test Criteria
- [ ] `npx tsc --noEmit` compiles
- [ ] Types match backend JSON response shape

---
Next Phase: phase-03-routes-pages.md
