# Phase 01: Marker Ingestion
Status: ⬜ Pending
Plan: K - Intro / Credits Skip
Dependencies: Plans A/D, existing scan pipeline

## Mục tiêu
Tạo được persistence layer và chapter parser đủ chắc để mỗi `media_file` có thể sinh ra danh sách marker `intro` / `credits` ổn định sau mỗi lần scan.

## Output của phase này
- DB có bảng `media_markers`
- `ffprobe` trả được chapter range + title
- Scan pipeline persist marker theo `media_file_id`
- Rescan không tạo duplicate và không để lại marker stale

## Tasks
### 1. Add Persistence Layer
- [ ] Add migration `019_intro_markers.go` tạo bảng `media_markers`
- [ ] Define fields: `media_file_id`, `marker_type`, `start_sec`, `end_sec`, `source`, `confidence`, timestamps
- [ ] Add unique index cho `(media_file_id, marker_type, source, start_sec, end_sec)`
- [ ] Register migration trong `backend/internal/database/migrate/registry.go`
- [ ] Decide có thêm `label` ngay từ đầu để giữ raw chapter title hay không; khuyến nghị: có

**Verify:** `migrate_test.go` pass và schema in-memory có bảng + index đúng tên

### 2. Add Domain Types
- [ ] Add model cho marker trong `backend/internal/model`
- [ ] Add repository cho create/upsert/list/delete marker theo `media_file_id`
- [ ] Follow repo pattern từ `subtitle_audio.go`: `NewXRepo`, `WithTx`, `ListByMediaFileID`, `DeleteByMediaFileID`
- [ ] Add helper list sorted by `start_sec ASC`

**Verify:** repository test tạo được marker, upsert không duplicate, delete-by-media-file xóa sạch đúng record

### 3. Parse Chapter Metadata
- [ ] Add `-show_chapters` vào ffprobe command args (hiện chỉ có `-show_streams` + `-show_format`)
- [ ] Add `Chapters []ChapterEntry` vào `DetailedProbeResult` struct để parse ffprobe JSON
- [ ] Add `ChapterInfo` / `MarkerCandidate` struct để tách raw ffprobe output khỏi normalized marker
- [ ] Extend `ProbeResult` với `Chapters []ChapterInfo` (fields: Title, StartSec, EndSec)
- [ ] Normalize chapter title thành `intro` hoặc `credits`
- [ ] Ignore chapter quá ngắn (`< 5s`) hoặc invalid range (`end <= start`)
- [ ] Parse chapter start/end từ ffprobe JSON (fields: `start_time`, `end_time` as strings → float64)
- [ ] Preserve raw title trong `label` field để debug heuristic mismatch

**Verify:** unit test với sample ffprobe JSON:
- chapter `Intro` -> marker `intro`
- chapter `Opening Credits` -> marker `intro`
- chapter `Credits` -> marker `credits`
- unknown title -> no marker

### 4. Integrate with Scan Pipeline
- [ ] Persist markers khi scan file mới
- [ ] Refresh markers khi rescan hoặc file bị replace
- [ ] Xóa marker cũ **chỉ cùng source**: `DELETE FROM media_markers WHERE media_file_id = ? AND source = 'chapter'` — không xóa `manual` hoặc `fingerprint`
- [ ] Hook marker persistence vào cùng transaction với subtitle/audio track refresh nếu có thể
- [ ] Chỉ persist `source = chapter` trong phase này
- [ ] Đảm bảo renamed file giữ marker vì marker bám theo `media_file_id`, không theo path

**Verify:** scan 1 file có chapter, DB có marker; scan lại file đó không tăng row count; replace file tại same path refresh marker mới

### 5. Verify
- [ ] Add unit tests cho chapter parsing + normalization
- [ ] Add repository tests cho upsert/list/delete
- [ ] Add scan pipeline regression test nếu feasible, hoặc ít nhất test helper persist batch

## Files dự kiến
- `backend/internal/database/migrate/019_intro_markers.go` - NEW
- `backend/internal/database/migrate/registry.go`
- `backend/internal/model/marker.go` hoặc file phù hợp - NEW
- `backend/internal/repository/media_marker.go` - NEW
- `backend/pkg/ffprobe/ffprobe.go`
- `backend/internal/scanner/pipeline.go`

## Notes
- Phase này là nền cho cả chapter support lẫn fingerprint backfill, nên schema đừng khóa cứng vào chapter.
- Nếu raw ffprobe chapter support thiếu ổn định, fallback là parse chapter title tối thiểu và defer heuristic nâng cao sang phase sau.
