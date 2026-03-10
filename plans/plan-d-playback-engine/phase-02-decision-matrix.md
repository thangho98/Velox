# Phase 02: Playback Decision Matrix
Status: ⬜ Pending
Plan: D - Playback Decision Engine
Dependencies: Phase 01

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

### 1. Playback Decision Engine
- [ ] Func `Decide(mediaFile, profile, prefs) PlaybackDecision`
- [ ] Return: `PlaybackDecision { Method, VideoAction, AudioAction, SubtitleAction, Container, EstimatedBitrate }`
- [ ] Methods: `DirectPlay | DirectStream | TranscodeAudio | FullTranscode`
- [ ] VideoAction: `Copy | Transcode(codec, quality)`
- [ ] AudioAction: `Copy | Transcode(codec, channels)`
- **File:** `internal/playback/engine.go` - NEW

### 2. FFmpeg Command Builder
- [ ] Func `BuildFFmpegArgs(decision, inputPath, outputPath/pipe) []string`
- [ ] Direct Stream: `-c:v copy -c:a copy` (just remux container)
- [ ] Audio transcode: `-c:v copy -c:a aac -b:a 192k -ac 2`
- [ ] Full transcode: `-c:v libx264 -preset fast -crf 22 -c:a aac`
- [ ] Output: HLS segments or pipe to stdout (for direct stream)
- **File:** `internal/playback/ffmpeg.go` - NEW

### 3. Playback Info API
- [ ] `GET /api/playback/{mediaID}/info` - return playback decision for current user/client
- [ ] Response: `{method: "DirectPlay|Transcode", stream_url, video_codec, audio_codec, subtitle_tracks, file_size, bitrate}`
- [ ] Frontend uses this to decide how to load the video
- **File:** `internal/handler/playback.go` - NEW

### 4. Stream Router
- [ ] Refactor stream handler to use playback engine
- [ ] `GET /api/stream/{id}` → engine decides: serve file directly OR start transcode
- [ ] `GET /api/stream/{id}?mediaFileId=X` → play specific version
- [ ] Attach decision to response headers: `X-Playback-Method: DirectPlay`
- **File:** `internal/handler/stream.go` - Refactor

### 5. Multi-Audio Track Switching (HLS)
- [ ] When media has multiple audio tracks (e.g., English + Vietnamese dub):
  - Direct Play: browser only plays default audio track (limitation)
  - HLS mode: generate `#EXT-X-MEDIA:TYPE=AUDIO` groups per language
  - Each audio track = separate HLS audio stream, switchable without re-buffering
- [ ] FFmpeg: `-map 0:v:0 -map 0:a:0 -map 0:a:1 ... -var_stream_map "a:0,agroup:audio,name:eng a:1,agroup:audio,name:vie v:0,agroup:audio"`
- [ ] Playback decision: if user wants non-default audio → force HLS mode (even if video codec is direct-playable)
- [ ] `GET /api/playback/{mediaID}/info` response includes `audio_tracks` with selectable options
- [ ] Default audio = user preference language > file default > first track

### 6. Multi-Version Selection
- [ ] If media has multiple files (720p + 1080p + 4K):
  - Pick version that can direct play if possible
  - Otherwise pick closest to user's max quality preference
- [ ] `GET /api/media/{id}/versions` - list available file versions
- [ ] User can manually select version in player

### 7. Transcode Session Lifecycle
- [ ] Track active transcode sessions: media_id, user_id, PID, started_at, progress
- [ ] Kill FFmpeg process when: user stops, disconnects, or switches media
- [ ] Cleanup stale sessions on server start
- [ ] Limit concurrent transcode sessions (config: `VELOX_MAX_TRANSCODES=2`)
- **File:** `internal/playback/session.go` - NEW

## Files to Create/Modify
- `internal/playback/engine.go` - NEW
- `internal/playback/ffmpeg.go` - NEW
- `internal/playback/session.go` - NEW
- `internal/handler/playback.go` - NEW
- `internal/handler/stream.go` - Major refactor

---
Next: phase-03-subtitle-serving.md
