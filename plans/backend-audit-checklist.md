# Backend Audit Checklist (Post-Refactor)

Created: 2026-04-01
Status: ✅ Done (P1-P4, P6 fixed; P5 skipped — files functional correct)

> Review kết quả audit backend sau Plan Q refactor.
> Mỗi item cần fix phải đảm bảo `go build ./...` và `go vet ./...` pass.

---

## Priority 1: Repository — ErrNotFound Wrapping (HIGH)

**Rule:** Tất cả `GetByID`-style methods phải wrap `sql.ErrNoRows` → `repository.ErrNotFound`.
**Impact:** Missing record trả HTTP 500 thay vì 404.

**Pattern cần áp dụng:**
```go
func (r *FooRepo) GetByID(ctx context.Context, id int64) (*model.Foo, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... WHERE id = ?", id)
    var foo model.Foo
    if err := row.Scan(&foo.Field1, ...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get foo by id %d: %w", id, err)
    }
    return &foo, nil
}
```

### repository/user.go
- [ ] `UserRepo.GetByID` (line ~45) — bare `return nil, err` leaks `sql.ErrNoRows`
- [ ] `UserRepo.GetByUsername` (line ~61) — bare `return nil, err` leaks `sql.ErrNoRows`

### repository/genre.go
- [ ] `GenreRepo.GetByID` (line ~27) — bare `return nil, err`
- [ ] `GenreRepo.GetByName` (line ~38) — bare `return nil, err`
- [ ] `GenreRepo.GetByTmdbID` (line ~49) — bare `return nil, err`

### repository/person.go
- [ ] `PersonRepo.GetByID` (line ~27) — bare `return nil, err`
- [ ] `PersonRepo.GetByName` (line ~38) — bare `return nil, err`
- [ ] `PersonRepo.GetByTmdbID` (line ~49) — bare `return nil, err`

### repository/library.go
- [ ] `LibraryRepo.GetByID` (line ~76) — bare `return nil, err`

### repository/media.go
- [ ] `MediaRepo.GetByID` via `scanMedia` (line ~70) — bare `return nil, err`

### repository/media_file.go
- [ ] `MediaFileRepo.GetByID` via `scanMediaFile` (line ~73) — bare `return nil, err`

### repository/subtitle.go
- [ ] `SubtitleRepo.GetByID` (line ~57) — bare `return nil, err`

### repository/audio_track.go
- [ ] `AudioTrackRepo.GetByID` (line ~44) — bare `return nil, err`

### repository/session.go
- [ ] `RefreshTokenRepo.GetByTokenHash` (line ~34) — bare `return nil, err`
- [ ] `SessionRepo.GetByID` (line ~111) — bare `return nil, err`

### repository/scan_job.go
- [ ] `ScanJobRepo.GetByID` (line ~28) — bare `return nil, err`

### repository/media_marker.go
- [ ] `MediaMarkerRepo.GetByID` via `scanMarker` (line ~46) — bare `return nil, err`

### repository/webhook.go
- [ ] `WebhookRepo.GetByID` (line ~30) — wraps with `fmt.Errorf` but does NOT convert to `ErrNotFound`

### repository/series.go
- [ ] `SeasonRepo.GetByID` (line ~416) — bare `return nil, err`
- [ ] `SeasonRepo.GetBySeriesAndNumber` (line ~430) — bare `return nil, err`
- [ ] `EpisodeRepo.GetByID` (line ~508) — bare `return nil, err`
- [ ] `EpisodeRepo.GetByMediaID` (line ~522) — bare `return nil, err`
- [ ] `EpisodeRepo.GetBySeasonAndNumber` (line ~536) — bare `return nil, err`

### Test update needed
- [ ] `repository/user_test.go` (line ~126) — test checks `sql.ErrNoRows` → update to check `repository.ErrNotFound`

**Note — Intentional exceptions (DO NOT fix):**
- `AppSettingsRepo.Get` returns `("", nil)` on ErrNoRows — correct by design (missing = empty default)
- `PretranscodeRepo` methods returning `(nil, nil)` on ErrNoRows — intentional "not found = nil result"
- `MediaMarkerRepo.FindLastByType` returns raw `sql.ErrNoRows` — callers check this intentionally

---

## Priority 2: Repository — RowsAffected Checks (MEDIUM)

**Rule:** Tất cả update/delete operations phải check `RowsAffected()`, return `ErrNotFound` nếu 0.
**Impact:** Silent success khi record không tồn tại, có thể mask bugs.

**Pattern cần áp dụng:**
```go
func (r *FooRepo) Delete(ctx context.Context, id int64) error {
    res, err := r.db.ExecContext(ctx, "DELETE FROM foo WHERE id = ?", id)
    if err != nil {
        return fmt.Errorf("delete foo %d: %w", id, err)
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return ErrNotFound
    }
    return nil
}
```

### repository/user.go
- [ ] `UserRepo.Update` (line ~101)
- [ ] `UserRepo.UpdatePassword` (line ~116)
- [ ] `UserRepo.Delete` (line ~124)

### repository/genre.go
- [ ] `GenreRepo.Update` (line ~61)
- [ ] `GenreRepo.Delete` (line ~67)

### repository/person.go
- [ ] `PersonRepo.Update` (line ~60)
- [ ] `PersonRepo.Delete` (line ~67)
- [ ] `PersonRepo.UpdateCredit` (line ~107)
- [ ] `PersonRepo.RemoveCredit` (line ~115)

### repository/library.go
- [ ] `LibraryRepo.Delete` (line ~118)

### repository/media.go
- [ ] `MediaRepo.Update` — full update (line ~116)

### repository/media_file.go
- [ ] `MediaFileRepo.Update` (line ~83)
- [ ] `MediaFileRepo.Delete` (line ~103)
- [ ] `MediaFileRepo.UpdatePath` (line ~163)

### repository/subtitle.go
- [ ] `SubtitleRepo.Delete` (line ~112)
- [ ] `SubtitleRepo.Update` (line ~126)

### repository/audio_track.go
- [ ] `AudioTrackRepo.Delete` (line ~93)
- [ ] `AudioTrackRepo.Update` (line ~98)

### repository/session.go
- [ ] `RefreshTokenRepo.Delete` (line ~63)
- [ ] `SessionRepo.Delete` (line ~153)
- [ ] `SessionRepo.UpdateLastActive` (line ~178)

### repository/webhook.go
- [ ] `WebhookRepo.Update` (line ~63)
- [ ] `WebhookRepo.Delete` (line ~75)

### repository/media_marker.go
- [ ] `MediaMarkerRepo.Delete` (line ~108)

### repository/scan_job.go
- [ ] `ScanJobRepo.UpdateStatus` (line ~51)
- [ ] `ScanJobRepo.Start` (line ~57)
- [ ] `ScanJobRepo.Complete` (line ~65)
- [ ] `ScanJobRepo.Fail` (line ~75)

### repository/series.go
- [ ] `SeriesRepo.Update` — full update (line ~57)
- [ ] `SeriesRepo.Delete` (line ~161)
- [ ] `SeasonRepo.Update` (line ~449)
- [ ] `SeasonRepo.Delete` (line ~454)
- [ ] `EpisodeRepo.Update` (line ~555)
- [ ] `EpisodeRepo.UpdateSeasonLink` (line ~563)
- [ ] `EpisodeRepo.Delete` (line ~569)

**Note — Intentional exceptions (DO NOT fix):**
- Bulk delete methods: `ClearMediaGenres`, `ClearSeriesGenres`, `DeleteByMediaFileID`, `DeleteByUserID`, `DeleteExpired` — 0 rows affected is semantically valid
- Upsert operations: `UserDataRepo.SaveProgress`, `SetRating` — always succeed by design
- `SubtitleRepo.DeleteByMediaFileID`, `AudioTrackRepo.DeleteByMediaFileID` — bulk cleanup

---

## Priority 3: Service — Raw SQL in admin.go (HIGH)

**Rule:** Services KHÔNG được chứa raw SQL. Mọi query phải qua repository layer.
**Impact:** Vi phạm layer architecture, khó test, khó maintain.

### service/admin.go
- [ ] Line ~71: `s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(file_size), 0) FROM media_files")` → Move to `MediaFileRepo.TotalSize(ctx) (int64, error)`
- [ ] Lines ~81-93: Multi-table JOIN query for library stats → Move to `LibraryRepo.GetStats(ctx) ([]LibraryStat, error)` or similar
- [ ] Line ~112: `fmt.Sprintf("SELECT COUNT(*) FROM %s", table)` → Move to a repo method per table, hoặc `AdminRepo.CountTable(ctx, table)` (cẩn thận SQL injection dù hiện tại chỉ dùng hardcoded table names)

### After moving queries
- [ ] Remove `repository.DBTX` field from `AdminService` struct — replace with proper repo dependencies
- [ ] Remove `database/sql` import from `admin.go` if no longer needed

---

## Priority 4: Error Context Wrapping (LOW)

**Rule:** Wrap errors với context: `fmt.Errorf("doing X: %w", err)`. Không bare `return err`.
**Impact:** Debug khó hơn trong production logs.

> **Note:** Có ~150+ bare `return err` instances across repository + service files.
> Khi fix Priority 1 và 2, nhiều chỗ sẽ được fix cùng lúc (vì thêm `fmt.Errorf` wrapper).
> Phần còn lại chủ yếu là 1-liner passthrough methods — ưu tiên thấp.

### Repository files cần wrap (worst offenders)
- [ ] `genre.go` — 8 bare returns (Update, Delete, Link/Unlink methods)
- [ ] `person.go` — 6 bare returns (Update, Delete, Credit methods)
- [ ] `session.go` — 8 bare returns (Delete, UpdateLastActive methods)
- [ ] `scan_job.go` — 8 bare returns (UpdateStatus, Start, Complete, Fail, etc.)
- [ ] `series.go` — 10 bare returns (Season/Episode Update, Delete methods)
- [ ] `media_file.go` — 9 bare returns (Create, Update, Delete, etc.)
- [ ] `subtitle.go` — 5 bare returns
- [ ] `audio_track.go` — 5 bare returns
- [ ] `user_data.go` — 5 bare returns
- [ ] `media_marker.go` — 4 bare returns

### Service files cần wrap (worst offenders)
- [ ] `settings.go` — ~12 bare returns
- [ ] `metadata.go` — ~6 bare returns
- [ ] `metadata_editor.go` — ~6 bare returns
- [ ] `pretranscode_admin.go` — ~8 bare returns
- [ ] `browse.go` — ~3 bare returns
- [ ] `cinema.go` — ~4 bare returns

---

## Priority 5: File Size Splits (LOW)

**Rule:** Max ~500 lines per file. Split by logical concern.

### metadata/matcher.go (647 lines)
- [ ] Xác định split points — e.g., movie matching vs episode matching vs helper functions
- [ ] Split thành 2-3 files trong cùng package

### scanner/pipeline.go (643 lines)
- [ ] Xác định split points — e.g., scan orchestration vs file processing vs metadata extraction
- [ ] Split thành 2-3 files trong cùng package

### repository/series.go (614 lines)
- [ ] Chứa 3 repos: `SeriesRepo`, `SeasonRepo`, `EpisodeRepo`
- [ ] Split thành `series.go` + `season.go` + `episode.go`

### repository/pretranscode.go (533 lines)
- [ ] Chứa 2 repos: `PretranscodeProfileRepo`, `PretranscodeFileRepo`
- [ ] Split thành `pretranscode_profile.go` + `pretranscode_file.go`

---

## Priority 6: Service Layer — Minor Issues (LOW)

### Missing context.Context parameter
- [ ] `stream.go:206` `RemuxToWriter` — long-running FFmpeg process, nên accept `ctx` để support cancellation
- [ ] `stream.go:147` `WaitForSegment` — nên accept `ctx` thay vì bare `timeout`

### database/sql import cleanup
Sau khi fix Priority 1 (ErrNotFound wrapping), các service files nên dùng `repository.ErrNotFound` thay vì `sql.ErrNoRows`:
- [ ] `auth.go` — replace `sql.ErrNoRows` checks with `repository.ErrNotFound`
- [ ] `browse.go` — same
- [ ] `marker.go` — same
- [ ] `media.go` — same
- [ ] `metadata.go` — same
- [ ] `metadata_editor_operations.go` — same
- [ ] `stream.go` — same
- [ ] `subtitle.go` — same
- [ ] `subtitle_tracks.go` — same
- [ ] `user_data.go` — same

### Handler — Business logic in playback (INFORMATIONAL)
> Không cần fix ngay. Trade-off hợp lý vì playback logic tightly coupled với HTTP response building.
- `handler/playback_info.go` — goroutine orchestration, pretranscode validation, admin policy
- `handler/playback_quality.go` — quality tier building
- `handler/playback_tracks.go` — language normalization, subtitle filtering

---

## Verification

Sau mỗi batch fix:
```sh
cd backend && go build ./... && go vet ./...
```

Sau khi fix tất cả:
```sh
cd backend && make test
```
