/**
 * Chromecast hook for React Native
 *
 * Wraps react-native-google-cast to provide a simple API for discovering,
 * connecting, and casting media to Chromecast devices.
 *
 * Uses the Default Media Receiver (CC1AD845).
 */

import { useState, useEffect, useCallback } from 'react'
import GoogleCast, {
  CastState,
  useRemoteMediaClient,
  useCastState,
  useCastSession,
  useMediaStatus,
  MediaPlayerState,
} from 'react-native-google-cast'
import { api } from '@velox/shared/api'

interface StreamUrlResponse {
  direct_url: string
  hls_url: string
  token: string
  expires_in: number
}

interface CastMediaOptions {
  mediaId: number
  title: string
  posterUrl?: string
  startTime?: number
}

export function useChromecast() {
  const client = useRemoteMediaClient()
  const castState = useCastState()
  const session = useCastSession()
  const mediaStatus = useMediaStatus()

  const [isCastingMedia, setIsCastingMedia] = useState(false)
  const [castError, setCastError] = useState<string | null>(null)

  const isConnected = castState === CastState.CONNECTED
  const isAvailable =
    castState !== CastState.NO_DEVICES_AVAILABLE && castState !== CastState.NOT_CONNECTED

  // Reset casting state when session ends
  useEffect(() => {
    if (!isConnected) {
      setIsCastingMedia(false)
    }
  }, [isConnected])

  // Derive current playback state from media status
  const isPlaying = mediaStatus?.playerState === MediaPlayerState.PLAYING
  const isPaused = mediaStatus?.playerState === MediaPlayerState.PAUSED
  const isBuffering = mediaStatus?.playerState === MediaPlayerState.BUFFERING
  const currentTime = mediaStatus?.streamPosition ?? 0
  const duration = mediaStatus?.mediaInfo?.streamDuration ?? 0
  const deviceName = (session as any)?.device?.friendlyName ?? null

  /** Show the native Cast device picker dialog */
  const showCastPicker = useCallback(() => {
    GoogleCast.showCastDialog()
  }, [])

  /** Cast a media item to the connected Chromecast device */
  const castMedia = useCallback(
    async ({ mediaId, title, posterUrl, startTime = 0 }: CastMediaOptions) => {
      if (!client) {
        setCastError('No cast device connected')
        return
      }

      setCastError(null)

      try {
        // Fetch stream URL with temporary api_key from backend
        const streamData = await api.post<StreamUrlResponse>(`/stream/${mediaId}/url`, {})

        const contentUrl = streamData.hls_url || streamData.direct_url
        const contentType = streamData.hls_url
          ? 'application/x-mpegURL'
          : 'video/mp4'

        // Append token as query param for authentication
        const separator = contentUrl.includes('?') ? '&' : '?'
        const authenticatedUrl = `${contentUrl}${separator}token=${streamData.token}`

        await client.loadMedia({
          mediaInfo: {
            contentUrl: authenticatedUrl,
            contentType,
            metadata: {
              type: 'generic',
              title,
              images: posterUrl ? [{ url: posterUrl }] : undefined,
            },
            streamDuration: undefined, // Let the receiver determine duration
          },
          startTime,
          autoplay: true,
        })

        setIsCastingMedia(true)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to cast media'
        setCastError(message)
        console.error('Cast load failed:', err)
      }
    },
    [client],
  )

  /** Play/pause remote playback */
  const togglePlayPause = useCallback(async () => {
    if (!client) return

    if (isPlaying) {
      await client.pause()
    } else {
      await client.play()
    }
  }, [client, isPlaying])

  /** Seek to a specific time (seconds) on the cast device */
  const seekTo = useCallback(
    async (time: number) => {
      if (!client) return
      await client.seek({ position: time })
    },
    [client],
  )

  /** Seek forward/backward by a delta (seconds) */
  const seekBy = useCallback(
    async (delta: number) => {
      if (!client) return
      const newTime = Math.max(0, Math.min(currentTime + delta, duration))
      await client.seek({ position: newTime })
    },
    [client, currentTime, duration],
  )

  /** Stop casting and disconnect */
  const stopCasting = useCallback(async () => {
    try {
      if (client) {
        await client.stop()
      }
      await GoogleCast.sessionManager.endCurrentSession(true)
    } catch {
      // Session may already be ended
    }
    setIsCastingMedia(false)
  }, [client])

  return {
    // Connection state
    available: isAvailable,
    connected: isConnected,
    casting: isCastingMedia,
    deviceName,
    error: castError,

    // Playback state (remote)
    isPlaying,
    isPaused,
    isBuffering,
    currentTime,
    duration,

    // Actions
    showCastPicker,
    castMedia,
    togglePlayPause,
    seekTo,
    seekBy,
    stopCasting,
  }
}
