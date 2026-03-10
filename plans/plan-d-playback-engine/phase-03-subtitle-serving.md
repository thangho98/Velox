# Phase 03: Subtitle Burn-in & Serving
Status: ⬜ Pending
Plan: D - Playback Decision Engine
Dependencies: Plan A Phase 05, Phase 02

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

### 1. Subtitle Extractor
- [ ] Func `ExtractSubtitle(videoPath, streamIndex, outputDir) (string, error)`
- [ ] Text subs: `ffmpeg -i video.mkv -map 0:s:{index} -c:s webvtt output.vtt`
- [ ] SRT→VTT converter (Go native, no FFmpeg needed for external .srt)
- [ ] ASS→VTT: extract via FFmpeg
- [ ] Cache extracted subs: `data/subtitles/{media_file_id}/{stream_index}.vtt`
- **File:** `pkg/subtitle/extract.go` - NEW

### 2. SRT → VTT Converter (Native Go)
- [ ] Parse SRT format: sequence number, timestamp `00:01:23,456 --> 00:01:25,789`, text
- [ ] Output VTT: `WEBVTT` header, timestamp `00:01:23.456 --> 00:01:25.789`, text
- [ ] Handle: BOM, CRLF, empty lines, HTML tags in text
- **File:** `pkg/subtitle/convert.go` - NEW

### 3. Subtitle Serving API
- [ ] `GET /api/media/{id}/subtitles/{subID}/serve` → serve VTT file
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
- **File:** `src/components/DualSubtitleOverlay.tsx` - NEW

### 6. Frontend: Subtitle & Audio Selection Flow
- [ ] Fetch subtitle + audio track lists on player mount
- [ ] Subtitle picker: groups by language, supports dual selection
- [ ] Audio picker: switch audio track (triggers HLS mode if non-default)
- [ ] Text subs: fetch VTT, render via DualSubtitleOverlay
- [ ] Image subs: selecting triggers server-side burn-in → switch to HLS stream
- [ ] Show indicator: "Subtitle requires transcoding" for image subs (PGS/VobSub)
- [ ] Default to user's preferred language (audio + subtitle)
- **File:** `src/components/SubtitlePicker.tsx`, `src/components/AudioPicker.tsx` - NEW

## Files to Create/Modify
- `pkg/subtitle/extract.go` - NEW
- `pkg/subtitle/convert.go` - NEW
- `internal/handler/subtitle.go` - Major update
- `internal/playback/engine.go` - Subtitle + audio decision
- `src/components/DualSubtitleOverlay.tsx` - NEW
- `src/components/SubtitlePicker.tsx` - NEW
- `src/components/AudioPicker.tsx` - NEW

---
✅ End of Plan D
Next Plan: plan-e-streaming-enhancement
