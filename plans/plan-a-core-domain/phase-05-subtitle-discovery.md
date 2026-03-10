# Phase 05: Subtitle Discovery
Status: ⬜ Pending
Plan: A - Core Domain & Ingestion
Dependencies: Phase 03

## Mục tiêu
Discover và index tất cả subtitles (embedded + external) + audio tracks.
Chưa extract/serve ở phase này, chỉ biết có gì. Data này phục vụ cho multi-audio switching và dual subtitle ở Plan D.

## Tasks

### 1. Subtitle Model
- [ ] Table `subtitles`: id, media_file_id, language, codec, title, is_embedded, stream_index (for embedded), file_path (for external), is_forced, is_default, is_sdh
- [ ] Migration: `006_subtitles.go`

### 2. Embedded Subtitle Discovery (FFprobe)
- [ ] Enhance `pkg/ffprobe` để return subtitle stream details
- [ ] Parse: codec_name, language (tags.language), title (tags.title), disposition.forced, disposition.default
- [ ] Index vào `subtitles` table during scan pipeline "Probe" stage

### 3. External Subtitle Discovery
- [ ] Khi scan video file, cũng scan cho sidecar subtitles cùng folder
- [ ] Match patterns: `movie.srt`, `movie.en.srt`, `movie.vi.srt`, `movie.forced.en.srt`
- [ ] Parse language code từ filename extension chain
- [ ] Supported formats: .srt, .vtt, .ass, .ssa, .sub
- **File:** `internal/scanner/subtitle.go` - NEW

### 4. Subtitle Repository
- [ ] `SubtitleRepo.ListByMediaFile(mediaFileID) ([]Subtitle, error)`
- [ ] `SubtitleRepo.Upsert(subtitle) error`
- [ ] `SubtitleRepo.DeleteByMediaFile(mediaFileID) error`
- **File:** `internal/repository/subtitle.go` - NEW

### 5. Audio Track Discovery
- [ ] Table `audio_tracks`: id, media_file_id, stream_index, codec, language, channels, bitrate, is_default, title
- [ ] Enhance `pkg/ffprobe` để return audio stream details (codec, language, channels, layout, bitrate)
- [ ] Index all audio tracks during scan pipeline "Probe" stage
- [ ] Many MKV files have multiple audio: original + dubbed (VN/JP/KR etc) + commentary
- [ ] Migration: `006_subtitles.go` (combine with subtitle table)

### 6. Audio/Subtitle API (Read-only for now)
- [ ] `GET /api/media/{id}/subtitles` - list available subtitles
- [ ] `GET /api/media/{id}/audio-tracks` - list available audio tracks
- [ ] Response: `[{id, language, codec, channels, is_default, title}]`
- [ ] Actual switching/serving handled in Plan D (playback engine)

## Files to Create/Modify
- `internal/database/migrate/migrations/006_subtitles.go` - NEW (includes audio_tracks table)
- `internal/model/subtitle.go` - NEW
- `internal/model/audio_track.go` - NEW
- `internal/repository/subtitle.go` - NEW
- `internal/repository/audio_track.go` - NEW
- `internal/scanner/subtitle.go` - NEW
- `pkg/ffprobe/ffprobe.go` - Enhance subtitle + audio parsing
- `internal/handler/subtitle.go` - NEW (list only)

---
✅ End of Plan A
Next Plan: plan-b-auth-sessions
