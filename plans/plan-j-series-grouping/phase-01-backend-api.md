# Phase 01: Backend Series API
Status: ⬜ Pending
Dependencies: None

## Objective
Expose existing `SeriesRepo` methods via HTTP endpoints so frontend can fetch grouped series data.

## Implementation Steps
1. [ ] Add `seriesRepo` field to `SeriesHandler` struct
2. [ ] Update `NewSeriesHandler(seriesRepo, seasonRepo, episodeRepo)` signature
3. [ ] Add `ListSeries` handler — `GET /api/series` with `?library_id=&limit=&offset=`
4. [ ] Add `GetSeries` handler — `GET /api/series/{id}` → single series
5. [ ] Add `SearchSeries` handler — `GET /api/series/search?q=&limit=`
6. [ ] Wire `seriesRepo` in `main.go` line 269
7. [ ] Register routes (search BEFORE `{id}` for correct Go routing)

## Files to Modify
- `backend/internal/handler/series.go` — Add seriesRepo + 3 new handlers
- `backend/cmd/server/main.go` — Wire seriesRepo, register routes

## Existing Code to Reuse
- `SeriesRepo.List(ctx, libraryID, limit, offset)` — already implemented
- `SeriesRepo.GetByID(ctx, id)` — already implemented
- `SeriesRepo.Search(ctx, query, limit)` — already implemented
- `parseID()`, `parseIntQuery()`, `respondJSON()`, `respondError()` — handler helpers

## Test Criteria
- [ ] `go build ./...` compiles
- [ ] `GET /api/series` returns array of Series (not episodes)
- [ ] `GET /api/series/{id}` returns single series with poster/backdrop
- [ ] `GET /api/series/search?q=malcolm` returns matching series

---
Next Phase: phase-02-frontend-types.md
