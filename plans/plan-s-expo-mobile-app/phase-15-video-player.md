# Phase 15: Video Player (ExoPlayer Direct Play)
Status: ⬜ Pending
Dependencies: Phase 14

## Objective
Fullscreen video player with expo-video. Direct Play cho hầu hết formats nhờ ExoPlayer.

## Context
- ExoPlayer (Android) supports: H.264, HEVC, VP9, AV1, AAC, AC3, EAC3, DTS, MKV, MP4, WebM
- Browser chỉ supports: H.264, (maybe HEVC), AAC, MP4 → cần transcode nhiều hơn
- Mobile gửi capabilities rộng hơn → backend trả Direct Play cho ~90% files

## API Contracts (verified from code)
```typescript
// PlaybackInfoRequest — POST /api/playback/{id}/info
interface PlaybackInfoRequest {
  video_codecs?: string[]      // ['h264', 'hevc', 'vp9', 'av1', ...]
  audio_codecs?: string[]      // ['aac', 'ac3', 'eac3', 'dts', ...]
  containers?: string[]        // ['mp4', 'mkv', 'webm', ...]
  max_height?: number          // ⚠️ ONLY max_height, NO max_width
  media_file_id?: number
  selected_audio_track?: number   // ⚠️ NOT selected_audio_track_id
  selected_subtitle?: string
  selected_subtitle_id?: number
}

// UpdateProgressRequest — PUT /api/profile/progress/{mediaId}
interface UpdateProgressRequest {
  position: number    // ⚠️ "position" NOT "position_seconds"
  completed: boolean  // ⚠️ "completed" NOT "duration_seconds"
}

// PlaybackInfo response
interface PlaybackInfo {
  position: number   // ⚠️ resume position from server — field is "position"
  skip_segments?: SkipSegment[]
  available_qualities?: QualityOption[]
  // ... stream_url, method, audio_tracks, subtitle_tracks, etc.
}
```

## Implementation Steps

### 1. Playback request builder
- [ ] Tạo `mobile/src/hooks/usePlaybackRequest.ts`:
  ```typescript
  import { mobileCapabilities } from '../platform/capabilities'
  import type { PlaybackInfoRequest } from '@velox/shared/types'

  export function buildPlaybackRequest(
    overrides?: Partial<PlaybackInfoRequest>,
  ): PlaybackInfoRequest {
    return {
      video_codecs: mobileCapabilities.videoCodecs,
      audio_codecs: mobileCapabilities.audioCodecs,
      containers: mobileCapabilities.containers,
      max_height: mobileCapabilities.maxResolution.height, // ← max_height only
      ...overrides,
    }
  }
  ```

### 2. Watch Screen
- [ ] `mobile/app/watch/[id].tsx`:
  ```typescript
  import { usePlaybackInfo, useUpdateProgress } from '@velox/shared/hooks/media'
  import { usePlayerStore } from '../../src/stores/player'
  import { buildPlaybackRequest } from '../../src/hooks/usePlaybackRequest'
  import * as ScreenOrientation from 'expo-screen-orientation'
  import { StatusBar } from 'expo-status-bar'

  export default function WatchScreen() {
    const { id } = useLocalSearchParams<{ id: string }>()
    const mediaId = Number(id)
    const request = buildPlaybackRequest()
    const { data: playbackInfo } = usePlaybackInfo(mediaId, request)
    const lastPosition = usePlayerStore((s) => s.getLastPosition(mediaId))

    // Lock orientation to landscape
    useEffect(() => {
      ScreenOrientation.lockAsync(ScreenOrientation.OrientationLock.LANDSCAPE)
      return () => { ScreenOrientation.unlockAsync() }
    }, [])

    // Resume position — single source of truth:
    // playbackInfo.position is the server-side position (already fetched in the
    // playback info call). Use local lastPosition only as instant fallback while
    // playbackInfo is loading (avoids extra GET /profile/progress call).
    const resumeFrom = playbackInfo?.position ?? lastPosition ?? 0

    if (!playbackInfo) return <LoadingScreen />

    return (
      <>
        <StatusBar hidden />
        <VideoPlayer
          playbackInfo={playbackInfo}
          mediaId={mediaId}
          resumeFrom={resumeFrom}
        />
      </>
    )
  }
  ```

### 3. VideoPlayer component
- [ ] Tạo `mobile/src/components/VideoPlayer.tsx`:
  ```typescript
  import { useVideoPlayer, VideoView } from 'expo-video'
  import { useAuthStore } from '../stores/auth'
  import { getPlatform } from '@velox/shared/platform'

  export function VideoPlayer({ playbackInfo, mediaId, resumeFrom }) {
    const token = useAuthStore((s) => s.accessToken)

    // Build absolute stream URL (mobile needs full URL, not relative)
    const streamUrl = useMemo(() => {
      const url = playbackInfo.stream_url
      if (url.startsWith('http')) return url
      // Relative URL → prepend server base (without /api suffix)
      const apiBase = getPlatform().getApiBaseUrl()
      const serverBase = apiBase.replace(/\/api$/, '')
      return `${serverBase}${url}`
    }, [playbackInfo.stream_url])

    const player = useVideoPlayer(streamUrl, (player) => {
      player.currentTime = resumeFrom
      player.play()
    })

    // ⚠️ Auth for stream URL:
    // Option A (recommended): api_key query param
    //   POST /api/stream/{id}/url → returns { direct_url, hls_url, api_key, expires_in }
    //   direct_url/hls_url already contain ?api_key=... — pass to player directly
    // Option B: Bearer header (if expo-video supports custom headers)

    return (
      <VideoView
        player={player}
        style={{ flex: 1 }}
        nativeControls={false}  // Custom controls in Phase 16
      />
    )
  }
  ```

### 4. Progress saving — correct contract
- [ ] Save progress every 10s:
  ```typescript
  const updateProgress = useUpdateProgress()
  const setLastPosition = usePlayerStore((s) => s.setLastPosition)

  useEffect(() => {
    const interval = setInterval(() => {
      const currentTime = player.currentTime
      const duration = player.duration
      if (currentTime > 0) {
        // ⚠️ Correct API contract: { position, completed }
        updateProgress.mutate({
          mediaId,
          data: {
            position: Math.floor(currentTime),
            completed: duration > 0 && currentTime >= duration * 0.9,
          },
        })
        // Save locally for instant resume
        setLastPosition(mediaId, currentTime)
      }
    }, 10000)
    return () => clearInterval(interval)
  }, [player, mediaId])

  // Save on unmount
  useEffect(() => {
    return () => {
      const currentTime = player.currentTime
      const duration = player.duration
      if (currentTime > 0) {
        updateProgress.mutate({
          mediaId,
          data: {
            position: Math.floor(currentTime),
            completed: duration > 0 && currentTime >= duration * 0.9,
          },
        })
        setLastPosition(mediaId, currentTime)
      }
    }
  }, [])
  ```

### 5. Background audio
- [ ] Đã configure trong `app.json` (Phase 09):
  ```json
  "ios": { "infoPlist": { "UIBackgroundModes": ["audio"] } }
  ```
- [ ] expo-video supports background audio natively on Android
- [ ] Install `expo-keep-awake` to prevent screen sleep:
  ```typescript
  import { useKeepAwake } from 'expo-keep-awake'
  // Inside WatchScreen:
  useKeepAwake()
  ```

## Files to Create
- `mobile/src/hooks/usePlaybackRequest.ts`
- `mobile/app/watch/[id].tsx`
- `mobile/src/components/VideoPlayer.tsx`

## Dependencies (npm packages)
- `expo-video` — video playback (ExoPlayer on Android)
- `expo-screen-orientation` — lock to landscape
- `expo-keep-awake` — prevent screen sleep

## Test Criteria
- [ ] Direct Play works for H.264 MP4 files
- [ ] Direct Play works for HEVC MKV files (ExoPlayer handles natively)
- [ ] HLS fallback works when backend returns transcode method
- [ ] Landscape lock during playback
- [ ] Progress saves every 10s: `{ position: N, completed: false }` format
- [ ] Progress saves on exit (back button)
- [ ] Resume from `playbackInfo.position` (server truth), with local `lastPosition` as instant fallback
- [ ] Background audio continues when app backgrounded

---
Next Phase: [phase-16-player-controls.md](phase-16-player-controls.md)
