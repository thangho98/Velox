import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react'
import Hls from 'hls.js'
import { LuMaximize2, LuPlay, LuX } from 'react-icons/lu'
import { useToast } from '@/components/Toast'
import type { LiveChannel, LiveProgram } from '@/hooks/stores/useLiveTV'
import { useFullscreen } from '@/hooks/useFullscreen'
import { useAuthStore } from '@/stores/auth'
import { getLiveStreamUrl } from '@velox/shared/api'

interface LiveTVPlayerProps {
  channel: LiveChannel | null
  currentProgram?: LiveProgram | null
  onWatch?: (channel: LiveChannel, autoFullscreen?: boolean) => void
  className?: string
}

export interface LiveTVPlayerHandle {
  play: () => void
  pause: () => void
  toggle: () => void
}

function formatTimeRange(start: string, end: string) {
  const fmt = (iso: string) =>
    new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  return `${fmt(start)} – ${fmt(end)}`
}

export const LiveTVPlayer = forwardRef<LiveTVPlayerHandle, LiveTVPlayerProps>(function LiveTVPlayer(
  { channel, currentProgram, onWatch, className = '' },
  ref,
) {
  const { info: showToastInfo } = useToast()
  const containerRef = useRef<HTMLDivElement>(null)
  const videoRef = useRef<HTMLVideoElement>(null)
  const hlsRef = useRef<Hls | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isBuffering, setIsBuffering] = useState(false)
  const [isPaused, setIsPaused] = useState(false)

  const { isFullscreen, toggleFullscreen } = useFullscreen(containerRef, videoRef, showToastInfo)

  useImperativeHandle(
    ref,
    () => ({
      play: () => {
        const video = videoRef.current
        if (!video) return
        video.play().catch((err) => console.warn('Play rejected', err))
      },
      pause: () => videoRef.current?.pause(),
      toggle: () => {
        const video = videoRef.current
        if (!video) return
        if (video.paused) {
          video.play().catch((err) => console.warn('Play rejected', err))
        } else {
          video.pause()
        }
      },
    }),
    [],
  )

  useEffect(() => {
    queueMicrotask(() => {
      setError(null)
      setIsBuffering(!!channel?.streamUrl)
      setIsPaused(false)
    })
  }, [channel])

  useEffect(() => {
    const video = videoRef.current
    if (!video || !channel?.id) return

    // Route manifest + segments through the backend proxy to bypass CORS
    // restrictions on upstream IPTV CDNs. Pass the access token so the proxy
    // can record "recently watched" for this user.
    const accessToken = useAuthStore.getState().accessToken
    const streamUrl = getLiveStreamUrl(channel.id, accessToken ?? undefined)
    let retryCount = 0

    const initPlayer = async () => {
      if (Hls.isSupported()) {
        const hls = new Hls({
          maxBufferLength: 30,
          maxMaxBufferLength: 60,
          manifestLoadingTimeOut: 20000,
          manifestLoadingMaxRetry: 3,
        })
        hlsRef.current = hls
        hls.attachMedia(video)

        hls.on(Hls.Events.MEDIA_ATTACHED, () => {
          hls.loadSource(streamUrl)
        })

        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          video.play().catch((e) => {
            console.warn('Play auto-rejected by browser', e)
            setIsBuffering(false)
            setIsPaused(true)
          })
        })

        hls.on(Hls.Events.ERROR, (_, data) => {
          if (data.fatal) {
            if (retryCount >= 3) {
              setError('Playback error: Stream unavailable after multiple retries.')
              hls.destroy()
              return
            }
            switch (data.type) {
              case Hls.ErrorTypes.NETWORK_ERROR:
                retryCount += 1
                console.warn(`Network error: Recovering (Attempt ${retryCount})...`)
                setIsBuffering(true)
                hls.startLoad()
                break
              case Hls.ErrorTypes.MEDIA_ERROR:
                retryCount += 1
                console.warn(`Media error: Recovering (Attempt ${retryCount})...`)
                setIsBuffering(true)
                hls.recoverMediaError()
                break
              default:
                setError('Playback error: Stream unavailable or format not supported.')
                hls.destroy()
                break
            }
          }
        })
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = streamUrl
        video.addEventListener('loadedmetadata', () => {
          video.play().catch((e) => {
            console.warn('Play auto-rejected by browser', e)
            setIsBuffering(false)
            setIsPaused(true)
          })
        })
        video.addEventListener('error', () => {
          setError('Playback error: Stream format not supported.')
        })
      } else {
        setError('HLS playback is not supported in this browser.')
      }
    }

    initPlayer()

    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy()
        hlsRef.current = null
      }
    }
  }, [channel?.id])

  const handleTogglePreview = () => {
    const video = videoRef.current
    if (!video) return
    if (video.paused) {
      video.play().catch((e) => console.warn('Play action rejected', e))
    } else {
      video.pause()
    }
  }

  if (!channel) {
    return (
      <div className={`relative flex items-center justify-center bg-black ${className}`}>
        <p className="font-medium tracking-wide text-fog-500">Select a channel to watch</p>
      </div>
    )
  }

  const metaParts = [channel.country, channel.groupTitle].filter(Boolean)

  return (
    <div ref={containerRef} className={`group relative overflow-hidden bg-black ${className}`}>
      {/* VIDEO */}
      <video
        ref={videoRef}
        className="absolute inset-0 h-full w-full bg-black object-cover"
        controls={false}
        autoPlay
        muted
        playsInline
        onPlaying={() => setIsBuffering(false)}
        onWaiting={() => setIsBuffering(true)}
        onPause={() => setIsPaused(true)}
        onPlay={() => setIsPaused(false)}
      />

      {/* Gradient overlays */}
      <div className="pointer-events-none absolute inset-0 z-10 bg-[linear-gradient(180deg,rgba(0,0,0,0.55)_0%,rgba(0,0,0,0)_24%,rgba(0,0,0,0)_45%,rgba(0,0,0,0.88)_100%)]" />
      <div className="pointer-events-none absolute inset-0 z-10 bg-[linear-gradient(90deg,rgba(0,0,0,0.5)_0%,rgba(0,0,0,0)_45%)]" />

      {/* TOP — LIVE badge only */}
      <div className="pointer-events-none absolute left-0 right-0 top-0 z-20 flex items-center justify-between p-4 sm:p-6">
        <span className="inline-flex items-center gap-2 bg-crimson-500 px-2.5 py-1 text-[10px] font-bold tracking-[0.28em] text-white shadow-card">
          <span className="live-dot h-1.5 w-1.5 rounded-full bg-white" />
          LIVE
        </span>
        <span className="hidden rounded-sm bg-black/55 px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.24em] text-fog-700 backdrop-blur-sm sm:inline-flex">
          Preview · Muted
        </span>
      </div>

      {/* ERROR overlay */}
      {error && (
        <div className="absolute inset-0 z-30 flex flex-col items-center justify-center bg-black/75 p-6 text-center">
          <div className="mb-3 rounded-full bg-crimson-500/20 p-4">
            <LuX className="text-crimson-500" size={36} />
          </div>
          <p className="mb-1 text-base font-semibold text-white">Stream Unavailable</p>
          <p className="max-w-sm text-sm text-fog-600">{error}</p>
        </div>
      )}

      {/* BUFFERING overlay */}
      {!error && isBuffering && (
        <div className="pointer-events-none absolute inset-0 z-20 flex items-center justify-center">
          <div className="h-12 w-12 animate-spin rounded-full border-4 border-ink-300 border-t-white shadow-lg" />
        </div>
      )}

      {/* Center Play when paused (autoplay-rejection fallback) */}
      {!error && isPaused && !isBuffering && (
        <button
          type="button"
          onClick={handleTogglePreview}
          className="absolute inset-0 z-20 flex items-center justify-center bg-black/40 transition-opacity"
          aria-label="Resume preview"
        >
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-white text-black shadow-pop transition-transform duration-300 hover:scale-[1.05] sm:h-20 sm:w-20">
            <LuPlay size={24} className="ml-1" fill="currentColor" />
          </span>
        </button>
      )}

      {/* BOTTOM — cinematic hero info */}
      <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 p-5 pt-20 sm:p-8 md:p-10 md:pt-24 lg:p-12 lg:pt-32">
        <div className="mx-auto flex w-full max-w-[1600px] flex-col items-start gap-6 lg:flex-row lg:items-end lg:justify-between">
          {/* Left — channel info */}
          <div className="min-w-0 flex-1">
            <p className="mb-2 text-[10px] font-bold uppercase tracking-[0.36em] text-crimson-500 sm:text-[11px]">
              Now Watching
            </p>
            <h1 className="text-[2rem] font-extrabold leading-[0.95] tracking-[-0.035em] text-white drop-shadow-xl sm:text-[2.8rem] md:text-5xl lg:text-6xl">
              {channel.name}
            </h1>
            {metaParts.length > 0 && (
              <p className="mt-3 text-xs font-semibold uppercase tracking-[0.24em] text-fog-700 sm:text-sm">
                {metaParts.join(' · ')}
              </p>
            )}
            {currentProgram && (
              <p className="mt-3 line-clamp-1 text-sm text-fog-600 sm:text-base">
                <span className="font-semibold text-white">{currentProgram.title}</span>
                <span className="mx-2 text-fog-500">·</span>
                <span className="font-medium text-fog-600">
                  {formatTimeRange(currentProgram.start_time, currentProgram.end_time)}
                </span>
              </p>
            )}
          </div>

          {/* Right — CTAs */}
          <div className="pointer-events-auto flex items-center gap-3">
            <button
              type="button"
              onClick={() => onWatch?.(channel, false)}
              className="flex h-11 items-center gap-2 rounded-sm bg-white px-6 text-sm font-bold tracking-[0.04em] text-black transition-colors hover:bg-fog-800 sm:h-12 sm:px-7"
            >
              <LuPlay fill="currentColor" size={16} />
              WATCH
            </button>
            <button
              type="button"
              onClick={() => onWatch?.(channel, true)}
              className="flex h-11 w-11 items-center justify-center rounded-sm bg-white/10 text-white backdrop-blur-sm transition-colors hover:bg-white/20 sm:h-12 sm:w-12"
              aria-label="Fullscreen watch"
              title="Fullscreen"
            >
              <LuMaximize2 size={18} />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
})
