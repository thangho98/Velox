# Phase 02: Playback Decision Matrix
Status: ✅ Done (7/7)
Plan: D - Playback Decision Engine
Dependencies: Phase 01, Plan A Ph 03-05 (media_files/audio_tracks/subtitles tables)

## Mục tiêu
Engine quyết định playback path tối ưu cho mỗi media + client combination.

## Decision Tree
```
Input: MediaFile + DeviceProfile + UserPrefs
  │
  ├─ Video codec supported? ──── Container supported? ──── Audio supported?
  │     YES                          YES                      YES
  │     → DIRECT PLAY (zero server cost)
  │
  ├─ Video codec supported? ──── Container NOT supported?
  │     YES                          YES (e.g., MKV)
  │     → DIRECT STREAM / REMUX (MKV→MP4, very low CPU)
  │
  ├─ Video codec supported? ──── Audio NOT supported?
  │     YES                          YES (e.g., DTS, TrueHD)
  │     → TRANSCODE AUDIO ONLY (low CPU)
  │
  └─ Video codec NOT supported? (e.g., HEVC on Chrome)
        → FULL TRANSCODE (high CPU)
```

## Tasks

### 1. Playback Decision Engine ✅ Done
- [x] `Decide(media MediaFileInfo, profile *DeviceProfile, prefs UserPreferences) PlaybackDecision`
- [x] Methods: `DirectPlay | DirectStream | TranscodeAudio | FullTranscode`
- [x] Priority order: resolution → codec → bitrate → container → audio → subtitles
- **File:** `internal/playback/engine.go`

### 2. FFmpeg Command Builder ✅ Done
- [x] `BuildFFmpegArgs`, `BuildRemuxArgs`, `BuildExtractSubtitleArgs`, `BuildBurnSubtitleArgs`
- [x] HLS args with keyframe alignment (`-force_key_frames`, `-sc_threshold 0`)
- **File:** `internal/playback/ffmpeg.go`

### 3. Playback Info API ✅ Done
- [x] `POST /api/playback/{id}/info` — returns decision + stream_url + audio/subtitle tracks
- [x] Accepts client capability override in request body
- [x] `GET /api/playback/capabilities` — UA-detected capabilities
- **File:** `internal/handler/playback.go`

### 4. Transcode Session Lifecycle ✅ Done
- [x] `SessionManager` with RWMutex, per-session progress, kill, cleanup goroutine
- [x] Session ID uses `crypto/rand` (collision-proof)
- **File:** `internal/playback/session.go`

### 5. Decision Engine Tests ✅ Done
- [x] Table-driven tests for all 10 matrix cases in plan.md
- **File:** `internal/playback/engine_test.go`

### 6. Stream Router Integration ✅ Done
- [x] `GET /api/stream/{id}` calls `Decide()`:
  - DirectPlay → `http.ServeContent` with HTTP range support
  - DirectStream → `RemuxToWriter` pipe (MKV→fragmented MP4)
  - TranscodeAudio / FullTranscode → 307 redirect to `/hls/master.m3u8`
- **File:** `internal/handler/stream.go`

### 7. Multi-Audio HLS ✅ Done
- [x] Non-default audio track → forces HLS mode
- [x] Separate per-audio-track transcode + manual `#EXT-X-MEDIA:TYPE=AUDIO` master playlist
  - Approach: explicit `-map 0:N` per stream, manual `writeMasterPlaylistWithAudio()` — more debuggable than `-var_stream_map`
- [x] `GenerateHLSWithAudio()` in transcoder handles fallback to simple HLS when ≤ 1 audio track
- **File:** `internal/transcoder/transcoder.go`

## Files
- `internal/playback/engine.go` ✅
- `internal/playback/ffmpeg.go` ✅
- `internal/playback/session.go` ✅
- `internal/handler/playback.go` ✅
- `internal/playback/engine_test.go` ✅
- `internal/handler/stream.go` ✅
- `internal/transcoder/transcoder.go` ✅

---
Next: phase-03-subtitle-serving.md
