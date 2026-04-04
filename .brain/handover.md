
# HANDOVER DOCUMENT

**Date:** 2026-04-04 (Session 2)

## Current State: Mobile App polish + Video Playback Fix

### Done this session (2026-04-04, Session 2):

#### 1. MediaDetailScreen — Full Rewrite (Match Web Layout)
   - **Fixed backdrop**: full-screen, fixed position behind ScrollView (not scrolling)
   - **Poster centered**: large (200x300) centered on backdrop, not left-aligned
   - **Title/metadata left-aligned**: matching web (was centered)
   - **Tech specs inline**: "Video 2076p HEVC  Audio EAC3  Container MATROSKA,WEBM  Size 34.62 GB" (was grid section)
   - **Overview inline**: no section header
   - **Compact Play button**: small with icon buttons inline (was full-width)
   - **Inline subtitle selector**: dropdown next to action buttons (was separate section)
   - **Removed**: PlayOptionsBottomSheet (play goes directly to video), separate subtitle section (CC/Search/Translate)
   - **Fixed "Ends at"**: was epoch-based, now correctly calculates `now + remaining`
   - **Gradient**: heavy overlay matching web cinematic feel (`expo-linear-gradient`)
   - **Subtitle label**: uses `subtitle.label` (like web) + matches by `subtitle.id` (not language)

#### 2. SeriesDetailScreen — Layout Update
   - **Poster centered**: matching MediaDetailScreen and web layout
   - **Title/metadata below poster**: was side-by-side
   - **Edit badge**: pencil icon + "Edit" text next to network name
   - **Removed custom header bar**: uses React Navigation transparent header
   - **Menu button**: moved to navigation header right side via `useLayoutEffect`

#### 3. Navigation Fixes
   - **Transparent header**: for both Media and SeriesDetail screens (backdrop shows through)
   - **Fixed SeriesDetail navigation**: renamed `Series` → `SeriesDetail` in RootStackParamList
   - **Added root-level screens**: Media + SeriesDetail in RootStack (was only in tab stacks)
   - **Updated all navigate calls**: `navigate('SeriesDetail', { id })` across 6+ files

#### 4. Video Playback — Major Fix (HEVC 4K + EAC3)
   - **Problem**: HEVC Main 10 4K + EAC3 pre-transcoded MP4 fails on Android tablet (black screen)
   - **Root cause**: expo-video can't decode HEVC Main 10 at 4K, PreTranscode MP4 served directly
   - **Mobile fix**: Fallback chain — direct play → HLS transcode (via `player.replaceAsync`)
   - **`fallbackRef`**: synchronous ref to avoid poll race conditions
   - **`toHlsUrl`**: converts direct URL to HLS with `mh=` param for resolution
   - **try-catch in pollState**: prevents "shared object released" crash during fallback
   - **Audio codecs**: added `ac3`, `eac3` to supported list (ExoPlayer supports them)
   - **Default `max_height: 1080`**: for auto mode on mobile

#### 5. Backend — Playback Quality Pipeline
   - **Pretranscode exact match**: only use pretranscode when height matches user's selection exactly (or auto mode)
   - **`mh` param in HLS URL**: server passes max_height to FFmpeg via `?mh=1440`
   - **`maxHeight` through HLS chain**: stream.go → PrepareHLS → GenerateHLS → runHLSFFmpeg
   - **FFmpeg scale filter**: `-vf scale=-2:{maxHeight}` when maxHeight > 0 and not videoCopy
   - **PreTranscode HLS fallback**: when client doesn't support MP4 but supports HLS
   - **`useStreamUrls`**: added `PreTranscode` to isHLS detection

#### 6. Quality Selection — Full Pipeline
   - **`maxQuality` state**: moved before `playbackRequest` (was declared after)
   - **`playbackRequest` reactive**: recalculates when `maxQuality` changes
   - **Quality change flow**: maxQuality → playbackRequest → useStreamUrls refetch → videoSource → replaceAsync
   - **Fallback reset**: quality change resets `fallbackRef` to 'normal'

#### 7. DualSubtitleOverlay — Dynamic Position
   - **Landscape**: `bottom: 14%` (low, near controls)
   - **Portrait**: `bottom: 22%` (higher, avoids controls overlap)
   - **Uses `useWindowDimensions`** for orientation detection

#### 8. Dockerfile Fix
   - **pnpm workspace**: Dockerfile now uses pnpm instead of npm ci
   - **Copies shared packages**: `packages/shared/` included in frontend build stage

### Files Changed (32 files):
- Backend: playback_info.go, stream.go, transcoder_hls.go, profile.go
- Mobile: MediaDetailScreen, SeriesDetailScreen, VideoPlayerScreen, App.tsx, DualSubtitleOverlay, + navigation files
- Shared: usePlayback.ts, useEpisodesProgress.ts, useProgress.ts
- Web: WatchPage.tsx, WatchPlaybackStatsOverlay.tsx
- Config: Dockerfile, pnpm-lock.yaml

### Commits:
- None yet (32 files changed, uncommitted)

### Pending:
1. Commit all changes
2. Verify 1440p transcode works after backend deploy
3. Settings screen update (match web)
4. Test quality selection end-to-end on mobile

### Important notes for next session:
- Admin Playback Mode must be "Auto" (not "Direct Play") for quality selection to work
- Backend needs rebuild+deploy for playback changes to take effect
- `expo-video` on Android: HEVC Main 10 4K fails silently (status: error, no error message)
- Pre-transcoded MP4 files may fail on mobile → always need HLS fallback
- Docker build: use pnpm (not npm) because of workspace:* dependencies
