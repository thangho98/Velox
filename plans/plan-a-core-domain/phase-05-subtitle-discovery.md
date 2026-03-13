# Phase 05: Subtitle Discovery
Status: ✅ Done
Plan: A - Core Domain & Ingestion
Dependencies: Phase 03

## Mục tiêu
Discover và index tất cả subtitles (embedded + external) + audio tracks.
Chưa extract/serve ở phase này, chỉ biết có gì. Data này phục vụ cho multi-audio switching và dual subtitle ở Plan D.

## Tasks

### 1. Subtitle Model
- [x] Table `subtitles`: id, media_file_id, language, codec, title, is_embedded, stream_index (for embedded), file_path (for external), is_forced, is_default, is_sdh
- [x] Migration: `006_subtitles.go`

### 2. Embedded Subtitle Discovery (FFprobe)
- [x] Enhance `pkg/ffprobe` để return subtitle stream details
- [x] Parse: codec_name, language (tags.language), title (tags.title), disposition.forced, disposition.default
- [x] Index vào `subtitles` table during scan pipeline "Probe" stage

### 3. External Subtitle Discovery
- [x] Khi scan video file, cũng scan cho sidecar subtitles cùng folder
- [x] Match patterns: `movie.srt`, `movie.en.srt`, `movie.vi.srt`, `movie.forced.en.srt`
- [x] Parse language code từ filename extension chain
- [x] Supported formats: .srt, .vtt, .ass, .ssa, .sub
- **File:** `internal/scanner/subtitle.go` - NEW

### 4. Subtitle Repository
- [x] `SubtitleRepo.ListByMediaFile(mediaFileID) ([]Subtitle, error)`
- [x] `SubtitleRepo.Upsert(subtitle) error`
- [x] `SubtitleRepo.DeleteByMediaFile(mediaFileID) error`
- **File:** `internal/repository/subtitle_audio.go` - NEW (combined with audio track repo)

### 5. Audio Track Discovery
- [x] Table `audio_tracks`: id, media_file_id, stream_index, codec, language, channels, bitrate, is_default, title
- [x] Enhance `pkg/ffprobe` để return audio stream details (codec, language, channels, layout, bitrate)
- [x] Index all audio tracks during scan pipeline "Probe" stage
- [x] Many MKV files have multiple audio: original + dubbed (VN/JP/KR etc) + commentary
- [x] Migration: `006_subtitles.go` (combine with subtitle table)

### 6. Audio/Subtitle API (Read-only for now)
- [x] `GET /api/media/{id}/subtitles` - list available subtitles
- [x] `GET /api/media/{id}/audio-tracks` - list available audio tracks
- [x] Response: `[{id, language, codec, channels, is_default, title}]`
- [x] Actual switching/serving handled in Plan D (playback engine)

## Files to Create/Modify
- `internal/database/migrate/registry.go` - Migration 006 (subtitles + audio_tracks tables)
- `internal/model/model.go` - Subtitle + AudioTrack structs
- `internal/repository/subtitle_audio.go` - NEW (combined repo)
- `internal/scanner/subtitle.go` - NEW
- `pkg/ffprobe/ffprobe.go` - Enhanced subtitle + audio parsing
- `internal/handler/subtitle.go` - NEW (list only)

---
✅ End of Plan A
Next Plan: plan-b-auth-sessions
