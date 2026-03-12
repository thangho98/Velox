# Plan D: Playback Decision Engine
Status: ✅ Done
Priority: 🟡 High

## Dependencies
| Dependency | Why |
|---|---|
| Plan A Ph 03-05 | `media_files`, `subtitles`, `audio_tracks` tables — engine reads codec/container/subtitle metadata from these |
| Plan B | `user_preferences` table (`audio_language`, `subtitle_language`, `max_streaming_quality`) — already exists in `model.UserPreferences` and `repository.UserPreferencesRepo` |
| Plan C (optional) | Working player to validate decisions end-to-end; backend engine itself does not need Plan C |

## Core Principle: Direct Play First
**Direct Play là ưu tiên số 1.** Zero CPU cost trên server.
HLS/transcode chỉ kick in khi:
- Video codec incompatible (HEVC trên Chrome)
- Audio codec incompatible (DTS/TrueHD → AAC)
- Container incompatible (MKV → remux MP4)
- Bitrate exceeds client limit
- User chọn non-default audio track (forces HLS mode)
- Image-based subtitle selected (PGS burn-in, forces full transcode)

## Playback Method → Transport Mapping

| Decision Method | Video | Audio | Transport | Stream URL format |
|---|---|---|---|---|
| `DirectPlay` | copy | copy | HTTP range (serve file) | `?format=direct` |
| `DirectStream` | copy | copy | Remux pipe (MKV→MP4) | `?format=remux` |
| `TranscodeAudio` | copy | transcode | HLS segments | `?format=hls` |
| `FullTranscode` | transcode | transcode | HLS segments | `?format=hls` |

> **Note:** `DirectStream` + incompatible audio collapses to `FullTranscode` — piped remux cannot transcode audio simultaneously. This is intentional and already implemented in `engine.go`.

## Current Status (as of 2026-03-12)

### Phase 01 — Client Capability Profiles
| Task | Status | Location |
|---|---|---|
| `DeviceProfile` struct | ✅ Done | `internal/playback/profile.go` |
| Built-in profiles (Chrome, Firefox, Safari, Mobile, Edge, Generic, SmartTV) | ✅ Done | `internal/playback/profiles_builtin.go` |
| UA detection (`DetectClient`, `GetClientInfo`) | ✅ Done | `internal/playback/detect.go` |
| `GET /api/playback/capabilities` endpoint | ✅ Done | `internal/handler/playback.go` |
| Frontend `capabilities.ts` (MediaSource API probe) | ✅ Done | `webapp/src/lib/capabilities.ts` — wired into `WatchPage.tsx` `playbackRequest` |
| User quality settings UI | ✅ Done | `webapp/src/stores/player.ts` `maxStreamingQuality` + `WatchPage.tsx` settings menu |

### Phase 02 — Playback Decision Matrix
| Task | Status | Location |
|---|---|---|
| `Decide()` engine | ✅ Done | `internal/playback/engine.go` |
| FFmpeg arg builder | ✅ Done | `internal/playback/ffmpeg.go` |
| Transcode session manager | ✅ Done | `internal/playback/session.go` |
| `POST /api/playback/{id}/info` endpoint | ✅ Done | `internal/handler/playback.go` — wired `UserPreferencesRepo`, `media_file_id` selection, correct `subType` derivation, audio/subtitle selection end-to-end |
| Decision engine tests (10 cases) | ✅ Done | `internal/playback/engine_test.go` |
| Stream router integration (use engine in `stream.go`) | ✅ Done | `internal/handler/stream.go` — DirectPlay/DirectStream/HLS redirect; `transcoder.RemuxToWriter`, `StreamService.GetPrimaryFile`+`RemuxToWriter` |
| Multi-audio HLS (`#EXT-X-MEDIA:TYPE=AUDIO`) | ✅ Done | `internal/transcoder/transcoder.go` — `GenerateHLSWithAudio`, `writeMasterPlaylistWithAudio`; fixed: absolute `-map 0:N`, BANDWIDTH=4000000, AUTOSELECT, NAME escape |
| Multi-version selection (`GET /api/media/{id}/versions`) | ✅ Done | `internal/handler/media.go` `GetVersions`, `internal/service/media.go` `ListVersions` |

### Phase 03 — Subtitle Serving & Dual Subs
| Task | Status | Location |
|---|---|---|
| Subtitle extractor (embedded → VTT via FFmpeg) | ✅ Done | `pkg/subtitle/extract.go` — `ExtractSubtitle()` with absolute stream index + file cache |
| SRT → VTT converter (native Go) | ✅ Done | inline `srtToVTT()` in `internal/handler/subtitle.go` |
| Subtitle serving API (`GET /api/media-files/{media_file_id}/subtitles/{subtitle_id}/serve`) | ✅ Done | `internal/handler/subtitle.go` — embedded: extract on demand; external .srt: convert; .vtt: direct |
| Frontend: subtitle serve URL in `<track>` (image subs filtered out) | ✅ Done | `webapp/src/pages/WatchPage.tsx` |
| Frontend: playback info unified (single POST replaces 3 broken endpoints) | ✅ Done | `webapp/src/hooks/stores/useMedia.ts` |
| Frontend: audio/subtitle selection wired to request + query key | ✅ Done | `webapp/src/pages/WatchPage.tsx`, `webapp/src/stores/player.ts` |
| Image subtitle burn-in (PGS/VobSub → FFmpeg `-vf subtitles`) | ✅ Done | `internal/playback/engine.go` `SubtitleStreamIndex`; `internal/playback/ffmpeg.go` `BuildFFmpegArgs` prepends `-vf` + `buildSubtitleArgs` returns `-sn` |
| Frontend dual subtitle overlay | ✅ Done | `webapp/src/components/DualSubtitleOverlay.tsx` — VTT parser, primary (white/bottom) + secondary (yellow/above) |
| Frontend subtitle/audio picker components | ✅ Done | `webapp/src/components/SubtitlePicker.tsx` (dual-mode), `AudioPicker.tsx` (codec/channel label + HLS hint) |

## Phase Summary

| Phase | Name | Status |
|---|---|---|
| 01 | Client Capability Profiles | ✅ 6/6 done |
| 02 | Playback Decision Matrix | ✅ 9/9 done |
| 03 | Subtitle Serving & Dual Subs | ✅ 9/9 done |

## Test Matrix (Required Before Merging Each Phase)

### Phase 01 + 02 — Decision Engine
Table-driven tests in `internal/playback/engine_test.go`:

| # | Media | Profile | Expected Method | Why |
|---|---|---|---|---|
| 1 | H.264 + AAC, MP4 | Chrome | DirectPlay | All compatible |
| 2 | HEVC + AAC, MP4 | Chrome | FullTranscode | HEVC not supported |
| 3 | H.264 + AAC, MKV | Chrome | DirectStream | Container incompatible, codec OK |
| 4 | H.264 + DTS, MKV | Chrome | FullTranscode | Audio incompatible + container → collapses |
| 5 | H.264 + DTS, MP4 | Chrome | TranscodeAudio | Video OK, audio incompatible |
| 6 | H.264 + AAC, MP4, 4K | profile MaxHeight=1080 | FullTranscode | Resolution limit |
| 7 | H.264 + AAC, MP4 | profile MaxBitrate=5000, media Bitrate=30000 | FullTranscode | Bitrate limit |
| 8 | H.264 + AAC, MP4, has PGS sub | Chrome, sub selected | FullTranscode | Burn-in required |
| 9 | H.264 + AAC, MP4, has SRT sub | Chrome, sub selected | DirectPlay + SubtitleCopy | Text sub, no transcode needed |
| 10 | HEVC + AC3, MKV | Safari | DirectPlay | Safari supports HEVC + AC3 + MKV |

### Phase 03 — Subtitle
- SRT → VTT: timestamp format, BOM, CRLF, HTML tags in text, empty cues
- Extractor: embedded VTT, embedded SRT (ASS via FFmpeg), external .srt
- Serving endpoint: Content-Type header, cache hit/miss

## Remaining Work (Prioritized)

1. **`engine_test.go`** — table-driven tests covering the matrix above (no deps, do first)
2. **Stream router** — integrate `Decide()` into `internal/handler/stream.go`
3. **`capabilities.ts`** — frontend MediaSource probe, include result in `POST /api/playback/{id}/info` body (not `/api/playback/capabilities` which is GET-only)
4. **Subtitle extractor + SRT→VTT converter** — `pkg/subtitle/`
5. **Subtitle serving API** — `GET /api/media-files/{media_file_id}/subtitles/{subtitle_id}/serve`
6. **Multi-audio HLS** — FFmpeg `#EXT-X-MEDIA` generation
7. **Dual subtitle overlay** — frontend component
