# Phase 03: Video Player
Status: ⬜ Pending
Plan: C - Web App MVP
Dependencies: Phase 01, 02

## Mục tiêu
Fullscreen video player với subtitle selection, resume, và progress sync.
Đây là core UX của media server - phải mượt.

## Tasks

### 1. Player Page
- [ ] Route: `/play/:id`
- [ ] Fullscreen layout (no navbar/sidebar)
- [ ] Back button overlay (top-left) → return to detail page
- [ ] Show media title overlay (auto-hide after 3s)
- **File:** `src/pages/PlayerPage.tsx`

### 2. Custom Video Player Component (hls.js)
- [ ] `src/components/VideoPlayer.tsx`
- [ ] **Custom controls** - hide native browser controls (`controls={false}`)
- [ ] Build: seekbar/scrubber, play/pause, volume, fullscreen, time display
- [ ] Style with Tailwind - dark theme, Netflix-like aesthetic
- [ ] Direct Play: `<video src="/api/stream/{id}?token=...">` for MP4/H.264
- [ ] HLS: `hls.js` for transcoded/multi-audio → `/api/stream/{id}/hls/master.m3u8?token=...`
- [ ] Auto-detect: try direct play first, fallback to HLS
- [ ] Auto-hide controls after 3s inactivity, show on mouse move

### 3. Resume Playback
- [ ] Fetch progress on mount: `GET /api/progress/{id}`
- [ ] If position > 0 and < 90%: show "Resume from X:XX:XX" prompt
- [ ] User choice: Resume / Start from Beginning
- [ ] `video.currentTime = position` khi resume

### 4. Progress Sync
- [ ] Report progress to backend mỗi 10 giây: `PUT /api/progress/{id}`
- [ ] Save on pause, on seek, on visibility change (tab switch)
- [ ] Mark completed when position > 90% duration
- [ ] Debounce: don't send if position hasn't changed

### 5. Subtitle Track Selection (MVP - single sub)
- [ ] Fetch available subtitles: `GET /api/media/{id}/subtitles`
- [ ] Subtitle picker UI (CC button in player controls)
- [ ] MVP: single subtitle via `<track>` elements
- [ ] Remember last selected language (user preference)
- [ ] "Off" option to disable subtitles
- [ ] Note: dual subtitle + custom overlay renderer is Plan D Phase 03

### 6. Keyboard Shortcuts
- [ ] Space: play/pause
- [ ] Left/Right arrow: seek -10s/+10s
- [ ] Up/Down: volume
- [ ] F: fullscreen toggle
- [ ] M: mute toggle
- [ ] Escape: exit fullscreen / back to detail

### 7. Next Episode (Series)
- [ ] If playing an episode: show "Next Episode" overlay 30s before end
- [ ] Auto-play next episode after 10s countdown (configurable)
- [ ] Show: next episode title + thumbnail
- [ ] Skip to next immediately button

## Files to Create
- `src/pages/PlayerPage.tsx`
- `src/components/VideoPlayer.tsx`
- `src/components/SubtitlePicker.tsx`
- `src/components/NextEpisode.tsx`
- `src/hooks/useProgress.ts`
- `src/hooks/useKeyboardShortcuts.ts`

---
Next: phase-04-home-polish.md
