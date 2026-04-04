# Phase 06: Advanced Playback + Cast
Status: ⬜ Pending
Dependencies: Phase 05

## Objective
Close the biggest playback parity gaps after the player core is stable: subtitles, audio track control, quality switching, intro skip, and Chromecast.

## Scope
- subtitle selection
- audio track selection
- aspect ratio / fit modes
- quality switching
- skip intro / credits markers
- Chromecast

## Product Notes
- Velox already has stronger playback infrastructure than Jellyfin/Emby in some areas:
  - pretranscode
  - multi-quality HLS
  - audio-only incompatibility without unnecessary video downgrade
- Native client must expose those strengths rather than flattening everything into a generic player UI.

## Tasks
1. [ ] Implement subtitle track rendering and selection on top of Media3 text tracks.
2. [ ] Implement audio track selection UI + player wiring.
3. [ ] Implement quality selection UX using backend `available_qualities`.
4. [ ] Implement aspect ratio controls:
   - contain
   - cover
   - fill
5. [ ] Implement skip intro / credits actions when backend markers exist.
6. [ ] Decide what to do with dual subtitle support in V1:
   - include if cheap and stable
   - otherwise explicitly defer
7. [ ] Decide what to do with subtitle translate/search:
   - include only if a backend-supported lightweight flow exists
   - otherwise defer beyond cutover
8. [ ] Implement native Google Cast integration:
   - device discovery
   - connect/disconnect
   - load media with authenticated URL
   - remote pause/seek/stop
9. [ ] Handle cast handoff from local player to remote session.
10. [ ] Add tests/manual scenarios for audio/subtitle/quality/cast edge cases.
11. [ ] Measure startup and switching latency on real devices.

## Verified Contract Sources
- `packages/shared/types/playback.ts`
- `packages/shared/hooks/media/usePlayback.ts`
- `packages/shared/hooks/media/useSubtitleOps.ts`
- `mobile/src/hooks/useChromecast.ts`
- `mobile/src/screens/VideoPlayerScreen.tsx`

## Implementation Steps

### 1. Subtitle / Audio / Quality
- [ ] map backend track payloads into native selection models
- [ ] implement picker UI for subtitles
- [ ] implement picker UI for audio tracks
- [ ] implement quality sheet with `Original` / `Instant` / `Live` semantics
- [ ] wire selections back into playback request builder and stream refresh path

### 2. Skip Intro / Credits
- [ ] surface backend markers from `skip_segments`
- [ ] show context-aware skip CTA
- [ ] ensure CTA disappears outside valid time windows

### 3. Aspect Ratio / Fit Modes
- [ ] contain
- [ ] cover
- [ ] fill
- [ ] persist last-used setting locally if appropriate

### 4. Fallback and Recovery
- [ ] design a retry flow when source/playback init fails
- [ ] avoid silent infinite retries
- [ ] log enough diagnostics for beta testing

### 5. Chromecast
- [ ] device discovery
- [ ] session connect/disconnect
- [ ] authenticated stream URL handoff
- [ ] remote controls for play/pause/seek/stop
- [ ] local-to-remote handoff behavior

## Files / Modules to Create
- [ ] `feature:player/.../SubtitlePicker*.kt`
- [ ] `feature:player/.../AudioTrackPicker*.kt`
- [ ] `feature:player/.../QualityPicker*.kt`
- [ ] `feature:player/.../SkipSegmentUi*.kt`
- [ ] `feature:cast/.../CastRepository*.kt`
- [ ] `feature:cast/.../CastViewModel*.kt`
- [ ] `feature:cast/.../CastButton*.kt`

## Test Criteria
- [ ] subtitle change works on real files
- [ ] audio track change works on real files
- [ ] quality switch respects backend-returned options
- [ ] skip intro/credits appears only when valid
- [ ] cast can start and stop a real session
- [ ] direct/local playback state and cast state never conflict permanently

## Done When
- [ ] Native client exposes the important playback controls people actually need, not just a bare player.

## Verification
- [ ] User can change subtitle and audio tracks during playback.
- [ ] Quality switch does not break playback session.
- [ ] Skip intro button appears only when markers exist.
- [ ] Chromecast can start and control playback on a real device.

## Exit Criteria
- Major playback parity blockers with RN are gone or explicitly deferred.
