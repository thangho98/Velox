# Phase 03: Scan Pipeline & File Identity
Status: ⬜ Pending
Plan: A - Core Domain & Ingestion
Dependencies: Phase 02

## Mục tiêu
Scan pipeline đúng nghĩa: job queue, state machine, file fingerprinting, name parsing, fsnotify.
Xương sống của toàn bộ media server.

## Tasks

### 1. Scan Job Model & State Machine
- [ ] Table `scan_jobs`: id, library_id, status (queued|scanning|completed|failed), total_files, scanned_files, errors, started_at, finished_at
- [ ] State transitions: queued → scanning → completed/failed
- [ ] Chỉ 1 scan job per library chạy cùng lúc
- [ ] `GET /api/libraries/{id}/scan-status` - poll job status
- **Migration:** `005_scan_jobs.go`

### 2. File Fingerprint
- [ ] Fingerprint = `fmt.Sprintf("%s:%d", filepath, filesize)` (fast, no hash needed for MVP)
- [ ] Func `ComputeFingerprint(path string) (string, error)`
- [ ] Khi scan: check fingerprint trong DB trước
- [ ] Nếu fingerprint match nhưng path khác → file đã rename/move → update path
- [ ] Nếu fingerprint mới → new file → process
- **File:** `internal/scanner/fingerprint.go` - NEW

### 3. Filename Parser
- [ ] `pkg/nameparser/parser.go` - NEW
- [ ] Parse movie: `Movie.Name.2024.1080p.BluRay.x264.mkv` → title="Movie Name", year=2024
- [ ] Parse series: `Show.Name.S02E05.Episode.Title.720p.mkv` → show="Show Name", season=2, episode=5
- [ ] Handle formats: `S01E01`, `1x01`, `Season 1/Episode 01`, folder-based
- [ ] Strip quality tags: 720p, 1080p, 4K, BluRay, WEB-DL, x264, x265, HEVC, AAC, DTS, REMUX
- [ ] Handle dots/underscores/dashes → space
- [ ] Return `ParsedMedia { Title, Year, Season, Episode, MediaType, Quality }`
- [ ] **Comprehensive test suite** (`pkg/nameparser/parser_test.go`):
  - Edge cases: special chars, unicode, multi-episode (S01E01E02), daily shows (2024.03.15)
  - Real-world samples: tham khảo Jellyfin regex patterns
  - Table-driven tests với 50+ cases từ tên file thực tế

### 4. Scan Pipeline Orchestrator
- [ ] Refactor `scanner.ScanLibrary()` thành pipeline stages:
  1. **Discover** - walk dir, filter video extensions, collect file list
  2. **Fingerprint** - compute fingerprint, skip known files
  3. **Probe** - ffprobe for new/changed files
  4. **Parse** - filename parser → title, year, season, episode
  5. **Match** - TMDb search (Phase 04)
  6. **Persist** - save to DB
- [ ] Each stage reports progress → update scan_job
- [ ] Errors don't stop pipeline (log + continue)
- **File:** `internal/scanner/pipeline.go` - NEW

### 5. Single File Scan
- [ ] Extract `ScanSingleFile(path, libraryID) error` từ pipeline
- [ ] Reuse cho file watcher events
- [ ] Same stages nhưng cho 1 file
- **File:** `internal/scanner/pipeline.go`

### 6. File Watcher (fsnotify)
- [ ] `internal/watcher/watcher.go` - NEW
- [ ] Watch tất cả library paths recursively
- [ ] CREATE event → debounce (5s) → ScanSingleFile
- [ ] REMOVE event → mark media_file as missing (set last_verified_at = null)
- [ ] RENAME event → update file_path via fingerprint match
- [ ] Watch new subdirectories khi được tạo
- [ ] Start on server boot, graceful stop on shutdown
- [ ] Config: `VELOX_FILE_WATCHER=true` (default)

### 7. Missing File Detection
- [ ] Periodic check (startup + scheduled): verify all media_files still exist on disk
- [ ] Func `VerifyFiles(libraryID) (missing []MediaFile, error)`
- [ ] Mark missing files: `last_verified_at = null`, optionally flag
- [ ] Don't auto-delete from DB (user might remount drive)
- **File:** `internal/scanner/verify.go` - NEW

### 8. Library Type
- [ ] Add `type` field to `libraries` table: movies | tvshows | mixed
- [ ] movies → only parse as movies
- [ ] tvshows → only parse as series
- [ ] mixed → auto-detect from filename pattern
- **Migration:** `005_scan_jobs.go` (combine)

## Files to Create/Modify
- `internal/database/migrate/migrations/005_scan_jobs.go` - NEW
- `internal/scanner/pipeline.go` - NEW (replaces old scanner.go logic)
- `internal/scanner/fingerprint.go` - NEW
- `internal/scanner/verify.go` - NEW
- `internal/watcher/watcher.go` - NEW
- `pkg/nameparser/parser.go` - NEW
- `internal/scanner/scanner.go` - Refactor into pipeline

---
Next: phase-04-tmdb-integration.md
