# Prompt: Fix Plan A Phase 02 Review Issues

## Context

You are working on **Velox**, a self-hosted home media server (Go backend + React frontend).

**Read these files FIRST before writing any code:**
- `CLAUDE.md` — Project rules, code style, architecture patterns
- `plans/plan-a-core-domain/review.md` — Full checklist of 16 issues to fix
- `docs/database-design.md` — Database schema reference

**Tech stack:** Go 1.26 + stdlib `net/http` + SQLite (WAL mode) + raw `database/sql`

**What's already done:**
- Migrations 001-006 in `backend/internal/database/migrate/registry.go`
- Models in `backend/internal/model/`
- Repositories in `backend/internal/repository/`
- Services in `backend/internal/service/`
- Handlers in `backend/internal/handler/`
- Scanner in `backend/internal/scanner/scanner.go`
- All tests currently pass (`make test`)

---

## Your Task

Fix all 16 issues from the code review. Work through them in order: Bugs → Design Issues → Code Quality → Minor. Run `make test` after each group to verify nothing breaks.

---

## Group 1: Bugs (3 items) — Fix these FIRST

### Bug #8 — `ListWithGenres` crashes on NULL GROUP_CONCAT

**File:** `backend/internal/repository/media.go` — `ListWithGenres` method

**Problem:** `GROUP_CONCAT(g.name, ',')` returns NULL when a media item has no genres. Scanning NULL into a Go `string` variable panics at runtime.

**Fix:**
- Use `sql.NullString` for `genreNames`
- When `genreNames` is not valid (NULL) or empty, return an empty `[]string{}` instead of `strings.Split("", ",")`  which produces `[""]`

### Bug #9 — LibraryRepo not selecting `type` column

**File:** `backend/internal/repository/library.go`

**Problem:** Migration 002 added `type TEXT DEFAULT 'mixed'` to `libraries`, and `model.Library` has a `Type` field, but the repo queries still only select `id, name, path, created_at`. The `Type` field is always empty.

**Fix:**
- Add `type` to SELECT in `List()`, `GetByID()`, and the query inside `Create()`
- Add `&l.Type` to all `Scan()` calls
- The `Create()` method should also accept and INSERT the `type` parameter

### Bug #2 — `ALTER TABLE DROP COLUMN` in down002

**File:** `backend/internal/database/migrate/registry.go` — `down002` function

**Problem:** Line `ALTER TABLE libraries DROP COLUMN type;` — SQLite before 3.35.0 doesn't support DROP COLUMN. CLAUDE.md rule #8 says: "no ALTER TABLE DROP COLUMN (recreate table instead)".

**Fix:** Replace `ALTER TABLE libraries DROP COLUMN type;` with table recreation:
```sql
CREATE TABLE libraries_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    path       TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO libraries_new SELECT id, name, path, created_at FROM libraries;
DROP TABLE libraries;
ALTER TABLE libraries_new RENAME TO libraries;
```

---

## Group 2: Design Issues (3 items)

### Design #1 — Orphan `_media_old` table

**File:** `backend/internal/database/migrate/registry.go` — `up002` function

**Problem:** Line `CREATE TABLE _media_old AS SELECT * FROM media;` creates a backup but never drops it.

**Fix:** Add `DROP TABLE IF EXISTS _media_old;` at the END of the `up002` SQL block (after creating indexes).

### Design #3 — Subtitle uniqueness uses non-unique INDEX

**File:** `backend/internal/database/migrate/registry.go` — `up006` function

**Problem:** Two indexes should be UNIQUE to prevent duplicate subtitle entries:
- `idx_sub_embedded` — should prevent duplicate embedded subtitles with same stream_index
- `idx_sub_external` — should prevent duplicate external subtitles with same file_path

**Fix:** Change these two lines:
```sql
-- FROM:
CREATE INDEX idx_sub_embedded ON subtitles(media_file_id, stream_index) WHERE is_embedded = 1;
CREATE INDEX idx_sub_external ON subtitles(media_file_id, file_path) WHERE is_embedded = 0;

-- TO:
CREATE UNIQUE INDEX idx_sub_embedded ON subtitles(media_file_id, stream_index) WHERE is_embedded = 1;
CREATE UNIQUE INDEX idx_sub_external ON subtitles(media_file_id, file_path) WHERE is_embedded = 0;
```

### Design #5 — Inconsistent timestamp types across models

**File:** `backend/internal/model/model.go` — `Library` struct

**Problem:** `Library.CreatedAt` uses `time.Time` but every other model uses `string` for timestamps. SQLite stores DATETIME as text, so `string` is simpler and consistent.

**Fix:**
- Change `Library.CreatedAt` from `time.Time` to `string`
- Remove the `"time"` import from `model.go` if no longer needed
- Verify `library.go` repo still scans correctly (it should — SQLite returns text)

---

## Group 3: Code Quality (6 items)

### Quality #7 — `ListWithGenres` returns `map[string]any`

**File:** `backend/internal/repository/media.go`

**Fix:**
- Create a `model.MediaListItem` struct in `backend/internal/model/media.go`:
```go
type MediaListItem struct {
    ID         int64    `json:"id"`
    Title      string   `json:"title"`
    SortTitle  string   `json:"sort_title"`
    PosterPath string   `json:"poster_path"`
    MediaType  string   `json:"media_type"`
    Genres     []string `json:"genres"`
}
```
- Update `ListWithGenres` to return `[]model.MediaListItem` instead of `[]map[string]any`
- Handle empty genres properly (empty slice, not `[""]`)

### Quality #12 — Scan boilerplate in MediaFileRepo

**File:** `backend/internal/repository/media.go`

**Fix:** Extract a private helper method:
```go
func scanMediaFile(scanner interface{ Scan(...any) error }) (*model.MediaFile, error) {
    var mf model.MediaFile
    var isPrimary int
    var lastVerified sql.NullString

    err := scanner.Scan(&mf.ID, &mf.MediaID, &mf.FilePath, &mf.FileSize, &mf.Duration,
        &mf.Width, &mf.Height, &mf.VideoCodec, &mf.AudioCodec, &mf.Container, &mf.Bitrate,
        &mf.Fingerprint, &isPrimary, &mf.AddedAt, &lastVerified)
    if err != nil {
        return nil, err
    }
    mf.IsPrimary = isPrimary == 1
    if lastVerified.Valid {
        mf.LastVerifiedAt = &lastVerified.String
    }
    return &mf, nil
}
```
- Use this helper in: `GetByID`, `GetPrimaryByMediaID`, `FindByFingerprint`, `FindByPath`, `ListByMediaID`
- For `ListByMediaID`, use it inside the `rows.Next()` loop

### Quality #13 — No error wrapping in service layer

**Files:** `backend/internal/service/media.go`, `backend/internal/service/stream.go`

**Fix:**
- Create `backend/internal/service/errors.go`:
```go
package service

import "errors"

var ErrNotFound = errors.New("not found")
```
- In service methods, wrap `sql.ErrNoRows`:
```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}
```
- Update handlers to check `errors.Is(err, service.ErrNotFound)` instead of `err == sql.ErrNoRows`
- Remove `"database/sql"` import from handler files

### Quality #14 — `respondJSON` ignores encode error

**File:** `backend/internal/handler/respond.go`

**Fix:**
```go
func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("json encode error: %v", err)
    }
}
```
Add `"log"` to imports.

### Quality #15 — Response format doesn't match API contract

**File:** `backend/internal/handler/respond.go`

**Problem:** Convention is `{"data": ...}` for success, `{"error": "message"}` for errors. Currently sends raw data.

**Fix:**
```go
func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(map[string]any{"data": data}); err != nil {
        log.Printf("json encode error: %v", err)
    }
}
```
Error responses already use `{"error": msg}` format — those are fine.

### Quality #16 — Scanner doesn't accept context

**File:** `backend/internal/scanner/scanner.go`

**Fix:**
- Change signature: `func (s *Scanner) ScanLibrary(ctx context.Context, libraryID int64) error`
- Pass `ctx` to all repo calls instead of `context.Background()`
- Update `LibraryRepo.GetByID` call too (see Minor #10 — if you fix #10 first, pass ctx; if not, keep as-is for now)
- Update all callers of `ScanLibrary` to pass context

---

## Group 4: Minor / TODO (4 items)

### Minor #10 — LibraryRepo methods lack `context.Context`

**File:** `backend/internal/repository/library.go`

**Fix:**
- Add `ctx context.Context` as first param to `List`, `GetByID`, `Create`, `Delete`
- Use `r.db.QueryContext(ctx, ...)` / `r.db.QueryRowContext(ctx, ...)` / `r.db.ExecContext(ctx, ...)`
- Update all callers (scanner, service/library.go, handler/library.go) to pass context

### Minor #11 — Genre `LinkToMedia` conflict strategy mismatch

**File:** `backend/internal/repository/genre.go:94`

**Fix:** Remove `ON CONFLICT DO NOTHING` from the INSERT statements in `LinkToMedia` and `LinkToSeries`. The table constraint already handles conflicts with `ON CONFLICT REPLACE`.
```go
// FROM:
"INSERT INTO media_genres (media_id, genre_id) VALUES (?, ?) ON CONFLICT DO NOTHING"
// TO:
"INSERT INTO media_genres (media_id, genre_id) VALUES (?, ?)"
```

### Minor #19 — Migration tests don't verify 005-006 tables

**File:** `backend/internal/database/migrate/migrate_test.go`

**Fix:** Add `scan_jobs`, `subtitles`, `audio_tracks` to the `coreTables` slice in `TestRealMigrations_FreshDB`:
```go
coreTables := []string{
    "libraries", "media", "media_files",
    "series", "seasons", "episodes",
    "genres", "media_genres", "people", "credits",
    "scan_jobs", "subtitles", "audio_tracks",
}
```

### Minor #20 — Unused repos commented out in main.go

**File:** `backend/cmd/server/main.go:106-110`

**Fix:** Leave as-is for now. These will be wired up in Phase 03-05. Just make sure the comments are accurate. No code change needed.

---

## Important Rules

1. **Run `make test` after each group** — all existing tests must still pass
2. **Follow existing code patterns** — look at how other repos/services are structured
3. **Error handling** — always wrap errors with context: `fmt.Errorf("doing X: %w", err)`
4. **No ORM** — use raw `database/sql` with `context.Context`
5. **Don't add features** — only fix the listed issues, no extra refactoring
6. **Don't change migration version numbers** — modify existing migration functions in place
7. **Check the review checklist** — mark items as `[x]` in `plans/plan-a-core-domain/review.md` as you complete them
8. **SQLite quirks** — no ALTER TABLE DROP COLUMN (recreate table), BOOLEAN = INTEGER 0/1
