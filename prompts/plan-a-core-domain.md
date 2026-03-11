# Prompt: Implement Plan A - Core Domain & Ingestion

## Context

You are working on **Velox**, a self-hosted home media server (like Jellyfin/Emby but lighter).

**Read these files FIRST before writing any code:**
- `CLAUDE.md` — Project rules, code style, architecture patterns
- `docs/database-design.md` — Full database schema (v1.1), ERD, all table definitions, indexes, constraints, query patterns
- `backend/internal/database/migrate/migrate.go` — Migration runner (already implemented)
- `backend/internal/database/migrate/registry.go` — Migration registry with `All()` function and migration 001

**Tech stack:** Go 1.26 + stdlib `net/http` + SQLite (WAL mode) + FFmpeg/FFprobe

**What's already done (Phase 01 ✅):**
- Migration runner with `Up()`, `Rollback()`, `Status()` — fully tested (10 tests)
- Migration 001: `libraries`, `media` (old flat schema), `progress` (old, no user_id)
- CLI: `velox migrate up|status|rollback`
- Basic handlers, services, repositories for the old schema
- Scanner with ffprobe integration (basic, needs refactor)

---

## Your Task

Implement **Plan A Phases 02-05** sequentially. Each phase builds on the previous. Do NOT skip phases. Run `make test` after each phase to verify.

---

## Phase 02: Core Data Model

### Migrations (append to `registry.go` — do NOT create separate files)

All migrations go into `backend/internal/database/migrate/registry.go` as new entries in the `All()` slice. Follow the exact pattern of `up001`/`down001`.

**Migration 002: `refactor_media_item`**
- ALTER `libraries`: add column `type TEXT DEFAULT 'mixed'`
- RECREATE `media` table (SQLite can't add multiple columns atomically):
  - Drop old `media` table (save data if any, but for dev it's OK to drop)
  - Create new `media`: id, library_id (FK CASCADE), media_type (TEXT NOT NULL DEFAULT 'movie'), title (TEXT NOT NULL), sort_title, tmdb_id (INTEGER), imdb_id (TEXT), overview, release_date, rating (REAL DEFAULT 0), poster_path, backdrop_path, created_at, updated_at
  - Indexes: idx_media_library, idx_media_tmdb (partial WHERE NOT NULL), idx_media_imdb (partial WHERE NOT NULL), idx_media_type, idx_media_title
- CREATE `media_files`: id, media_id (FK CASCADE), file_path (TEXT NOT NULL UNIQUE), file_size, duration, width, height, video_codec, audio_codec, container, bitrate, fingerprint (TEXT — `"{file_size}:{xxhash64_first_64KB}"`), is_primary (INTEGER DEFAULT 1), added_at, last_verified_at
  - Indexes: idx_mf_media, idx_mf_fingerprint
- Down: drop media_files, recreate old media schema, remove type from libraries

**Migration 003: `series_model`**
- CREATE `series`: id, library_id (FK CASCADE), title, sort_title, tmdb_id (UNIQUE), imdb_id, overview, status, first_air_date, poster_path, backdrop_path, created_at, updated_at
  - Indexes: idx_series_library, idx_series_tmdb (partial)
- CREATE `seasons`: id, series_id (FK CASCADE), season_number, title, overview, poster_path, episode_count, created_at
  - Indexes: idx_seasons_series, UNIQUE(series_id, season_number)
- CREATE `episodes`: id, series_id (FK CASCADE), season_id (FK CASCADE), media_id (FK CASCADE, UNIQUE), episode_number, title, overview, still_path, air_date, created_at
  - Indexes: idx_ep_series, idx_ep_season, UNIQUE(season_id, episode_number)

**Migration 004: `genres_people`**
- CREATE `genres`: id, name (TEXT NOT NULL UNIQUE), tmdb_id (UNIQUE)
- CREATE `media_genres`: media_id, series_id, genre_id (FK CASCADE)
  - CHECK ((media_id IS NOT NULL AND series_id IS NULL) OR (media_id IS NULL AND series_id IS NOT NULL)) — exactly one owner
  - UNIQUE(media_id, genre_id) WHERE media_id IS NOT NULL
  - UNIQUE(series_id, genre_id) WHERE series_id IS NOT NULL
- CREATE `people`: id, name, tmdb_id (UNIQUE), profile_path
- CREATE `credits`: id, media_id, series_id, person_id (FK CASCADE), character, role (TEXT NOT NULL), display_order
  - CHECK exactly one owner (same XOR as media_genres)
  - Indexes: idx_credits_media (partial), idx_credits_series (partial), idx_credits_person

### Go Models

Create these files in `backend/internal/model/`:

**`media.go`** — Refactor existing model.go, split out:
```go
type Media struct {
    ID           int64   `json:"id"`
    LibraryID    int64   `json:"library_id"`
    MediaType    string  `json:"media_type"` // "movie" | "episode"
    Title        string  `json:"title"`
    SortTitle    string  `json:"sort_title"`
    TmdbID       *int64  `json:"tmdb_id,omitempty"`
    ImdbID       *string `json:"imdb_id,omitempty"`
    Overview     string  `json:"overview"`
    ReleaseDate  string  `json:"release_date"`
    Rating       float64 `json:"rating"`
    PosterPath   string  `json:"poster_path"`
    BackdropPath string  `json:"backdrop_path"`
    CreatedAt    string  `json:"created_at"`
    UpdatedAt    string  `json:"updated_at"`
}

type MediaFile struct {
    ID             int64   `json:"id"`
    MediaID        int64   `json:"media_id"`
    FilePath       string  `json:"file_path"`
    FileSize       int64   `json:"file_size"`
    Duration       float64 `json:"duration"`
    Width          int     `json:"width"`
    Height         int     `json:"height"`
    VideoCodec     string  `json:"video_codec"`
    AudioCodec     string  `json:"audio_codec"`
    Container      string  `json:"container"`
    Bitrate        int     `json:"bitrate"`
    Fingerprint    string  `json:"fingerprint"`
    IsPrimary      bool    `json:"is_primary"`
    AddedAt        string  `json:"added_at"`
    LastVerifiedAt *string `json:"last_verified_at"`
}
```

**`series.go`** — Series, Season, Episode structs
**`genre.go`** — Genre struct
**`person.go`** — Person, Credit structs

### Repositories

Create in `backend/internal/repository/`. Use raw `database/sql` with `context.Context` as first param. Follow the pattern of existing repos.

- `media.go` — Refactor: Upsert, GetByID, List (with filters), Delete
- `media_file.go` — NEW: Upsert, FindByFingerprint, FindByPath, MarkMissing, ListByMediaID
- `series.go` — NEW: CRUD Series, GetWithSeasons, Season CRUD, Episode CRUD, GetBySeriesAndNumber
- `genre.go` — NEW: CRUD, LinkToMedia, LinkToSeries, ListByMedia, ListBySeries
- `person.go` — NEW: CRUD, AddCredit, ListCredits

### Service Layer

- `service/media.go` — Update to use new models
- `service/series.go` — NEW: series business logic

### Tests

Write migration tests: apply 001→002→003→004, verify tables exist, verify rollback drops them. Add to existing `migrate_test.go`.

---

## Phase 03: Scan Pipeline & File Identity

### Migration 005: `scan_jobs`
Add to registry.go:
- CREATE `scan_jobs`: id, library_id (FK CASCADE), status (DEFAULT 'queued'), total_files, scanned_files, new_files, errors, error_log, started_at, finished_at, created_at
- Also: ALTER libraries ADD type if not already done in 002

### File Fingerprint (`internal/scanner/fingerprint.go`)
```go
// ComputeFingerprint reads first 64KB of file, computes xxhash64, returns "{size}:{hash}"
func ComputeFingerprint(path string) (string, error)
```
Use `github.com/cespare/xxhash/v2` for xxHash64. Add to go.mod.

### Filename Parser (`pkg/nameparser/parser.go`)
```go
type ParsedMedia struct {
    Title     string
    Year      int
    Season    int  // -1 if not a series
    Episode   int  // -1 if not a series
    MediaType string // "movie" | "episode"
    Quality   string // "1080p", "4K", etc.
}
func Parse(filename string) ParsedMedia
```

**Critical: Write 50+ table-driven test cases** covering:
- Standard movies: `The.Matrix.1999.1080p.BluRay.x264.mkv`
- Series: `Breaking.Bad.S05E16.Felina.720p.mkv`
- Multi-episode: `S01E01E02`, `S01E01-E03`
- Daily shows: `Show.2024.03.15.mkv`
- Folder-based: parsing from parent directory name
- Unicode titles, special characters
- Various quality tags, codec tags, release groups
- Edge cases: no year, no quality, extra dots/dashes

### Scan Pipeline (`internal/scanner/pipeline.go`)
Refactor existing scanner into pipeline stages:
1. **Discover** — walk directory, filter video extensions (.mkv, .mp4, .avi, .mov, .wmv, .flv, .webm, .m4v, .ts)
2. **Fingerprint** — compute fingerprint, check DB: known file → skip, known fingerprint different path → rename detected → update path, new → continue
3. **Probe** — ffprobe for technical metadata
4. **Parse** — filename parser for title/year/season/episode
5. **Persist** — save Media + MediaFile to DB

Stage "Match" (TMDb) will be added in Phase 04.

### File Watcher (`internal/watcher/watcher.go`)
Use `github.com/fsnotify/fsnotify`:
- Watch all library paths recursively
- CREATE → debounce 5s → ScanSingleFile
- REMOVE → mark media_file missing
- RENAME → fingerprint match → update path
- Start on server boot, graceful shutdown
- Config: `VELOX_FILE_WATCHER` env (default true)

### Missing File Verification (`internal/scanner/verify.go`)
- `VerifyFiles(libraryID)` — check all media_files exist on disk
- Mark missing: `last_verified_at = NULL`
- Run at startup + expose as API

### API Endpoints
- `POST /api/libraries/{id}/scan` — trigger scan (create scan_job)
- `GET /api/libraries/{id}/scan-status` — poll scan job
- `GET /api/scan-jobs` — list recent scan jobs

---

## Phase 04: TMDb Integration

### TMDb Client (`pkg/tmdb/`)
- `client.go` — HTTP client, API key from `VELOX_TMDB_API_KEY`, rate limiter (40 req/10s), retry on 429
- `search.go` — SearchMovie(title, year), SearchTV(title, year)
- `movie.go` — GetMovie(tmdbID) with credits + genres
- `tv.go` — GetTVShow(tmdbID), GetSeason(tmdbID, seasonNum)
- `image.go` — PosterURL/BackdropURL builders, download + cache to `data/images/`

### NFO Parser (`internal/scanner/nfo.go`)
- Parse `.nfo` XML files next to video files
- Extract: title, year, tmdb_id, overview, rating
- Check for local images: poster.jpg, fanart.jpg, folder.jpg
- Priority: NFO > local images > TMDb API

### Metadata Matcher (`internal/scanner/matcher.go`)
Integrate into scan pipeline as new stage between Parse and Persist:
1. Check NFO first
2. If no NFO → TMDb search by title + year
3. Pick best match (exact title+year = high confidence)
4. Fetch full details (GetMovie/GetTVShow)
5. Save metadata + genres + cast to DB
6. Download poster + backdrop

### APIs
- `PUT /api/media/{id}/identify` — manual override with tmdb_id
- `POST /api/media/{id}/refresh` — re-fetch metadata
- `GET /api/images/{type}/{filename}` — serve cached images

---

## Phase 05: Subtitle & Audio Track Discovery

### Migration 006: `subtitles_audio_tracks`
- CREATE `subtitles`: id, media_file_id (FK CASCADE), language, codec, title, is_embedded, stream_index, file_path, is_forced, is_default, is_sdh
  - UNIQUE(media_file_id, stream_index) WHERE is_embedded = 1
  - UNIQUE(media_file_id, file_path) WHERE is_embedded = 0
- CREATE `audio_tracks`: id, media_file_id (FK CASCADE), stream_index, codec, language, channels, channel_layout, bitrate, title, is_default
  - UNIQUE(media_file_id, stream_index)

### FFprobe Enhancement (`pkg/ffprobe/ffprobe.go`)
Enhance to return detailed subtitle + audio stream info:
- Subtitle: codec_name, language, title, forced, default, sdh detection
- Audio: codec_name, language, channels, channel_layout, bitrate, title, default

### External Subtitle Scanner (`internal/scanner/subtitle.go`)
- Scan for sidecar files: .srt, .vtt, .ass, .ssa, .sub
- Match patterns: `video.srt`, `video.en.srt`, `video.vi.forced.srt`
- Parse language code from filename

### Repositories
- `repository/subtitle.go` — ListByMediaFile, Upsert, DeleteByMediaFile
- `repository/audio_track.go` — ListByMediaFile, Upsert, DeleteByMediaFile

### APIs (read-only)
- `GET /api/media/{id}/subtitles`
- `GET /api/media/{id}/audio-tracks`

---

## Important Rules

1. **All migrations in `registry.go`** — append to `All()` slice, do NOT create separate migration files
2. **Run `make test` after each phase** — ensure all existing tests still pass + new tests pass
3. **Follow existing code patterns** — look at how migration 001, existing handlers/services/repos are structured
4. **Error handling** — always wrap errors with context: `fmt.Errorf("scanning library %d: %w", id, err)`
5. **No ORM** — use raw `database/sql` with `context.Context`
6. **Commit after each phase** — with descriptive message like "feat: implement core data model (Plan A Phase 02)"
7. **Read `docs/database-design.md`** for exact column types, constraints, indexes, and design rationale
8. **SQLite quirks** — no ALTER TABLE DROP COLUMN (recreate table instead), no ENUM (use TEXT + CHECK), BOOLEAN = INTEGER 0/1
9. **Fingerprint** — `"{file_size}:{xxhash64_of_first_64KB}"`, NOT path-based
10. **XOR constraints** on polymorphic tables (media_genres, credits): exactly one of media_id/series_id must be non-null
