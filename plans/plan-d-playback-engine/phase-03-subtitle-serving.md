# Phase 03: Subtitle Burn-in & Serving
Status: ✅ Done
Plan: D - Playback Decision Engine
Dependencies: Plan A Ph 04 (subtitles table), Plan A Ph 05 (scanner populates subtitle rows), Phase 02

## Mục tiêu
Serve subtitles: text-based → external VTT track, image-based (PGS) → burn-in during transcode.

## Subtitle Decision Tree
```
Text-based (SRT, ASS, VTT)?
  ├─ External file (.srt) → Convert to VTT, serve as <track>
  ├─ Embedded → Extract with FFmpeg → Convert to VTT → serve as <track>
  └─ Browser renders natively (zero CPU cost)

Image-based (PGS, VobSub)?
  ├─ Can only be burned into video stream
  ├─ Forces transcode even if video codec is compatible
  └─ High CPU cost, but no alternative for PGS
```

## Tasks

### 1. Subtitle Extractor ✅ Done
- [x] Func `ExtractSubtitle(videoPath, streamIndex, outputDir) (string, error)`
- [x] Uses absolute stream index: `ffmpeg -i video.mkv -map 0:N -c:s webvtt output.vtt`
- [x] Cache extracted subs: `~/.velox/subtitles/{media_file_id}/{stream_index}.vtt`
- **File:** `pkg/subtitle/extract.go`

### 2. SRT → VTT Converter (Native Go) ✅ Done
- [x] Timestamp conversion: `00:01:23,456 --> 00:01:25,789` → `00:01:23.456 --> 00:01:25.789`
- [x] Handle: BOM strip, CRLF normalization
- [x] Export: `SRTToVTT(data []byte) []byte`; used in `handler/subtitle.go` `Serve()`
- **File:** `pkg/subtitle/convert.go`

### 3. Subtitle Serving API
- Existing routes already use `media_file_id` as the key (`GET /api/media-files/{media_file_id}/subtitles`); new serve route must follow the same pattern
- [ ] `GET /api/media-files/{media_file_id}/subtitles/{subtitle_id}/serve` → serve VTT file
- [ ] If embedded + not yet extracted → extract on demand → serve
- [ ] If external .srt → convert to VTT on the fly (or cache)
- [ ] If external .vtt → serve directly
- [ ] Content-Type: `text/vtt; charset=utf-8`
- **File:** `internal/handler/subtitle.go` - Update

### 4. Image Subtitle Burn-in
- [ ] When user selects PGS/VobSub subtitle → modify FFmpeg args
- [ ] Add `-filter_complex "[0:v][0:s:{index}]overlay"` for burn-in
- [ ] This forces full transcode regardless of video codec compatibility
- [ ] Update PlaybackDecision: if image sub selected → force transcode
- **File:** `internal/playback/engine.go` - Update

### 5. Dual Subtitle Rendering (Frontend)
- [ ] Custom subtitle overlay component (NOT browser native `<track>`)
- [ ] Render 2 VTT tracks simultaneously:
  - Primary subtitle: bottom of screen (white, larger)
  - Secondary subtitle: top-bottom offset or above primary (yellow, smaller)
- [ ] Parse VTT in JS: sync cues to `video.currentTime`
- [ ] User selects: Primary language + Secondary language (or "Off")
- [ ] Customizable: font size, color, background opacity, position
- [ ] Use case: language learning (e.g., Japanese primary + Vietnamese secondary)
- **File:** `webapp/src/components/DualSubtitleOverlay.tsx` - NEW

### 6. Frontend: Subtitle & Audio Selection Flow
- [ ] Fetch subtitle + audio track lists on player mount
- [ ] Subtitle picker: groups by language, supports dual selection
- [ ] Audio picker: switch audio track (triggers HLS mode if non-default)
- [ ] Text subs: fetch VTT, render via DualSubtitleOverlay
- [ ] Image subs: selecting triggers server-side burn-in → switch to HLS stream
- [ ] Show indicator: "Subtitle requires transcoding" for image subs (PGS/VobSub)
- [ ] Default to user's preferred language (audio + subtitle)
- **File:** `webapp/src/components/SubtitlePicker.tsx`, `webapp/src/components/AudioPicker.tsx` - NEW

## Files to Create/Modify
- `pkg/subtitle/extract.go` - NEW
- `pkg/subtitle/convert.go` - NEW
- `internal/handler/subtitle.go` - Major update
- `internal/playback/engine.go` - already handles burn-in decision; needs `-filter_complex` args
- `webapp/src/components/DualSubtitleOverlay.tsx` - NEW
- `webapp/src/components/SubtitlePicker.tsx` - NEW
- `webapp/src/components/AudioPicker.tsx` - NEW

---
✅ End of Plan D
Next Plan: plan-e-streaming-enhancement
