import { useParams, useNavigate } from 'react-router'
import { useEffect, useRef, useState, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import Hls from 'hls.js'
import {
  LuChevronLeft,
  LuPlay,
  LuPause,
  LuVolumeX,
  LuVolume1,
  LuVolume2,
  LuMaximize2,
  LuMinimize2,
  LuSettings,
  LuSkipForward,
  LuCaptions,
  LuMusic,
  LuZap,
  LuRotateCcw,
  LuRotateCw,
  LuInfo,
  LuListMusic,
  LuRepeat,
  LuRepeat2,
  LuActivity,
  LuExternalLink,
  LuExpand,
  LuLock,
  LuLockOpen,
  LuChevronRight,
  LuCheck,
  LuX,
} from 'react-icons/lu'
import {
  useMediaWithFiles,
  useUpdateProgress,
  useStreamUrls,
  useSubtitles,
  useAudioTracks,
  useEpisodes,
  usePlaybackInfo,
  streamingKeys,
} from '@/hooks/stores/useMedia'
import { usePreferences } from '@/hooks/stores/useAuth'
import { usePlayerStore } from '@/stores/player'
import { useAuthStore } from '@/stores/auth'
import { getCapabilities } from '@/lib/capabilities'
import { DualSubtitleOverlay } from '@/components/DualSubtitleOverlay'
import { SubtitlePicker } from '@/components/SubtitlePicker'
import { AudioPicker } from '@/components/AudioPicker'
import { TrickplayPreview } from '@/components/TrickplayPreview'
import { useToast } from '@/components/Toast'
import type { PlaybackSubtitleTrack } from '@/types/api'

const SEEK_STEP = 10
const VOLUME_STEP = 0.1

const QUALITY_OPTIONS = [
  { height: 1080, bitrateKbps: 60000, label: '1080p - 60 Mbps' },
  { height: 1080, bitrateKbps: 50000, label: '1080p - 50 Mbps' },
  { height: 1080, bitrateKbps: 40000, label: '1080p - 40 Mbps' },
  { height: 1080, bitrateKbps: 30000, label: '1080p - 30 Mbps' },
  { height: 1080, bitrateKbps: 25000, label: '1080p - 25 Mbps' },
  { height: 1080, bitrateKbps: 20000, label: '1080p - 20 Mbps' },
  { height: 1080, bitrateKbps: 15000, label: '1080p - 15 Mbps' },
  { height: 1080, bitrateKbps: 12000, label: '1080p - 12 Mbps' },
  { height: 1080, bitrateKbps: 10000, label: '1080p - 10 Mbps' },
  { height: 1080, bitrateKbps: 8000, label: '1080p - 8 Mbps' },
  { height: 1080, bitrateKbps: 6000, label: '1080p - 6 Mbps' },
  { height: 1080, bitrateKbps: 5000, label: '1080p - 5 Mbps' },
  { height: 1080, bitrateKbps: 4000, label: '1080p - 4 Mbps' },
  { height: 720, bitrateKbps: 4000, label: '720p - 4 Mbps' },
  { height: 720, bitrateKbps: 3000, label: '720p - 3 Mbps' },
  { height: 720, bitrateKbps: 2000, label: '720p - 2 Mbps' },
  { height: 720, bitrateKbps: 1500, label: '720p - 1.5 Mbps' },
  { height: 720, bitrateKbps: 1000, label: '720p - 1 Mbps' },
  { height: 480, bitrateKbps: 1000, label: '480p - 1 Mbps' },
  { height: 480, bitrateKbps: 720, label: '480p - 720 Kbps' },
  { height: 480, bitrateKbps: 420, label: '480p - 420 Kbps' },
  { height: 360, bitrateKbps: 420, label: '360p' },
  { height: 240, bitrateKbps: 250, label: '240p' },
  { height: 144, bitrateKbps: 120, label: '144p' },
]

function normalizeLanguageCode(language: string | null | undefined): string {
  const value = (language ?? '').trim().toLowerCase()
  switch (value) {
    case 'en':
    case 'eng':
      return 'eng'
    case 'vi':
    case 'vie':
      return 'vie'
    case 'zh':
    case 'zho':
    case 'chi':
      return 'zho'
    default:
      return value
  }
}

function languageMatches(lhs: string | null | undefined, rhs: string | null | undefined): boolean {
  if (!lhs || !rhs) return false
  return normalizeLanguageCode(lhs) === normalizeLanguageCode(rhs)
}

export function WatchPage() {
  const { id } = useParams<{ id: string }>()
  const mediaId = Number(id)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const videoRef = useRef<HTMLVideoElement>(null)
  const hlsRef = useRef<Hls | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const progressBarRef = useRef<HTMLDivElement>(null)
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastProgressUpdate = useRef(Date.now())
  const seekFeedbackTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const qualityIndicatorTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lowBandwidthToastShown = useRef(false)

  const { data: media, isLoading: mediaLoading } = useMediaWithFiles(mediaId)
  const { data: preferences } = usePreferences()
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
    subtitleLanguage,
    setSubtitleLanguage,
    secondarySubtitleLanguage,
    setSecondarySubtitleLanguage,
    subtitleSize,
    setSubtitleSize,
    subtitleColor,
    setSubtitleColor,
    subtitleBackground,
    setSubtitleBackground,
    audioLanguage,
    audioTrackId,
    setAudioTrack,
    maxQuality,
    setMaxQuality,
    aspectRatio,
    setAspectRatio,
    repeatMode,
    setRepeatMode,
  } = usePlayerStore()

  const clientCaps = getCapabilities()
  const effectiveSubtitleLanguage = subtitleLanguage ?? preferences?.subtitle_language ?? null
  const qualityMaxHeight = maxQuality === 'auto' ? undefined : maxQuality.height

  const playbackRequest = {
    video_codecs: clientCaps.videoCodecs,
    audio_codecs: clientCaps.audioCodecs,
    containers: clientCaps.containers,
    max_height: qualityMaxHeight,
    selected_subtitle: effectiveSubtitleLanguage ?? 'off',
    selected_audio_track: audioTrackId ?? 0,
  }
  const { data: streamUrls, isLoading: streamLoading } = useStreamUrls(mediaId, playbackRequest)
  const { data: subtitles = [] } = useSubtitles(mediaId, playbackRequest)
  const { data: audioTracks = [] } = useAudioTracks(mediaId, playbackRequest)
  const { data: playbackInfo } = usePlaybackInfo(mediaId, playbackRequest)

  const isEpisode = media?.media.media_type === 'episode'
  const seriesId = media?.media.series_id ?? 0
  const seasonId = media?.media.season_id ?? 0
  const { data: seasonEpisodes = [] } = useEpisodes(
    isEpisode ? seriesId : 0,
    isEpisode ? seasonId : 0,
  )
  const nextEpisode = (() => {
    if (!isEpisode || seasonEpisodes.length === 0) return null
    const currentIdx = seasonEpisodes.findIndex((ep) =>
      ep.media_files?.some((f) => f.media_id === mediaId),
    )
    if (currentIdx === -1 || currentIdx === seasonEpisodes.length - 1) return null
    return seasonEpisodes[currentIdx + 1]
  })()
  const nextEpisodeMediaId = nextEpisode?.media_files?.[0]?.media_id

  useEffect(() => {
    setAudioTrack(audioLanguage, null)
    // audio track IDs are file-specific; never carry them across media items
  }, [mediaId, audioLanguage, setAudioTrack])

  useEffect(() => {
    if (audioTrackId == null || audioTracks.length === 0) return
    const selectedTrack = audioTracks.find((track) => track.id === audioTrackId)
    if (!selectedTrack || selectedTrack.is_default) {
      setAudioTrack(audioLanguage, null)
    }
  }, [audioLanguage, audioTrackId, audioTracks, setAudioTrack])

  const primaryFileId = streamUrls?.primary_file_id ?? media?.files[0]?.id
  const subtitleServeUrl = (sub: PlaybackSubtitleTrack | undefined) => {
    if (!sub || !primaryFileId) return null
    const base = `/api/media-files/${primaryFileId}/subtitles/${sub.id}/serve`
    return accessToken ? `${base}?token=${encodeURIComponent(accessToken)}` : base
  }
  const primarySub = effectiveSubtitleLanguage
    ? subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage) && !s.is_image)
    : undefined
  const secondarySub = secondarySubtitleLanguage
    ? subtitles.find((s) => languageMatches(s.language, secondarySubtitleLanguage) && !s.is_image)
    : undefined

  // Player state
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [buffered, setBuffered] = useState(0)
  const [showControls, setShowControls] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isBuffering, setIsBuffering] = useState(true)
  const [availableLevels, setAvailableLevels] = useState<
    { level: number; height: number; bitrate: number }[]
  >([])
  const [currentLevel, setCurrentLevel] = useState(-1)
  const [bandwidth, setBandwidth] = useState<number | null>(null)
  const [showQualityIndicator, setShowQualityIndicator] = useState(false)
  const allowsImageSubtitles =
    playbackInfo?.method === 'FullTranscode' || playbackInfo?.method === 'TranscodeAudio'

  // Wall clock
  const [wallClock, setWallClock] = useState(() => getWallClock())
  useEffect(() => {
    const t = setInterval(() => setWallClock(getWallClock()), 30000)
    return () => clearInterval(t)
  }, [])

  // Seek feedback
  const [seekFeedback, setSeekFeedback] = useState<{ dir: 'back' | 'fwd'; n: number } | null>(null)

  // Progress bar hover/drag
  const [isHoveringBar, setIsHoveringBar] = useState(false)
  const [hoverTime, setHoverTime] = useState(0)
  const [hoverX, setHoverX] = useState(0)
  const [isDraggingBar, setIsDraggingBar] = useState(false)

  // Menus
  const [showSubtitleMenu, setShowSubtitleMenu] = useState(false)
  const [showAudioMenu, setShowAudioMenu] = useState(false)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [settingsView, setSettingsView] = useState<'main' | 'quality'>('main')
  const [showStats, setShowStats] = useState(false)

  // Bottom tab: 'none' | 'info' | 'chapters'
  const [activeTab, setActiveTab] = useState<'none' | 'info' | 'chapters'>('none')

  // Screen lock
  const [isLocked, setIsLocked] = useState(false)

  // Up Next
  const [upNextDismissed, setUpNextDismissed] = useState(false)
  const showUpNext =
    isEpisode &&
    nextEpisodeMediaId != null &&
    duration > 0 &&
    currentTime / duration > 0.88 &&
    !upNextDismissed

  // ── Callbacks ──────────────────────────────────────────────────────────────
  const togglePlay = useCallback(() => {
    const video = videoRef.current
    if (!video) return
    const willPlay = !isPlaying
    if (isPlaying) video.pause()
    else video.play().catch(() => setError('Playback failed'))
    setIsPlaying(willPlay)
    // Close all overlay menus
    setShowSubtitleMenu(false)
    setShowAudioMenu(false)
    setShowSpeedMenu(false)
    setShowSettings(false)
    // Start controls auto-hide timer when playing
    if (willPlay) {
      if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
      controlsTimeoutRef.current = setTimeout(() => setShowControls(false), 3500)
    }
  }, [isPlaying])

  const showSeekFeedback = useCallback((dir: 'back' | 'fwd', n: number) => {
    setSeekFeedback({ dir, n })
    if (seekFeedbackTimeout.current) clearTimeout(seekFeedbackTimeout.current)
    seekFeedbackTimeout.current = setTimeout(() => setSeekFeedback(null), 700)
  }, [])

  const seek = useCallback(
    (seconds: number) => {
      const video = videoRef.current
      if (!video) return
      const newTime = Math.max(0, Math.min(video.duration, video.currentTime + seconds))
      video.currentTime = newTime
      setCurrentTime(newTime)
      showSeekFeedback(seconds > 0 ? 'fwd' : 'back', Math.abs(seconds))
    },
    [showSeekFeedback],
  )

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
    type FullscreenDoc = Document & {
      webkitFullscreenElement?: Element
      webkitExitFullscreen?: () => void
    }
    type FullscreenEl = HTMLElement & {
      webkitRequestFullscreen?: () => Promise<void>
    }
    const doc = document as FullscreenDoc
    const el = document.documentElement as FullscreenEl
    const isFs = !!(document.fullscreenElement || doc.webkitFullscreenElement)

    if (isFs) {
      if (document.exitFullscreen) document.exitFullscreen().catch(console.error)
      else doc.webkitExitFullscreen?.()
    } else {
      if (el.requestFullscreen) {
        el.requestFullscreen().catch((err: Error) => {
          showToastInfo(`Fullscreen: ${err.message}`)
          console.error('[fullscreen]', err)
        })
      } else if (el.webkitRequestFullscreen) {
        el.webkitRequestFullscreen()
      } else {
        showToastInfo('Fullscreen not supported in this browser')
      }
    }
  }, [showToastInfo])

  const resetControlsTimeout = useCallback(() => {
    setShowControls(true)
    if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) setShowControls(false)
    }, 3500)
  }, [isPlaying])

  // ── Progress bar ────────────────────────────────────────────────────────────
  const getTimeFromClientX = useCallback(
    (clientX: number) => {
      if (!progressBarRef.current || !duration) return 0
      const rect = progressBarRef.current.getBoundingClientRect()
      const x = Math.max(0, Math.min(clientX - rect.left, rect.width))
      return (x / rect.width) * duration
    },
    [duration],
  )

  const handleBarMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation()
      setIsDraggingBar(true)
      const time = getTimeFromClientX(e.clientX)
      if (videoRef.current) {
        videoRef.current.currentTime = time
        setCurrentTime(time)
      }
    },
    [getTimeFromClientX],
  )

  const handleBarMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!progressBarRef.current) return
      const rect = progressBarRef.current.getBoundingClientRect()
      setHoverX(e.clientX - rect.left)
      setHoverTime(getTimeFromClientX(e.clientX))
      if (isDraggingBar && videoRef.current) {
        const time = getTimeFromClientX(e.clientX)
        videoRef.current.currentTime = time
        setCurrentTime(time)
      }
    },
    [getTimeFromClientX, isDraggingBar],
  )

  useEffect(() => {
    if (!isDraggingBar) return
    const onMove = (e: MouseEvent) => {
      const time = getTimeFromClientX(e.clientX)
      if (videoRef.current) {
        videoRef.current.currentTime = time
        setCurrentTime(time)
      }
    }
    const onUp = () => setIsDraggingBar(false)
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isDraggingBar, getTimeFromClientX])

  // ── HLS init ───────────────────────────────────────────────────────────────
  useEffect(() => {
    const video = videoRef.current
    if (!video || !streamUrls) return
    if (hlsRef.current) {
      hlsRef.current.destroy()
      hlsRef.current = null
    }

    const useHls =
      playbackInfo?.method === 'FullTranscode' || playbackInfo?.method === 'TranscodeAudio'
    const rawUrl = useHls ? streamUrls.abr || streamUrls.hls : streamUrls.direct
    if (!rawUrl) return
    setIsBuffering(true)
    const streamUrl = accessToken
      ? rawUrl + (rawUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(accessToken)
      : rawUrl

    if (useHls && (streamUrls.abr || streamUrls.hls) && Hls.isSupported()) {
      const token = accessToken
      const hls = new Hls({
        maxBufferLength: 30,
        maxMaxBufferLength: 600,
        enableWorker: true,
        xhrSetup: (xhr) => {
          if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`)
        },
      })
      hlsRef.current = hls
      hls.on(Hls.Events.MANIFEST_PARSED, (_e, data) => {
        setAvailableLevels(
          data.levels.map((l, i) => ({ level: i, height: l.height || 0, bitrate: l.bitrate || 0 })),
        )
        setCurrentLevel(hls.currentLevel)
        if (audioLanguage && hls.audioTracks.length > 1) {
          const idx = hls.audioTracks.findIndex(
            (t) =>
              t.lang === audioLanguage || t.name?.toLowerCase() === audioLanguage.toLowerCase(),
          )
          if (idx >= 0 && idx !== hls.audioTrack) hls.audioTrack = idx
        }
        // HLS VOD: update duration once manifest is parsed
        const v = videoRef.current
        if (v && v.duration && isFinite(v.duration)) {
          setDuration(v.duration)
        }
        // Resume position (initial load or after subtitle/quality change)
        const seekTo = usePlayerStore.getState().lastPositions[mediaId] ?? 0
        if (seekTo > 0 && v) {
          const dur = v.duration
          if (!dur || !isFinite(dur) || seekTo < dur * 0.95) {
            v.currentTime = seekTo
          }
        }
        // Auto-play when ready
        v?.play().catch(() => {})
      })
      hls.on(Hls.Events.LEVEL_LOADED, (_e, data) => {
        if (data.details?.totalduration) {
          setDuration(data.details.totalduration)
        }
      })
      hls.on(Hls.Events.LEVEL_SWITCHED, (_e, data) => {
        setCurrentLevel(data.level)
        setShowQualityIndicator(true)
        if (qualityIndicatorTimeout.current) clearTimeout(qualityIndicatorTimeout.current)
        qualityIndicatorTimeout.current = setTimeout(() => setShowQualityIndicator(false), 3000)
      })
      hls.on(Hls.Events.FRAG_LOADED, (_e, data) => {
        const stats = data.frag.stats
        if (stats?.loaded && stats?.loading) {
          const dur = stats.loading.end - stats.loading.start
          if (dur > 0) {
            const mbps = (stats.loaded * 8) / dur / 1e6
            setBandwidth(mbps)
            if (mbps < 1.5 && mbps > 0 && !lowBandwidthToastShown.current) {
              lowBandwidthToastShown.current = true
              showToastInfo('Kết nối yếu, chất lượng video có thể giảm')
            }
          }
        }
      })
      hls.on(Hls.Events.ERROR, (_e, data) => {
        if (data.fatal) {
          if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
            setError('Network error...')
            hls.startLoad()
          } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
            setError('Media error...')
            hls.recoverMediaError()
          } else {
            setError('Fatal playback error')
            hls.destroy()
          }
        }
      })
      hls.loadSource(streamUrl)
      hls.attachMedia(video)
    } else {
      video.src = streamUrl
      // Resume position (initial load or after subtitle/quality change)
      const seekTo = usePlayerStore.getState().lastPositions[mediaId] ?? 0
      if (seekTo > 0) {
        video.currentTime = seekTo
      }
      video.play().catch(() => {})
    }
    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy()
        hlsRef.current = null
      }
    }
  }, [streamUrls, playbackInfo?.method, accessToken, audioLanguage])

  // Resume position is read from usePlayerStore.getState().lastPositions[mediaId]
  // directly in the HLS init effect — no cross-effect refs needed.

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    video.volume = volume
    video.muted = isMuted
    video.playbackRate = playbackRate
  }, [volume, isMuted, playbackRate])

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    for (let i = 0; i < video.textTracks.length; i++) {
      const track = video.textTracks[i]
      if (track.kind === 'subtitles' || track.kind === 'captions') {
        track.mode = !subtitleLanguage
          ? 'disabled'
          : track.language === subtitleLanguage
            ? 'showing'
            : 'disabled'
      }
    }
  }, [subtitleLanguage])

  useEffect(() => {
    const video = videoRef.current as HTMLVideoElement & {
      audioTracks?: { length: number; [index: number]: { language: string; enabled: boolean } }
    }
    if (!video?.audioTracks) return
    for (let i = 0; i < video.audioTracks.length; i++) {
      const track = video.audioTracks[i]
      if (audioLanguage) track.enabled = track.language === audioLanguage
    }
  }, [audioLanguage])

  // Video event listeners — streamUrls in deps ensures this re-runs when the
  // video element first appears (it's conditionally rendered after data loads).
  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    const onTimeUpdate = () => {
      setCurrentTime(video.currentTime)
      if (video.duration && !isNaN(video.duration) && isFinite(video.duration)) {
        setDuration(video.duration)
      }
      setLastPosition(mediaId, video.currentTime)
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
    const onProgress = () => {
      if (video.buffered.length > 0) setBuffered(video.buffered.end(video.buffered.length - 1))
    }
    const onWaiting = () => setIsBuffering(true)
    const onPlaying = () => setIsBuffering(false)
    const onCanPlay = () => setIsBuffering(false)
    const onDurationChange = () => {
      if (video.duration && !isNaN(video.duration) && isFinite(video.duration)) {
        setDuration(video.duration)
      }
    }
    video.addEventListener('timeupdate', onTimeUpdate)
    video.addEventListener('progress', onProgress)
    video.addEventListener('waiting', onWaiting)
    video.addEventListener('playing', onPlaying)
    video.addEventListener('canplay', onCanPlay)
    video.addEventListener('loadedmetadata', onDurationChange)
    video.addEventListener('durationchange', onDurationChange)
    video.addEventListener('loadeddata', onDurationChange)
    return () => {
      video.removeEventListener('timeupdate', onTimeUpdate)
      video.removeEventListener('progress', onProgress)
      video.removeEventListener('waiting', onWaiting)
      video.removeEventListener('playing', onPlaying)
      video.removeEventListener('canplay', onCanPlay)
      video.removeEventListener('loadedmetadata', onDurationChange)
      video.removeEventListener('durationchange', onDurationChange)
      video.removeEventListener('loadeddata', onDurationChange)
    }
  }, [mediaId, setLastPosition, updateProgress, streamUrls])

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
      // When locked, only allow 'l' to unlock (long-press not needed for keyboard)
      if (isLocked) {
        if (e.key === 'l') {
          e.preventDefault()
          setIsLocked(false)
          resetControlsTimeout()
        }
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
          seek(-SEEK_STEP * 2)
          break
        case 'l':
          e.preventDefault()
          seek(SEEK_STEP * 2)
          break
        case '0':
        case 'Home':
          e.preventDefault()
          if (videoRef.current) videoRef.current.currentTime = 0
          break
        case 'Escape':
          if (isFullscreen) {
            e.preventDefault()
            document.exitFullscreen()
          }
          break
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [
    isFullscreen,
    isLocked,
    togglePlay,
    seek,
    changeVolume,
    toggleFullscreen,
    toggleMute,
    resetControlsTimeout,
  ])

  useEffect(() => {
    const onChange = () =>
      setIsFullscreen(
        !!(
          document.fullscreenElement ??
          (document as Document & { webkitFullscreenElement?: Element }).webkitFullscreenElement
        ),
      )
    document.addEventListener('fullscreenchange', onChange)
    document.addEventListener('webkitfullscreenchange', onChange)
    return () => {
      document.removeEventListener('fullscreenchange', onChange)
      document.removeEventListener('webkitfullscreenchange', onChange)
    }
  }, [])

  const currentQualityLabel =
    maxQuality === 'auto'
      ? 'Auto'
      : (QUALITY_OPTIONS.find(
          (q) => q.height === maxQuality.height && q.bitrateKbps === maxQuality.bitrateKbps,
        )?.label ?? `${maxQuality.height}p`)

  const getActiveSubtitleTrack = (): PlaybackSubtitleTrack | null => {
    if (!effectiveSubtitleLanguage) return null
    return subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage)) || null
  }

  const progressPercent = duration ? (currentTime / duration) * 100 : 0
  const bufferPercent = duration ? (buffered / duration) * 100 : 0
  const remainingTime = duration > 0 ? duration - currentTime : 0
  const VolumeIcon = isMuted || volume === 0 ? LuVolumeX : volume < 0.5 ? LuVolume1 : LuVolume2

  // ── Loading/Error ──────────────────────────────────────────────────────────
  if (mediaLoading || streamLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#141414] text-white">
        <div className="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white" />
      </div>
    )
  }

  if (error || !media || !media.files.length) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#141414] text-white">
        <div className="text-center">
          <p className="text-lg text-red-400">{error || 'Media not found'}</p>
          <button
            onClick={() => navigate(-1)}
            className="mt-4 text-sm text-white/60 hover:text-white"
          >
            Go back
          </button>
        </div>
      </div>
    )
  }

  // ── Render ─────────────────────────────────────────────────────────────────
  return (
    <div
      ref={containerRef}
      className={`relative h-screen w-full bg-[#141414] select-none overflow-hidden ${
        !showControls && isPlaying ? 'cursor-none' : ''
      }`}
      onMouseMove={() => {
        if (!isLocked) resetControlsTimeout()
      }}
      onClick={(e) => {
        if (isLocked) {
          e.stopPropagation()
          e.preventDefault()
        }
      }}
    >
      {/* Video */}
      <video
        ref={videoRef}
        className={`h-full w-full object-${aspectRatio}`}
        playsInline
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onEnded={() => {
          if (repeatMode === 'one') {
            const v = videoRef.current
            if (v) {
              v.currentTime = 0
              v.play().catch(() => {})
            }
            return
          }
          if (repeatMode === 'all' && nextEpisodeMediaId) {
            navigate(`/watch/${nextEpisodeMediaId}`)
            return
          }
          setIsPlaying(false)
          updateProgress({ mediaId, data: { position: duration, completed: true } })
        }}
        onError={() => setError('Video playback error')}
      />

      {/* Subtitle overlay */}
      <DualSubtitleOverlay
        videoRef={videoRef}
        primaryUrl={subtitleServeUrl(primarySub)}
        secondaryUrl={subtitleServeUrl(secondarySub)}
        currentTime={currentTime}
        style={{ size: subtitleSize, color: subtitleColor, background: subtitleBackground }}
      />

      {/* Buffering spinner */}
      {isBuffering && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <div className="h-12 w-12 animate-spin rounded-full border-2 border-white/20 border-t-white" />
        </div>
      )}

      {/* Seek feedback */}
      {seekFeedback && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <div className="flex items-center gap-2 rounded-full bg-black/50 px-6 py-3 text-white text-base font-medium backdrop-blur-sm">
            {seekFeedback.dir === 'back' ? <LuRotateCcw size={20} /> : <LuRotateCw size={20} />}
            {seekFeedback.dir === 'back' ? '-' : '+'}
            {seekFeedback.n}s
          </div>
        </div>
      )}

      {/* Quality indicator */}
      {showQualityIndicator && availableLevels.length > 0 && (
        <div className="pointer-events-none absolute left-1/2 top-5 -translate-x-1/2 rounded-full bg-black/60 px-4 py-1 text-sm text-white/90">
          {(() => {
            const h = availableLevels.find(
              (l) =>
                l.level === (currentLevel === -1 ? hlsRef.current?.currentLevel : currentLevel),
            )?.height
            return (
              <>
                {h ? `${h}p` : 'Auto'}
                {bandwidth !== null && ` · ${bandwidth.toFixed(1)} Mbps`}
              </>
            )
          })()}
        </div>
      )}

      {/* Playback Info overlay (Emby-style) */}
      {showStats && (
        <div
          className="absolute left-4 top-20 z-30 w-80 rounded-xl bg-black/70 backdrop-blur-md ring-1 ring-white/10 overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Close button */}
          <button
            onClick={() => setShowStats(false)}
            className="absolute right-2.5 top-2.5 rounded-lg p-1 text-white/40 hover:bg-white/10 hover:text-white"
          >
            <LuX size={16} />
          </button>

          {playbackInfo ? (
            <div className="space-y-0">
              {/* Stream */}
              <div className="border-b border-white/10 px-4 py-3">
                <p className="mb-2 text-sm font-bold text-white">Stream</p>
                <p className="font-mono text-xs leading-relaxed text-white/80">
                  {playbackInfo.container?.toUpperCase() || '—'}
                  {playbackInfo.bitrate > 0 &&
                    ` (${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.bitrate} kbps`})`}
                </p>
                {playbackInfo.method === 'DirectPlay' && (
                  <p className="font-mono text-xs leading-relaxed text-white/80">
                    <span className="text-white/50">→ </span>Direct Play
                  </p>
                )}
                {playbackInfo.method !== 'DirectPlay' && (
                  <p className="font-mono text-xs leading-relaxed text-white/80">
                    <span className="text-white/50">→ </span>
                    HLS
                    {playbackInfo.estimated_bitrate > 0
                      ? ` (${playbackInfo.estimated_bitrate >= 1000 ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.estimated_bitrate} kbps`})`
                      : playbackInfo.bitrate > 0
                        ? ` (${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.bitrate} kbps`})`
                        : ''}
                  </p>
                )}
                {playbackInfo.method === 'TranscodeAudio' && (
                  <p className="mt-1 text-[11px] text-white/50">
                    Converting audio to compatible codec
                  </p>
                )}
              </div>

              {/* Video */}
              <div className="border-b border-white/10 px-4 py-3">
                <p className="mb-2 text-sm font-bold text-white">Video</p>
                <p className="font-mono text-xs leading-relaxed text-white/80">
                  {playbackInfo.height > 0 ? `${playbackInfo.height}p` : '—'}{' '}
                  {playbackInfo.video_codec?.toUpperCase() || ''}
                </p>
                <p className="font-mono text-xs leading-relaxed text-white/80">
                  {playbackInfo.video_profile && `${playbackInfo.video_profile} `}
                  {playbackInfo.video_level > 0 && `${playbackInfo.video_level} `}
                  {playbackInfo.bitrate > 0 &&
                    `${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(0)} mbps` : `${playbackInfo.bitrate} kbps`} `}
                  {(() => {
                    if (playbackInfo.video_fps > 0) {
                      return `${Number.isInteger(playbackInfo.video_fps) ? playbackInfo.video_fps : playbackInfo.video_fps.toFixed(3)} fps`
                    }
                    // Fallback: estimate from decoded frames / currentTime
                    const v = videoRef.current
                    if (v && 'getVideoPlaybackQuality' in v && v.currentTime > 2) {
                      const q = v.getVideoPlaybackQuality()
                      if (q.totalVideoFrames > 0) {
                        return `${(q.totalVideoFrames / v.currentTime).toFixed(3)} fps`
                      }
                    }
                    return ''
                  })()}
                </p>
                {playbackInfo.method === 'FullTranscode' && (
                  <p className="font-mono text-xs leading-relaxed text-white/80">
                    <span className="text-white/50">→ </span>
                    Transcode ({playbackInfo.video_codec?.toUpperCase() || 'H264'}
                    {playbackInfo.estimated_bitrate > 0 &&
                      ` ${playbackInfo.estimated_bitrate >= 1000 ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(0)} mbps` : `${playbackInfo.estimated_bitrate} kbps`}`}
                    )
                  </p>
                )}
                {(playbackInfo.method === 'DirectPlay' ||
                  playbackInfo.method === 'DirectStream' ||
                  playbackInfo.method === 'TranscodeAudio') && (
                  <p className="font-mono text-xs leading-relaxed text-white/80">
                    <span className="text-white/50">→ </span>Direct Play
                  </p>
                )}
                <p className="mt-1.5 font-mono text-xs text-white/80">
                  Dropped Frames{' '}
                  <span
                    className={(() => {
                      const dropped =
                        videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0
                      return dropped > 0 ? 'text-yellow-400' : 'text-white/80'
                    })()}
                  >
                    {videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0}
                  </span>
                </p>
              </div>

              {/* Audio */}
              {(() => {
                const selectedAudio =
                  playbackInfo.audio_tracks?.find((t) => t.selected) ??
                  playbackInfo.audio_tracks?.find((t) => t.is_default) ??
                  playbackInfo.audio_tracks?.[0]
                if (!selectedAudio) return null
                const isTranscoding =
                  playbackInfo.method === 'FullTranscode' ||
                  playbackInfo.method === 'TranscodeAudio'
                const channelLabel =
                  selectedAudio.channels >= 6
                    ? selectedAudio.channels === 6
                      ? '5.1'
                      : selectedAudio.channels === 8
                        ? '7.1'
                        : `${selectedAudio.channels}ch`
                    : selectedAudio.channels > 0
                      ? `${selectedAudio.channels}.0`
                      : ''
                return (
                  <div className="px-4 py-3">
                    <p className="mb-2 text-sm font-bold text-white">Audio</p>
                    <p className="font-mono text-xs leading-relaxed text-white/80">
                      {selectedAudio.language || 'Unknown'}{' '}
                      {selectedAudio.codec?.toUpperCase() || ''} {channelLabel}
                      {selectedAudio.is_default && ' (Default)'}
                    </p>
                    <p className="font-mono text-xs leading-relaxed text-white/80">
                      {selectedAudio.bitrate > 0 &&
                        `${selectedAudio.bitrate >= 1000 ? `${Math.round(selectedAudio.bitrate / 1000)} kbps` : `${selectedAudio.bitrate} bps`} `}
                      {selectedAudio.sample_rate > 0 && `${selectedAudio.sample_rate} Hz`}
                    </p>
                    <p className="font-mono text-xs leading-relaxed text-white/80">
                      <span className="text-white/50">→ </span>
                      {isTranscoding
                        ? `Transcode (${playbackInfo.audio_codec?.toUpperCase() || 'AAC'})`
                        : 'Direct Play'}
                    </p>
                  </div>
                )
              })()}
            </div>
          ) : (
            <div className="px-4 py-4">
              <p className="text-xs text-white/40">Loading stream info…</p>
            </div>
          )}
        </div>
      )}

      {/* Up Next card */}
      {showUpNext && (
        <div
          className="absolute bottom-44 right-6 z-20 w-64 rounded-xl bg-[#1e1e1e] p-4 shadow-2xl ring-1 ring-white/10"
          onClick={(e) => e.stopPropagation()}
        >
          <p className="mb-1 text-xs text-white/50">Up next</p>
          <p className="mb-3 text-sm font-semibold text-white line-clamp-2">{nextEpisode?.title}</p>
          <div className="flex gap-2">
            <button
              onClick={() => navigate(`/watch/${nextEpisodeMediaId}`)}
              className="flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-netflix-red px-3 py-2 text-sm font-medium text-white hover:bg-netflix-red/90"
            >
              <LuPlay size={13} className="fill-white" /> Play Next
            </button>
            <button
              onClick={() => setUpNextDismissed(true)}
              className="rounded-lg bg-white/10 px-3 py-2 text-sm text-white/70 hover:bg-white/15"
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {/* Screen lock overlay */}
      {isLocked && (
        <div
          className="absolute inset-0 z-40"
          onClick={(e) => {
            e.stopPropagation()
            e.preventDefault()
          }}
        >
          {/* Unlock button — always visible when locked */}
          <button
            onClick={(e) => {
              e.stopPropagation()
              setIsLocked(false)
              resetControlsTimeout()
            }}
            className="absolute bottom-6 left-1/2 -translate-x-1/2 flex items-center gap-2 rounded-full bg-black/60 px-5 py-2.5 text-white/80 backdrop-blur-sm transition-opacity hover:text-white"
          >
            <LuLockOpen size={18} />
            <span className="text-sm font-medium">Tap to unlock</span>
          </button>
        </div>
      )}

      {/* ── Controls overlay ─────────────────────────────────────────────────── */}
      <div
        className={`absolute inset-0 flex flex-col justify-between transition-opacity duration-300 ${
          isLocked
            ? 'opacity-0 pointer-events-none'
            : showControls || isHoveringBar || isDraggingBar
              ? 'opacity-100'
              : 'opacity-0 pointer-events-none'
        }`}
        onClick={togglePlay}
      >
        {/* ── Top bar ──────────────────────────────────────────────────────── */}
        <div
          className="flex items-center justify-between px-5 py-4"
          style={{ background: 'linear-gradient(to bottom, rgba(0,0,0,0.7) 0%, transparent 100%)' }}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Back / breadcrumb */}
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-1.5 text-white/80 transition-colors hover:text-white"
          >
            <LuChevronLeft size={22} />
            <span className="text-sm font-medium">Back</span>
          </button>

          {/* Volume (top right) */}
          <div className="flex items-center gap-2">
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
              className="h-0.5 w-28 cursor-pointer accent-white"
            />
            <button
              onClick={toggleMute}
              className="text-white/80 hover:text-white transition-colors"
            >
              <VolumeIcon size={20} />
            </button>
          </div>
        </div>

        {/* ── Center: pause indicator (brief) ──────────────────────────────── */}
        {!isPlaying && !isBuffering && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="rounded-full bg-black/30 p-5 backdrop-blur-sm">
              <LuPlay size={44} className="text-white fill-white ml-1" />
            </div>
          </div>
        )}

        {/* ── Bottom panel ─────────────────────────────────────────────────── */}
        <div
          style={{
            background:
              'linear-gradient(to top, rgba(0,0,0,0.92) 0%, rgba(0,0,0,0.7) 70%, transparent 100%)',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Info panel (slides in above controls) */}
          {activeTab === 'info' && media.media.overview && (
            <div className="px-6 pb-2 pt-4">
              <p className="max-w-2xl text-sm leading-relaxed text-white/70">
                {media.media.overview}
              </p>
            </div>
          )}

          <div className="px-6 pb-4 pt-3 space-y-2">
            {/* Row 1: Title + icon buttons */}
            <div className="flex items-start justify-between gap-4">
              <h1 className="text-xl font-bold text-white leading-tight drop-shadow">
                {media.media.title}
              </h1>

              {/* Right icon buttons */}
              <div className="flex shrink-0 items-center gap-1.5">
                {/* Subtitles — always visible so users can search for subs */}
                <div className="relative">
                  <button
                    onClick={() => {
                      setShowSubtitleMenu(!showSubtitleMenu)
                      setShowAudioMenu(false)
                      setShowSpeedMenu(false)
                      setShowSettings(false)
                    }}
                    className={`flex h-9 w-9 items-center justify-center rounded-lg border transition-colors ${
                      getActiveSubtitleTrack()
                        ? 'border-white bg-white/20 text-white'
                        : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                    }`}
                    title="Subtitles"
                  >
                    <LuCaptions size={18} />
                  </button>
                  {showSubtitleMenu && (
                    <div className="absolute bottom-full right-0 mb-2">
                      <SubtitlePicker
                        subtitles={subtitles}
                        primaryLanguage={effectiveSubtitleLanguage}
                        secondaryLanguage={secondarySubtitleLanguage}
                        onSelectPrimary={(lang) => {
                          setSubtitleLanguage(lang)
                          setShowSubtitleMenu(false)
                        }}
                        onSelectSecondary={setSecondarySubtitleLanguage}
                        dualMode={true}
                        allowImageSubtitles={allowsImageSubtitles}
                        mediaId={mediaId}
                        onSubtitleAdded={() => {
                          queryClient.invalidateQueries({ queryKey: streamingKeys.all })
                        }}
                      />
                    </div>
                  )}
                </div>

                {/* Audio tracks */}
                {audioTracks.length > 0 && (
                  <div className="relative">
                    <button
                      onClick={() => {
                        setShowAudioMenu(!showAudioMenu)
                        setShowSubtitleMenu(false)
                        setShowSpeedMenu(false)
                        setShowSettings(false)
                      }}
                      className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white"
                      title="Audio"
                    >
                      <LuMusic size={17} />
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

                {/* Speed */}
                <div className="relative">
                  <button
                    onClick={() => {
                      setShowSpeedMenu(!showSpeedMenu)
                      setShowSubtitleMenu(false)
                      setShowAudioMenu(false)
                      setShowSettings(false)
                    }}
                    className={`flex h-9 min-w-[36px] items-center justify-center rounded-lg border px-2 transition-colors ${
                      playbackRate !== 1
                        ? 'border-white bg-white/20 text-white'
                        : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                    }`}
                    title="Speed"
                  >
                    <span className="text-xs font-bold tabular-nums">
                      {playbackRate === 1 ? '1×' : `${playbackRate}×`}
                    </span>
                  </button>
                  {showSpeedMenu && (
                    <div className="absolute bottom-full right-0 mb-2 w-44 rounded-xl bg-[#1e1e1e] py-2 shadow-2xl ring-1 ring-white/10">
                      <p className="px-4 pb-1.5 pt-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                        Playback Speed
                      </p>
                      {[0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2].map((rate) => (
                        <button
                          key={rate}
                          onClick={() => {
                            setPlaybackRate(rate)
                            if (videoRef.current) videoRef.current.playbackRate = rate
                            setShowSpeedMenu(false)
                          }}
                          className={`flex w-full items-center gap-3 px-4 py-2 text-sm transition-colors ${
                            playbackRate === rate
                              ? 'text-white'
                              : 'text-white/60 hover:bg-white/8 hover:text-white'
                          }`}
                        >
                          <span className="w-3 text-center text-white">
                            {playbackRate === rate ? '✓' : ''}
                          </span>
                          {rate === 1 ? 'Normal' : `${rate}x`}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Settings */}
                <div className="relative">
                  <button
                    onClick={() => {
                      const willShow = !showSettings
                      setShowSettings(willShow)
                      if (willShow) setSettingsView('main')
                      setShowSubtitleMenu(false)
                      setShowAudioMenu(false)
                      setShowSpeedMenu(false)
                    }}
                    className={`flex h-9 w-9 items-center justify-center rounded-lg border transition-colors ${
                      showSettings
                        ? 'border-white bg-white/20 text-white'
                        : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                    }`}
                    title="Settings"
                  >
                    <LuSettings size={17} />
                  </button>
                  {showSettings && (
                    <div className="absolute bottom-full right-0 mb-2 w-56 rounded-xl bg-[#1e1e1e] shadow-2xl ring-1 ring-white/10 overflow-hidden">
                      {settingsView === 'quality' ? (
                        /* Quality submenu */
                        <div className="flex flex-col">
                          <button
                            onClick={() => setSettingsView('main')}
                            className="flex items-center gap-2 border-b border-white/10 px-4 py-2.5 text-xs font-semibold text-white/70 hover:text-white"
                          >
                            <LuChevronLeft size={14} />
                            Quality
                          </button>
                          <div className="max-h-[50vh] overflow-y-auto py-1">
                            {QUALITY_OPTIONS.map((q) => {
                              const isSelected =
                                maxQuality !== 'auto' &&
                                maxQuality.height === q.height &&
                                maxQuality.bitrateKbps === q.bitrateKbps
                              return (
                                <button
                                  key={`${q.height}-${q.bitrateKbps}`}
                                  onClick={() => {
                                    setMaxQuality({
                                      height: q.height,
                                      bitrateKbps: q.bitrateKbps,
                                    })
                                    setSettingsView('main')
                                  }}
                                  className={`flex w-full items-center justify-between px-4 py-2 text-xs ${
                                    isSelected
                                      ? 'bg-white/10 text-white'
                                      : 'text-white/70 hover:bg-white/5 hover:text-white'
                                  }`}
                                >
                                  {q.label}
                                  {isSelected && <LuCheck size={14} className="text-white" />}
                                </button>
                              )
                            })}
                            <button
                              onClick={() => {
                                setMaxQuality('auto')
                                setSettingsView('main')
                              }}
                              className={`flex w-full items-center justify-between px-4 py-2 text-xs ${
                                maxQuality === 'auto'
                                  ? 'bg-white/10 text-white'
                                  : 'text-white/70 hover:bg-white/5 hover:text-white'
                              }`}
                            >
                              Auto
                              {maxQuality === 'auto' && (
                                <LuCheck size={14} className="text-white" />
                              )}
                            </button>
                          </div>
                        </div>
                      ) : (
                        /* Main settings view */
                        <div className="space-y-3 p-3">
                          {/* Aspect Ratio */}
                          <div>
                            <p className="mb-1.5 flex items-center gap-1.5 px-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                              <LuExpand size={10} /> Aspect Ratio
                            </p>
                            <div className="grid grid-cols-3 gap-1">
                              {(['contain', 'cover', 'fill'] as const).map((r) => (
                                <button
                                  key={r}
                                  onClick={() => setAspectRatio(r)}
                                  className={`rounded-lg py-1.5 text-xs font-medium capitalize ${
                                    aspectRatio === r
                                      ? 'bg-white/20 text-white'
                                      : 'bg-white/5 text-white/70 hover:bg-white/15 hover:text-white'
                                  }`}
                                >
                                  {r === 'contain'
                                    ? 'Auto'
                                    : r.charAt(0).toUpperCase() + r.slice(1)}
                                </button>
                              ))}
                            </div>
                          </div>

                          {/* Subtitle Appearance */}
                          <div className="border-t border-white/10 pt-3">
                            <p className="mb-1.5 flex items-center gap-1.5 px-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                              <LuCaptions size={10} /> Subtitles
                            </p>
                            <div className="grid grid-cols-3 gap-1 mb-2">
                              {(['small', 'medium', 'large'] as const).map((s) => (
                                <button
                                  key={s}
                                  onClick={() => setSubtitleSize(s)}
                                  className={`rounded-lg py-1.5 text-xs font-medium capitalize ${
                                    subtitleSize === s
                                      ? 'bg-white/20 text-white'
                                      : 'bg-white/5 text-white/70 hover:bg-white/15 hover:text-white'
                                  }`}
                                >
                                  {s === 'small' ? 'S' : s === 'medium' ? 'M' : 'L'}
                                </button>
                              ))}
                            </div>
                            <div className="grid grid-cols-3 gap-1 mb-2">
                              {[
                                { value: 'none' as const, label: 'None' },
                                { value: 'semi' as const, label: 'Semi' },
                                { value: 'solid' as const, label: 'Solid' },
                              ].map(({ value, label }) => (
                                <button
                                  key={value}
                                  onClick={() => setSubtitleBackground(value)}
                                  className={`rounded-lg py-1.5 text-xs font-medium ${
                                    subtitleBackground === value
                                      ? 'bg-white/20 text-white'
                                      : 'bg-white/5 text-white/70 hover:bg-white/15 hover:text-white'
                                  }`}
                                >
                                  {label}
                                </button>
                              ))}
                            </div>
                            <div className="flex items-center gap-1.5">
                              {['#ffffff', '#fde047', '#4ade80', '#60a5fa'].map((c) => (
                                <button
                                  key={c}
                                  onClick={() => setSubtitleColor(c)}
                                  className={`h-5 w-5 rounded-full border-2 ${
                                    subtitleColor === c ? 'border-white' : 'border-white/20'
                                  }`}
                                  style={{ background: c }}
                                />
                              ))}
                            </div>
                          </div>

                          {/* Quality — clickable row that opens submenu */}
                          <div className="border-t border-white/10 pt-3">
                            <button
                              onClick={() => setSettingsView('quality')}
                              className="flex w-full items-center justify-between rounded-lg px-3 py-1.5 text-xs text-white/70 hover:bg-white/10 hover:text-white"
                            >
                              <span className="flex items-center gap-1.5">
                                <LuZap size={13} /> Quality
                              </span>
                              <span className="flex items-center gap-1 text-white/50">
                                {currentQualityLabel}
                                <LuChevronRight size={14} />
                              </span>
                            </button>
                          </div>

                          {/* Repeat Mode */}
                          <div className="border-t border-white/10 pt-3">
                            <p className="mb-1.5 flex items-center gap-1.5 px-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                              <LuRepeat size={10} /> Repeat
                            </p>
                            <div className="grid grid-cols-3 gap-1">
                              {(['none', 'one', 'all'] as const).map((m) => (
                                <button
                                  key={m}
                                  onClick={() => setRepeatMode(m)}
                                  className={`flex items-center justify-center gap-1 rounded-lg py-1.5 text-xs font-medium ${
                                    repeatMode === m
                                      ? 'bg-white/20 text-white'
                                      : 'bg-white/5 text-white/70 hover:bg-white/15 hover:text-white'
                                  }`}
                                >
                                  {m === 'none' && <LuRepeat size={12} />}
                                  {m === 'one' && <LuRepeat2 size={12} />}
                                  {m === 'all' && <LuRepeat size={12} />}
                                  <span className="capitalize">{m}</span>
                                </button>
                              ))}
                            </div>
                          </div>

                          {/* Playback Info */}
                          <div className="border-t border-white/10 pt-3">
                            <button
                              onClick={() => {
                                setShowStats(true)
                                setShowSettings(false)
                              }}
                              className="flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-white/70 hover:bg-white/10 hover:text-white"
                            >
                              <LuActivity size={13} /> Playback Info
                            </button>
                          </div>

                          {/* More */}
                          <div className="border-t border-white/10 pt-3">
                            <button
                              onClick={() => navigate(-1)}
                              className="flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-white/70 hover:bg-white/10 hover:text-white"
                            >
                              <LuExternalLink size={13} /> Back
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>

                {/* Next episode */}
                {nextEpisodeMediaId && (
                  <button
                    onClick={() => navigate(`/watch/${nextEpisodeMediaId}`)}
                    className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white"
                    title="Next episode"
                  >
                    <LuSkipForward size={17} />
                  </button>
                )}

                {/* Screen lock */}
                <button
                  onClick={() => {
                    setIsLocked(true)
                    setShowControls(false)
                  }}
                  className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white"
                  title="Lock screen"
                >
                  <LuLock size={17} />
                </button>

                {/* Fullscreen */}
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    toggleFullscreen()
                  }}
                  className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white"
                  title={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
                >
                  {isFullscreen ? <LuMinimize2 size={17} /> : <LuMaximize2 size={17} />}
                </button>
              </div>
            </div>

            {/* Row 2: Progress bar */}
            <div
              ref={progressBarRef}
              className="group relative flex h-5 cursor-pointer items-center"
              onMouseEnter={() => setIsHoveringBar(true)}
              onMouseLeave={() => setIsHoveringBar(false)}
              onMouseMove={handleBarMouseMove}
              onMouseDown={handleBarMouseDown}
            >
              <TrickplayPreview
                mediaId={mediaId}
                currentHoverTime={hoverTime}
                visible={isHoveringBar && duration > 0}
                positionX={hoverX}
              />
              {isHoveringBar && duration > 0 && (
                <div
                  className="pointer-events-none absolute -top-8 -translate-x-1/2 rounded bg-black/80 px-2 py-0.5 text-xs text-white"
                  style={{ left: hoverX }}
                >
                  {formatTime(hoverTime)}
                </div>
              )}
              <div
                className={`relative w-full rounded-full bg-white/20 transition-all duration-150 ${isHoveringBar || isDraggingBar ? 'h-1.5' : 'h-[3px]'}`}
              >
                <div
                  className="absolute inset-y-0 left-0 rounded-full bg-white/20"
                  style={{ width: `${bufferPercent}%` }}
                />
                <div
                  className="absolute inset-y-0 left-0 rounded-full bg-white"
                  style={{ width: `${progressPercent}%` }}
                />
                <div
                  className={`absolute top-1/2 -translate-x-1/2 -translate-y-1/2 h-3.5 w-3.5 rounded-full bg-white shadow transition-opacity ${isHoveringBar || isDraggingBar ? 'opacity-100' : 'opacity-0'}`}
                  style={{ left: `${progressPercent}%` }}
                />
              </div>
            </div>

            {/* Row 3: Transport + time */}
            <div className="flex items-center justify-between">
              {/* Left: time + transport */}
              <div className="flex items-center gap-4">
                <span className="text-sm tabular-nums text-white/70">
                  {formatTime(currentTime)}
                </span>

                <div className="flex items-center gap-3">
                  <button
                    onClick={() => seek(-SEEK_STEP)}
                    className="flex items-center gap-0.5 text-white/75 transition-colors hover:text-white"
                    title="Rewind 10s"
                  >
                    <LuRotateCcw size={22} />
                    <span className="text-[11px] font-bold">10</span>
                  </button>

                  <button
                    onClick={togglePlay}
                    className="text-white transition-transform hover:scale-105"
                  >
                    {isPlaying ? (
                      <LuPause size={26} className="fill-white" />
                    ) : (
                      <LuPlay size={26} className="fill-white" />
                    )}
                  </button>

                  <button
                    onClick={() => seek(SEEK_STEP)}
                    className="flex items-center gap-0.5 text-white/75 transition-colors hover:text-white"
                    title="Forward 10s"
                  >
                    <span className="text-[11px] font-bold">10</span>
                    <LuRotateCw size={22} />
                  </button>
                </div>
              </div>

              {/* Right: remaining */}
              <span className="text-sm tabular-nums text-white/50">
                {duration > 0 ? `-${formatTime(remainingTime)}` : wallClock}
              </span>
            </div>

            {/* Row 4: Tabs */}
            <div className="flex items-center gap-5 pt-1">
              <button
                onClick={() => setActiveTab(activeTab === 'info' ? 'none' : 'info')}
                className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
                  activeTab === 'info' ? 'text-white' : 'text-white/45 hover:text-white/80'
                }`}
              >
                <LuInfo size={14} />
                Info
              </button>
              <button
                onClick={() => setActiveTab(activeTab === 'chapters' ? 'none' : 'chapters')}
                className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
                  activeTab === 'chapters' ? 'text-white' : 'text-white/45 hover:text-white/80'
                }`}
              >
                <LuListMusic size={14} />
                Chapters
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function getWallClock(): string {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatTime(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const hours = Math.floor(mins / 60)
  if (hours > 0) {
    return `${hours}:${(mins % 60).toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
