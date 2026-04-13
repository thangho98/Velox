import { useParams, useNavigate } from 'react-router'
import { useEffect, useEffectEvent, useRef, useState, useCallback, useMemo } from 'react'
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
  useProgress,
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
import { api } from '@velox/shared/api'
import { useToast } from '@/components/Toast'
import { createLogger } from '@/lib/logger'

const log = createLogger('WatchPage')
import { WatchDetailSheet } from '@/components/watch/WatchDetailSheet'
import { WatchPlaybackStatsOverlay } from '@/components/watch/WatchPlaybackStatsOverlay'
import { WatchTopBar } from '@/components/watch/WatchTopBar'
import { SkipIntroCredits } from '@/components/watch/SkipIntroCredits'
import { useChromecast } from '@/hooks/useChromecast'
import { useFullscreen } from '@/hooks/useFullscreen'
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts'
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

export default function WatchPage() {
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

  const [forceFullTranscode, setForceFullTranscode] = useState(false)

  const buildHlsSessionUrl = useCallback(
    (baseUrl: string, startOffset: number, forceTranscode: boolean) => {
      const url = new URL(baseUrl, window.location.origin)
      url.searchParams.set('start', startOffset.toFixed(3))
      if (forceTranscode) {
        url.searchParams.delete('vcopy')
      }
      return `${url.pathname}${url.search}${url.hash}`
    },
    [],
  )

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

  const playbackRequest = useMemo(
    () => ({
      video_codecs: clientCaps.videoCodecs,
      audio_codecs: clientCaps.audioCodecs,
      containers: clientCaps.containers,
      max_height: qualityMaxHeight,
      selected_subtitle: effectiveSubtitleLanguage ?? 'off',
      selected_subtitle_id: subtitleTrackId ?? 0,
      selected_audio_track: audioTrackId ?? 0,
    }),
    [
      clientCaps.videoCodecs,
      clientCaps.audioCodecs,
      clientCaps.containers,
      qualityMaxHeight,
      effectiveSubtitleLanguage,
      subtitleTrackId,
      audioTrackId,
    ],
  )

  const { data: subtitles = [] } = useSubtitles(mediaId, playbackRequest)
  const { data: audioTracks = [] } = useAudioTracks(mediaId, playbackRequest)
  const { data: playbackInfo } = usePlaybackInfo(mediaId, playbackRequest)

  // Stream URLs use a stable key that excludes TEXT subtitle selection so switching
  // text subtitles (SRT/ASS) does not reload the video. IMAGE subtitles
  // (PGS/VobSub) require server-side burn-in, so their ID IS included in the key
  // to trigger a fresh stream URL fetch with the correct ?si= parameter.
  const burnInSubtitleId =
    (subtitleTrackId
      ? subtitles.find((s) => s.id === subtitleTrackId && s.is_image)?.id
      : undefined) ??
    (effectiveSubtitleLanguage
      ? subtitles.find((s) => languageMatches(s.language, effectiveSubtitleLanguage) && s.is_image)
          ?.id
      : undefined) ??
    0
  const streamRequestKey = useMemo(
    () => ({
      video_codecs: clientCaps.videoCodecs,
      audio_codecs: clientCaps.audioCodecs,
      containers: clientCaps.containers,
      max_height: qualityMaxHeight,
      selected_audio_track: audioTrackId ?? 0,
      selected_subtitle_id: burnInSubtitleId,
    }),
    [
      clientCaps.videoCodecs,
      clientCaps.audioCodecs,
      clientCaps.containers,
      qualityMaxHeight,
      audioTrackId,
      burnInSubtitleId,
    ],
  )
  const { data: streamUrls, isLoading: streamLoading } = useStreamUrls(mediaId, streamRequestKey)

  // 3-tier fallback chain owned by the client:
  //   direct → pretranscode → hls
  // Backend hints the starting tier via streamUrls.prefer; client advances the
  // chain on playback failure (network error, audio decode failure, etc.).
  type PlaybackSource = 'direct' | 'pretranscode' | 'hls'
  const [playbackSource, setPlaybackSource] = useState<PlaybackSource | null>(null)

  // Initialize source from backend hint as soon as URLs are available.
  // Skip a tier when its URL is missing (e.g. no pretranscode → direct → hls).
  useEffect(() => {
    if (!streamUrls) return
    if (playbackSource !== null) return
    const pickFirstAvailable = (): PlaybackSource => {
      const order: PlaybackSource[] =
        streamUrls.prefer === 'pretranscode'
          ? ['pretranscode', 'direct', 'hls']
          : streamUrls.prefer === 'hls'
            ? ['hls', 'pretranscode', 'direct']
            : ['direct', 'pretranscode', 'hls']
      for (const tier of order) {
        if (tier === 'direct' && streamUrls.direct) return tier
        if (tier === 'pretranscode' && streamUrls.pretranscode) return tier
        if (tier === 'hls' && streamUrls.hls) return tier
      }
      return 'hls'
    }
    const picked = pickFirstAvailable()
    log.info(`[Playback Source] Init: preferred=${streamUrls.prefer}, chosen=${picked}`)
    setPlaybackSource(picked)
  }, [streamUrls, playbackSource])

  // advanceFallback moves to the next tier when current source fails.
  const advanceFallback = useCallback(
    (reason: string) => {
      log.warn(`[Fallback Triggered] Reason: ${reason}`)
      setPlaybackSource((current) => {
        if (!streamUrls) return current
        const chain: PlaybackSource[] = ['direct', 'pretranscode', 'hls']
        const startIdx = current ? chain.indexOf(current) + 1 : 0
        for (let i = startIdx; i < chain.length; i++) {
          const tier = chain[i]
          if (tier === 'direct' && streamUrls.direct) {
            log.warn(`fallback → direct (${reason})`)
            return tier
          }
          if (tier === 'pretranscode' && streamUrls.pretranscode) {
            log.warn(`fallback → pretranscode (${reason})`)
            return tier
          }
          if (tier === 'hls' && streamUrls.hls) {
            log.warn(`fallback → hls (${reason})`)
            return tier
          }
        }
        log.warn(`fallback exhausted (${reason})`)
        return current
      })
    },
    [streamUrls],
  )

  const isHlsPlayback = playbackSource === 'hls'

  const isEpisode = media?.media.media_type === 'episode'
  const seriesId = media?.series_id ?? 0
  const seasonId = media?.season_id ?? 0

  // Handle back navigation - go to detail page instead of history back
  const handleBack = () => {
    if (isEpisode && seriesId > 0) {
      navigate(`/series/${seriesId}`)
    } else {
      navigate(`/movies/${mediaId}`)
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

  const nextNextEpisode = (() => {
    if (!isEpisode || seasonEpisodes.length === 0) return null
    const currentIdx = seasonEpisodes.findIndex((ep) => ep.media_id === mediaId)
    if (currentIdx === -1 || currentIdx >= seasonEpisodes.length - 2) return null
    return seasonEpisodes[currentIdx + 2]
  })()
  const nextNextEpisodeMediaId = nextNextEpisode?.media_id

  const { data: nextEpisodeProgress } = useProgress(nextEpisodeMediaId ?? 0)
  const [showWatchedWarning, setShowWatchedWarning] = useState(false)

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
  const { isFullscreen, toggleFullscreen } = useFullscreen(containerRef, videoRef, showToastInfo)
  const [error, setError] = useState<string | null>(null)
  const [isBuffering, setIsBuffering] = useState(true)
  const [availableLevels, setAvailableLevels] = useState<
    { level: number; height: number; bitrate: number }[]
  >([])
  const [currentLevel, setCurrentLevel] = useState(-1)
  const [bandwidth, setBandwidth] = useState<number | null>(null)
  const [showQualityIndicator, setShowQualityIndicator] = useState(false)
  // Image subtitles (PGS/VobSub) require server-side burn-in via HLS transcode.
  // Only available when actually using HLS transcode — not direct play or pretranscode MP4.
  const allowsImageSubtitles =
    playbackInfo?.method === 'FullTranscode' || playbackInfo?.method === 'TranscodeAudio'

  // When an image subtitle (PGS/VobSub) is selected or deselected while HLS is active,
  // reload the HLS session at the current position so the server burns it in.
  const burnInSubId = burnedInPrimarySub?.id ?? 0
  const prevBurnInSubRef = useRef(burnInSubId)
  useEffect(() => {
    if (prevBurnInSubRef.current === burnInSubId) return
    prevBurnInSubRef.current = burnInSubId
    if (!isHlsPlayback || !allowsImageSubtitles) return
    const pos = videoRef.current?.currentTime ?? 0
    const globalPos = streamSourceOffsetRef.current + pos
    setLastPosition(mediaId, globalPos)
    // Trigger HLS session reload at current position
    setHlsStartOffset(globalPos > 0.25 ? globalPos : null)
  }, [burnInSubId, isHlsPlayback, allowsImageSubtitles, mediaId, setLastPosition])

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
    setPlaybackSource(null)
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

  // Fallback trigger: detect silent audio decode failure during direct play.
  // Chrome does NOT fire <video onError> when only the audio track fails to decode
  // (e.g. AC3/EAC3/DTS unsupported codecs) — video continues playing without sound.
  // We poll webkitAudioDecodedByteCount; if zero bytes were decoded after a few
  // seconds of playback, fall back to HLS where backend transcodes audio to AAC.
  useEffect(() => {
    // Only run while we're on a non-HLS tier (direct or pretranscode).
    // HLS tier has audio guaranteed AAC by backend transcode → no need to check.
    if (playbackSource !== 'direct' && playbackSource !== 'pretranscode') return
    const v = videoRef.current as
      | (HTMLVideoElement & { webkitAudioDecodedByteCount?: number })
      | null
    if (!v) return
    if (v.webkitAudioDecodedByteCount === undefined) {
      log.debug('webkitAudioDecodedByteCount unsupported — skip silent-audio detector')
      return
    }

    let zeroChecks = 0
    const interval = setInterval(() => {
      if (v.paused || v.currentTime < 2) return
      const decoded = v.webkitAudioDecodedByteCount ?? 0
      if (decoded > 0) {
        log.debug(`audio decode OK — decoded=${decoded} bytes at ${v.currentTime.toFixed(1)}s`)
        clearInterval(interval)
        return
      }
      zeroChecks++
      log.debug(`audio decode check ${zeroChecks}/3 — decoded=0 at ${v.currentTime.toFixed(1)}s`)
      if (zeroChecks >= 3) {
        clearInterval(interval)
        advanceFallback(
          `silent audio: 0 bytes decoded after ${v.currentTime.toFixed(1)}s on ${playbackSource}`,
        )
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [playbackSource, advanceFallback])

  // On initial HLS load, use server-side seek for resume position.
  // startPosition alone doesn't work with EVENT playlists (segments don't exist yet).
  const hlsResumeApplied = useRef(false)
  useEffect(() => {
    if (!isHlsPlayback || hlsResumeApplied.current) return
    hlsResumeApplied.current = true
    // Server progress is the cross-device source of truth; local store may be
    // more recent on the same device. Take whichever is further ahead.
    const serverPos = playbackInfo?.position ?? 0
    const localPos = usePlayerStore.getState().lastPositions[mediaId] ?? 0
    const resumePos = Math.max(serverPos, localPos)
    if (resumePos > 5) {
      setHlsStartOffset(resumePos)
    }
  }, [isHlsPlayback, mediaId, playbackInfo?.position])

  // Seek feedback
  const [seekFeedback, setSeekFeedback] = useState<{ dir: 'back' | 'fwd'; n: number } | null>(null)
  const seekAccumulatorRef = useRef(0)
  const seekExecuteTimeoutRef = useRef<NodeJS.Timeout | null>(null)

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

  // ── Local actions ──────────────────────────────────────────────────────────
  const scrollSeasonCarousel = (direction: 'prev' | 'next') => {
    const carousel = seasonCarouselRef.current
    if (!carousel) return

    const amount = Math.max(carousel.clientWidth * 0.82, 360)
    carousel.scrollBy({
      left: direction === 'next' ? amount : -amount,
      behavior: 'smooth',
    })
  }

  const toggleDetailPanel = (panel: Exclude<DetailPanel, 'none'>) => {
    const nextPanel = activeTab === panel ? 'none' : panel
    if (nextPanel !== 'none' && isPlaying) {
      videoRef.current?.pause()
      setIsPlaying(false)
    }
    setActiveTab(nextPanel)
  }

  const togglePlay = () => {
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
    // Start or restart controls auto-hide timer
    if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
    controlsTimeoutRef.current = setTimeout(() => setShowControls(false), 3000)
  }

  const handleNextEpisode = useCallback(() => {
    if (!nextEpisodeMediaId) return
    if (nextEpisodeProgress?.completed) {
      if (isPlaying) {
        videoRef.current?.pause()
        setIsPlaying(false)
      }
      setShowWatchedWarning(true)
    } else {
      updateProgress({ mediaId, data: { position: currentTime, completed: false } })
      navigate(`/watch/${nextEpisodeMediaId}`)
    }
  }, [
    nextEpisodeMediaId,
    nextEpisodeProgress,
    navigate,
    isPlaying,
    updateProgress,
    mediaId,
    currentTime,
  ])

  const showSeekFeedback = (dir: 'back' | 'fwd', n: number) => {
    setSeekFeedback({ dir, n })
    if (seekFeedbackTimeout.current) clearTimeout(seekFeedbackTimeout.current)
    seekFeedbackTimeout.current = setTimeout(() => setSeekFeedback(null), 700)
  }

  const getEffectiveDuration = (video?: HTMLVideoElement | null) => {
    const nativeDuration =
      video && Number.isFinite(video.duration) && video.duration > 0 ? video.duration : 0
    return Math.max(nativeDuration, knownDurationRef.current, duration, playbackInfo?.duration ?? 0)
  }

  const clampSeekTarget = (targetTime: number, video?: HTMLVideoElement | null) => {
    const effectiveDuration = getEffectiveDuration(video)
    if (effectiveDuration <= 0) return Math.max(0, targetTime)
    const seekCeiling = effectiveDuration > 0.25 ? effectiveDuration - 0.25 : effectiveDuration
    return Math.max(0, Math.min(seekCeiling, targetTime))
  }

  const requestHlsSessionReload = (targetTime: number) => {
    const globalTarget = clampSeekTarget(targetTime, videoRef.current)
    const nextOffset = globalTarget > 0.25 ? globalTarget : null
    log.info(
      `sessionReload — target=${targetTime.toFixed(2)}, clamped=${globalTarget.toFixed(2)}, nextOffset=${nextOffset?.toFixed(2) ?? 'null'}`,
    )

    // Destroy old HLS instance BEFORE setting new offset to prevent the old
    // instance's pending requests from racing with the new session's FFmpeg.
    if (hlsRef.current) {
      hlsRef.current.destroy()
      hlsRef.current = null
    }

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
  }

  const applySeek = (targetTime: number) => {
    const video = videoRef.current
    if (!video) return 0

    const clampedTime = clampSeekTarget(targetTime, video)
    const sessionOffset = streamSourceOffsetRef.current
    const localTarget = clampedTime - sessionOffset
    // Session reload (server-side seek) is safe when video is copied (not transcoded):
    // - TranscodeAudio: video copy + audio transcode (lightweight)
    // - HLS fallback after a failed direct/pretranscode tier: video copy via HLS
    // HLS V2 architecture cleanly restarts transcoder processes.
    // We can safely reload the session for both VideoCopy and FullTranscode without resource leakage.
    const canReloadSession = isHlsPlayback && Boolean(streamUrls?.hls)
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
  }

  const accumulateSeek = (seconds: number) => {
    if (seconds === 0) return

    // Standardize addition direction
    if (seconds > 0) {
      if (seekAccumulatorRef.current <= 0) seekAccumulatorRef.current = seconds
      else seekAccumulatorRef.current += seconds
    } else {
      if (seekAccumulatorRef.current >= 0) seekAccumulatorRef.current = seconds
      else seekAccumulatorRef.current += seconds
    }

    showSeekFeedback(seconds > 0 ? 'fwd' : 'back', Math.abs(seekAccumulatorRef.current))

    if (seekExecuteTimeoutRef.current) clearTimeout(seekExecuteTimeoutRef.current)
    seekExecuteTimeoutRef.current = setTimeout(() => {
      applySeek((videoRef.current?.currentTime ?? 0) + seekAccumulatorRef.current)
      seekAccumulatorRef.current = 0
      setSeekFeedback(null)
    }, 800)
  }

  const seek = (seconds: number) => {
    accumulateSeek(seconds)
  }

  const changeVolume = (delta: number) => {
    const video = videoRef.current
    if (!video) return
    const newVolume = Math.max(0, Math.min(1, volume + delta))
    setVolume(newVolume)
    video.volume = newVolume
  }

  const resetControlsTimeout = () => {
    setShowControls(true)
    if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
    controlsTimeoutRef.current = setTimeout(() => {
      setShowControls(false)
    }, 3000)
  }

  // ── Progress bar ────────────────────────────────────────────────────────────
  const getTimeFromClientX = (clientX: number) => {
    if (!progressBarRef.current || !duration) return 0
    const rect = progressBarRef.current.getBoundingClientRect()
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width))
    return (x / rect.width) * duration
  }

  const handleBarMouseDown = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIsDraggingBar(true)
    const time = clampSeekTarget(getTimeFromClientX(e.clientX), videoRef.current)
    dragSeekTimeRef.current = time
    setDragSeekTime(time)
  }

  const handleBarMouseMove = (e: React.MouseEvent) => {
    if (!progressBarRef.current) return
    const rect = progressBarRef.current.getBoundingClientRect()
    const clampedTime = clampSeekTarget(getTimeFromClientX(e.clientX), videoRef.current)
    setHoverX(e.clientX - rect.left)
    setHoverTime(clampedTime)
    if (isDraggingBar) {
      dragSeekTimeRef.current = clampedTime
      setDragSeekTime(clampedTime)
    }
  }

  const handleDragMove = useEffectEvent((clientX: number) => {
    const time = clampSeekTarget(getTimeFromClientX(clientX), videoRef.current)
    dragSeekTimeRef.current = time
    setDragSeekTime(time)
  })

  const handleDragEnd = useEffectEvent(() => {
    const targetTime = dragSeekTimeRef.current
    if (targetTime != null) {
      applySeek(targetTime)
    }
    dragSeekTimeRef.current = null
    setDragSeekTime(null)
    setIsDraggingBar(false)
  })

  useEffect(() => {
    if (!isDraggingBar) return
    const onMove = (e: MouseEvent) => handleDragMove(e.clientX)
    const onUp = () => handleDragEnd()
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isDraggingBar])

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
    // Start with 0; FRAG_BUFFERED will set the correct offset once buffer data is available.
    streamSourceOffsetRef.current = 0
    hlsSeekableEndRef.current = sessionOffset
    setBufferedRange({
      start: sessionOffset,
      end: sessionOffset,
    })

    const useHls = isHlsPlayback
    // Pick URL by current playback tier — direct, pretranscode, or hls.
    const rawUrl = useHls
      ? sessionOffset > 0.25 && streamUrls.hls
        ? buildHlsSessionUrl(streamUrls.hls, sessionOffset, forceFullTranscode)
        : buildHlsSessionUrl(streamUrls.hls || '', 0, forceFullTranscode)
      : playbackSource === 'pretranscode'
        ? streamUrls.pretranscode
        : streamUrls.direct
    if (!rawUrl) return
    setIsBuffering(true)
    log.info(`player init — source=${playbackSource}, sessionOffset=${sessionOffset.toFixed(2)}s`)
    const hlsInitTimer = log.time('HLS init → first frame')
    const streamUrl = accessToken
      ? rawUrl + (rawUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(accessToken)
      : rawUrl

    if (useHls && streamUrls.hls && Hls.isSupported()) {
      // Resume is handled via server-side seek (?start=X in the URL).
      // FFmpeg begins encoding from the resume position, so HLS.js starts
      // from segment 0 of the offset session — no client-side startPosition needed.
      const hls = new Hls({
        // Tuning for high-bitrate live transcodes (BluRay REMUX ~16 Mbps,
        // ~12-15 MB per 6s segment). Chrome MSE SourceBuffer quota is ~150 MB.
        // We must limit our buffer limits so that we NEVER hit true QuotaExceededError.
        maxBufferLength: 10,
        maxMaxBufferLength: 15,
        backBufferLength: 5,
        // Cap physical bytes at 90 MB so hls.js manages memory BEFORE the browser throws error.
        maxBufferSize: 90 * 1000 * 1000,
        enableWorker: true,
        startPosition: -1,
        xhrSetup: (xhr) => {
          // Read fresh token on every request — prevents 401 when token refreshes during playback
          const freshToken = useAuthStore.getState().accessToken
          if (freshToken) xhr.setRequestHeader('Authorization', `Bearer ${freshToken}`)
        },
      })
      hlsRef.current = hls
      hls.on(Hls.Events.MANIFEST_PARSED, (_e, data) => {
        log.debug(
          `MANIFEST_PARSED — ${data.levels.length} levels, ${hls.audioTracks.length} audio tracks`,
        )
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
        // Sync currentTime state to the server-side offset
        if (sessionOffset > 0) setCurrentTime(sessionOffset)
        // Auto-play when ready
        log.debug(
          `MANIFEST_PARSED — initDur=${initDur.toFixed(2)}s, sessionOffset=${sessionOffset}, calling play()`,
        )
        v?.play().catch((err) => log.warn(`auto-play blocked: ${err.message}`))
      })
      hls.on(Hls.Events.LEVEL_LOADED, (_e, data) => {
        const playlistDur = data.details?.totalduration ?? 0
        const playlistEdge = data.details?.edge ?? playlistDur
        hlsPlaylistLiveRef.current = data.details?.live ?? false
        hlsSeekableEndRef.current = sessionOffset + playlistEdge
        log.debug(
          `LEVEL_LOADED — playlistDur=${playlistDur.toFixed(2)}s, edge=${playlistEdge.toFixed(2)}s, live=${data.details?.live}, fragments=${data.details?.fragments?.length ?? 0}`,
        )
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
      let consecutiveEmptyBuffers = 0

      hls.on(Hls.Events.FRAG_LOADED, (_e, data) => {
        log.debug(
          `FRAG_LOADED — sn=${data.frag.sn}, start=${data.frag.start.toFixed(2)}s, dur=${data.frag.duration.toFixed(2)}s, size=${((data.frag.stats?.loaded ?? 0) / 1024).toFixed(0)}KB`,
        )

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
      // Both video copy (-copyts -start_at_zero) and full transcode reset
      // timestamps to ~0. Detect the actual buffer start once the first
      // fragment is buffered and set streamSourceOffsetRef accordingly.
      const bufferSyncDone = { current: false }
      hls.on(Hls.Events.FRAG_BUFFERED, (_e, data) => {
        const v = videoRef.current
        if (!v) return
        const bufRanges = []
        for (let i = 0; i < v.buffered.length; i++) {
          bufRanges.push(`[${v.buffered.start(i).toFixed(2)}-${v.buffered.end(i).toFixed(2)}]`)
        }
        log.debug(
          `FRAG_BUFFERED — sn=${data.frag.sn}, video.currentTime=${v.currentTime.toFixed(2)}, readyState=${v.readyState}, buffered=${bufRanges.join(',') || 'empty'}, offset=${streamSourceOffsetRef.current.toFixed(2)}`,
        )

        if (v.buffered.length === 0) {
          consecutiveEmptyBuffers++
          if (consecutiveEmptyBuffers >= 5) {
            log.warn(
              `Stuck loading segments with 0 valid buffers appended (${consecutiveEmptyBuffers} times). Forcing full transcode recovery.`,
            )
            hls.destroy()
            setForceFullTranscode(true)
            setHlsStartOffset(
              streamSourceOffsetRef.current + Math.max(0, sessionOffset, v.currentTime) + 1.0,
            )
            return
          }
        } else {
          consecutiveEmptyBuffers = 0
        }

        if (bufferSyncDone.current) return
        if (v.buffered.length === 0) return
        const bufStart = v.buffered.start(0)

        if (sessionOffset > 0) {
          if (bufStart < sessionOffset * 0.8) {
            // Timestamps were reset to near 0 for full transcodes.
            // BE processes seek offset by SegLength chunks (floor(sessionOffset / 6.0) * 6.0).
            // So the video output natively starts exactly at trueStartOffset, NOT sessionOffset.
            const segLength = 6.0
            const trueStartOffset = Math.floor(sessionOffset / segLength) * segLength

            // Set the correct baseline so subtitles sync with the decoded visual frames
            streamSourceOffsetRef.current = trueStartOffset
            log.debug(
              `bufferSync — timestamps reset detected (bufStart=${bufStart.toFixed(2)}). True offset set to ${trueStartOffset.toFixed(2)} (requested=${sessionOffset.toFixed(2)})`,
            )

            // Seek the video forward to the exact remainder so playback resumes seamlessly
            const gap = sessionOffset - trueStartOffset
            const targetTime = bufStart + gap
            if (gap > 0.1 && Math.abs(v.currentTime - targetTime) > 0.5) {
              log.debug(
                `bufferSync — jumping relative gap to resume point: currentTime ${v.currentTime.toFixed(2)} → ${targetTime.toFixed(2)}`,
              )
              v.currentTime = targetTime
            } else if (bufStart > 1 && Math.abs(v.currentTime - bufStart) > 2) {
              v.currentTime = bufStart + 0.5
            }
          } else {
            // Fallback: timestamps NOT reset — video.currentTime IS the global PTS.
            // This happens when FFmpeg didn't fully shift timestamps. offset stays 0.
            log.warn(
              `bufferSync — timestamps NOT reset (bufStart=${bufStart.toFixed(2)}, sessionOffset=${sessionOffset.toFixed(2)}), globalTime = video.currentTime`,
            )
            if (bufStart > 1 && Math.abs(v.currentTime - bufStart) > 2) {
              // Jump to buffer start (+0.5s to land on a keyframe, avoiding readyState=2 stall).
              const jumpTo = bufStart + 0.5
              log.debug(
                `bufferSync — jumping currentTime ${v.currentTime.toFixed(2)} → ${jumpTo.toFixed(2)} (bufStart=${bufStart.toFixed(2)})`,
              )
              v.currentTime = jumpTo
            }
          }
        } else {
          if (bufStart > 1 && Math.abs(v.currentTime - bufStart) > 2) {
            // Jump to buffer start (+0.5s to land on a keyframe, avoiding readyState=2 stall).
            const jumpTo = bufStart + 0.5
            log.debug(
              `bufferSync — jumping currentTime ${v.currentTime.toFixed(2)} → ${jumpTo.toFixed(2)} (bufStart=${bufStart.toFixed(2)})`,
            )
            v.currentTime = jumpTo
          }
        }
        bufferSyncDone.current = true
        hlsInitTimer(`readyState=${v.readyState}, bufStart=${bufStart.toFixed(2)}`)
      })
      hls.on(Hls.Events.ERROR, (_e, data) => {
        const v = videoRef.current
        log.error(
          `[HLS ERROR] fatal=${data.fatal}, type=${data.type}, details=${data.details}, reason=${data.reason}, responseCode=${data.response?.code}, responseText=${data.response?.text?.substring(0, 50)}, readyState=${v?.readyState}, currentTime=${v?.currentTime?.toFixed(2)}, networkState=${v?.networkState}`,
        )

        // If HLS fails to append a segment, it gets stuck in an infinite loop downloading it again.
        // Explicitly recover the media source to flush the bad state.
        if (
          data.details === Hls.ErrorDetails.BUFFER_APPEND_ERROR ||
          data.details === Hls.ErrorDetails.BUFFER_APPENDING_ERROR ||
          data.details === Hls.ErrorDetails.BUFFER_STALLED_ERROR ||
          data.details === Hls.ErrorDetails.FRAG_PARSING_ERROR
        ) {
          log.warn(`Forcing hls.recoverMediaError() due to ${data.details}`)
          hls.recoverMediaError()

          // Nudge timeline forward to jump over the discontinuity gap if we're stalled
          if (v && data.details === Hls.ErrorDetails.BUFFER_STALLED_ERROR) {
            v.currentTime += 0.1
          }
          return
        }

        if (data.fatal) {
          const resumePos = videoRef.current?.currentTime ?? 0
          if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
            hls.startLoad(resumePos)
          } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
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
      // Resume: take whichever is further ahead (server = cross-device, local = same device)
      const seekToGlobal = Math.max(
        playbackInfo?.position ?? 0,
        usePlayerStore.getState().lastPositions[mediaId] ?? 0,
      )
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
  }, [
    audioLanguage,
    buildHlsSessionUrl,
    hlsStartOffset,
    isHlsPlayback,
    playbackSource,
    streamUrls,
    forceFullTranscode,
  ])

  // Stop backend transcode when leaving the page or switching media.
  // Uses beforeunload for hard reload / tab close.
  useEffect(() => {
    const streamSessionId = streamUrls?.stream_session_id
    if (!streamSessionId) return

    const stopTranscode = () => {
      api.delete(`/stream/sessions/${streamSessionId}`).catch(() => {})
    }
    window.addEventListener('beforeunload', stopTranscode)
    return () => {
      window.removeEventListener('beforeunload', stopTranscode)
      stopTranscode()
    }
  }, [streamUrls?.stream_session_id])

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
    // Native <track> is only for iOS native fullscreen (webkitEnterFullscreen).
    // - SRT/ASS (text): DualSubtitleOverlay renders → track.mode = 'disabled'
    // - PGS/VobSub (image): browser renders via <track> → track.mode = 'hidden'
    const suppressTracks = () => {
      if (!primarySub) {
        // No subtitle selected — disable all tracks
        for (let i = 0; i < video.textTracks.length; i++) {
          const track = video.textTracks[i]
          if (track.kind === 'subtitles' || track.kind === 'captions') {
            track.mode = 'disabled'
          }
        }
        return
      }
      // A subtitle is selected — only enable tracks matching the selected subtitle
      for (let i = 0; i < video.textTracks.length; i++) {
        const track = video.textTracks[i]
        if (track.kind === 'subtitles' || track.kind === 'captions') {
          const isMatch = track.language === effectiveSubtitleLanguage
          if (primarySub.is_image) {
            // Image subtitle: browser renders it → 'hidden' so cues are available
            track.mode = isMatch ? 'hidden' : 'disabled'
          } else {
            // Text subtitle: DualSubtitleOverlay renders it → 'disabled' to prevent double
            track.mode = 'disabled'
          }
        }
      }
    }
    suppressTracks()
    // Also suppress any tracks added later (e.g. HLS.js or browser auto-adding)
    video.textTracks.addEventListener('addtrack', suppressTracks)
    return () => {
      video.textTracks.removeEventListener('addtrack', suppressTracks)
    }
  }, [effectiveSubtitleLanguage, primarySub])

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
  const handleVideoTimeUpdate = useEffectEvent((video: HTMLVideoElement) => {
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
  })

  const handleVideoProgress = useEffectEvent((video: HTMLVideoElement) => {
    if (video.buffered.length > 0) {
      const bufferIndex = video.buffered.length - 1
      const offset = streamSourceOffsetRef.current
      setBufferedRange({
        start: offset + video.buffered.start(bufferIndex),
        end: offset + video.buffered.end(bufferIndex),
      })
    }
  })

  const syncVideoDuration = useEffectEvent((video: HTMLVideoElement) => {
    if (video.duration && !isNaN(video.duration) && isFinite(video.duration)) {
      setDuration((prev) =>
        Math.max(prev, streamSourceOffsetRef.current + video.duration, knownDurationRef.current),
      )
    }
  })

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    const onTimeUpdate = () => handleVideoTimeUpdate(video)
    const onProgress = () => handleVideoProgress(video)
    const onWaiting = () => {
      const bufRanges = []
      for (let i = 0; i < video.buffered.length; i++) {
        bufRanges.push(
          `[${video.buffered.start(i).toFixed(2)}-${video.buffered.end(i).toFixed(2)}]`,
        )
      }
      log.debug(
        `EVENT waiting — readyState=${video.readyState}, currentTime=${video.currentTime.toFixed(2)}, buffered=${bufRanges.join(',') || 'empty'}, networkState=${video.networkState}`,
      )
      setIsBuffering(true)
    }
    const onPlaying = () => {
      log.debug(
        `EVENT playing — readyState=${video.readyState}, currentTime=${video.currentTime.toFixed(2)}, paused=${video.paused}`,
      )
      setIsBuffering(false)
    }
    const onCanPlay = () => {
      log.debug(
        `EVENT canplay — readyState=${video.readyState}, currentTime=${video.currentTime.toFixed(2)}, paused=${video.paused}`,
      )
      setIsBuffering(false)
    }
    const onDurationChange = () => syncVideoDuration(video)
    const onSeeking = () =>
      log.debug(
        `EVENT seeking — to ${video.currentTime.toFixed(2)}, readyState=${video.readyState}`,
      )
    const onSeeked = () =>
      log.debug(`EVENT seeked — at ${video.currentTime.toFixed(2)}, readyState=${video.readyState}`)
    const onStalled = () =>
      log.warn(
        `EVENT stalled — currentTime=${video.currentTime.toFixed(2)}, readyState=${video.readyState}, networkState=${video.networkState}`,
      )
    const onSuspend = () =>
      log.debug(
        `EVENT suspend — currentTime=${video.currentTime.toFixed(2)}, readyState=${video.readyState}`,
      )
    const onLoadedData = () =>
      log.debug(
        `EVENT loadeddata — readyState=${video.readyState}, videoSize=${video.videoWidth}x${video.videoHeight}`,
      )
    video.addEventListener('timeupdate', onTimeUpdate)
    video.addEventListener('progress', onProgress)
    video.addEventListener('waiting', onWaiting)
    video.addEventListener('playing', onPlaying)
    video.addEventListener('canplay', onCanPlay)
    video.addEventListener('loadedmetadata', onDurationChange)
    video.addEventListener('durationchange', onDurationChange)
    video.addEventListener('loadeddata', onLoadedData)
    video.addEventListener('seeking', onSeeking)
    video.addEventListener('seeked', onSeeked)
    video.addEventListener('stalled', onStalled)
    video.addEventListener('suspend', onSuspend)
    return () => {
      video.removeEventListener('timeupdate', onTimeUpdate)
      video.removeEventListener('progress', onProgress)
      video.removeEventListener('waiting', onWaiting)
      video.removeEventListener('playing', onPlaying)
      video.removeEventListener('canplay', onCanPlay)
      video.removeEventListener('loadedmetadata', onDurationChange)
      video.removeEventListener('durationchange', onDurationChange)
      video.removeEventListener('loadeddata', onLoadedData)
      video.removeEventListener('seeking', onSeeking)
      video.removeEventListener('seeked', onSeeked)
      video.removeEventListener('stalled', onStalled)
      video.removeEventListener('suspend', onSuspend)
    }
  }, [streamUrls])

  useKeyboardShortcuts({
    togglePlay,
    seek,
    applySeek,
    changeVolume,
    toggleFullscreen,
    toggleMute,
    resetControlsTimeout,
    isFullscreen,
    isLocked,
    setIsLocked,
    seekStep: SEEK_STEP,
    volumeStep: VOLUME_STEP,
  })

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

  // ── Gesture tracking ───────────────────────────────────────────────────────
  const lastTapRef = useRef<{ time: number; x: number; y: number } | null>(null)

  const handlePointerUp = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      // Ignore clicks on controls or links
      if ((e.target as Element).closest('button, input, a, [role="button"], .pointer-events-auto'))
        return
      if (isLocked) return

      const now = Date.now()
      const lastTap = lastTapRef.current
      const rect = e.currentTarget.getBoundingClientRect()
      const x = e.clientX - rect.left
      const y = e.clientY - rect.top

      // Loosen the strict distance check for rapid subsequent taps.
      // And if we are already actively seeking, ANY tap in the zone within the 800ms window counts!
      const isStrictDoubleTap =
        lastTap &&
        now - lastTap.time < 300 &&
        Math.abs(e.clientX - lastTap.x) < 60 &&
        Math.abs(e.clientY - lastTap.y) < 60
      const zoneWidth = rect.width / 3
      const isActivelySeekingBack = seekAccumulatorRef.current < 0 && x < zoneWidth
      const isActivelySeekingFwd = seekAccumulatorRef.current > 0 && x > rect.width - zoneWidth

      if (isStrictDoubleTap || isActivelySeekingBack || isActivelySeekingFwd) {
        // Treat as a double tap or a continuation of an active seek
        lastTapRef.current = null

        if (e.pointerType === 'mouse') {
          toggleFullscreen()
        } else {
          // Touch double tap / seeking continuation
          if (x < zoneWidth) {
            accumulateSeek(-5)
          } else if (x > rect.width - zoneWidth) {
            accumulateSeek(5)
          } else {
            // Center double tap
            if (!showControls) {
              setShowControls(true)
              if (isPlaying) togglePlay()
            } else {
              togglePlay()
            }
          }
        }
      } else {
        // Single tap candidate
        lastTapRef.current = { time: now, x: e.clientX, y: e.clientY }

        const v = videoRef.current
        const isInsideViewport = () => {
          if (!v) return true
          const vw = v.videoWidth || 16
          const vh = v.videoHeight || 9
          const aspect = vw / vh
          const screenAspect = rect.width / rect.height
          let viewW = rect.width
          let viewH = rect.height
          if (aspectRatio === 'contain') {
            if (aspect > screenAspect) viewH = rect.width / aspect
            else viewW = rect.height * aspect
          }
          const pillar = Math.max(0, (rect.width - viewW) / 2)
          const letter = Math.max(0, (rect.height - viewH) / 2)
          return x >= pillar && x <= rect.width - pillar && y >= letter && y <= rect.height - letter
        }

        if (e.pointerType === 'mouse') {
          // Instant play/pause for mouse
          if (isInsideViewport()) togglePlay()
        } else {
          // Delayed single tap for touch
          setTimeout(() => {
            if (lastTapRef.current?.time === now) {
              lastTapRef.current = null
              if (showControls) {
                setShowControls(false)
              } else {
                setShowControls(true)
                if (isPlaying && isInsideViewport()) {
                  togglePlay()
                }
              }
            }
          }, 300)
        }
      }
    },
    [isLocked, showControls, isPlaying, aspectRatio, toggleFullscreen, togglePlay],
  )

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
      className={`fixed inset-0 bg-[#141414] select-none touch-none overflow-hidden ${
        !showControls && isPlaying ? 'cursor-none' : ''
      }`}
      onMouseMove={() => {
        if (!isLocked) resetControlsTimeout()
      }}
      onMouseLeave={() => {
        if (!isLocked) {
          if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
          setShowControls(false)
        }
      }}
      onPointerUp={handlePointerUp}
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
            handleNextEpisode()
            return
          }
          setIsPlaying(false)
          updateProgress({ mediaId, data: { position: duration, completed: true } })
        }}
        onError={(e) => {
          const rawErr = videoRef.current?.error
          log.error(
            `[Native <video> ERROR] code=${rawErr?.code}, message=${rawErr?.message}, source=${playbackSource}`,
          )
          // Try the next tier in the fallback chain (direct → pretranscode → hls).
          // If the chain is exhausted (already on hls and it failed), surface the error.
          if (playbackSource === 'hls') {
            setError(`Video playback error (Code: ${rawErr?.code || 'unknown'})`)
          } else {
            advanceFallback(
              `<video> onError on ${playbackSource ?? 'unknown'} (code=${rawErr?.code})`,
            )
          }
        }}
      >
        {/* Native <track> for iOS native fullscreen (webkitEnterFullscreen fallback) */}
        {primarySub && !primarySub.is_image && subtitleServeUrl(primarySub) && (
          <track
            kind="subtitles"
            src={subtitleServeUrl(primarySub)!}
            srcLang={primarySub.language ?? 'und'}
            label={primarySub.label ?? primarySub.language ?? 'Subtitles'}
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
        <>
          <style>{`
            @keyframes seekArrowFade {
              0%, 100% { opacity: 0.2; }
              50% { opacity: 1; }
            }
            .seek-arrow-1 { animation: seekArrowFade 0.6s infinite 0s; }
            .seek-arrow-2 { animation: seekArrowFade 0.6s infinite 0.15s; }
            .seek-arrow-3 { animation: seekArrowFade 0.6s infinite 0.3s; }
            @keyframes scaleFadeIn {
              0% { opacity: 0; transform: scale(0.95); }
              100% { opacity: 1; transform: scale(1); }
            }
            .animate-scale-fade {
              animation: scaleFadeIn 0.15s ease-out forwards;
            }
          `}</style>
          <div
            className={`pointer-events-none absolute inset-y-0 flex w-[40%] items-center animate-scale-fade ${
              seekFeedback.dir === 'back'
                ? 'left-0 justify-start pl-[10%]'
                : 'right-0 justify-end pr-[10%]'
            }`}
          >
            <div className="flex flex-col items-center justify-center text-white text-base sm:text-lg font-bold drop-shadow-lg">
              {seekFeedback.dir === 'back' ? (
                <div className="flex -space-x-3 mb-1">
                  <LuChevronLeft className="seek-arrow-3" size={32} />
                  <LuChevronLeft className="seek-arrow-2" size={32} />
                  <LuChevronLeft className="seek-arrow-1" size={32} />
                </div>
              ) : (
                <div className="flex -space-x-3 mb-1">
                  <LuChevronRight className="seek-arrow-1" size={32} />
                  <LuChevronRight className="seek-arrow-2" size={32} />
                  <LuChevronRight className="seek-arrow-3" size={32} />
                </div>
              )}
              <span>
                {seekFeedback.dir === 'back' ? '-' : '+'}
                {seekFeedback.n}s
              </span>
            </div>
          </div>
        </>
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
              onClick={handleNextEpisode}
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
      >
        <div onClick={(e) => e.stopPropagation()} onPointerUp={(e) => e.stopPropagation()}>
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
          onPointerUp={(e) => e.stopPropagation()}
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
                          </div>
                        )}
                      </div>
                    )}
                  </div>

                  {/* Next episode */}
                  {nextEpisodeMediaId && (
                    <button
                      onClick={handleNextEpisode}
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

      {/* Watched Warning Overlay */}
      {showWatchedWarning && (
        <div className="absolute inset-0 z-[1000] flex items-center justify-center bg-black/80 backdrop-blur-sm">
          <div className="w-[400px] max-w-[90vw] rounded-2xl bg-[#1e1e1e] border border-white/10 p-6 shadow-2xl">
            <h3 className="text-xl font-bold text-white mb-2">Đã xem tập này</h3>
            <p className="text-sm text-gray-400 mb-6">
              Bạn đã xem xong tập phim tiếp theo. Bạn muốn xem lại từ đầu hay chuyển qua tập kế
              tiếp?
            </p>
            <div className="flex justify-end gap-3">
              <button
                className="flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2 text-sm font-medium text-white hover:bg-white/20 transition-colors"
                onClick={(e) => {
                  e.stopPropagation()
                  setShowWatchedWarning(false)
                  updateProgress({
                    mediaId: nextEpisodeMediaId!,
                    data: { position: 0, completed: false },
                  })
                  navigate(`/watch/${nextEpisodeMediaId}`)
                }}
              >
                <LuRotateCcw size={16} />
                <span>Xem lại từ đầu</span>
              </button>
              {nextNextEpisodeMediaId && (
                <button
                  className="flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2 text-sm font-medium text-white hover:bg-white/20 transition-colors"
                  onClick={(e) => {
                    e.stopPropagation()
                    setShowWatchedWarning(false)
                    updateProgress({ mediaId, data: { position: currentTime, completed: false } })
                    navigate(`/watch/${nextNextEpisodeMediaId}`)
                  }}
                >
                  <LuSkipForward size={16} />
                  <span>Tập kế tiếp</span>
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
