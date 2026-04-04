# Phase 05: Player Core + Progress Sync
Status: ⬜ Pending
Dependencies: Phase 04

## Objective
Build the first real native playback path on top of Media3, preserving backend-driven playback decisions and reusing lessons from the existing Kotlin player module.

## Core Principles
- Backend remains source of truth for playback method and stream URL.
- Player internals are native Kotlin, not bridged from RN.
- Keep the first version focused on reliable playback before advanced polish.
- Preserve Velox-specific playback advantages:
  - do not hide pretranscode availability
  - do not assume transcode means low quality
  - do not break the audio-only fallback strategy

## Tasks
1. [ ] Port playback DTOs from current TS contracts:
   - playback info
   - stream urls
   - audio tracks
   - subtitle tracks
   - quality options
   - progress payloads
2. [ ] Create `PlaybackRepository`:
   - get playback info
   - get stream URL/token where needed
   - update progress
3. [ ] Extract or rewrite the current `VeloxExoPlayer` logic into `core:player`.
4. [ ] Decide final player wrapper API:
   - play/pause
   - seek
   - observe duration/current position/buffer
   - observe tracks
   - observe errors
5. [ ] Build fullscreen player screen in Compose:
   - video surface
   - loading state
   - error state
   - basic play/pause/seek
6. [ ] Implement resume logic:
   - server progress first
   - local fallback if needed
7. [ ] Implement progress sync:
   - periodic save
   - on pause/background
   - on exit
8. [ ] Implement quality source selection using backend playback info.
9. [ ] Support movie and episode entry paths.
10. [ ] Add orientation/system UI behavior for phone/tablet playback.
11. [ ] Add unit/integration tests around playback request building and progress sync.
12. [ ] Validate playback on real sample files:
   - MP4 direct play
   - MKV direct play
   - HLS fallback

## Verified Contract Sources
- `packages/shared/hooks/media/usePlayback.ts`
- `packages/shared/types/playback.ts`
- `packages/shared/hooks/media/useProgress.ts`
- `mobile/modules/velox-player/android/VeloxExoPlayer.kt`
- `mobile/modules/velox-player/android/VeloxPlayerView.kt`

## Implementation Steps

### 1. Port Playback Contracts
- [ ] `PlaybackInfoRequest`
- [ ] `PlaybackInfo`
- [ ] `StreamUrls`
- [ ] `PlaybackAudioTrack`
- [ ] `PlaybackSubtitleTrack`
- [ ] `QualityOption`
- [ ] progress DTOs

### 2. Build Player Request Builder
- [ ] define Android capability profile
- [ ] include codec/container/max-height information
- [ ] include selected audio/subtitle fields when present
- [ ] keep builder testable outside UI

### 3. Build Native Player Engine
- [ ] move Media3 integration into `core:player`
- [ ] expose playback state as `StateFlow`
- [ ] expose tracks, duration, current time, errors
- [ ] keep player lifecycle independent of Compose screen recreation

### 4. Build Player Screen
- [ ] fullscreen video surface
- [ ] loading state
- [ ] initial controls
- [ ] error state with retry/fallback path
- [ ] movie/episode title context

### 5. Build Progress Sync
- [ ] read resume position from server
- [ ] keep local in-memory/current session fallback
- [ ] periodic save
- [ ] save on pause/background/dispose

### 6. Validate Fallback Behavior
- [ ] direct play path
- [ ] remux/direct stream path
- [ ] HLS/transcode path
- [ ] direct failure -> fallback behavior if supported

## Files / Modules to Create
- [ ] `core:model/.../playback/*.kt`
- [ ] `core:network/.../PlaybackApi.kt`
- [ ] `core:network/.../PlaybackRepositoryImpl.kt`
- [ ] `core:player/.../VeloxPlayerEngine.kt`
- [ ] `core:player/.../PlayerSessionState.kt`
- [ ] `core:player/.../PlaybackRequestBuilder.kt`
- [ ] `feature:player/.../PlayerViewModel.kt`
- [ ] `feature:player/.../PlayerScreen.kt`

## Test Criteria
- [ ] playback request builder sends expected fields
- [ ] resume position works after app restart
- [ ] progress PUT format matches backend expectation
- [ ] direct play, direct stream, and HLS paths are all manually verified
- [ ] local player survives rotation/background transitions expected for Android

## Done When
- [ ] Android users can watch real media daily from the native client.

## Verification
- [ ] User can start playback from movie detail and episode detail.
- [ ] Seek/play/pause/resume work reliably.
- [ ] Progress persists and resumes after app restart.
- [ ] Player errors surface clearly and do not soft-lock the app.

## Exit Criteria
- Native player is good enough for internal daily use on Android mobile.
