# Phase 03: Trickplay Thumbnails
Status: ✅ Done
Plan: E - Streaming Enhancement

## Tasks

### 1. Sprite Sheet Generation
- [ ] FFmpeg: extract frame every 10s → tile into 10x10 sprite sheets (320px wide)
- [ ] `ffmpeg -i video.mkv -vf "fps=1/10,scale=320:-1,tile=10x10" sprite_%d.jpg`
- [ ] Generate async (background task, low priority)
- [ ] Store: `data/trickplay/{media_id}/`

### 2. VTT Manifest
- [ ] Generate VTT file mapping timestamp → sprite coordinates (x,y,w,h)
- [ ] Compatible with video.js/hls.js thumbnail plugins
- [ ] `GET /api/media/{id}/trickplay/manifest.vtt`
- [ ] `GET /api/media/{id}/trickplay/{sprite}.jpg`

### 3. Frontend: Thumbnail Preview
- [ ] Show thumbnail on timeline hover/scrub
- [ ] Load sprite sheet, crop to correct position
- [ ] Smooth UX: preload sprite sheets

### 4. Config & Background Task
- [ ] `VELOX_TRICKPLAY_ENABLED=false` (default off - disk intensive)
- [ ] `VELOX_TRICKPLAY_INTERVAL=10` (seconds)
- [ ] Generate after media indexed (scheduled task)
- [ ] Track generation status, skip if already done

---
✅ End of Plan E
🎯 MILESTONE 2: Full Streaming
Next Plan: plan-f-admin-operations
