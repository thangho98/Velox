import { useParams, useNavigate, Link } from 'react-router'
import { useEffect, useRef, useState, useCallback } from 'react'
import Hls from 'hls.js'
import {
  useMediaWithFiles,
  useUpdateProgress,
  useStreamUrls,
  useSubtitles,
  useAudioTracks,
} from '@/hooks/stores/useMedia'
import { usePlayerStore } from '@/stores/player'
import { useAuthStore } from '@/stores/auth'
import { getCapabilities } from '@/lib/capabilities'
import { DualSubtitleOverlay } from '@/components/DualSubtitleOverlay'
import { SubtitlePicker } from '@/components/SubtitlePicker'
import { AudioPicker } from '@/components/AudioPicker'
import { QualityPicker } from '@/components/QualityPicker'
import { TrickplayPreview } from '@/components/TrickplayPreview'
import { useToast } from '@/components/Toast'
import type { PlaybackSubtitleTrack } from '@/types/api'

// Keyboard shortcut constants
const SEEK_STEP = 10 // seconds
const VOLUME_STEP = 0.1

export function WatchPage() {
  const { id } = useParams<{ id: string }>()
  const mediaId = Number(id)
  const navigate = useNavigate()
  const videoRef = useRef<HTMLVideoElement>(null)
  const hlsRef = useRef<Hls | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const { data: media, isLoading: mediaLoading } = useMediaWithFiles(mediaId)
  const { mutate: updateProgress } = useUpdateProgress()
  const { accessToken } = useAuthStore()
  const { info: showToastInfo } = useToast()

  const {
    volume,
    isMuted,
    setVolume,
    toggleMute,
    playbackRate,
    setPlaybackRate,
    setLastPosition,
    getLastPosition,
    subtitleLanguage,
    setSubtitleLanguage,
    secondarySubtitleLanguage,
    setSecondarySubtitleLanguage,
    audioLanguage,
    audioTrackId,
    setAudioTrack,
    maxStreamingQuality,
    setMaxStreamingQuality,
  } = usePlayerStore()

  // Probe client codec support once (cached in localStorage for 7 days)
  const clientCaps = getCapabilities()
  const qualityMaxHeight =
    maxStreamingQuality === '1080p'
      ? 1080
      : maxStreamingQuality === '720p'
        ? 720
        : maxStreamingQuality === '480p'
          ? 480
          : undefined // 'auto' = no constraint

  // Include client capabilities and current selections so the engine can pick the optimal path
  const playbackRequest = {
    video_codecs: clientCaps.videoCodecs,
    audio_codecs: clientCaps.audioCodecs,
    containers: clientCaps.containers,
    max_height: qualityMaxHeight,
    selected_subtitle: subtitleLanguage ?? 'off',
    selected_audio_track: audioTrackId ?? 0,
  }
  const { data: streamUrls, isLoading: streamLoading } = useStreamUrls(mediaId, playbackRequest)
  const { data: subtitles = [] } = useSubtitles(mediaId, playbackRequest)
  const { data: audioTracks = [] } = useAudioTracks(mediaId, playbackRequest)

  // Use the file ID from playback response (respects version selection); fall back to first file
  const primaryFileId = streamUrls?.primary_file_id ?? media?.files[0]?.id
  const subtitleServeUrl = (sub: PlaybackSubtitleTrack | undefined) =>
    sub && primaryFileId ? `/api/media-files/${primaryFileId}/subtitles/${sub.id}/serve` : null
  const primarySub = subtitleLanguage
    ? subtitles.find((s) => s.language === subtitleLanguage && !s.is_image)
    : undefined
  const secondarySub = secondarySubtitleLanguage
    ? subtitles.find((s) => s.language === secondarySubtitleLanguage && !s.is_image)
    : undefined

  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [buffered, setBuffered] = useState(0)
  const [showControls, setShowControls] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isBuffering, setIsBuffering] = useState(false)
  const [availableLevels, setAvailableLevels] = useState<{ level: number; height: number }[]>([])
  const [currentLevel, setCurrentLevel] = useState(-1) // -1 = auto
  const [bandwidth, setBandwidth] = useState<number | null>(null)
  const [showQualityIndicator, setShowQualityIndicator] = useState(false)
  const qualityIndicatorTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Trickplay preview state
  const [isHoveringTimeline, setIsHoveringTimeline] = useState(false)
  const [hoverTime, setHoverTime] = useState(0)
  const [hoverPositionX, setHoverPositionX] = useState(0)
  const progressBarRef = useRef<HTMLDivElement>(null)
  const [showSettings, setShowSettings] = useState(false)
  const [showSubtitleMenu, setShowSubtitleMenu] = useState(false)
  const [showAudioMenu, setShowAudioMenu] = useState(false)
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastProgressUpdate = useRef(0)

  // Define callbacks first (before useEffect that uses them)
  const togglePlay = useCallback(() => {
    const video = videoRef.current
    if (!video) return

    if (isPlaying) {
      video.pause()
    } else {
      video.play().catch(() => {
        setError('Playback failed')
      })
    }
    setIsPlaying(!isPlaying)
  }, [isPlaying])

  const seek = useCallback((seconds: number) => {
    const video = videoRef.current
    if (!video) return

    const newTime = Math.max(0, Math.min(video.duration, video.currentTime + seconds))
    video.currentTime = newTime
    setCurrentTime(newTime)
  }, [])

  const changeVolume = useCallback(
    (delta: number) => {
      const video = videoRef.current
      if (!video) return

      const newVolume = Math.max(0, Math.min(1, volume + delta))
      setVolume(newVolume)
      video.volume = newVolume
    },
    [volume, setVolume],
  )

  const toggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) {
      document.exitFullscreen()
    } else {
      containerRef.current?.requestFullscreen()
    }
  }, [])

  // Initialize HLS
  useEffect(() => {
    const video = videoRef.current
    if (!video || !streamUrls) return

    // Clean up previous HLS instance
    if (hlsRef.current) {
      hlsRef.current.destroy()
      hlsRef.current = null
    }

    // Prefer ABR (multi-quality) over single-quality HLS over direct play
    const rawUrl = streamUrls.abr || streamUrls.hls || streamUrls.direct
    if (!rawUrl) {
      return
    }
    // Append access token — browser cannot send Authorization header for video src / HLS fetches.
    const streamUrl = accessToken
      ? rawUrl + (rawUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(accessToken)
      : rawUrl

    if ((streamUrls.abr || streamUrls.hls) && Hls.isSupported()) {
      // HLS playback with hls.js
      const token = accessToken
      const hls = new Hls({
        maxBufferLength: 30,
        maxMaxBufferLength: 600,
        enableWorker: true,
        // Attach Bearer token to every XHR hls.js makes (master, variant, segments).
        xhrSetup: (xhr) => {
          if (token) {
            xhr.setRequestHeader('Authorization', `Bearer ${token}`)
          }
        },
      })
      hlsRef.current = hls

      hls.on(Hls.Events.MANIFEST_PARSED, (_event, data) => {
        const levels = data.levels.map((level, index) => ({
          level: index,
          height: level.height || 0,
        }))
        setAvailableLevels(levels)
        setCurrentLevel(hls.currentLevel)

        // Switch to the user-selected audio track if non-default.
        // HLS.js picks DEFAULT=YES by default; we override using the stored language.
        if (audioLanguage && hls.audioTracks.length > 1) {
          const idx = hls.audioTracks.findIndex(
            (t) =>
              t.lang === audioLanguage || t.name?.toLowerCase() === audioLanguage.toLowerCase(),
          )
          if (idx >= 0 && idx !== hls.audioTrack) {
            hls.audioTrack = idx
          }
        }
      })

      hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
        setCurrentLevel(data.level)
        // Show quality indicator when level switches
        setShowQualityIndicator(true)
        if (qualityIndicatorTimeoutRef.current) {
          clearTimeout(qualityIndicatorTimeoutRef.current)
        }
        qualityIndicatorTimeoutRef.current = setTimeout(() => {
          setShowQualityIndicator(false)
        }, 3000)
      })

      hls.on(Hls.Events.FRAG_LOADED, (_event, data) => {
        // Calculate bandwidth from fragment stats
        const stats = data.frag.stats
        if (stats && stats.loaded && stats.loading) {
          const duration = stats.loading.end - stats.loading.start
          if (duration > 0) {
            const bitsLoaded = stats.loaded * 8
            const bitrateMbps = bitsLoaded / duration / 1000 / 1000
            setBandwidth(bitrateMbps)
            // Show warning if bandwidth is low (< 1.5 Mbps)
            if (bitrateMbps < 1.5 && bitrateMbps > 0) {
              showToastInfo('Kết nối yếu, chất lượng video có thể giảm')
            }
          }
        }
      })

      hls.on(Hls.Events.ERROR, (_event, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              setError('Network error. Trying to recover...')
              hls.startLoad()
              break
            case Hls.ErrorTypes.MEDIA_ERROR:
              setError('Media error. Trying to recover...')
              hls.recoverMediaError()
              break
            default:
              setError('Fatal playback error')
              hls.destroy()
              break
          }
        }
      })

      hls.loadSource(streamUrl)
      hls.attachMedia(video)
    } else {
      // Direct playback
      video.src = streamUrl
    }

    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy()
        hlsRef.current = null
      }
    }
  }, [streamUrls, accessToken, audioLanguage])

  // Resume from saved position
  useEffect(() => {
    const video = videoRef.current
    if (!video || duration === 0) return

    const savedPosition = getLastPosition(mediaId)
    if (savedPosition > 0 && savedPosition < duration * 0.95) {
      video.currentTime = savedPosition
    }
  }, [mediaId, getLastPosition, duration])

  // Apply player settings
  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    video.volume = volume
    video.muted = isMuted
    video.playbackRate = playbackRate
  }, [volume, isMuted, playbackRate])

  // Update subtitle track
  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    // Find and enable the selected subtitle track
    const textTracks = video.textTracks
    for (let i = 0; i < textTracks.length; i++) {
      const track = textTracks[i]
      if (track.kind === 'subtitles' || track.kind === 'captions') {
        if (subtitleLanguage === null) {
          track.mode = 'disabled'
        } else if (track.language === subtitleLanguage) {
          track.mode = 'showing'
        } else {
          track.mode = 'disabled'
        }
      }
    }
  }, [subtitleLanguage])

  // Update audio track
  useEffect(() => {
    const video = videoRef.current as HTMLVideoElement & {
      audioTracks?: { length: number; [index: number]: { language: string; enabled: boolean } }
    }
    if (!video || !video.audioTracks) return

    const tracks = video.audioTracks
    for (let i = 0; i < tracks.length; i++) {
      const track = tracks[i]
      if (audioLanguage === null) {
        // Keep default
      } else if (track.language === audioLanguage) {
        track.enabled = true
      } else {
        track.enabled = false
      }
    }
  }, [audioLanguage])

  // Progress tracking and buffering
  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    const handleTimeUpdate = () => {
      setCurrentTime(video.currentTime)
      setLastPosition(mediaId, video.currentTime)

      // Update progress every 10 seconds or at end
      const now = Date.now()
      if (now - lastProgressUpdate.current >= 10000 || video.currentTime >= video.duration * 0.95) {
        updateProgress({
          mediaId,
          data: {
            position: video.currentTime,
            completed: video.currentTime / video.duration > 0.9,
          },
        })
        lastProgressUpdate.current = now
      }
    }

    const handleProgress = () => {
      if (video.buffered.length > 0) {
        setBuffered(video.buffered.end(video.buffered.length - 1))
      }
    }

    const handleWaiting = () => setIsBuffering(true)
    const handlePlaying = () => setIsBuffering(false)
    const handleCanPlay = () => setIsBuffering(false)

    const handleLoadedMetadata = () => {
      setDuration(video.duration)
    }

    video.addEventListener('timeupdate', handleTimeUpdate)
    video.addEventListener('progress', handleProgress)
    video.addEventListener('waiting', handleWaiting)
    video.addEventListener('playing', handlePlaying)
    video.addEventListener('canplay', handleCanPlay)
    video.addEventListener('loadedmetadata', handleLoadedMetadata)

    return () => {
      video.removeEventListener('timeupdate', handleTimeUpdate)
      video.removeEventListener('progress', handleProgress)
      video.removeEventListener('waiting', handleWaiting)
      video.removeEventListener('playing', handlePlaying)
      video.removeEventListener('canplay', handleCanPlay)
      video.removeEventListener('loadedmetadata', handleLoadedMetadata)
    }
  }, [mediaId, setLastPosition, updateProgress])

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const video = videoRef.current
      if (!video) return

      // Ignore if user is typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return
      }

      switch (e.key) {
        case ' ':
        case 'k':
          e.preventDefault()
          togglePlay()
          break
        case 'ArrowLeft':
          e.preventDefault()
          seek(-SEEK_STEP)
          break
        case 'ArrowRight':
          e.preventDefault()
          seek(SEEK_STEP)
          break
        case 'ArrowUp':
          e.preventDefault()
          changeVolume(VOLUME_STEP)
          break
        case 'ArrowDown':
          e.preventDefault()
          changeVolume(-VOLUME_STEP)
          break
        case 'f':
          e.preventDefault()
          toggleFullscreen()
          break
        case 'm':
          e.preventDefault()
          toggleMute()
          break
        case 'j':
          e.preventDefault()
          seek(-SEEK_STEP * 2) // Double backward
          break
        case 'l':
          e.preventDefault()
          seek(SEEK_STEP * 2) // Double forward
          break
        case '0':
        case 'Home':
          e.preventDefault()
          video.currentTime = 0
          break
        case 'End':
          e.preventDefault()
          video.currentTime = video.duration
          break
        case 'Escape':
          if (isFullscreen) {
            e.preventDefault()
            document.exitFullscreen()
          }
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isFullscreen, togglePlay, seek, changeVolume, toggleFullscreen, toggleMute])

  // Fullscreen change handler
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  const handleMouseMove = () => {
    setShowControls(true)
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current)
    }
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) setShowControls(false)
    }, 3000)
  }

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = Number(e.target.value)
    if (videoRef.current) {
      videoRef.current.currentTime = time
      setCurrentTime(time)
    }
  }

  const changeQualityLevel = (level: number) => {
    if (hlsRef.current) {
      hlsRef.current.currentLevel = level
      setCurrentLevel(level)
    }
  }

  const getActiveSubtitleTrack = (): PlaybackSubtitleTrack | null => {
    if (subtitleLanguage === null) return null
    return subtitles.find((s) => s.language === subtitleLanguage) || null
  }

  if (mediaLoading || streamLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-black text-white">
        <div className="flex flex-col items-center gap-4">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
          <p className="text-gray-400">Loading...</p>
        </div>
      </div>
    )
  }

  if (error || !media || !media.files.length) {
    return (
      <div className="flex h-screen items-center justify-center bg-black text-white">
        <div className="text-center">
          <p className="text-lg text-red-400">{error || 'Media not found'}</p>
          <button onClick={() => navigate('/')} className="mt-4 text-netflix-blue hover:underline">
            Go home
          </button>
        </div>
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className="relative h-screen w-full bg-black"
      onMouseMove={handleMouseMove}
      onClick={togglePlay}
    >
      {/* Video element — no native <track>; DualSubtitleOverlay handles rendering */}
      <video
        ref={videoRef}
        className="h-full w-full"
        playsInline
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onEnded={() => {
          setIsPlaying(false)
          updateProgress({
            mediaId,
            data: { position: duration, completed: true },
          })
        }}
        onError={() => setError('Video playback error')}
      />

      {/* Dual subtitle overlay — renders primary (white/bottom) and secondary (yellow/above) */}
      <DualSubtitleOverlay
        videoRef={videoRef}
        primaryUrl={subtitleServeUrl(primarySub)}
        secondaryUrl={subtitleServeUrl(secondarySub)}
        currentTime={currentTime}
      />

      {/* Loading/Buffering indicator */}
      {isBuffering && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/30">
          <div className="h-12 w-12 animate-spin rounded-full border-2 border-white border-t-transparent" />
        </div>
      )}

      {/* Quality & Bandwidth indicator */}
      {showQualityIndicator && availableLevels.length > 0 && (
        <div className="absolute right-4 top-4 rounded bg-black/60 px-3 py-1.5 text-sm text-white backdrop-blur-sm transition-opacity">
          {(() => {
            const height =
              currentLevel === -1
                ? availableLevels.find((l) => l.level === hlsRef.current?.currentLevel)?.height
                : availableLevels.find((l) => l.level === currentLevel)?.height
            return (
              <span>
                {height ? `${height}p` : 'Auto'}
                {bandwidth !== null && ` · ${bandwidth.toFixed(1)} Mbps`}
              </span>
            )
          })()}
        </div>
      )}

      {/* Controls Overlay */}
      <div
        className={`absolute inset-0 flex flex-col justify-between bg-gradient-to-t from-black/80 via-transparent to-black/40 p-6 transition-opacity ${
          showControls ? 'opacity-100' : 'opacity-0'
        }`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Top bar */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              to="/"
              className="flex items-center gap-2 rounded-full bg-black/50 px-4 py-2 text-white backdrop-blur-sm transition-colors hover:bg-black/70"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
              Back
            </Link>
          </div>
          <h1 className="text-lg font-medium text-white">{media.media.title}</h1>
          <div className="w-20" />
        </div>

        {/* Center play button (shown when paused) */}
        {!isPlaying && !isBuffering && (
          <div className="absolute inset-0 flex items-center justify-center">
            <button
              onClick={togglePlay}
              className="rounded-full bg-white/20 p-6 text-white backdrop-blur-sm transition-colors hover:bg-white/30"
            >
              <svg className="h-12 w-12" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
            </button>
          </div>
        )}

        {/* Bottom controls */}
        <div className="space-y-3">
          {/* Progress bar with buffer indicator and trickplay */}
          <div
            ref={progressBarRef}
            className="group relative"
            onMouseEnter={() => setIsHoveringTimeline(true)}
            onMouseLeave={() => setIsHoveringTimeline(false)}
            onMouseMove={(e) => {
              if (!progressBarRef.current || !duration) return
              const rect = progressBarRef.current.getBoundingClientRect()
              const x = e.clientX - rect.left
              const percentage = Math.max(0, Math.min(1, x / rect.width))
              setHoverTime(percentage * duration)
              setHoverPositionX(x)
            }}
          >
            {/* Trickplay preview */}
            <TrickplayPreview
              mediaId={mediaId}
              currentHoverTime={hoverTime}
              visible={isHoveringTimeline && duration > 0}
              positionX={hoverPositionX}
            />

            {/* Buffered bar */}
            <div className="absolute h-1 w-full rounded-full bg-white/20">
              <div
                className="h-full rounded-full bg-white/40"
                style={{ width: `${duration ? (buffered / duration) * 100 : 0}%` }}
              />
            </div>
            {/* Seek bar */}
            <input
              type="range"
              min={0}
              max={duration || 100}
              value={currentTime}
              onChange={handleSeek}
              className="relative z-10 w-full cursor-pointer appearance-none bg-transparent [&::-webkit-slider-runnable-track]:h-1 [&::-webkit-slider-runnable-track]:rounded-full [&::-webkit-slider-runnable-track]:bg-transparent [&::-webkit-slider-thumb]:-mt-[3px] [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-netflix-red [&::-webkit-slider-thumb]:opacity-0 [&::-webkit-slider-thumb]:transition-opacity group-hover:[&::-webkit-slider-thumb]:opacity-100"
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              {/* Play/Pause */}
              <button
                onClick={togglePlay}
                className="text-white transition-colors hover:text-netflix-red"
              >
                {isPlaying ? (
                  <svg className="h-8 w-8" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
                  </svg>
                ) : (
                  <svg className="h-8 w-8" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                )}
              </button>

              {/* Rewind/Forward */}
              <button
                onClick={() => seek(-SEEK_STEP)}
                className="text-white transition-colors hover:text-gray-300"
              >
                <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11 18V6l-8.5 6 8.5 6zm.5-6l8.5 6V6l-8.5 6z" />
                </svg>
              </button>
              <button
                onClick={() => seek(SEEK_STEP)}
                className="text-white transition-colors hover:text-gray-300"
              >
                <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M4 18l8.5-6L4 6v12zm9-12v12l8.5-6L13 6z" />
                </svg>
              </button>

              {/* Time */}
              <span className="text-sm text-white">
                {formatTime(currentTime)} / {formatTime(duration)}
              </span>

              {/* Volume */}
              <div className="flex items-center gap-2">
                <button
                  onClick={toggleMute}
                  className="text-white transition-colors hover:text-gray-300"
                >
                  {isMuted || volume === 0 ? (
                    <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
                    </svg>
                  ) : volume < 0.5 ? (
                    <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M18.5 12c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM5 9v6h4l5 5V4L9 9H5z" />
                    </svg>
                  ) : (
                    <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z" />
                    </svg>
                  )}
                </button>
                <input
                  type="range"
                  min={0}
                  max={1}
                  step={0.05}
                  value={isMuted ? 0 : volume}
                  onChange={(e) => {
                    const v = Number(e.target.value)
                    setVolume(v)
                    if (videoRef.current) {
                      videoRef.current.volume = v
                      videoRef.current.muted = false
                    }
                  }}
                  className="w-20 cursor-pointer"
                />
              </div>
            </div>

            {/* Right controls */}
            <div className="flex items-center gap-2">
              {/* Subtitle selector */}
              {subtitles.length > 0 && (
                <div className="relative">
                  <button
                    onClick={() => {
                      setShowSubtitleMenu(!showSubtitleMenu)
                      setShowAudioMenu(false)
                      setShowSettings(false)
                    }}
                    className={`rounded px-3 py-1.5 text-sm font-medium transition-colors ${
                      getActiveSubtitleTrack()
                        ? 'bg-netflix-red text-white'
                        : 'bg-white/10 text-white hover:bg-white/20'
                    }`}
                  >
                    CC
                  </button>
                  {showSubtitleMenu && (
                    <div className="absolute bottom-full right-0 mb-2">
                      <SubtitlePicker
                        subtitles={subtitles}
                        primaryLanguage={subtitleLanguage}
                        secondaryLanguage={secondarySubtitleLanguage}
                        onSelectPrimary={(lang) => {
                          setSubtitleLanguage(lang)
                          setShowSubtitleMenu(false)
                        }}
                        onSelectSecondary={setSecondarySubtitleLanguage}
                        dualMode={true}
                      />
                    </div>
                  )}
                </div>
              )}

              {/* Audio track selector */}
              {audioTracks.length > 0 && (
                <div className="relative">
                  <button
                    onClick={() => {
                      setShowAudioMenu(!showAudioMenu)
                      setShowSubtitleMenu(false)
                      setShowSettings(false)
                    }}
                    className="rounded bg-white/10 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-white/20"
                  >
                    Audio
                  </button>
                  {showAudioMenu && (
                    <div className="absolute bottom-full right-0 mb-2">
                      <AudioPicker
                        tracks={audioTracks}
                        selectedLanguage={audioLanguage}
                        onSelect={(lang, trackId) => {
                          setAudioTrack(lang, trackId)
                          setShowAudioMenu(false)
                        }}
                      />
                    </div>
                  )}
                </div>
              )}

              {/* Quality picker (HLS only) */}
              {availableLevels.length > 0 && (
                <QualityPicker
                  levels={availableLevels}
                  currentLevel={currentLevel}
                  onChange={changeQualityLevel}
                  currentHeight={availableLevels.find((l) => l.level === currentLevel)?.height}
                />
              )}

              {/* Settings (Quality & Speed) */}
              <div className="relative">
                <button
                  onClick={() => {
                    setShowSettings(!showSettings)
                    setShowSubtitleMenu(false)
                    setShowAudioMenu(false)
                  }}
                  className="rounded p-2 text-white transition-colors hover:bg-white/10"
                >
                  <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58a.49.49 0 0 0 .12-.61l-1.92-3.32a.488.488 0 0 0-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54a.484.484 0 0 0-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L3.16 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.09.63-.09.94s.02.64.07.94l-2.03 1.58a.49.49 0 0 0-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z" />
                  </svg>
                </button>
                {showSettings && (
                  <div className="absolute bottom-full right-0 mb-2 min-w-[180px] rounded-lg bg-black/90 p-2 shadow-xl">
                    {/* Max streaming quality */}
                    <div className="mb-2 border-b border-white/10 pb-2">
                      <p className="mb-1 px-3 text-xs text-gray-400">Max Quality</p>
                      {(['auto', '1080p', '720p', '480p'] as const).map((q) => (
                        <button
                          key={q}
                          onClick={() => setMaxStreamingQuality(q)}
                          className={`w-full rounded px-3 py-1.5 text-left text-sm ${
                            maxStreamingQuality === q
                              ? 'bg-netflix-red text-white'
                              : 'text-white hover:bg-white/10'
                          }`}
                        >
                          {q === 'auto' ? 'Auto' : q}
                        </button>
                      ))}
                    </div>

                    {/* Playback speed */}
                    <div className="mb-2 border-b border-white/10 pb-2">
                      <p className="mb-1 px-3 text-xs text-gray-400">Playback Speed</p>
                      {[0.5, 0.75, 1, 1.25, 1.5, 2].map((rate) => (
                        <button
                          key={rate}
                          onClick={() => {
                            setPlaybackRate(rate)
                            if (videoRef.current) {
                              videoRef.current.playbackRate = rate
                            }
                          }}
                          className={`w-full rounded px-3 py-1.5 text-left text-sm ${
                            playbackRate === rate
                              ? 'bg-netflix-red text-white'
                              : 'text-white hover:bg-white/10'
                          }`}
                        >
                          {rate === 1 ? 'Normal' : `${rate}x`}
                        </button>
                      ))}
                    </div>

                    {/* Quality selector (HLS only) */}
                    {availableLevels.length > 0 && (
                      <div>
                        <p className="mb-1 px-3 text-xs text-gray-400">Quality</p>
                        <button
                          onClick={() => changeQualityLevel(-1)}
                          className={`w-full rounded px-3 py-1.5 text-left text-sm ${
                            currentLevel === -1
                              ? 'bg-netflix-red text-white'
                              : 'text-white hover:bg-white/10'
                          }`}
                        >
                          Auto
                        </button>
                        {[...availableLevels].reverse().map((level) => (
                          <button
                            key={level.level}
                            onClick={() => changeQualityLevel(level.level)}
                            className={`w-full rounded px-3 py-1.5 text-left text-sm ${
                              currentLevel === level.level
                                ? 'bg-netflix-red text-white'
                                : 'text-white hover:bg-white/10'
                            }`}
                          >
                            {level.height}p
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* Fullscreen */}
              <button
                onClick={toggleFullscreen}
                className="rounded p-2 text-white transition-colors hover:bg-white/10"
              >
                {isFullscreen ? (
                  <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z" />
                  </svg>
                ) : (
                  <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Keyboard shortcuts hint (shown briefly on load) */}
      <KeyboardShortcutsHint show={showControls} />
    </div>
  )
}

function KeyboardShortcutsHint({ show }: { show: boolean }) {
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => setVisible(false), 5000)
    return () => clearTimeout(timer)
  }, [])

  if (!show || !visible) return null

  return (
    <div className="absolute bottom-24 left-6 rounded-lg bg-black/80 p-4 text-xs text-white backdrop-blur-sm">
      <p className="mb-2 font-semibold text-gray-400">Keyboard Shortcuts</p>
      <div className="grid grid-cols-2 gap-x-4 gap-y-1">
        <span className="text-gray-500">Space</span>
        <span>Play/Pause</span>
        <span className="text-gray-500">← →</span>
        <span>Seek ±10s</span>
        <span className="text-gray-500">↑ ↓</span>
        <span>Volume</span>
        <span className="text-gray-500">F</span>
        <span>Fullscreen</span>
        <span className="text-gray-500">M</span>
        <span>Mute</span>
      </div>
    </div>
  )
}

function formatTime(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const hours = Math.floor(mins / 60)
  const displayMins = mins % 60

  if (hours > 0) {
    return `${hours}:${displayMins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
