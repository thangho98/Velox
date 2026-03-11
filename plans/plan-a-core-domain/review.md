# Plan A Code Review — Phase 02

Reviewed: 2026-03-11
Status: ✅ All Fixed

---

## Bugs (Fix Before Phase 03)

- [x] **#8 — `ListWithGenres` crashes on NULL GROUP_CONCAT**
  - File: `backend/internal/repository/media.go:206-208`
  - `GROUP_CONCAT` returns NULL when no genres match. Scanning NULL into `string` panics.
  - Fix: Use `sql.NullString` for `genreNames`, return `[]` when empty.

- [x] **#9 — LibraryRepo not selecting `type` column**
  - File: `backend/internal/repository/library.go:17-43`
  - After migration 002 adds `type` to libraries, the repo only queries `id, name, path, created_at`.
  - `Library.Type` will always be empty string.
  - Fix: Add `type` to SELECT in `List()`, `GetByID()`, `Create()`.

- [x] **#2 — `ALTER TABLE DROP COLUMN` in down002**
  - File: `backend/internal/database/migrate/registry.go:195`
  - SQLite < 3.35.0 doesn't support DROP COLUMN. Violates CLAUDE.md rule #8.
  - Fix: Recreate `libraries` table without `type` column instead.

---

## Design Issues

- [x] **#1 — Orphan `_media_old` table**
  - File: `backend/internal/database/migrate/registry.go:105`
  - Migration 002 creates `_media_old` backup but never drops it.
  - Fix: Add `DROP TABLE IF EXISTS _media_old;` at end of `up002`.

- [x] **#3 — Subtitle uniqueness uses INDEX, not UNIQUE INDEX**
  - File: `backend/internal/database/migrate/registry.go:375-376`
  - Spec requires `UNIQUE(media_file_id, stream_index) WHERE is_embedded = 1`.
  - Current code uses `CREATE INDEX` (non-unique) — duplicates not prevented.
  - Fix: Change to `CREATE UNIQUE INDEX`.

- [x] **#5 — Inconsistent timestamp types across models**
  - `Library.CreatedAt` is `time.Time` (`backend/internal/model/model.go:11`)
  - All other models use `string` for timestamps.
  - Fix: Standardize on `string` for all models (simpler with SQLite DATETIME text).

---

## Code Quality

- [x] **#7 — `ListWithGenres` returns `map[string]any`**
  - File: `backend/internal/repository/media.go:178-221`
  - Should return typed struct. Also `strings.Split("", ",")` returns `[""]` not `[]`.
  - Fix: Create `MediaWithGenres` struct, handle empty genre list.

- [x] **#12 — Massive scan boilerplate in MediaFileRepo**
  - File: `backend/internal/repository/media.go`
  - 6 methods each scan same 15 columns with `isPrimary` + `lastVerified` conversion.
  - Fix: Extract `scanMediaFile(row)` helper to eliminate ~100 lines duplication.

- [x] **#13 — No error wrapping in service layer**
  - Files: `backend/internal/service/media.go`, `backend/internal/service/stream.go`
  - Raw `sql.ErrNoRows` leaks to handlers. Handler must know DB internals.
  - Fix: Define `service.ErrNotFound` sentinel, wrap repo errors with context.

- [x] **#14 — `respondJSON` ignores encode error**
  - File: `backend/internal/handler/respond.go:12`
  - `json.NewEncoder(w).Encode(data)` error silently dropped.
  - Fix: Log encode errors (can't change status after WriteHeader).

- [x] **#15 — Response format doesn't match API contract**
  - File: `backend/internal/handler/respond.go:9-13`
  - Convention says `{"data": ...}` for success, but `respondJSON` sends raw data.
  - Fix: Wrap success responses in `{"data": ...}`.

- [x] **#16 — Scanner doesn't accept context**
  - File: `backend/internal/scanner/scanner.go:36`
  - `ScanLibrary` has no context param. Long scans can't be cancelled.
  - Uses `context.Background()` internally (line 56, 73).
  - Fix: Add `ctx context.Context` parameter, pass through to repo calls.

---

## Minor / Phase 03+ TODO

- [x] **#10 — LibraryRepo methods lack `context.Context`**
  - File: `backend/internal/repository/library.go`
  - All new repos use context pattern, LibraryRepo (pre-existing) doesn't.
  - Fix: Add `ctx context.Context` as first param to all LibraryRepo methods.

- [x] **#11 — Genre `LinkToMedia` conflict strategy mismatch**
  - File: `backend/internal/repository/genre.go:94-98`
  - Uses `ON CONFLICT DO NOTHING` but table constraint has `ON CONFLICT REPLACE`.
  - Not a bug (REPLACE takes precedence) but intent is unclear.
  - Fix: Remove `ON CONFLICT DO NOTHING` from INSERT or align with table constraint.

- [x] **#19 — Migration tests don't verify 005-006 tables**
  - File: `backend/internal/database/migrate/migrate_test.go:279-283`
  - `TestRealMigrations_FreshDB` checks 10 tables but skips `scan_jobs`, `subtitles`, `audio_tracks`.
  - Fix: Add these 3 tables to `coreTables` list.

- [x] **#20 — Unused repos commented out in main.go**
  - File: `backend/cmd/server/main.go:106-110`
  - Commented-out `seriesRepo`, `seasonRepo`, `episodeRepo`, `genreRepo`, `personRepo`.
  - Fix: Wire up when Phase 03-05 handlers are added. (No change needed - Phase 03+)

---

## Stats

| Severity | Count |
|----------|-------|
| Bug | 3 |
| Design Issue | 3 |
| Code Quality | 6 |
| Minor/TODO | 4 |
| **Total** | **16** |

## Summary

All 16 issues have been fixed and verified:
- All tests pass (`make test`)
- Build succeeds (`make build`)
- No lint errors (`go vet`)
