# Phase 01: Client Capability Profiles
Status: ✅ Done (6/6)
Plan: D - Playback Decision Engine

## Mục tiêu
Define client capabilities (what codecs/containers/subs the browser can handle)
để playback engine quyết định path tối ưu.

## Tasks

### 1. Device Profile Model ✅ Done
- [x] Struct `DeviceProfile` with `SupportedVideoCodecs`, `SupportedAudioCodecs`, `SupportedContainers`, `SupportedSubtitleFormats`, `MaxWidth`, `MaxHeight`, `MaxBitrate`, `CanBurnSubtitles`, `SupportsHLS`
- **File:** `internal/playback/profile.go`

### 2. Built-in Profiles ✅ Done
- [x] `ChromeDesktop`, `FirefoxDesktop`, `SafariDesktop`, `MobileSafari`, `EdgeDesktop`, `GenericBrowser`, `SmartTV`
- **File:** `internal/playback/profiles_builtin.go`

### 3. Client Detection ✅ Done
- [x] UA detection: `DetectClient()`, `DetectClientFromUA()`, `GetClientInfo()`
- [x] `GET /api/playback/capabilities` endpoint
- **File:** `internal/playback/detect.go`, `internal/handler/playback.go`

### 4. MediaSource Probe (Frontend) ✅ Done
- [x] `webapp/src/lib/capabilities.ts` — probe via `MediaSource.isTypeSupported()`
- [x] Test MIME types: `video/mp4; codecs="avc1.640028"`, `video/mp4; codecs="hev1.1.6.L93.B0"`, `video/webm; codecs="vp9"`, `video/webm; codecs="av01.0.04M.08"`
- [x] Cache result in localStorage (7-day TTL)
- [x] Sent as body to `POST /api/playback/{id}/info` via `WatchPage.tsx` `playbackRequest`

### 5. User Quality Settings ✅ Done
- [x] Player UI: quality selector dropdown (Auto / 1080p / 720p / 480p) in settings menu
- [x] `maxStreamingQuality` persisted in `usePlayerStore` (localStorage)
- [x] Translates to `max_height` in `PlaybackInfoRequest`; backend reads from DB prefs + client override

## Files
- `internal/playback/profile.go` ✅
- `internal/playback/profiles_builtin.go` ✅
- `internal/playback/detect.go` ✅
- `webapp/src/lib/capabilities.ts` ✅
- `webapp/src/stores/player.ts` ✅
- `webapp/src/pages/WatchPage.tsx` ✅

---
Next: phase-02-decision-matrix.md
