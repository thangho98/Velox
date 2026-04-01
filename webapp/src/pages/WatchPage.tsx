import { useParams, useNavigate } from 'react-router'
import { useEffect, useRef, useState, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import Hls from 'hls.js'
import {
  LuActivity,
  LuChevronLeft,
  LuChevronRight,
  LuPlay,
  LuPause,
  LuMaximize2,
  LuMinimize2,
  LuSettings,
  LuSkipForward,
  LuCaptions,
  LuMusic,
  LuZap,
  LuRotateCcw,
  LuRotateCw,
  LuRepeat,
  LuRepeat2,
  LuExternalLink,
  LuExpand,
  LuInfo,
  LuLock,
  LuLockOpen,
  LuListMusic,
  LuCheck,
} from 'react-icons/lu'
import {
  useMediaWithFiles,
  useUpdateProgress,
  useStreamUrls,
  useSubtitles,
  useAudioTracks,
  useSeasons,
  useEpisodes,
  usePlaybackInfo,
  streamingKeys,
} from '@/hooks/stores/useMedia'
import { usePreferences } from '@/hooks/stores/useAuth'
import { usePlayerStore } from '@/stores/player'
import { useAuthStore } from '@/stores/auth'
import { getCapabilities } from '@/lib/capabilities'
import { tmdbImage } from '@/lib/image'
import { DualSubtitleOverlay } from '@/components/DualSubtitleOverlay'
import { SubtitlePicker } from '@/components/SubtitlePicker'
import { AudioPicker } from '@/components/AudioPicker'
import { TrickplayPreview } from '@/components/TrickplayPreview'
import { useToast } from '@/components/Toast'
import { WatchDetailSheet } from '@/components/watch/WatchDetailSheet'
import { WatchPlaybackStatsOverlay } from '@/components/watch/WatchPlaybackStatsOverlay'
import { WatchTopBar } from '@/components/watch/WatchTopBar'
import { SkipIntroCredits } from '@/components/watch/SkipIntroCredits'
import { useChromecast } from '@/hooks/useChromecast'
import { useTranslation } from '@/hooks/useTranslation'
import {
  DETAIL_PANEL_ANIMATION_MS,
  formatChannelLayout,
  formatLanguageLabel,
  formatResolutionLabel,
  formatRuntimeMinutes,
  formatTime,
  getWallClock,
  languageMatches,
} from '@/components/watch/watchHelpers'
import type { PlaybackSubtitleTrack } from '@/types/api'

const SEEK_STEP = 5
const VOLUME_STEP = 0.1
type DetailPanel = 'none' | 'info' | 'season'

import type { QualityOption } from '@/types/api'

export function WatchPage() {
  const { id } = useParams<{ id: string }>()
  const mediaId = Number(id)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const videoRef = useRef<HTMLVideoElement>(null)
  const hlsRef = useRef<Hls | null>(null)
  const hlsPlaylistLiveRef = useRef(false)
  const hlsSeekableEndRef = useRef(0)
  const streamSourceOffsetRef = useRef(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const progressBarRef = useRef<HTMLDivElement>(null)
  const seasonCarouselRef = useRef<HTMLDivElement>(null)
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastProgressUpdate = useRef(0)
  const dragSeekTimeRef = useRef<number | null>(null)
  const seekFeedbackTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const qualityIndicatorTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lowBandwidthToastShown = useRef(false)
  const { t } = useTranslation('watch')

  const {
    available: castAvailable,
    connected: castConnected,
    casting,
    castMedia,
    requestSession: requestCast,
    stopCasting,
  } = useChromecast()
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
    subtitleTrackId,
    setSubtitleLanguage,
    secondarySubtitleLanguage,
    secondarySubtitleTrackId,
    setSecondarySubtitleLanguage,
    setSubtitleTrackId,
    setSecondarySubtitleTrackId,
    subtitleSize,
    setSubtitleSize,
    subtitleColor,
    setSubtitleColor,
    subtitleBackground,
    setSubtitleBackground,
    getSubtitleOffset,
    setSubtitleOffset,
    resetSubtitleOffset,
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
  const subtitleOffsetSeconds = getSubtitleOffset(mediaId)

  const clientCaps = getCapabilities()
  const effectiveSubtitleLanguage = subtitleLanguage ?? preferences?.subtitle_language ?? null
  const qualityMaxHeight = maxQuality === 'auto' ? undefined : maxQuality

  const playbackRequest = {
    video_codecs: clientCaps.videoCodecs,
    audio_codecs: clientCaps.audioCodecs,
    containers: clientCaps.containers,
    max_height: qualityMaxHeight,
    selected_subtitle: effectiveSubtitleLanguage ?? 'off',
    selected_subtitle_id: subtitleTrackId ?? 0,
    selected_audio_track: audioTrackId ?? 0,
  }
  const { data: streamUrls, isLoading: streamLoading } = useStreamUrls(mediaId, playbackRequest)
  const { data: subtitles = [] } = useSubtitles(mediaId, playbackRequest)
  const { data: audioTracks = [] } = useAudioTracks(mediaId, playbackRequest)
  const { data: playbackInfo } = usePlaybackInfo(mediaId, playbackRequest)
  const isHlsPlayback =
    playbackInfo?.method === 'FullTranscode' || playbackInfo?.method === 'TranscodeAudio'

  const isEpisode = media?.media.media_type === 'episode'
  const seriesId = media?.series_id ?? 0
  const seasonId = media?.season_id ?? 0

  // Handle back navigation - go to detail page instead of history back
  const handleBack = () => {
    if (isEpisode && seriesId > 0) {
      navigate(`/series/${seriesId}`)
    } else {
      navigate(`/movie/${mediaId}`)
    }
  }
  const { data: seasons = [] } = useSeasons(isEpisode ? seriesId : 0)
  const [seasonPanelSeasonId, setSeasonPanelSeasonId] = useState(0)
  const { data: seasonEpisodes = [] } = useEpisodes(
    isEpisode ? seriesId : 0,
    isEpisode ? seasonId : 0,
  )
  const { data: seasonPanelEpisodes = [] } = useEpisodes(
    isEpisode ? seriesId : 0,
    isEpisode ? seasonPanelSeasonId : 0,
  )
  const nextEpisode = (() => {
    if (!isEpisode || seasonEpisodes.length === 0) return null
    const currentIdx = seasonEpisodes.findIndex((ep) => ep.media_id === mediaId)
    if (currentIdx === -1 || currentIdx === seasonEpisodes.length - 1) return null
    return seasonEpisodes[currentIdx + 1]
  })()
  const nextEpisodeMediaId = nextEpisode?.media_id

  useEffect(() => {
    if (!isEpisode || seasonId <= 0) {
      setSeasonPanelSeasonId(0)
      return
    }
    setSeasonPanelSeasonId((current) => (current > 0 ? current : seasonId))
  }, [isEpisode, seasonId])

  useEffect(() => {
    setAudioTrack(audioLanguage, null)
    // audio track IDs are file-specific; never carry them across media items
  }, [mediaId, audioLanguage, setAudioTrack])

  useEffect(() => {
    setSubtitleTrackId(null)
    setSecondarySubtitleTrackId(null)
    // subtitle track IDs are file-specific; never carry them across media items
  }, [mediaId, setSecondarySubtitleTrackId, setSubtitleTrackId])

  useEffect(() => {
    if (audioTrackId == null || audioTracks.length === 0) return
    const selectedTrack = audioTracks.find((track) => track.id === audioTrackId)
    if (!selectedTrack || selectedTrack.is_default) {
      setAudioTrack(audioLanguage, null)
    }
  }, [audioLanguage, audioTrackId, audioTracks, setAudioTrack])

  useEffect(() => {
    if (subtitleTrackId == null) return
    const selectedTrack = subtitles.find((track) => track.id === subtitleTrackId)
    if (!selectedTrack || !languageMatches(selectedTrack.language, effectiveSubtitleLanguage)) {
      setSubtitleTrackId(null)
    }
  }, [effectiveSubtitleLanguage, setSubtitleTrackId, subtitleTrackId, subtitles])

  useEffect(() => {
    if (secondarySubtitleTrackId == null) return
    const selectedTrack = subtitles.find((track) => track.id === secondarySubtitleTrackId)
    if (!selectedTrack || !languageMatches(selectedTrack.language, secondarySubtitleLanguage)) {
      setSecondarySubtitleTrackId(null)
    }
  }, [secondarySubtitleLanguage, secondarySubtitleTrackId, setSecondarySubtitleTrackId, subtitles])

  // ── Media Session API — lock screen / notification controls ──────────────
  useEffect(() => {
    if (!('mediaSession' in navigator) || !media) return

    const title = media.media.title ?? ''
    const artist =
      isEpisode && media.season_number && media.episode_number
        ? `S${media.season_number}E${media.episode_number}`
        : ''
    const posterUrl = media.media.poster_path ? tmdbImage(media.media.poster_path, 'w342') : ''

    navigator.mediaSession.metadata = new MediaMetadata({
      title,
      artist,
      album: 'Velox',
      artwork: posterUrl ? [{ src: posterUrl, sizes: '342x513', type: 'image/jpeg' }] : [],
    })

    navigator.mediaSession.setActionHandler('play', () => videoRef.current?.play())
    navigator.mediaSession.setActionHandler('pause', () => videoRef.current?.pause())
    navigator.mediaSession.setActionHandler('seekbackward', (details) => {
      const v = videoRef.current
      if (v) v.currentTime = Math.max(0, v.currentTime - (details.seekOffset ?? 10))
    })
    navigator.mediaSession.setActionHandler('seekforward', (details) => {
      const v = videoRef.current
      if (v) v.currentTime = Math.min(v.duration || 0, v.currentTime + (details.seekOffset ?? 10))
    })
    navigator.mediaSession.setActionHandler('seekto', (details) => {
      const v = videoRef.current
      if (v && details.seekTime != null) v.currentTime = details.seekTime
    })

    if (isEpisode && nextEpisodeMediaId) {
      navigator.mediaSession.setActionHandler('nexttrack', () => {
        navigate(`/watch/${nextEpisodeMediaId}`)
      })
    } else {
      navigator.mediaSession.setActionHandler('nexttrack', null)
    }

    return () => {
      navigator.mediaSession.metadata = null
      navigator.mediaSession.setActionHandler('play', null)
      navigator.mediaSession.setActionHandler('pause', null)
      navigator.mediaSession.setActionHandler('seekbackward', null)
      navigator.mediaSession.setActionHandler('seekforward', null)
      navigator.mediaSession.setActionHandler('seekto', null)
      navigator.mediaSession.setActionHandler('nexttrack', null)
    }
  }, [media, isEpisode, nextEpisodeMediaId, navigate])

  const primaryFileId = streamUrls?.primary_file_id ?? media?.files[0]?.id
  const subtitleServeUrl = (sub: PlaybackSubtitleTrack | undefined) => {
    if (!sub || !primaryFileId) return null
    const base = `/api/media-files/${primaryFileId}/subtitles/${sub.id}/serve`
    return accessToken ? `${base}?token=${encodeURIComponent(accessToken)}` : base
  }
  const primarySub =
    (subtitleTrackId ? subtitles.find((s) => s.id === subtitleTrackId) : undefined) ??
    (effectiveSubtitleLanguage
      ? subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage) && !s.is_image)
      : undefined)
  const burnedInPrimarySub =
    (subtitleTrackId ? subtitles.find((s) => s.id === subtitleTrackId && s.is_image) : undefined) ??
    (effectiveSubtitleLanguage
      ? subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage) && s.is_image)
      : undefined)
  const secondarySub =
    (secondarySubtitleTrackId
      ? subtitles.find((s) => s.id === secondarySubtitleTrackId)
      : undefined) ??
    (secondarySubtitleLanguage
      ? subtitles.find((s) => languageMatches(s.language, secondarySubtitleLanguage) && !s.is_image)
      : undefined)
  const primaryMediaFile = media?.files.find((file) => file.is_primary) ?? media?.files[0]
  const selectedAudioTrack =
    playbackInfo?.audio_tracks?.find((track) => track.selected) ??
    playbackInfo?.audio_tracks?.find((track) => track.is_default) ??
    playbackInfo?.audio_tracks?.[0]
  const infoLogoUrl = tmdbImage(media?.media.logo_path, 'w500') ?? null
  const infoYear = media?.media.release_date
    ? new Date(media.media.release_date).getFullYear()
    : null
  const infoRuntime = formatRuntimeMinutes(
    playbackInfo?.duration ?? primaryMediaFile?.duration ?? 0,
  )
  const infoCC = subtitles.length > 0
  const infoResolution = formatResolutionLabel(
    playbackInfo?.height ?? primaryMediaFile?.height ?? 0,
  )
  const infoVideoCodec = (
    playbackInfo?.video_codec ??
    primaryMediaFile?.video_codec ??
    ''
  ).toUpperCase()
  const infoAudioLanguage = formatLanguageLabel(selectedAudioTrack?.language)
  const infoAudioCodec = selectedAudioTrack?.codec?.toUpperCase() ?? ''
  const infoAudioChannels = formatChannelLayout(selectedAudioTrack?.channels ?? 0)
  const infoEpisodeLabel =
    isEpisode && media?.season_number && media?.episode_number
      ? `Episode S${media.season_number}E${media.episode_number}`
      : isEpisode && media?.episode_number
        ? `Episode ${media.episode_number}`
        : null

  // Player state
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  // knownDuration: ffprobe-reported total duration — used as floor so the player
  // never shows a partial duration while HLS transcoding is still in progress.
  const knownDurationRef = useRef(0)
  const [bufferedRange, setBufferedRange] = useState({ start: 0, end: 0 })
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

  // Media Session — update position state for lock screen progress bar
  useEffect(() => {
    if (!('mediaSession' in navigator) || !videoRef.current) return
    const v = videoRef.current
    if (Number.isFinite(v.duration) && v.duration > 0) {
      navigator.mediaSession.setPositionState({
        duration: v.duration,
        playbackRate: v.playbackRate,
        position: Math.min(v.currentTime, v.duration),
      })
    }
  }, [currentTime, duration])

  // Wall clock
  const [wallClock, setWallClock] = useState(() => getWallClock())
  useEffect(() => {
    const t = setInterval(() => setWallClock(getWallClock()), 30000)
    return () => clearInterval(t)
  }, [])

  // Sync knownDurationRef and duration state from playback info (ffprobe value).
  // This is the floor: duration state never drops below this even while HLS
  // transcoding is in progress and the live-like playlist only has partial segments.
  useEffect(() => {
    const d = playbackInfo?.duration ?? 0
    if (d > 0) {
      knownDurationRef.current = d
      setDuration((prev) => (prev < d ? d : prev))
    }
  }, [playbackInfo?.duration])

  useEffect(() => {
    setHlsStartOffset(null)
    setBufferedRange({ start: 0, end: 0 })
    streamSourceOffsetRef.current = 0
  }, [mediaId])

  useEffect(() => {
    if (!isHlsPlayback) {
      setHlsStartOffset(null)
      streamSourceOffsetRef.current = 0
    }
  }, [isHlsPlayback])

  // Seek feedback
  const [seekFeedback, setSeekFeedback] = useState<{ dir: 'back' | 'fwd'; n: number } | null>(null)

  // Progress bar hover/drag
  const [isHoveringBar, setIsHoveringBar] = useState(false)
  const [hoverTime, setHoverTime] = useState(0)
  const [hoverX, setHoverX] = useState(0)
  const [isDraggingBar, setIsDraggingBar] = useState(false)
  const [dragSeekTime, setDragSeekTime] = useState<number | null>(null)
  const [hlsStartOffset, setHlsStartOffset] = useState<number | null>(null)

  // Menus
  const [showSubtitleMenu, setShowSubtitleMenu] = useState(false)
  const [showAudioMenu, setShowAudioMenu] = useState(false)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [settingsView, setSettingsView] = useState<'main' | 'quality'>('main')
  const [showStats, setShowStats] = useState(false)

  // Bottom tab: 'none' | 'info' | 'season'
  const [activeTab, setActiveTab] = useState<DetailPanel>('none')
  const [displayPanel, setDisplayPanel] = useState<DetailPanel>('none')
  const [isPanelVisible, setIsPanelVisible] = useState(false)
  const desiredPanel: DetailPanel =
    activeTab === 'season' && isEpisode
      ? 'season'
      : activeTab === 'info' && Boolean(media?.media.overview)
        ? 'info'
        : 'none'
  const isDetailPanelActive = activeTab !== 'none'

  useEffect(() => {
    if (desiredPanel === 'none') {
      setIsPanelVisible(false)
      const timeout = setTimeout(() => setDisplayPanel('none'), DETAIL_PANEL_ANIMATION_MS)
      return () => clearTimeout(timeout)
    }

    setDisplayPanel(desiredPanel)
    const frame = requestAnimationFrame(() => setIsPanelVisible(true))
    return () => cancelAnimationFrame(frame)
  }, [desiredPanel])

  // Screen lock
  const [isLocked, setIsLocked] = useState(false)
  const [showLockedUI, setShowLockedUI] = useState(false)
  const lockedUITimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Up Next — trigger when credits start (Netflix-style) or fallback to 90%
  const [upNextDismissed, setUpNextDismissed] = useState(false)
  const creditsSegment = playbackInfo?.skip_segments?.find((s) => s.type === 'credits')
  const upNextThreshold = creditsSegment
    ? creditsSegment.start // Credits detected → show at credits start
    : duration * 0.9 // No credits → fallback to 90%
  const showUpNext =
    isEpisode &&
    nextEpisodeMediaId != null &&
    duration > 0 &&
    currentTime >= upNextThreshold &&
    !upNextDismissed

  // ── Callbacks ──────────────────────────────────────────────────────────────
  const scrollSeasonCarousel = useCallback((direction: 'prev' | 'next') => {
    const carousel = seasonCarouselRef.current
    if (!carousel) return

    const amount = Math.max(carousel.clientWidth * 0.82, 360)
    carousel.scrollBy({
      left: direction === 'next' ? amount : -amount,
      behavior: 'smooth',
    })
  }, [])

  const toggleDetailPanel = useCallback(
    (panel: Exclude<DetailPanel, 'none'>) => {
      const nextPanel = activeTab === panel ? 'none' : panel
      if (nextPanel !== 'none' && isPlaying) {
        videoRef.current?.pause()
        setIsPlaying(false)
      }
      setActiveTab(nextPanel)
    },
    [activeTab, isPlaying],
  )

  const togglePlay = useCallback(() => {
    const video = videoRef.current
    if (!video) return
    const willPlay = !isPlaying
    if (isPlaying) video.pause()
    else video.play().catch(() => setError('Playback failed'))
    setIsPlaying(willPlay)
    if (willPlay) {
      setActiveTab('none')
    }
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

  const getEffectiveDuration = useCallback(
    (video?: HTMLVideoElement | null) => {
      const nativeDuration =
        video && Number.isFinite(video.duration) && video.duration > 0 ? video.duration : 0
      return Math.max(
        nativeDuration,
        knownDurationRef.current,
        duration,
        playbackInfo?.duration ?? 0,
      )
    },
    [duration, playbackInfo?.duration],
  )

  const clampSeekTarget = useCallback(
    (targetTime: number, video?: HTMLVideoElement | null) => {
      const effectiveDuration = getEffectiveDuration(video)
      if (effectiveDuration <= 0) return Math.max(0, targetTime)
      const seekCeiling = effectiveDuration > 0.25 ? effectiveDuration - 0.25 : effectiveDuration
      return Math.max(0, Math.min(seekCeiling, targetTime))
    },
    [getEffectiveDuration],
  )

  const buildHlsSessionUrl = useCallback((baseUrl: string, startOffset: number) => {
    const url = new URL(baseUrl, window.location.origin)
    url.searchParams.set('start', startOffset.toFixed(3))
    return `${url.pathname}${url.search}${url.hash}`
  }, [])

  const requestHlsSessionReload = useCallback(
    (targetTime: number) => {
      const globalTarget = clampSeekTarget(targetTime, videoRef.current)
      const nextOffset = globalTarget > 0.25 ? globalTarget : null

      setLastPosition(mediaId, globalTarget)
      setCurrentTime(globalTarget)
      setBufferedRange({
        start: globalTarget,
        end: globalTarget,
      })
      setIsBuffering(true)
      setHlsStartOffset(nextOffset)
      showToastInfo(`Đang tua tới ${formatTime(globalTarget)}...`)

      return globalTarget
    },
    [clampSeekTarget, mediaId, setLastPosition, showToastInfo],
  )

  const applySeek = useCallback(
    (targetTime: number) => {
      const video = videoRef.current
      if (!video) return 0

      const clampedTime = clampSeekTarget(targetTime, video)
      const sessionOffset = streamSourceOffsetRef.current
      const localTarget = clampedTime - sessionOffset
      // Session reload (server-side seek) is only safe for TranscodeAudio (video copy)
      // which is lightweight and doesn't consume transcode slots. FullTranscode would
      // tie up a VAAPI/CPU slot for a second concurrent job — clamp seek instead.
      const canReloadSession =
        isHlsPlayback && playbackInfo?.method === 'TranscodeAudio' && Boolean(streamUrls?.hls)
      const isBeforeCurrentSession = isHlsPlayback && localTarget < 0
      const isBeyondReadyEdge =
        isHlsPlayback &&
        hlsPlaylistLiveRef.current &&
        hlsSeekableEndRef.current > 0 &&
        clampedTime > hlsSeekableEndRef.current + 1

      if (canReloadSession && (isBeforeCurrentSession || isBeyondReadyEdge)) {
        return requestHlsSessionReload(clampedTime)
      }

      // For FullTranscode: clamp seek to the ready edge of the current session
      if (isBeyondReadyEdge && hlsSeekableEndRef.current > 0) {
        const clampedLocal = hlsSeekableEndRef.current - sessionOffset
        video.currentTime = Math.max(0, clampedLocal)
        setCurrentTime(hlsSeekableEndRef.current)
        return hlsSeekableEndRef.current
      }

      video.currentTime = Math.max(0, localTarget)
      setCurrentTime(clampedTime)
      return clampedTime
    },
    [
      clampSeekTarget,
      isHlsPlayback,
      playbackInfo?.method,
      requestHlsSessionReload,
      streamUrls?.hls,
    ],
  )

  const seek = useCallback(
    (seconds: number) => {
      applySeek(currentTime + seconds)
      showSeekFeedback(seconds > 0 ? 'fwd' : 'back', Math.abs(seconds))
    },
    [applySeek, currentTime, showSeekFeedback],
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
    const isFs = !!(document.fullscreenElement || doc.webkitFullscreenElement)
    const container = containerRef.current as FullscreenEl | null
    const video = videoRef.current as
      | (HTMLVideoElement & {
          webkitEnterFullscreen?: () => void
          webkitExitFullscreen?: () => void
          webkitDisplayingFullscreen?: boolean
        })
      | null

    if (isFs || video?.webkitDisplayingFullscreen) {
      // Exit fullscreen
      screen.orientation?.unlock?.()
      if (video?.webkitDisplayingFullscreen) {
        video.webkitExitFullscreen?.()
      } else if (document.exitFullscreen) {
        document.exitFullscreen().catch(console.error)
      } else {
        doc.webkitExitFullscreen?.()
      }
    } else {
      // Enter fullscreen — use container element (keeps custom controls + subtitles visible)
      const lockLandscape = () => {
        const o = screen.orientation as ScreenOrientation & { lock?: (s: string) => Promise<void> }
        o?.lock?.('landscape').catch(() => {})
      }
      if (container?.requestFullscreen) {
        container
          .requestFullscreen()
          .then(lockLandscape)
          .catch((err: Error) => {
            // Fallback: iOS native video fullscreen (auto-rotates but loses custom UI)
            if (video?.webkitEnterFullscreen) {
              video.webkitEnterFullscreen()
            } else {
              showToastInfo(`Fullscreen: ${err.message}`)
              console.error('[fullscreen]', err)
            }
          })
      } else if (container?.webkitRequestFullscreen) {
        container.webkitRequestFullscreen()
        lockLandscape()
      } else if (video?.webkitEnterFullscreen) {
        video.webkitEnterFullscreen()
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
      const time = clampSeekTarget(getTimeFromClientX(e.clientX), videoRef.current)
      dragSeekTimeRef.current = time
      setDragSeekTime(time)
    },
    [clampSeekTarget, getTimeFromClientX],
  )

  const handleBarMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!progressBarRef.current) return
      const rect = progressBarRef.current.getBoundingClientRect()
      const clampedTime = clampSeekTarget(getTimeFromClientX(e.clientX), videoRef.current)
      setHoverX(e.clientX - rect.left)
      setHoverTime(clampedTime)
      if (isDraggingBar) {
        dragSeekTimeRef.current = clampedTime
        setDragSeekTime(clampedTime)
      }
    },
    [clampSeekTarget, getTimeFromClientX, isDraggingBar],
  )

  useEffect(() => {
    if (!isDraggingBar) return
    const onMove = (e: MouseEvent) => {
      const time = clampSeekTarget(getTimeFromClientX(e.clientX), videoRef.current)
      dragSeekTimeRef.current = time
      setDragSeekTime(time)
    }
    const onUp = () => {
      const targetTime = dragSeekTimeRef.current
      if (targetTime != null) {
        applySeek(targetTime)
      }
      dragSeekTimeRef.current = null
      setDragSeekTime(null)
      setIsDraggingBar(false)
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [applySeek, clampSeekTarget, getTimeFromClientX, isDraggingBar])

  // ── HLS init ───────────────────────────────────────────────────────────────
  useEffect(() => {
    const video = videoRef.current
    if (!video || !streamUrls) return
    if (hlsRef.current) {
      hlsRef.current.destroy()
      hlsRef.current = null
    }
    hlsPlaylistLiveRef.current = false
    const sessionOffset = hlsStartOffset ?? 0
    streamSourceOffsetRef.current = sessionOffset
    hlsSeekableEndRef.current = sessionOffset
    setBufferedRange({
      start: sessionOffset,
      end: sessionOffset,
    })

    const useHls = isHlsPlayback
    const rawUrl = useHls
      ? sessionOffset > 0.25 && streamUrls.hls
        ? buildHlsSessionUrl(streamUrls.hls, sessionOffset)
        : streamUrls.abr || streamUrls.hls
      : streamUrls.direct
    if (!rawUrl) return
    setIsBuffering(true)
    const streamUrl = accessToken
      ? rawUrl + (rawUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(accessToken)
      : rawUrl

    if (useHls && (streamUrls.abr || streamUrls.hls) && Hls.isSupported()) {
      const hls = new Hls({
        maxBufferLength: 30,
        maxMaxBufferLength: 600,
        enableWorker: true,
        xhrSetup: (xhr) => {
          // Read fresh token on every request — prevents 401 when token refreshes during playback
          const freshToken = useAuthStore.getState().accessToken
          if (freshToken) xhr.setRequestHeader('Authorization', `Bearer ${freshToken}`)
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
        // Set duration from playback info immediately (true duration from ffprobe),
        // so the player shows correct total even while transcoding is in progress.
        const knownDur = playbackInfo?.duration ?? 0
        const v = videoRef.current
        const videoDur = v?.duration && isFinite(v.duration) ? v.duration : 0
        const initDur = Math.max(knownDur, sessionOffset + videoDur)
        if (initDur > 0) setDuration(initDur)
        // Resume position (initial load or after subtitle/quality change)
        const seekToGlobal = usePlayerStore.getState().lastPositions[mediaId] ?? 0
        if (seekToGlobal > 0 && v) {
          const maxSeekableGlobal =
            hlsPlaylistLiveRef.current && hlsSeekableEndRef.current > 0
              ? hlsSeekableEndRef.current
              : Math.max(knownDur, sessionOffset + videoDur)
          const localMaxSeekable = Math.max(0, maxSeekableGlobal - sessionOffset)
          const localSeekTo = Math.max(0, seekToGlobal - sessionOffset)
          const safeSeekTo =
            localMaxSeekable > 0.25
              ? Math.min(localSeekTo, localMaxSeekable - 0.25)
              : Math.min(localSeekTo, localMaxSeekable)
          if (safeSeekTo > 0) {
            v.currentTime = safeSeekTo
          }
        }
        // Auto-play when ready
        v?.play().catch(() => {})
      })
      hls.on(Hls.Events.LEVEL_LOADED, (_e, data) => {
        const playlistDur = data.details?.totalduration ?? 0
        const playlistEdge = data.details?.edge ?? playlistDur
        hlsPlaylistLiveRef.current = data.details?.live ?? false
        hlsSeekableEndRef.current = sessionOffset + playlistEdge
        // While transcoding in background, the playlist only contains segments
        // encoded so far. Use playback info duration (from ffprobe) if it's larger
        // so the player always shows the correct total duration.
        const knownDur = playbackInfo?.duration ?? 0
        const trueDur = Math.max(knownDur, sessionOffset + playlistDur)
        if (trueDur > 0) setDuration(trueDur)
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
      hlsPlaylistLiveRef.current = false
      hlsSeekableEndRef.current = 0
      streamSourceOffsetRef.current = 0
      setBufferedRange({ start: 0, end: 0 })
      video.src = streamUrl
      // Resume position (initial load or after subtitle/quality change)
      const seekToGlobal = usePlayerStore.getState().lastPositions[mediaId] ?? 0
      if (seekToGlobal > 0) {
        const nativeDuration =
          Number.isFinite(video.duration) && video.duration > 0 ? video.duration : 0
        const effectiveDuration = Math.max(
          nativeDuration,
          knownDurationRef.current,
          playbackInfo?.duration ?? 0,
        )
        const seekCeiling = effectiveDuration > 0.25 ? effectiveDuration - 0.25 : effectiveDuration
        video.currentTime = Math.max(0, Math.min(seekCeiling, seekToGlobal))
      }
      video.play().catch(() => {})
    }
    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy()
        hlsRef.current = null
      }
      hlsPlaylistLiveRef.current = false
      hlsSeekableEndRef.current = 0
      streamSourceOffsetRef.current = 0
    }
    // Note: accessToken intentionally excluded — token refresh must NOT restart video.
    // HLS uses useAuthStore.getState() for fresh tokens per-request.
    // playbackInfo.duration and clampSeekTarget intentionally excluded so
    // progressive duration updates do not recreate the HLS instance mid-playback.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [audioLanguage, buildHlsSessionUrl, hlsStartOffset, isHlsPlayback, streamUrls])

  // Resume position is read from usePlayerStore.getState().lastPositions[mediaId]
  // directly in the HLS init effect — no cross-effect refs needed.

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    video.volume = volume
    video.muted = isMuted
    video.playbackRate = playbackRate
  }, [volume, isMuted, playbackRate, streamUrls]) // streamUrls ensures re-sync after video remount

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
      const globalTime = streamSourceOffsetRef.current + video.currentTime
      const effectiveDuration = getEffectiveDuration(video)
      setCurrentTime(globalTime)
      if (video.duration && !isNaN(video.duration) && isFinite(video.duration)) {
        setDuration((prev) =>
          Math.max(prev, streamSourceOffsetRef.current + video.duration, knownDurationRef.current),
        )
      }
      setLastPosition(mediaId, globalTime)
      const now = Date.now()
      if (now - lastProgressUpdate.current >= 10000 || globalTime >= effectiveDuration * 0.95) {
        updateProgress({
          mediaId,
          data: {
            position: globalTime,
            completed: effectiveDuration > 0 ? globalTime / effectiveDuration > 0.9 : false,
          },
        })
        lastProgressUpdate.current = now
      }
    }
    const onProgress = () => {
      if (video.buffered.length > 0) {
        const bufferIndex = video.buffered.length - 1
        const offset = streamSourceOffsetRef.current
        setBufferedRange({
          start: offset + video.buffered.start(bufferIndex),
          end: offset + video.buffered.end(bufferIndex),
        })
      }
    }
    const onWaiting = () => setIsBuffering(true)
    const onPlaying = () => setIsBuffering(false)
    const onCanPlay = () => setIsBuffering(false)
    const onDurationChange = () => {
      if (video.duration && !isNaN(video.duration) && isFinite(video.duration)) {
        setDuration((prev) =>
          Math.max(prev, streamSourceOffsetRef.current + video.duration, knownDurationRef.current),
        )
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
  }, [getEffectiveDuration, mediaId, setLastPosition, updateProgress, streamUrls])

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
          applySeek(0)
          break
        case 'Escape':
          if (isFullscreen) {
            e.preventDefault()
            screen.orientation?.unlock?.()
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
    applySeek,
    seek,
    changeVolume,
    toggleFullscreen,
    toggleMute,
    resetControlsTimeout,
  ])

  useEffect(() => {
    const onChange = () => {
      const nowFs = !!(
        document.fullscreenElement ??
        (document as Document & { webkitFullscreenElement?: Element }).webkitFullscreenElement
      )
      setIsFullscreen(nowFs)
      if (!nowFs) screen.orientation?.unlock?.()
    }
    document.addEventListener('fullscreenchange', onChange)
    document.addEventListener('webkitfullscreenchange', onChange)

    // Track iOS native video fullscreen (webkitEnterFullscreen fallback)
    const video = videoRef.current
    const onBeginFs = () => setIsFullscreen(true)
    const onEndFs = () => {
      setIsFullscreen(false)
      screen.orientation?.unlock?.()
    }
    video?.addEventListener('webkitbeginfullscreen', onBeginFs)
    video?.addEventListener('webkitendfullscreen', onEndFs)

    return () => {
      document.removeEventListener('fullscreenchange', onChange)
      document.removeEventListener('webkitfullscreenchange', onChange)
      video?.removeEventListener('webkitbeginfullscreen', onBeginFs)
      video?.removeEventListener('webkitendfullscreen', onEndFs)
    }
  }, [])

  const qualityOptions: QualityOption[] = playbackInfo?.available_qualities ?? []
  const currentQualityLabel =
    maxQuality === 'auto'
      ? t('controls.auto')
      : (qualityOptions.find((q) => q.height === maxQuality)?.label ?? `${maxQuality}p`)

  const getActiveSubtitleTrack = (): PlaybackSubtitleTrack | null => {
    if (!effectiveSubtitleLanguage) return null
    return (
      (subtitleTrackId ? subtitles.find((s) => s.id === subtitleTrackId) : undefined) ??
      subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage)) ??
      null
    )
  }

  const displayTime = dragSeekTime ?? currentTime
  const progressPercent = duration ? (displayTime / duration) * 100 : 0
  const bufferWidthPercent = duration ? Math.max(0, (bufferedRange.end / duration) * 100) : 0
  const remainingTime = duration > 0 ? duration - displayTime : 0

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
          <button onClick={handleBack} className="mt-4 text-sm text-white/60 hover:text-white">
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
      className={`fixed inset-0 bg-[#141414] select-none overflow-hidden ${
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
        className="h-full w-full"
        style={{ objectFit: aspectRatio as 'contain' | 'cover' | 'fill' }}
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
      >
        {/* Native <track> for iOS native fullscreen (webkitEnterFullscreen fallback) */}
        {primarySub && !primarySub.is_image && subtitleServeUrl(primarySub) && (
          <track
            kind="subtitles"
            src={subtitleServeUrl(primarySub)!}
            srcLang={primarySub.language ?? 'und'}
            label={primarySub.label ?? primarySub.language ?? 'Subtitles'}
            default
          />
        )}
      </video>

      {/* Subtitle overlay */}
      <DualSubtitleOverlay
        videoRef={videoRef}
        primaryUrl={subtitleServeUrl(primarySub)}
        secondaryUrl={subtitleServeUrl(secondarySub)}
        currentTime={currentTime}
        offsetSeconds={subtitleOffsetSeconds}
        primaryRenderedInVideo={Boolean(burnedInPrimarySub)}
        style={{ size: subtitleSize, color: subtitleColor, background: subtitleBackground }}
      />

      {/* Buffering spinner */}
      {isBuffering && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <div className="-translate-y-20">
            <div className="h-12 w-12 animate-spin rounded-full border-2 border-white/20 border-t-white" />
          </div>
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
          {availableLevels.find((l) => l.level === currentLevel)?.height
            ? `${availableLevels.find((l) => l.level === currentLevel)?.height}p`
            : t('controls.auto')}
          {bandwidth !== null && ` · ${bandwidth.toFixed(1)} Mbps`}
        </div>
      )}

      {showStats &&
        (playbackInfo ? (
          <div onClick={(e) => e.stopPropagation()}>
            <WatchPlaybackStatsOverlay
              onClose={() => setShowStats(false)}
              playbackInfo={playbackInfo}
              videoRef={videoRef}
            />
          </div>
        ) : (
          <div
            className="absolute left-4 top-20 z-30 w-80 overflow-hidden rounded-xl bg-black/70 px-4 py-4 backdrop-blur-md ring-1 ring-white/10"
            onClick={(e) => e.stopPropagation()}
          >
            <p className="text-xs text-white/40">Loading stream info…</p>
          </div>
        ))}

      {/* Up Next card — z-50 so it stays above lock overlay */}
      {showUpNext && (
        <div
          className="absolute bottom-56 right-6 z-50 w-64 rounded-xl bg-[#1e1e1e] p-4 shadow-2xl ring-1 ring-white/10"
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

      {/* Skip Intro/Credits CTA */}
      <SkipIntroCredits
        segments={playbackInfo?.skip_segments}
        currentTime={currentTime}
        onSkip={(toTime) => {
          applySeek(toTime)
        }}
        visible
        hideCredits={isEpisode && nextEpisodeMediaId != null}
      />

      {/* Screen lock overlay — tap to show unlock icon, auto-hides after 3s */}
      {isLocked && (
        <div
          className="absolute inset-0 z-40"
          onClick={(e) => {
            e.stopPropagation()
            e.preventDefault()
            if (lockedUITimeoutRef.current) clearTimeout(lockedUITimeoutRef.current)
            setShowLockedUI(true)
            lockedUITimeoutRef.current = setTimeout(() => setShowLockedUI(false), 3000)
          }}
        >
          <button
            onClick={(e) => {
              e.stopPropagation()
              setIsLocked(false)
              setShowLockedUI(false)
              if (lockedUITimeoutRef.current) clearTimeout(lockedUITimeoutRef.current)
              resetControlsTimeout()
            }}
            className={`absolute left-6 top-1/2 -translate-y-1/2 rounded-full bg-black/50 p-3 text-white/80 backdrop-blur-sm transition-opacity duration-300 hover:text-white ${showLockedUI ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
          >
            <LuLockOpen size={22} />
          </button>
        </div>
      )}

      {/* ── Controls overlay ─────────────────────────────────────────────────── */}
      <div
        className={`absolute inset-0 flex flex-col justify-between transition-opacity duration-300 ${
          isLocked
            ? 'opacity-0 pointer-events-none'
            : isDetailPanelActive || showControls || isHoveringBar || isDraggingBar
              ? 'opacity-100'
              : 'opacity-0 pointer-events-none'
        }`}
        onClick={togglePlay}
      >
        <div onClick={(e) => e.stopPropagation()}>
          <WatchTopBar
            isMuted={isMuted}
            volume={volume}
            onBack={handleBack}
            onMuteToggle={toggleMute}
            onVolumeChange={(nextVolume) => {
              setVolume(nextVolume)
              if (videoRef.current) {
                videoRef.current.volume = nextVolume
                videoRef.current.muted = false
              }
            }}
            castAvailable={castAvailable}
            castConnected={castConnected}
            casting={casting}
            onCastClick={() => {
              if (casting) {
                stopCasting()
              } else if (castConnected) {
                castMedia(
                  mediaId,
                  media?.media.title ?? '',
                  media?.media.poster_path ? tmdbImage(media.media.poster_path, 'w342') : undefined,
                  currentTime,
                )
              } else {
                requestCast()
              }
            }}
          />
        </div>

        {/* ── Center: pause indicator (brief) ──────────────────────────────── */}
        {!isPlaying && !isBuffering && !isDetailPanelActive && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="rounded-full bg-black/30 p-5 backdrop-blur-sm">
              <LuPlay size={44} className="text-white fill-white ml-1" />
            </div>
          </div>
        )}

        {/* ── Bottom panel ─────────────────────────────────────────────────── */}
        <div
          className="relative"
          style={{
            background:
              'linear-gradient(to top, rgba(0,0,0,0.92) 0%, rgba(0,0,0,0.7) 70%, transparent 100%)',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <WatchDetailSheet
            activeTab={activeTab}
            displayPanel={displayPanel}
            isEpisode={isEpisode}
            isPanelVisible={isPanelVisible}
            infoAudioChannels={infoAudioChannels}
            infoAudioCodec={infoAudioCodec}
            infoAudioLanguage={infoAudioLanguage}
            infoCC={infoCC}
            infoEpisodeLabel={infoEpisodeLabel}
            infoLogoUrl={infoLogoUrl}
            infoResolution={infoResolution}
            infoRuntime={infoRuntime}
            infoVideoCodec={infoVideoCodec}
            infoYear={infoYear}
            media={media}
            mediaId={mediaId}
            onEpisodeSelect={(episodeMediaId, isCurrentEpisode) => {
              if (isCurrentEpisode) {
                setActiveTab('none')
                return
              }
              navigate(`/watch/${episodeMediaId}`)
            }}
            onScrollSeasonCarousel={scrollSeasonCarousel}
            onSeasonSelect={setSeasonPanelSeasonId}
            onToggleDetailPanel={toggleDetailPanel}
            seasonCarouselRef={seasonCarouselRef}
            seasonPanelEpisodes={seasonPanelEpisodes}
            seasonPanelSeasonId={seasonPanelSeasonId}
            seasons={seasons}
          />

          <div
            className={`px-3 sm:px-6 transition-[opacity,transform] duration-[380ms] ease-[cubic-bezier(0.22,1,0.36,1)] ${
              displayPanel === 'none'
                ? 'translate-y-0 pb-3 pt-2 opacity-100 sm:pb-4 sm:pt-3'
                : 'pointer-events-none translate-y-5 pb-3 pt-2 opacity-0 sm:pb-4 sm:pt-3'
            }`}
          >
            <div className="space-y-1.5 sm:space-y-2">
              {/* Row 1: Title + icon buttons */}
              <div className="flex items-center justify-between gap-2 sm:gap-4">
                <h1 className="min-w-0 truncate text-sm font-bold text-white leading-tight drop-shadow sm:text-xl">
                  {media.media.title}
                </h1>

                {/* Right icon buttons */}
                <div className="flex shrink-0 items-center gap-1 sm:gap-1.5">
                  {/* Subtitles — always visible so users can search for subs */}
                  <div className="relative">
                    <button
                      onClick={() => {
                        setShowSubtitleMenu(!showSubtitleMenu)
                        setShowAudioMenu(false)
                        setShowSpeedMenu(false)
                        setShowSettings(false)
                      }}
                      className={`flex h-9 w-9 items-center justify-center rounded-lg border transition-colors sm:h-10 sm:w-10 ${
                        getActiveSubtitleTrack()
                          ? 'border-white bg-white/20 text-white'
                          : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                      }`}
                      title={t('controls.subtitles')}
                    >
                      <LuCaptions size={18} />
                    </button>
                    {showSubtitleMenu && (
                      <div className="fixed inset-x-3 bottom-44 z-50 sm:absolute sm:inset-auto sm:bottom-full sm:right-0 sm:z-auto sm:mb-2">
                        <SubtitlePicker
                          subtitles={subtitles}
                          primaryLanguage={effectiveSubtitleLanguage}
                          primaryTrackId={subtitleTrackId}
                          secondaryLanguage={secondarySubtitleLanguage}
                          secondaryTrackId={secondarySubtitleTrackId}
                          onSelectPrimary={(lang, trackId) => {
                            setSubtitleLanguage(lang, trackId ?? null)
                          }}
                          onSelectPrimarySource={(trackId) => setSubtitleTrackId(trackId)}
                          onSelectSecondary={(lang, trackId) =>
                            setSecondarySubtitleLanguage(lang, trackId ?? null)
                          }
                          onSelectSecondarySource={(trackId) =>
                            setSecondarySubtitleTrackId(trackId)
                          }
                          dualMode={true}
                          allowImageSubtitles={allowsImageSubtitles}
                          mediaId={mediaId}
                          onSubtitleAdded={() => {
                            queryClient.refetchQueries({ queryKey: streamingKeys.all })
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
                        className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
                        title={t('controls.audio')}
                      >
                        <LuMusic size={17} />
                      </button>
                      {showAudioMenu && (
                        <div className="fixed inset-x-3 bottom-44 z-50 sm:absolute sm:inset-auto sm:bottom-full sm:right-0 sm:z-auto sm:mb-2">
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
                      className={`flex h-9 min-w-[36px] items-center justify-center rounded-lg border px-1.5 transition-colors sm:h-10 sm:min-w-[40px] sm:px-2 ${
                        playbackRate !== 1
                          ? 'border-white bg-white/20 text-white'
                          : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                      }`}
                      title={t('controls.speed')}
                    >
                      <span className="text-xs font-bold tabular-nums">
                        {playbackRate === 1 ? '1×' : `${playbackRate}×`}
                      </span>
                    </button>
                    {showSpeedMenu && (
                      <div className="fixed inset-x-3 bottom-44 z-50 rounded-xl bg-[#1e1e1e] py-2 shadow-2xl ring-1 ring-white/10 sm:absolute sm:inset-auto sm:bottom-full sm:right-0 sm:z-auto sm:mb-2 sm:w-44">
                        <p className="px-4 pb-1.5 pt-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                          {t('controls.playbackSpeed')}
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
                      className={`flex h-9 w-9 items-center justify-center rounded-lg border transition-colors sm:h-10 sm:w-10 ${
                        showSettings
                          ? 'border-white bg-white/20 text-white'
                          : 'border-white/30 bg-white/5 text-white/70 hover:border-white/60 hover:text-white'
                      }`}
                      title={t('controls.settings')}
                    >
                      <LuSettings size={17} />
                    </button>
                    {showSettings && (
                      <div className="fixed inset-x-3 bottom-44 z-50 overflow-hidden rounded-xl bg-[#1e1e1e] shadow-2xl ring-1 ring-white/10 sm:absolute sm:inset-auto sm:bottom-full sm:right-0 sm:z-auto sm:mb-2 sm:w-56">
                        {settingsView === 'quality' ? (
                          /* Quality submenu — resolution-based (Netflix style) */
                          <div className="flex flex-col">
                            <button
                              onClick={() => setSettingsView('main')}
                              className="flex items-center gap-2 border-b border-white/10 px-4 py-2.5 text-xs font-semibold text-white/70 hover:text-white"
                            >
                              <LuChevronLeft size={14} />
                              {t('controls.quality')}
                            </button>
                            <div className="max-h-[50vh] overflow-y-auto py-1">
                              {qualityOptions.map((q) => {
                                const isSelected = maxQuality !== 'auto' && maxQuality === q.height
                                return (
                                  <button
                                    key={`${q.source}-${q.height}`}
                                    onClick={() => {
                                      setMaxQuality(q.height)
                                      setSettingsView('main')
                                    }}
                                    className={`flex w-full items-center justify-between px-4 py-2 text-xs ${
                                      isSelected
                                        ? 'bg-white/10 text-white'
                                        : 'text-white/70 hover:bg-white/5 hover:text-white'
                                    }`}
                                  >
                                    <span className="flex items-center gap-1.5">
                                      {q.label}
                                      {q.instant && q.source !== 'original' && (
                                        <LuZap size={11} className="text-yellow-400" />
                                      )}
                                    </span>
                                    {isSelected && <LuCheck size={14} className="text-white" />}
                                  </button>
                                )
                              })}
                              {/* Auto option — always at bottom */}
                              <button
                                onClick={() => {
                                  setMaxQuality('auto')
                                  setSettingsView('main')
                                }}
                                className={`flex w-full items-center justify-between border-t border-white/10 px-4 py-2 text-xs ${
                                  maxQuality === 'auto'
                                    ? 'bg-white/10 text-white'
                                    : 'text-white/70 hover:bg-white/5 hover:text-white'
                                }`}
                              >
                                {t('controls.auto')}
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
                                <LuExpand size={10} /> {t('controls.aspectRatio')}
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
                                      ? t('controls.auto')
                                      : r.charAt(0).toUpperCase() + r.slice(1)}
                                  </button>
                                ))}
                              </div>
                            </div>

                            {/* Subtitle Appearance */}
                            <div className="border-t border-white/10 pt-3">
                              <p className="mb-1.5 flex items-center gap-1.5 px-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                                <LuCaptions size={10} /> {t('controls.subtitles')}
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
                                  { value: 'none' as const, labelKey: 'controls.none' },
                                  { value: 'semi' as const, labelKey: 'controls.semi' },
                                  { value: 'solid' as const, labelKey: 'controls.solid' },
                                ].map(({ value, labelKey }) => (
                                  <button
                                    key={value}
                                    onClick={() => setSubtitleBackground(value)}
                                    className={`rounded-lg py-1.5 text-xs font-medium ${
                                      subtitleBackground === value
                                        ? 'bg-white/20 text-white'
                                        : 'bg-white/5 text-white/70 hover:bg-white/15 hover:text-white'
                                    }`}
                                  >
                                    {t(labelKey)}
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
                              <div className="mt-3 rounded-lg bg-white/[0.04] px-3 py-2">
                                <div className="mb-2 flex items-center justify-between text-[11px] text-white/65">
                                  <span>Delay</span>
                                  <span className="font-semibold tabular-nums text-white/85">
                                    {subtitleOffsetSeconds > 0 ? '+' : ''}
                                    {subtitleOffsetSeconds.toFixed(2)}s
                                  </span>
                                </div>
                                <div className="grid grid-cols-3 gap-1">
                                  <button
                                    onClick={() =>
                                      setSubtitleOffset(
                                        mediaId,
                                        Number((subtitleOffsetSeconds - 0.25).toFixed(2)),
                                      )
                                    }
                                    className="rounded-lg bg-white/5 py-1.5 text-xs font-medium text-white/70 transition-colors hover:bg-white/15 hover:text-white"
                                  >
                                    -0.25s
                                  </button>
                                  <button
                                    onClick={() => resetSubtitleOffset(mediaId)}
                                    className="rounded-lg bg-white/5 py-1.5 text-xs font-medium text-white/70 transition-colors hover:bg-white/15 hover:text-white"
                                  >
                                    Reset
                                  </button>
                                  <button
                                    onClick={() =>
                                      setSubtitleOffset(
                                        mediaId,
                                        Number((subtitleOffsetSeconds + 0.25).toFixed(2)),
                                      )
                                    }
                                    className="rounded-lg bg-white/5 py-1.5 text-xs font-medium text-white/70 transition-colors hover:bg-white/15 hover:text-white"
                                  >
                                    +0.25s
                                  </button>
                                </div>
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
                                onClick={handleBack}
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
                      className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
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
                    className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
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
                    className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/30 bg-white/5 text-white/70 transition-colors hover:border-white/60 hover:text-white sm:h-10 sm:w-10"
                    title={isFullscreen ? t('controls.exitFullscreen') : t('controls.fullscreen')}
                  >
                    {isFullscreen ? <LuMinimize2 size={17} /> : <LuMaximize2 size={17} />}
                  </button>
                </div>
              </div>

              {/* Row 2: Progress bar */}
              <div
                ref={progressBarRef}
                className="group relative flex h-8 sm:h-5 cursor-pointer items-center"
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
                    style={{ width: `${bufferWidthPercent}%` }}
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
                <div className="flex items-center gap-2 sm:gap-4">
                  <span className="text-xs tabular-nums text-white/70 sm:text-sm">
                    {formatTime(displayTime)}
                  </span>

                  <div className="flex items-center gap-2 sm:gap-3">
                    <button
                      onClick={() => seek(-SEEK_STEP)}
                      className="flex items-center gap-0.5 p-2 text-white/75 transition-colors hover:text-white"
                      title="Rewind 5s"
                    >
                      <LuRotateCcw className="size-5 sm:size-[22px]" />
                      <span className="text-[11px] font-bold">5</span>
                    </button>

                    <button
                      onClick={togglePlay}
                      className="flex h-10 w-10 sm:h-11 sm:w-11 items-center justify-center text-white transition-transform hover:scale-105"
                    >
                      {isPlaying ? (
                        <LuPause className="fill-white size-6 sm:size-[30px]" />
                      ) : (
                        <LuPlay className="fill-white ml-0.5 size-6 sm:size-[30px]" />
                      )}
                    </button>

                    <button
                      onClick={() => seek(SEEK_STEP)}
                      className="flex items-center gap-0.5 p-2 text-white/75 transition-colors hover:text-white"
                      title="Forward 5s"
                    >
                      <span className="text-[11px] font-bold">5</span>
                      <LuRotateCw className="size-5 sm:size-[22px]" />
                    </button>
                  </div>
                </div>

                {/* Right: remaining */}
                <span className="text-xs tabular-nums text-white/50 sm:text-sm">
                  {duration > 0 ? `-${formatTime(remainingTime)}` : wallClock}
                </span>
              </div>

              {/* Row 4: Tabs */}
              <div className="flex items-center gap-5 pt-1">
                <button
                  onClick={() => toggleDetailPanel('info')}
                  className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
                    activeTab === 'info' ? 'text-white' : 'text-white/45 hover:text-white/80'
                  }`}
                >
                  <LuInfo size={14} />
                  {t('detail.info')}
                </button>
                {isEpisode && (
                  <button
                    onClick={() => toggleDetailPanel('season')}
                    className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
                      activeTab === 'season' ? 'text-white' : 'text-white/45 hover:text-white/80'
                    }`}
                  >
                    <LuListMusic size={12} />
                    {t('detail.season')}
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
