# Phase 16: Player Controls + Subtitles + Quality
Status: ⬜ Pending
Dependencies: Phase 15

## Objective
Custom controls overlay, subtitle/audio/quality pickers, skip intro, next episode, gestures, PiP.

## Implementation Steps

### 1. Controls Overlay
- [ ] Tạo `mobile/src/components/PlayerControls.tsx`:
  - Tap video → toggle controls visibility
  - Auto-hide after 3s of no interaction
  - **Top bar:** Back button, title, settings icon
  - **Center:** Play/pause button (large), seek ±10s (smaller, left/right)
  - **Bottom bar:** Progress scrubber (seekable timeline), current time / total time
  - Semi-transparent dark overlay behind controls
  - Animated show/hide (opacity + translateY)

### 2. Subtitle Picker
- [ ] Tạo `mobile/src/components/SubtitlePicker.tsx`:
  - Bottom sheet (use `@gorhom/bottom-sheet` or simple Modal)
  - List available subtitles from `playbackInfo.subtitle_tracks`
  - Each item: language name (from `@velox/shared/lib/languages`) + format badge (SRT/ASS/PGS)
  - "Off" option at top
  - Selected state: checkmark icon
  - On selection:
    1. Save to `usePlayerStore.setSubtitleLanguage(lang, trackId)`
    2. **Refetch playback info** with `selected_subtitle_id` in request:
       ```typescript
       // PlaybackInfoRequest accepts selected_subtitle_id (number)
       // Backend returns subtitle_tracks with selection applied
       buildPlaybackRequest({
         selected_subtitle_id: trackId,
         selected_audio_track: currentAudioTrackId,
       })
       ```
    3. ExoPlayer renders SRT/ASS/PGS natively — no custom rendering needed
  - ⚠️ Webapp sends `selected_subtitle_id` in PlaybackInfoRequest to tell backend which subtitle to serve. Mobile must do the same.

### 3. Audio Track Picker
- [ ] Tạo `mobile/src/components/AudioPicker.tsx`:
  - Bottom sheet (same style as subtitle picker)
  - List audio tracks from `playbackInfo.audio_tracks`
  - Each item: language name + codec badge (AAC/AC3/DTS) + "selected" checkmark
  - On selection:
    1. Save to `usePlayerStore.setAudioTrack(lang, trackId)`
    2. **Refetch playback info** with `selected_audio_track` in request:
       ```typescript
       // ⚠️ Field is "selected_audio_track" (NOT "selected_audio_track_id")
       buildPlaybackRequest({
         selected_audio_track: trackId,
         selected_subtitle_id: currentSubtitleId,
       })
       ```
    3. Changing audio track may change transcode method → new stream URL

### 4. Quality Picker
- [ ] Tạo `mobile/src/components/QualityPicker.tsx`:
  - Bottom sheet
  - `playbackInfo.available_qualities` list
  - Each item: resolution label (e.g. "1080p"), ⚡ badge for pretranscode
  - "Auto" option at bottom with separator
  - Save to `usePlayerStore.setMaxQuality()`
  - Changing quality → rebuild playback request → refetch → new stream URL

### 5. Skip Intro / Credits
- [ ] Trong PlayerControls, check `playbackInfo.skip_segments`:
  ```typescript
  const introSegment = playbackInfo.skip_segments?.find(s => s.type === 'intro')
  const creditsSegment = playbackInfo.skip_segments?.find(s => s.type === 'credits')

  // Show "Skip Intro" button when currentTime is within intro segment
  if (introSegment && currentTime >= introSegment.start && currentTime < introSegment.end) {
    // Show skip button → seek to introSegment.end
  }
  ```
  - "Skip Intro" button: bottom-right, auto-dismiss when segment ends
  - "Skip Credits" button: same logic for credits segment

### 6. Next Episode Auto-play
- [ ] For series playback:
  - Fetch next episode info from API
  - 30s before end → show "Next Episode" overlay card:
    - Episode thumbnail + title
    - Countdown: "Playing in 10s..."
    - "Play Now" / "Cancel" buttons
  - Auto-play when countdown reaches 0 (if not cancelled)
  - Navigate to new watch screen with next episode ID

### 7. Gesture Controls
- [ ] Implement gesture handlers (use `react-native-gesture-handler`):
  - **Horizontal swipe:** Seek forward/backward (proportional to swipe distance)
  - **Vertical swipe (left half):** Brightness control
  - **Vertical swipe (right half):** Volume control
  - **Double-tap left:** Seek -10s (with ripple animation)
  - **Double-tap right:** Seek +10s (with ripple animation)
  - Visual feedback: seek amount text overlay (e.g. "+10s", "-30s")

### 8. Picture-in-Picture (PiP)
- [ ] `expo-video` VideoView hỗ trợ PiP trên Android qua 2 props:
  ```typescript
  <VideoView
    player={player}
    style={{ flex: 1 }}
    nativeControls={false}
    allowsPictureInPicture={true}
    startsPictureInPictureAutomatically={true}
    // allowsPictureInPicture: bật khả năng PiP (user có thể trigger)
    // startsPictureInPictureAutomatically: auto-enter PiP khi app ra background
    //   → mặc định false, phải set true để home button trigger PiP
  />
  ```
- [ ] Thêm PiP + background playback config trong `app.json` plugins:
  ```json
  [
    "expo-video",
    {
      "supportsBackgroundPlayback": true,
      "supportsPictureInPicture": true
    }
  ]
  ```
  Sau đó re-run `npx expo prebuild --clean` để regenerate native config.
- [ ] Behavior: khi user nhấn home button trong lúc video đang play → auto-enter PiP mode (Android 8+). PiP window shows video only, không có custom controls.
- [ ] Nếu `expo-video` PiP chưa ổn định ở thời điểm build → **defer PiP sang post-MVP**. Ghi TODO và move on — PiP là nice-to-have, không block MVP.

## Files to Create
- `mobile/src/components/PlayerControls.tsx`
- `mobile/src/components/SubtitlePicker.tsx`
- `mobile/src/components/AudioPicker.tsx`
- `mobile/src/components/QualityPicker.tsx`
- `mobile/src/components/NextEpisodeOverlay.tsx`

## Files to Modify
- `mobile/src/components/VideoPlayer.tsx` — integrate controls + pickers + gestures
- `mobile/app/watch/[id].tsx` — pass next episode info
- `mobile/app.json` — add `expo-video` plugin config (PiP + background playback)
  → re-run `npx expo prebuild --clean` after

## Dependencies (npm packages)
All already installed in Phase 09:
- `react-native-gesture-handler` — swipe gestures
- `react-native-reanimated` — animations
- `expo-brightness` — brightness gesture control
- `expo-video` — PiP support via `allowsPictureInPicture` prop

Optional (install if needed):
- `@gorhom/bottom-sheet` — bottom sheet for pickers (alternative: simple Modal)

## Test Criteria
- [ ] Controls show on tap, auto-hide after 3s
- [ ] Play/pause, seek ±10s work
- [ ] Progress scrubber: drag to seek
- [ ] Subtitle selection works (SRT, ASS, PGS rendered by ExoPlayer)
- [ ] Audio track switching works
- [ ] Quality change triggers stream URL change
- [ ] "Skip Intro" button appears at correct time range
- [ ] Next episode overlay shows 30s before end
- [ ] Auto-play next episode works
- [ ] Gestures: swipe seek, double-tap seek, volume, brightness
- [ ] PiP works on Android (home button during playback)

---
Next Phase: [phase-17-profile-settings.md](phase-17-profile-settings.md)
