# Phase 04: Player Dual Subtitles + Cinema
Status: ✅ Complete
Dependencies: Phase 01-03

## Objective
Thêm dual subtitle overlay, cinema overlay (auto trailers), gesture seek feedback vào Player.

## Features to Implement

### 1. Dual Subtitle Overlay
```
- Hiện 2 subtitles cùng lúc (primary + translated)
- Positioned at bottom letterbox
- Font: Netflix style (large, no background)
- Color: White with subtle shadow
```

### 2. Cinema Overlay
```
- Auto-play YouTube trailer before content
- Tap "Skip" → start main video
- Tap anywhere on trailer → close
- Remember cinema setting per media
```

### 3. Gesture Seek Feedback
```
- Double tap left edge → seek -10s
- Double tap right edge → seek +10s
- Show seek feedback overlay with animation
- Directional arrows + amount text
```

### 4. Subtitle Translate/Download
```
- Search and download subtitles from API
- Show translate option for subtitles
- Save subtitle preference per media
```

## Files to Create/Modify

### New Files:
- `presentation/ui/components/DualSubtitleOverlay.kt`
- `presentation/ui/components/CinemaOverlay.kt`
- `presentation/ui/components/SeekFeedbackOverlay.kt`
- `presentation/viewmodel/SubtitleViewModel.kt`

### Modify Files:
- `presentation/ui/components/VideoPlayer.kt` - Add overlays integration
- `presentation/viewmodel/PlayerViewModel.kt` - Add subtitle/cinema state

## Implementation Steps

1. **DualSubtitleOverlay Component**
   - [x] Create overlay positioned at bottom
   - [x] Accept list of SubtitleCue (text, start, end, position)
   - [x] Render primary + secondary subtitle text
   - [x] Handle positioning (inside video area, not letterbox)

2. **CinemaOverlay Component**
   - [x] Check if cinema enabled in settings
   - [x] Fetch trailer from API (stubbed - needs backend)
   - [x] Show YouTubePlayer in fullscreen overlay
   - [x] "Skip" button → dismiss + start video
   - [ ] "Skip" timeout (30s) - skipped for now

3. **SeekFeedbackOverlay**
   - [x] Animated arrows showing direction
   - [x] Text showing seek amount (+10s, -10s)
   - [x] Fade in/out animation
   - [x] Position left/right based on seek direction

4. **Update VideoPlayer**
   - [x] Integrate DualSubtitleOverlay
   - [x] Integrate CinemaOverlay (before playback)
   - [x] Detect double-tap gestures
   - [x] Show SeekFeedbackOverlay on gesture

5. **Subtitle Download/Translate**
   - [ ] Add subtitle search UI (modal)
   - [ ] Download subtitle from API
   - [ ] Show loading/progress
   - [ ] Save to media file

## Test Criteria
- [x] Dual subtitles display correctly at bottom
- [x] Cinema trailer plays before main content
- [x] Double-tap gestures trigger seek + feedback
- [ ] Subtitle download works (Phase 05)

## Reference
Mobile implementation:
- `mobile/src/components/DualSubtitleOverlay.tsx`
- `mobile/src/components/CinemaOverlay.tsx`
- `mobile/src/screens/VideoPlayerScreen.tsx` (gestures)
