import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router'
import Hls from 'hls.js'
import {
  LuArrowLeft,
  LuMaximize2,
  LuMinimize2,
  LuPause,
  LuPlay,
  LuVolume2,
  LuVolumeX,
  LuX,
} from 'react-icons/lu'
import { useLiveTVChannel, useLiveTVEpg } from '@/hooks/stores/useLiveTV'
import type { LiveChannel } from '@/hooks/stores/useLiveTV'
import { useFullscreen } from '@/hooks/useFullscreen'
import { useToast } from '@/components/Toast'
import { useAuthStore } from '@/stores/auth'
import { getLiveStreamUrl } from '@velox/shared/api'

const IDLE_MS = 3000

export default function WatchLivePage() {
  const { channelId } = useParams<{ channelId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const toast = useToast()

  const id = Number(channelId)
  const navState = (location.state ?? {}) as { channel?: LiveChannel; autoFullscreen?: boolean }

  // Fetch specific channel instead of searching in full list to support pagination > 100
  const { data: fetchedChannel, isLoading: isChannelLoading } = useLiveTVChannel(
    !navState.channel || navState.channel.id !== id ? id : null,
  )

  const channel = useMemo<LiveChannel | null>(() => {
    if (navState.channel && navState.channel.id === id) return navState.channel
    return fetchedChannel ?? null
  }, [navState.channel, fetchedChannel, id])

  const { data: epgPrograms = [] } = useLiveTVEpg(id || null)
  const currentProgram = useMemo(() => {
    const now = Date.now()
    return (
      epgPrograms.find((p) => {
        const s = new Date(p.start_time).getTime()
        const e = new Date(p.end_time).getTime()
        return s <= now && now < e
      }) ?? null
    )
  }, [epgPrograms])

  const containerRef = useRef<HTMLDivElement>(null)
  const videoRef = useRef<HTMLVideoElement>(null)
  const hlsRef = useRef<Hls | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isBuffering, setIsBuffering] = useState(true)
  const [isPaused, setIsPaused] = useState(false)
  const [isMuted, setIsMuted] = useState(false)
  const [showControls, setShowControls] = useState(true)

  const { isFullscreen, toggleFullscreen } = useFullscreen(containerRef, videoRef, toast.info)

  // Auto-fullscreen on mount if navigated via the expand button
  useEffect(() => {
    if (navState.autoFullscreen && !isFullscreen) {
      // Small timeout to ensure DOM the ref is attached, and the browser acknowledges the gesture
      setTimeout(() => {
        toggleFullscreen()
      }, 50)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // HLS setup
  useEffect(() => {
    const video = videoRef.current
    if (!video || !channel?.id) return

    const accessToken = useAuthStore.getState().accessToken
    const streamUrl = getLiveStreamUrl(channel.id, accessToken ?? undefined)
    let retryCount = 0

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
          console.warn('Play auto-rejected', e)
          setIsPaused(true)
          setIsBuffering(false)
        })
      })

      hls.on(Hls.Events.ERROR, (_, data) => {
        if (!data.fatal) return
        if (retryCount >= 3) {
          setError('Stream unavailable after multiple retries.')
          hls.destroy()
          return
        }
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            retryCount += 1
            setIsBuffering(true)
            hls.startLoad()
            break
          case Hls.ErrorTypes.MEDIA_ERROR:
            retryCount += 1
            setIsBuffering(true)
            hls.recoverMediaError()
            break
          default:
            setError('Stream unavailable or format not supported.')
            hls.destroy()
            break
        }
      })
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = streamUrl
      video.addEventListener('loadedmetadata', () => {
        video.play().catch((e) => {
          console.warn('Play auto-rejected', e)
          setIsPaused(true)
        })
      })
      video.addEventListener('error', () => setError('Stream format not supported.'))
    } else {
      setError('HLS playback is not supported in this browser.')
    }

    return () => {
      hlsRef.current?.destroy()
      hlsRef.current = null
    }
  }, [channel?.id])

  // 60-second connection/buffering timeout
  useEffect(() => {
    if (isBuffering && !error) {
      const timer = setTimeout(() => {
        setError('Connection timeout. The stream may be offline or unreachable.')
        if (hlsRef.current) {
          hlsRef.current.destroy()
        }
      }, 60000)
      return () => clearTimeout(timer)
    }
  }, [isBuffering, error])

  const togglePlay = useCallback(() => {
    const v = videoRef.current
    if (!v) return
    if (v.paused) {
      v.play().catch((e) => console.warn('Play rejected', e))
    } else {
      v.pause()
    }
  }, [])

  const toggleMute = useCallback(() => {
    const v = videoRef.current
    if (!v) return
    const next = !v.muted
    v.muted = next
    setIsMuted(next)
  }, [])

  const handleBack = useCallback(() => navigate('/livetv'), [navigate])

  // Idle auto-hide controls
  useEffect(() => {
    let timer: number | undefined
    const show = () => {
      setShowControls(true)
      window.clearTimeout(timer)
      timer = window.setTimeout(() => setShowControls(false), IDLE_MS)
    }
    show()
    window.addEventListener('mousemove', show)
    window.addEventListener('touchstart', show)
    window.addEventListener('keydown', show)
    return () => {
      window.clearTimeout(timer)
      window.removeEventListener('mousemove', show)
      window.removeEventListener('touchstart', show)
      window.removeEventListener('keydown', show)
    }
  }, [])

  // Keyboard shortcuts
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (isFullscreen) return
        handleBack()
      } else if (e.key === ' ' || e.key.toLowerCase() === 'k') {
        e.preventDefault()
        togglePlay()
      } else if (e.key.toLowerCase() === 'f') {
        e.preventDefault()
        toggleFullscreen()
      } else if (e.key.toLowerCase() === 'm') {
        e.preventDefault()
        toggleMute()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isFullscreen, handleBack, togglePlay, toggleFullscreen, toggleMute])

  if (!channel && isChannelLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-black">
        <div className="flex flex-col items-center gap-4">
          <div className="h-10 w-10 animate-spin rounded-full border-2 border-fog-400 border-t-white" />
          <p className="text-sm text-fog-500">Loading channel...</p>
        </div>
      </div>
    )
  }

  if (!channel) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-black p-6 text-center">
        <div className="mb-3 rounded-full bg-crimson-500/20 p-4">
          <LuX className="text-crimson-500" size={36} />
        </div>
        <p className="mb-1 text-base font-semibold text-white">Channel not found</p>
        <p className="mb-6 max-w-sm text-sm text-fog-600">
          The channel may no longer be available.
        </p>
        <button
          type="button"
          onClick={handleBack}
          className="flex h-10 items-center gap-2 rounded-sm bg-white px-5 text-sm font-bold text-black transition-colors hover:bg-fog-800"
        >
          <LuArrowLeft size={16} /> Back to channels
        </button>
      </div>
    )
  }

  const metaParts = [channel.country, channel.groupTitle].filter(Boolean)

  return (
    <div
      ref={containerRef}
      className={`fixed inset-0 flex bg-black text-white ${
        showControls ? 'cursor-default' : 'cursor-none'
      }`}
    >
      <video
        ref={videoRef}
        className="absolute inset-0 h-full w-full bg-black object-contain"
        autoPlay
        playsInline
        onPlaying={() => setIsBuffering(false)}
        onWaiting={() => setIsBuffering(true)}
        onPause={() => setIsPaused(true)}
        onPlay={() => setIsPaused(false)}
        onVolumeChange={() => setIsMuted(!!videoRef.current?.muted)}
        onClick={togglePlay}
      />

      {/* Gradients — only dim tight edges, fade out with controls */}
      <div
        className={`pointer-events-none absolute inset-0 z-10 bg-[linear-gradient(180deg,rgba(0,0,0,0.55)_0%,rgba(0,0,0,0)_12%,rgba(0,0,0,0)_72%,rgba(0,0,0,0.82)_100%)] transition-opacity duration-300 ${
          showControls ? 'opacity-100' : 'opacity-0'
        }`}
      />

      {/* ERROR */}
      {error && (
        <div className="absolute inset-0 z-40 flex flex-col items-center justify-center bg-black/80 p-6 text-center">
          <div className="mb-3 rounded-full bg-crimson-500/20 p-4">
            <LuX className="text-crimson-500" size={40} />
          </div>
          <p className="mb-2 text-base font-semibold text-white">Stream Unavailable</p>
          <p className="mb-6 max-w-sm text-sm text-fog-600">{error}</p>
          <button
            type="button"
            onClick={handleBack}
            className="flex h-10 items-center gap-2 rounded-sm bg-white px-5 text-sm font-bold text-black transition-colors hover:bg-fog-800"
          >
            <LuArrowLeft size={16} /> Back to channels
          </button>
        </div>
      )}

      {/* BUFFERING */}
      {!error && isBuffering && (
        <div className="pointer-events-none absolute inset-0 z-20 flex items-center justify-center">
          <div className="h-14 w-14 animate-spin rounded-full border-4 border-ink-300 border-t-white" />
        </div>
      )}

      {/* PAUSED center play */}
      {!error && isPaused && !isBuffering && (
        <button
          type="button"
          onClick={togglePlay}
          className="absolute inset-0 z-20 flex items-center justify-center bg-black/40"
          aria-label="Play"
        >
          <span className="flex h-20 w-20 items-center justify-center rounded-full bg-white text-black shadow-pop transition-transform duration-300 hover:scale-[1.05] sm:h-24 sm:w-24">
            <LuPlay size={28} className="ml-1" fill="currentColor" />
          </span>
        </button>
      )}

      {/* TOP BAR */}
      <div
        className={`pointer-events-none absolute inset-x-0 top-0 z-30 flex items-center justify-between p-4 transition-opacity duration-300 sm:px-8 sm:pt-6 ${
          showControls ? 'opacity-100' : 'opacity-0'
        }`}
      >
        <button
          type="button"
          onClick={handleBack}
          className="pointer-events-auto flex items-center gap-2 text-white/70 transition-colors hover:text-white"
        >
          <LuArrowLeft size={32} />
        </button>

        <span className="inline-flex items-center gap-2 bg-crimson-500 px-2.5 py-1 text-[10px] font-bold tracking-[0.28em] text-white shadow-card">
          <span className="live-dot h-1.5 w-1.5 rounded-full bg-white" />
          LIVE
        </span>
      </div>

      {/* BOTTOM — info + controls */}
      <div
        className={`pointer-events-none absolute inset-x-0 bottom-0 z-30 flex flex-col justify-end px-4 pb-4 pt-32 transition-opacity duration-[400ms] sm:px-8 sm:pb-8 sm:pt-48 ${
          showControls ? 'opacity-100' : 'opacity-0'
        }`}
      >
        <div className="mx-auto flex w-full max-w-[2000px] flex-col gap-4 sm:gap-6">
          <div className="flex items-end justify-between gap-4">
            {/* Left section: Play/Pause and Title */}
            <div className="flex items-center gap-4 sm:gap-6">
              <button
                type="button"
                onClick={togglePlay}
                className="pointer-events-auto flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-white text-black transition-transform hover:scale-105 active:scale-95 sm:h-14 sm:w-14"
                aria-label={isPaused ? 'Play' : 'Pause'}
              >
                {isPaused ? (
                  <LuPlay size={24} className="ml-1" fill="currentColor" />
                ) : (
                  <LuPause size={24} fill="currentColor" />
                )}
              </button>

              <div className="flex flex-col">
                <h1 className="line-clamp-1 text-base font-bold text-white drop-shadow-md sm:text-2xl">
                  {channel.name}
                </h1>
                <div className="mt-1 flex flex-wrap items-center gap-1.5 text-[10px] font-medium text-white/60 sm:gap-2 sm:text-sm">
                  {currentProgram && <span className="text-white/90">{currentProgram.title}</span>}
                  {currentProgram && metaParts.length > 0 && <span>•</span>}
                  {metaParts.length > 0 && <span>{metaParts.join(' · ')}</span>}
                </div>
              </div>
            </div>

            {/* Right section: Controls */}
            <div className="pointer-events-auto flex items-center gap-2 sm:gap-3">
              <button
                type="button"
                onClick={toggleMute}
                className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
                aria-label={isMuted ? 'Unmute' : 'Mute'}
                title="M"
              >
                {isMuted ? <LuVolumeX size={18} /> : <LuVolume2 size={18} />}
              </button>
              <button
                type="button"
                onClick={toggleFullscreen}
                className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
                aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
                title="F"
              >
                {isFullscreen ? <LuMinimize2 size={18} /> : <LuMaximize2 size={18} />}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
