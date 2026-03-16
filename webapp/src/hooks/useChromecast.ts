import { useState, useEffect, useCallback, useRef } from 'react'
import { api } from '@/lib/fetch'

// Google Cast SDK types (loaded externally)
declare global {
  interface Window {
    __onGCastApiAvailable?: (isAvailable: boolean) => void
    cast?: {
      framework: {
        CastContext: {
          getInstance: () => CastContext
        }
        RemotePlayerController: new (player: RemotePlayer) => RemotePlayerController
        RemotePlayer: new () => RemotePlayer
        CastContextEventType: { SESSION_STATE_CHANGED: string }
        SessionState: { SESSION_STARTED: string; SESSION_ENDED: string; SESSION_RESUMED: string }
      }
    }
    chrome?: {
      cast: {
        media: {
          MediaInfo: new (contentId: string, contentType: string) => MediaInfo
          GenericMediaMetadata: new () => GenericMediaMetadata
          LoadRequest: new (mediaInfo: MediaInfo) => LoadRequest
        }
        AutoJoinPolicy: { ORIGIN_SCOPED: string }
        Image: new (url: string) => { url: string }
      }
    }
  }
}

interface CastContext {
  setOptions: (options: Record<string, unknown>) => void
  requestSession: () => Promise<void>
  getCurrentSession: () => CastSession | null
  addEventListener: (type: string, listener: (event: { sessionState: string }) => void) => void
}

interface CastSession {
  getMediaSession: () => MediaSession | null
  loadMedia: (request: LoadRequest) => Promise<void>
  endSession: (stopCasting: boolean) => void
}

interface MediaSession {
  getEstimatedTime: () => number
  playerState: string
  media: { duration: number }
}

interface MediaInfo {
  metadata: GenericMediaMetadata | null
  streamType: string
}

interface GenericMediaMetadata {
  metadataType: number
  title: string
  images: Array<{ url: string }>
}

interface LoadRequest {
  currentTime: number
  autoplay: boolean
}

interface RemotePlayer {
  isConnected: boolean
  currentTime: number
  duration: number
  isPaused: boolean
  volumeLevel: number
}

interface RemotePlayerController {
  playOrPause: () => void
  seek: () => void
  stop: () => void
  setVolumeLevel: () => void
}

const DEFAULT_MEDIA_RECEIVER = 'CC1AD845'

export function useChromecast() {
  const [available, setAvailable] = useState(false)
  const [connected, setConnected] = useState(false)
  const [casting, setCasting] = useState(false)
  const castContextRef = useRef<CastContext | null>(null)
  const playerRef = useRef<RemotePlayer | null>(null)
  const controllerRef = useRef<RemotePlayerController | null>(null)

  useEffect(() => {
    // Wait for Cast SDK to load
    window.__onGCastApiAvailable = (isAvailable: boolean) => {
      if (!isAvailable || !window.cast || !window.chrome?.cast) return

      const context = window.cast.framework.CastContext.getInstance()
      context.setOptions({
        receiverApplicationId: DEFAULT_MEDIA_RECEIVER,
        autoJoinPolicy: window.chrome.cast.AutoJoinPolicy.ORIGIN_SCOPED,
      })

      castContextRef.current = context
      const player = new window.cast.framework.RemotePlayer()
      const controller = new window.cast.framework.RemotePlayerController(player)
      playerRef.current = player
      controllerRef.current = controller

      setAvailable(true)

      context.addEventListener(
        window.cast.framework.CastContextEventType.SESSION_STATE_CHANGED,
        (event: { sessionState: string }) => {
          const fw = window.cast!.framework
          const isConn =
            event.sessionState === fw.SessionState.SESSION_STARTED ||
            event.sessionState === fw.SessionState.SESSION_RESUMED
          setConnected(isConn)
          if (!isConn) setCasting(false)
        },
      )
    }

    // If SDK already loaded before hook mounts
    if (window.cast && window.chrome?.cast) {
      window.__onGCastApiAvailable(true)
    }
  }, [])

  const requestSession = useCallback(async () => {
    if (!castContextRef.current) return
    try {
      await castContextRef.current.requestSession()
    } catch {
      // User cancelled or no devices
    }
  }, [])

  const castMedia = useCallback(
    async (mediaId: number, title: string, posterUrl?: string, startTime = 0) => {
      const session = castContextRef.current?.getCurrentSession()
      if (!session || !window.chrome?.cast) return

      // Get stream URL with api_key
      const streamData = await api.post<{
        direct_url: string
        api_key: string
      }>(`/stream/${mediaId}/url`, {})

      const mediaInfo = new window.chrome.cast.media.MediaInfo(streamData.direct_url, 'video/mp4')
      mediaInfo.streamType = 'BUFFERED'

      const metadata = new window.chrome.cast.media.GenericMediaMetadata()
      metadata.metadataType = 0
      metadata.title = title
      if (posterUrl) {
        metadata.images = [new window.chrome.cast.Image(posterUrl)]
      }
      mediaInfo.metadata = metadata

      const request = new window.chrome.cast.media.LoadRequest(mediaInfo)
      request.currentTime = startTime
      request.autoplay = true

      try {
        await session.loadMedia(request)
        setCasting(true)
      } catch (err) {
        console.error('Cast load failed:', err)
      }
    },
    [],
  )

  const stopCasting = useCallback(() => {
    const session = castContextRef.current?.getCurrentSession()
    if (session) {
      session.endSession(true)
      setCasting(false)
    }
  }, [])

  return {
    available,
    connected,
    casting,
    requestSession,
    castMedia,
    stopCasting,
  }
}
