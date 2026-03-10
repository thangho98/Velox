# Phase 01: Client Capability Profiles
Status: ⬜ Pending
Plan: D - Playback Decision Engine

## Mục tiêu
Define client capabilities (what codecs/containers/subs the browser can handle)
để playback engine quyết định path tối ưu.

## Tasks

### 1. Device Profile Model
- [ ] Struct `DeviceProfile`:
  - `SupportedVideoCodecs: []string` (h264, vp9, av1, hevc*)
  - `SupportedAudioCodecs: []string` (aac, opus, mp3, flac, ac3)
  - `SupportedContainers: []string` (mp4, webm, hls)
  - `SupportedSubtitleFormats: []string` (vtt, srt)
  - `MaxWidth, MaxHeight, MaxBitrate`
  - `CanBurnSubtitles: bool` (false for browsers - need server-side)
  - `SupportsHLS: bool`
- **File:** `internal/playback/profile.go` - NEW

### 2. Built-in Profiles
- [ ] `ChromeDesktop`: h264+vp9+av1, aac+opus, mp4+webm+hls, vtt
- [ ] `FirefoxDesktop`: h264+vp9+av1, aac+opus+flac, mp4+webm+hls, vtt
- [ ] `SafariDesktop`: h264+hevc, aac+alac, mp4+hls, vtt
- [ ] `MobileSafari`: h264+hevc, aac, mp4+hls, vtt (lower bitrate limit)
- [ ] `GenericBrowser`: h264, aac, mp4+hls, vtt (safe fallback)
- **File:** `internal/playback/profiles_builtin.go` - NEW

### 3. Client Detection
- [ ] Parse `User-Agent` header to determine browser family
- [ ] `GET /api/playback/capabilities` → frontend can also self-report capabilities
- [ ] Frontend: use `MediaSource.isTypeSupported()` to detect codec support
- [ ] POST capabilities to server or include in playback request
- **File:** `internal/playback/detect.go` - NEW

### 4. MediaSource Probe (Frontend)
- [ ] `src/lib/capabilities.ts` - detect client capabilities via MediaSource API
- [ ] Test: `video/mp4; codecs="avc1.640028"` (H.264 High)
- [ ] Test: `video/mp4; codecs="hev1.1.6.L93.B0"` (HEVC)
- [ ] Test: `video/webm; codecs="vp9"` (VP9)
- [ ] Send results to server once, cache in localStorage

### 5. User Quality Settings
- [ ] User preference: max streaming quality (auto/original/1080p/720p/480p)
- [ ] User preference: prefer direct play (minimize server load) vs prefer compatibility
- [ ] Per-session override: quality selector in player UI
- [ ] Apply as constraints on playback decision

## Files to Create
- `internal/playback/profile.go`
- `internal/playback/profiles_builtin.go`
- `internal/playback/detect.go`
- `src/lib/capabilities.ts`

---
Next: phase-02-decision-matrix.md
