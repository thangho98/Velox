import { useState, useEffect, useCallback, useRef } from 'react'
import {
  View,
  Text,
  Image,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Pressable,
  ScrollView,
  Modal,
} from 'react-native'
import { Volume2, Sun, X, ChevronLeft, Settings, Pause, Play, Lock, RotateCcw, RotateCw, SkipForward, Music, Maximize2, Minimize2, Captions } from 'lucide-react-native'
import { CastButton } from '../components/CastButton'
import { useChromecast } from '../hooks/useChromecast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'
import { useRoute, RouteProp, useNavigation } from '@react-navigation/native'
import { useVideoPlayer, VideoView } from 'expo-video'
import type { VideoPlayerStatus } from 'expo-video'
import { GestureDetector, Gesture, GestureHandlerRootView } from 'react-native-gesture-handler'

// SubtitleTrack type (from expo-video)
type SubtitleTrack = {
  id?: string
  language: string
  label: string
}
import { useStreamUrl, useEpisodes, usePlaybackInfo, useMedia } from '@velox/shared/hooks'
import { useProgress, useUpdateProgress } from '@velox/shared/hooks/media/useProgress'
import { useCinemaSettings } from '@velox/shared/hooks/settings'
import type { CinemaItem } from '@velox/shared/hooks/media/useCinema'
import type { PlaybackSubtitleTrack } from '@velox/shared/types'
import { mediaImage, parseSubtitleLabel, languageMatches, buildVisibleSubtitles } from '@velox/shared/lib'
import { api } from '@velox/shared/api'
import { DualSubtitleOverlay } from '../components/DualSubtitleOverlay'
import { CinemaOverlay } from '../components/CinemaOverlay'
import { usePlayerStore } from '../stores/player'
import { useAuthStore } from '../stores/auth'
import { getServerUrl } from '../platform/mobile-adapter'
import type { RootStackParamList } from '../../App'

type VideoRouteProp = RouteProp<RootStackParamList, 'Episode'> | RouteProp<RootStackParamList, 'Media'>

// Aspect ratio type
type AspectRatio = 'contain' | 'cover' | 'fill'

// Repeat mode type
type RepeatMode = 'off' | 'one' | 'all'

const CONTROLS_HIDE_TIMEOUT = 3000
const STATE_POLL_INTERVAL = 500

/** Current wall clock as HH:MM string. */
function getWallClock(): string {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

interface QualityOption {
  label: string
  height: number
}

export function VideoPlayerScreen() {
  const route = useRoute<VideoRouteProp>()
  const navigation = useNavigation()
  const { id: mediaId, seriesId, startFromBeginning } = route.params as {
    id: number
    seriesId?: number
    startFromBeginning?: boolean
  }

  // Responsive layout
  const layout = useResponsiveLayout()
  const SCREEN_WIDTH = layout.width
  const SCREEN_HEIGHT = layout.height

  // Scaled sizes based on device
  const iconSize = layout.largeControls ? 36 : layout.device === 'tablet' ? 28 : 24
  const topIconSize = layout.largeControls ? 28 : layout.device === 'tablet' ? 22 : 18
  const playPauseIconSize = layout.largeControls ? 48 : layout.device === 'tablet' ? 40 : 36
  const progressBarHeight = layout.largeControls ? 6 : layout.device === 'tablet' ? 5 : 4
  const touchTargetSize = layout.largeControls ? 64 : layout.device === 'tablet' ? 52 : 44
  const seekButtonSize = layout.largeControls ? 80 : layout.device === 'tablet' ? 70 : 60
  const playButtonSize = layout.largeControls ? 100 : layout.device === 'tablet' ? 90 : 80
  const controlsGap = layout.largeControls ? 64 : layout.device === 'tablet' ? 56 : 48
  const bottomPadding = layout.largeControls ? 32 : 24

  // Chromecast
  const chromecast = useChromecast()

  // Auth token for subtitle URL construction
  const accessToken = useAuthStore((s) => s.accessToken)

  // --- Cinema Mode (pre-roll trailers/intro before main video) ---
  const { data: cinemaSettings } = useCinemaSettings()
  const cinemaEnabled = cinemaSettings?.enabled ?? false
  const [cinemaPhase, setCinemaPhase] = useState<'cinema' | 'playing'>('cinema')
  const [cinemaItems, setCinemaItems] = useState<CinemaItem[]>([])
  const [cinemaLoading, setCinemaLoading] = useState(true)

  useEffect(() => {
    if (!cinemaEnabled) {
      setCinemaPhase('playing')
      setCinemaLoading(false)
      return
    }
    const fetchCinema = async () => {
      try {
        const endpoint = seriesId
          ? `/series/${seriesId}/cinema`
          : `/media/${mediaId}/cinema`
        const data = await api.get<{ items: CinemaItem[] }>(endpoint)
        const items = data?.items ?? []
        if (items.length === 0) {
          setCinemaPhase('playing')
        } else {
          setCinemaItems(items)
          setCinemaPhase('cinema')
        }
      } catch {
        setCinemaPhase('playing')
      } finally {
        setCinemaLoading(false)
      }
    }
    fetchCinema()
  }, [mediaId, seriesId, cinemaEnabled])

  const handleCinemaComplete = useCallback(() => {
    setCinemaPhase('playing')
  }, [])

  // Stream URL state
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { mutateAsync: fetchStreamUrl, data: streamData } = useStreamUrl(mediaId)
  const { data: progress } = useProgress(mediaId)
  const updateProgress = useUpdateProgress()

  // Media detail for title
  const { data: mediaDetail } = useMedia(mediaId)

  // Episode data for next episode
  const { data: episodes } = useEpisodes(seriesId || 0, 0)

  // Playback info (for subtitle tracks)
  const { data: playbackInfo } = usePlaybackInfo(mediaId)

  // Build VTT subtitle serve URL from a PlaybackSubtitleTrack
  const subtitleServeUrl = (sub: PlaybackSubtitleTrack | undefined): string | null => {
    if (!sub || !playbackInfo?.primary_file_id) return null
    const serverUrl = getServerUrl()
    if (!serverUrl) return null
    const base = `${serverUrl}/api/media-files/${playbackInfo.primary_file_id}/subtitles/${sub.id}/serve`
    return accessToken ? `${base}?token=${encodeURIComponent(accessToken)}` : base
  }

  // Video source
  const videoSource = streamData?.hls_url || streamData?.direct_url || null

  // Player state
  const playerRef = useRef<ReturnType<typeof useVideoPlayer> | null>(null)
  const [playerStatus, setPlayerStatus] = useState<VideoPlayerStatus>('idle')
  const [isPlaying, setIsPlaying] = useState(false)
  const [isBuffering, setIsBuffering] = useState(false)
  const userWantsToPlayRef = useRef(true) // user intent: true = wants to play
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showControls, setShowControls] = useState(true)
  const [showSubtitleMenu, setShowSubtitleMenu] = useState(false)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)
  const [showAudioMenu, setShowAudioMenu] = useState(false)
  const [showNextEpisode, setShowNextEpisode] = useState(false)
  const [nextEpisodeCountdown, setNextEpisodeCountdown] = useState(15)
  const [isLocked, setIsLocked] = useState(false)
  const [showStats, setShowStats] = useState(false)
  const [selectedSubtitle, setSelectedSubtitle] = useState<string | null>(null)
  const [availableSubtitles, setAvailableSubtitles] = useState<SubtitleTrack[]>([])
  const [availableAudioTracks, setAvailableAudioTracks] = useState<string[]>([])
  const [selectedAudioTrack, setSelectedAudioTrack] = useState<string | null>(null)
  const [playbackSpeed, setPlaybackSpeed] = useState(1)
  const [isPlayerReady, setIsPlayerReady] = useState(false)

  // Wall clock (HH:MM)
  const [wallClock, setWallClock] = useState(() => getWallClock())

  // Subtitle appearance from player store
  const subtitleSize = usePlayerStore((s) => s.subtitleSize)
  const subtitleColor = usePlayerStore((s) => s.subtitleColor)
  const subtitleBackground = usePlayerStore((s) => s.subtitleBackground)
  const setSubtitleSize = usePlayerStore((s) => s.setSubtitleSize)
  const setSubtitleColor = usePlayerStore((s) => s.setSubtitleColor)
  const setSubtitleBackground = usePlayerStore((s) => s.setSubtitleBackground)


  // Video stats (enhanced with expo-video documented APIs)
  const [videoStats, setVideoStats] = useState({
    bitrate: 0,
    resolution: '',
    codec: '',
    frameRate: 0,
    bufferHealth: 0,
    qualityLevels: 0,
    abrSwitches: 0,
    connectionQuality: 'N/A' as string,
    audioInfo: '',
  })
  const abrSwitchCountRef = useRef(0)

  // Gesture feedback
  const [gestureSeek, setGestureSeek] = useState<number | null>(null)
  const [gestureVolume, setGestureVolume] = useState<number | null>(null)
  const [gestureBrightness, setGestureBrightness] = useState<number | null>(null)

  // Resolve primary & secondary subtitle VTT URLs from backend tracks
  const backendSubtitleTracks = playbackInfo?.subtitle_tracks ?? []
  const selectedNativeSub = availableSubtitles.find((t) => t.id === selectedSubtitle)
  const primaryBackendSub = selectedNativeSub
    ? backendSubtitleTracks.find(
        (t) => t.language === selectedNativeSub.language && !t.is_image,
      )
    : undefined
  const primarySubtitleVttUrl = subtitleServeUrl(primaryBackendSub)
  const secondarySubLang = usePlayerStore((s) => s.secondarySubtitleLanguage)
  const setSecondarySubtitleLanguage = usePlayerStore((s) => s.setSecondarySubtitleLanguage)
  const secondaryBackendSub = secondarySubLang
    ? backendSubtitleTracks.find((t) => t.language === secondarySubLang && !t.is_image)
    : undefined
  const secondarySubtitleVttUrl = subtitleServeUrl(secondaryBackendSub)

  // Episode metadata
  const currentEpisode = episodes?.find((ep) => ep.id === mediaId)
  const mediaTitle = currentEpisode?.title || mediaDetail?.title || ''

  // Skip markers (would come from backend - intro/outro timestamps)
  // @ts-ignore - intro_end and outro_start from episode metadata if available
  const introEnd = currentEpisode?.intro_end
  // @ts-ignore
  const outroStart = currentEpisode?.outro_start
  const [showSkipIntro, setShowSkipIntro] = useState(false)
  const [showSkipOutro, setShowSkipOutro] = useState(false)

  const videoViewRef = useRef<{ enterFullscreen: () => void; exitFullscreen: () => void } | null>(null)
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const progressSaveTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hasSeekedRef = useRef(false)
  const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Quality options (for HLS streams)
  const qualityOptions: QualityOption[] = [
    { label: 'Auto', height: 0 },
    { label: '1080p', height: 1080 },
    { label: '720p', height: 720 },
    { label: '480p', height: 480 },
    { label: '360p', height: 360 },
  ]
  const [selectedQuality, setSelectedQuality] = useState('Auto')

  // Speed options
  const speedOptions = [
    { label: '0.5x', value: 0.5 },
    { label: '0.75x', value: 0.75 },
    { label: '1x (Normal)', value: 1 },
    { label: '1.25x', value: 1.25 },
    { label: '1.5x', value: 1.5 },
    { label: '2x', value: 2 },
  ]

  // Aspect ratio options
  const aspectRatioOptions: { label: string; value: AspectRatio }[] = [
    { label: 'Contain', value: 'contain' },
    { label: 'Cover', value: 'cover' },
    { label: 'Fill', value: 'fill' },
  ]
  const [aspectRatio, setAspectRatio] = useState<AspectRatio>('contain')

  // Repeat mode options
  const repeatModeOptions: { label: string; value: RepeatMode; icon: string }[] = [
    { label: 'Off', value: 'off', icon: '🔁' },
    { label: 'One', value: 'one', icon: '🔂' },
    { label: 'All', value: 'all', icon: '🔁' },
  ]
  const [repeatMode, setRepeatMode] = useState<RepeatMode>('off')

  // Subtitle delay offset (stored per mediaId, in seconds)
  const [subtitleDelay, setSubtitleDelay] = useState(0)


  // Aspect ratio modal
  const [showAspectRatioMenu, setShowAspectRatioMenu] = useState(false)

  // Double tap detection
  const lastTapRef = useRef<number>(0)
  const tapXRef = useRef<number>(0)

  // Fetch stream URL
  useEffect(() => {
    const loadStream = async () => {
      setIsLoading(true)
      setError(null)
      try {
        await fetchStreamUrl()
        setIsLoading(false)
      } catch (e) {
        setError('Failed to load video')
        setIsLoading(false)
      }
    }
    loadStream()
  }, [mediaId])

  // Create player with setup callback
  const player = useVideoPlayer(videoSource, (p) => {
    playerRef.current = p
    p.timeUpdateEventInterval = 1
    p.volume = 1
    p.muted = false
  })

  // Poll player state
  useEffect(() => {
    if (!player) return

    const pollState = () => {
      setPlayerStatus(player.status)
      setIsPlaying(player.playing)
      // Buffering = user wants to play but player is not playing (stalled/loading)
      setIsBuffering(userWantsToPlayRef.current && !player.playing && player.status !== 'idle')
      setCurrentTime(player.currentTime)
      setDuration(player.duration)
      setVolume(player.volume)
      setIsMuted(player.muted)
      // @ts-ignore - availableSubtitleTracks type mismatch with local SubtitleTrack
      setAvailableSubtitles(player.availableSubtitleTracks)
      // @ts-ignore - availableAudioTracks not in types but exists on player
      setAvailableAudioTracks(player.availableAudioTracks || [])
      // @ts-ignore - playbackRate not in types but exists
      setPlaybackSpeed(player.playbackRate || 1)

      // Update video stats from documented expo-video APIs
      if (showStats) {
        const track = player.videoTrack
        const bitrateRaw = track?.peakBitrate ?? track?.averageBitrate ?? track?.bitrate ?? 0
        const bufferSec =
          player.bufferedPosition >= 0 && player.currentTime >= 0
            ? Math.max(0, player.bufferedPosition - player.currentTime)
            : 0

        // Connection quality heuristic based on buffer health
        let connectionQuality = 'N/A'
        if (player.bufferedPosition >= 0) {
          if (bufferSec >= 10) connectionQuality = 'Excellent'
          else if (bufferSec >= 5) connectionQuality = 'Good'
          else if (bufferSec >= 2) connectionQuality = 'Poor'
          else connectionQuality = 'Buffering...'
        }

        // Audio info from current audio track
        const audio = player.audioTrack
        const audioInfo = audio ? (audio.label || audio.name || audio.language || 'Unknown') : 'N/A'

        setVideoStats({
          bitrate: bitrateRaw,
          resolution:
            track?.size ? `${track.size.width}\u00D7${track.size.height}` : '',
          codec: track?.mimeType ?? '',
          frameRate: track?.frameRate ?? 0,
          bufferHealth: bufferSec,
          qualityLevels: player.availableVideoTracks?.length ?? 0,
          abrSwitches: abrSwitchCountRef.current,
          connectionQuality,
          audioInfo,
        })
      }

      if (player.status === 'readyToPlay' && !isPlayerReady) {
        setIsPlayerReady(true)
      }
    }

    const interval = setInterval(pollState, STATE_POLL_INTERVAL)
    pollState() // Initial poll

    return () => clearInterval(interval)
  }, [player, showStats])

  // Track ABR (adaptive bitrate) switches via videoTrackChange event
  useEffect(() => {
    if (!player) return
    abrSwitchCountRef.current = 0

    const subscription = player.addListener('videoTrackChange', () => {
      abrSwitchCountRef.current += 1
    })

    return () => subscription.remove()
  }, [player])

  // Check for next episode when near end
  useEffect(() => {
    if (!seriesId || !episodes || episodes.length === 0) return
    if (!isPlaying) return

    const timeRemaining = duration - currentTime
    const isNearEnd = timeRemaining > 0 && timeRemaining < 30 // Last 30 seconds

    if (isNearEnd && !showNextEpisode) {
      // Find next episode
      const currentEpisode = episodes.find((ep) => ep.id === mediaId)
      if (currentEpisode) {
        const currentIndex = episodes.findIndex((ep) => ep.id === mediaId)
        if (currentIndex < episodes.length - 1) {
          setShowNextEpisode(true)
          setNextEpisodeCountdown(15)
        }
      }
    }
  }, [currentTime, duration, isPlaying, seriesId, episodes, mediaId])

  // Check for intro/outro skip markers
  useEffect(() => {
    if (!isPlaying || !currentTime) return

    // Check if in intro region
    if (introEnd && currentTime < introEnd && currentTime > 0) {
      setShowSkipIntro(true)
    } else {
      setShowSkipIntro(false)
    }

    // Check if in outro region
    if (outroStart && duration > 0 && currentTime >= outroStart) {
      setShowSkipOutro(true)
    } else {
      setShowSkipOutro(false)
    }
  }, [currentTime, introEnd, outroStart, duration, isPlaying])

  // Next episode countdown
  useEffect(() => {
    if (!showNextEpisode) {
      if (countdownIntervalRef.current) {
        clearInterval(countdownIntervalRef.current)
      }
      return
    }

    countdownIntervalRef.current = setInterval(() => {
      setNextEpisodeCountdown((prev) => {
        if (prev <= 1) {
          // Auto-play next episode
          const currentIndex = episodes?.findIndex((ep) => ep.id === mediaId) ?? -1
          if (currentIndex >= 0 && episodes && currentIndex < episodes.length - 1) {
            const nextEp = episodes[currentIndex + 1]
            handlePlayNextEpisode(nextEp.id)
          }
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => {
      if (countdownIntervalRef.current) {
        clearInterval(countdownIntervalRef.current)
      }
    }
  }, [showNextEpisode, episodes, mediaId])

  // Resume from previous position
  useEffect(() => {
    if (!player || !progress || progress.position <= 0 || hasSeekedRef.current) return
    if (playerStatus !== 'readyToPlay') return
    if (startFromBeginning) return

    if (Math.abs(player.currentTime - progress.position / 1000) > 5) {
      player.currentTime = progress.position / 1000
      hasSeekedRef.current = true
    }
  }, [player, progress, playerStatus, startFromBeginning])

  // Schedule controls hide
  const scheduleHideControls = useCallback(() => {
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current)
    }
    if (!isLocked) {
      controlsTimeoutRef.current = setTimeout(() => {
        if (isPlaying) {
          setShowControls(false)
        }
      }, CONTROLS_HIDE_TIMEOUT)
    }
  }, [isPlaying, isLocked])

  // Handle tap
  const handleTapScreen = (event: { nativeEvent: { locationX: number; locationY: number } }) => {
    if (isLocked) {
      return
    }

    const now = Date.now()
    const tapX = event.nativeEvent.locationX
    const isDoubleTap = now - lastTapRef.current < 300 && Math.abs(tapX - tapXRef.current) < 50

    if (isDoubleTap) {
      handleDoubleTap(tapX)
    } else {
      setShowControls(true)
      scheduleHideControls()
    }

    lastTapRef.current = now
    tapXRef.current = tapX
  }

  // Handle double tap for seek
  const handleDoubleTap = (x: number) => {
    if (isLocked) {
      setIsLocked(false)
      return
    }
    // Wider seek zones on tablet/TV
    const seekZoneDivider = layout.largeControls ? 2.5 : layout.device === 'tablet' ? 2.2 : 2
    if (x < SCREEN_WIDTH / seekZoneDivider) {
      handleSeek(-10)
    } else if (x > SCREEN_WIDTH - SCREEN_WIDTH / seekZoneDivider) {
      handleSeek(10)
    }
  }

  // Skip intro/outro handlers
  const handleSkipIntro = () => {
    if (introEnd && player) {
      player.currentTime = introEnd
    }
    setShowSkipIntro(false)
    scheduleHideControls()
  }

  const handleSkipOutro = () => {
    if (outroStart && player) {
      player.currentTime = outroStart
    }
    setShowSkipOutro(false)
    scheduleHideControls()
  }

  // Playback controls
  const togglePlayPause = () => {
    if (!player) return
    if (player.playing || userWantsToPlayRef.current) {
      player.pause()
      userWantsToPlayRef.current = false
    } else {
      player.play()
      userWantsToPlayRef.current = true
    }
    scheduleHideControls()
  }

  const handleSeek = (seconds: number) => {
    if (!player) return
    const newTime = Math.max(0, Math.min(player.currentTime + seconds, duration))
    player.currentTime = newTime
    scheduleHideControls()
  }

  const handleSeekTo = (time: number) => {
    if (!player) return
    player.currentTime = time
    scheduleHideControls()
  }

  const toggleMute = () => {
    if (!player) return
    player.muted = !player.muted
    scheduleHideControls()
  }

  const handleVolumeChange = (delta: number) => {
    if (!player) return
    const newVolume = Math.max(0, Math.min(1, player.volume + delta))
    player.volume = newVolume
    scheduleHideControls()
  }

  const toggleFullscreen = () => {
    if (isFullscreen) {
      videoViewRef.current?.exitFullscreen()
    } else {
      videoViewRef.current?.enterFullscreen()
    }
    setIsFullscreen(!isFullscreen)
    scheduleHideControls()
  }

  // Subtitle handling
  const handleSubtitleSelect = (track: SubtitleTrack | null) => {
    if (!player) return
    player.subtitleTrack = track
    setSelectedSubtitle(track?.id || null)
    setShowSubtitleMenu(false)
    scheduleHideControls()
  }

  // Quality handling
  const handleQualitySelect = (qualityLabel: string) => {
    setSelectedQuality(qualityLabel)
    scheduleHideControls()
  }

  // Speed handling
  const handleSpeedSelect = (speed: number) => {
    if (!player) return
    // @ts-ignore - playbackRate not in types
    player.playbackRate = speed
    setPlaybackSpeed(speed)
    setShowSpeedMenu(false)
    scheduleHideControls()
  }

  // Audio track handling
  const handleAudioSelect = (trackId: string | null) => {
    if (!player) return
    // @ts-ignore - selectedAudioTrack not in types
    player.selectedAudioTrack = trackId ? { value: trackId } : null
    setSelectedAudioTrack(trackId)
    setShowAudioMenu(false)
    scheduleHideControls()
  }

  // Lock/unlock controls
  const toggleLock = () => {
    setIsLocked(!isLocked)
    if (!isLocked) {
      setShowControls(false)
    } else {
      setShowControls(true)
    }
  }

  // Toggle stats overlay
  const toggleStats = () => {
    setShowStats(!showStats)
  }

  // Aspect ratio handling
  const handleAspectRatioSelect = (ratio: AspectRatio) => {
    setAspectRatio(ratio)
    setShowAspectRatioMenu(false)
    scheduleHideControls()
  }

  // Repeat mode handling
  const handleRepeatModeToggle = () => {
    const modes: RepeatMode[] = ['off', 'one', 'all']
    const currentIndex = modes.indexOf(repeatMode)
    const nextIndex = (currentIndex + 1) % modes.length
    setRepeatMode(modes[nextIndex])
    scheduleHideControls()
  }

  // Handle repeat mode when playback ends
  useEffect(() => {
    if (!player) return

    const handlePlaybackEnd = () => {
      if (repeatMode === 'one') {
        player.currentTime = 0
        player.play()
      } else if (repeatMode === 'all' && seriesId) {
        handlePlayNextEpisode(nextEpisode?.id ?? 0)
      }
    }

    // @ts-ignore -- expo-video status event listener (type definitions lag behind runtime)
    const listener = player.addListener('statusDidChange', () => {
      // @ts-ignore -- 'ended' exists at runtime
      if (player.status === 'ended') {
        handlePlaybackEnd()
      }
    })

    return () => {
      // @ts-ignore
      if (listener?.remove) listener.remove()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [player, repeatMode, seriesId])

  // Subtitle delay handling
  const handleSubtitleDelayAdjust = (delta: number) => {
    setSubtitleDelay((prev) => Math.round((prev + delta) * 4) / 4) // Round to nearest 0.25
    scheduleHideControls()
  }

  const handleSubtitleDelayReset = () => {
    setSubtitleDelay(0)
    scheduleHideControls()
  }

  // Cancel next episode
  const handleCancelNextEpisode = () => {
    setShowNextEpisode(false)
    if (countdownIntervalRef.current) {
      clearInterval(countdownIntervalRef.current)
    }
  }

  // Play next episode
  const handlePlayNextEpisode = (nextEpisodeId: number) => {
    handleCancelNextEpisode()
    saveCurrentProgress()
    // @ts-ignore - navigation types
    navigation.navigate('Episode', { id: nextEpisodeId, seriesId })
  }

  // Save progress periodically
  useEffect(() => {
    if (!player || !isPlaying || currentTime <= 0) return

    if (progressSaveTimeoutRef.current) {
      clearTimeout(progressSaveTimeoutRef.current)
    }

    progressSaveTimeoutRef.current = setTimeout(() => {
      const positionMs = Math.floor(currentTime * 1000)
      const completed = duration > 0 && positionMs >= (duration * 1000 - 30000)
      updateProgress.mutate({
        mediaId,
        data: { position: positionMs, completed },
      })
    }, 5000)

    return () => {
      if (progressSaveTimeoutRef.current) {
        clearTimeout(progressSaveTimeoutRef.current)
      }
    }
  }, [currentTime, isPlaying, duration, player, mediaId])

  // Save progress on pause/exit
  const saveCurrentProgress = useCallback(() => {
    if (!player || currentTime <= 0) return
    const positionMs = Math.floor(currentTime * 1000)
    const completed = duration > 0 && positionMs >= (duration * 1000 - 30000)
    updateProgress.mutate({
      mediaId,
      data: { position: positionMs, completed },
    })
  }, [player, currentTime, duration, mediaId])

  // Handle back navigation
  const handleBack = () => {
    handleCancelNextEpisode()
    saveCurrentProgress()
    navigation.goBack()
  }

  // Cast current media to Chromecast
  const handleCastMedia = useCallback(() => {
    if (!chromecast.connected) return

    const title = mediaTitle || `Media ${mediaId}`
    const posterUrl = currentEpisode?.still_path
      ? mediaImage(currentEpisode.still_path, 'w500') || undefined
      : undefined

    chromecast.castMedia({
      mediaId,
      title,
      posterUrl,
      startTime: currentTime,
    })

    // Pause local playback when casting starts
    if (player?.playing) {
      player.pause()
    }
  }, [chromecast, currentEpisode, mediaId, currentTime, player])

  // Auto-cast when Chromecast connects while on this screen
  useEffect(() => {
    if (chromecast.connected && !chromecast.casting && isPlayerReady) {
      handleCastMedia()
    }
  }, [chromecast.connected])

  // Resume local playback position when casting stops
  useEffect(() => {
    if (!chromecast.casting && chromecast.currentTime > 0 && player) {
      player.currentTime = chromecast.currentTime
    }
  }, [chromecast.casting])

  // Format time
  const formatTime = (seconds: number): string => {
    if (!seconds || isNaN(seconds) || seconds < 0) return '0:00'
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    if (hrs > 0) {
      return `${hrs}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  // Pan gesture for seeking, volume, and brightness
  const panGesture = Gesture.Pan()
    .onStart((event) => {
      if (isLocked) return
      // Store initial position
    })
    .onUpdate((event) => {
      if (isLocked) return

      const { translationX, translationY, x, absoluteX } = event

      // Horizontal swipe = seek
      if (Math.abs(translationX) > Math.abs(translationY) && Math.abs(translationX) > 30) {
        const seekDelta = Math.floor(translationX / 10) * 5
        setGestureSeek(seekDelta)
        setGestureVolume(null)
        setGestureBrightness(null)
      }
      // Vertical swipe on right side = volume
      else if (absoluteX > SCREEN_WIDTH / 2 && Math.abs(translationY) > 20) {
        setGestureSeek(null)
        setGestureBrightness(null)
        const volumeDelta = -translationY / 200
        setGestureVolume(Math.min(1, Math.max(0, volume + volumeDelta)))
      }
      // Vertical swipe on left side = brightness (visual feedback only)
      else if (absoluteX <= SCREEN_WIDTH / 2 && Math.abs(translationY) > 20) {
        setGestureSeek(null)
        setGestureVolume(null)
        // Brightness: -1 to 1 range for visual feedback
        const brightnessDelta = -translationY / 200
        setGestureBrightness(Math.min(1, Math.max(-1, brightnessDelta)))
      }
    })
    .onEnd((event) => {
      if (isLocked) return

      const { translationX, translationY, absoluteX } = event

      // Apply seek
      if (Math.abs(translationX) > 50 && Math.abs(translationX) > Math.abs(translationY)) {
        const seekSeconds = Math.floor(translationX / 20) * 10
        handleSeek(seekSeconds)
      }
      // Apply volume change
      else if (Math.abs(translationY) > 50 && absoluteX > SCREEN_WIDTH / 2) {
        if (player) {
          const volumeDelta = -translationY / 200
          player.volume = Math.min(1, Math.max(0, volume + volumeDelta))
        }
      }
      // Brightness gesture (left side) - just visual feedback, no actual control

      setGestureSeek(null)
      setGestureVolume(null)
      setGestureBrightness(null)
    })

  // Wall clock update every 30 seconds
  useEffect(() => {
    const timer = setInterval(() => setWallClock(getWallClock()), 30_000)
    return () => clearInterval(timer)
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      saveCurrentProgress()
      if (controlsTimeoutRef.current) clearTimeout(controlsTimeoutRef.current)
      if (progressSaveTimeoutRef.current) clearTimeout(progressSaveTimeoutRef.current)
      if (countdownIntervalRef.current) clearInterval(countdownIntervalRef.current)
    }
  }, [saveCurrentProgress])

  // Cinema phase: show pre-roll overlay before main playback
  if (cinemaPhase === 'cinema') {
    if (cinemaLoading) {
      return (
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#e50914" />
          <Text style={styles.loadingText}>Preparing cinema experience...</Text>
        </View>
      )
    }
    if (cinemaItems.length > 0) {
      return (
        <CinemaOverlay
          items={cinemaItems}
          onComplete={handleCinemaComplete}
          serverUrl={getServerUrl() ?? 'http://localhost:8080'}
          accessToken={accessToken}
        />
      )
    }
  }

  // Loading state
  if (isLoading) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#e50914" />
        <Text style={styles.loadingText}>Loading video...</Text>
      </View>
    )
  }

  // Error state
  if (error || !streamData?.hls_url && !streamData?.direct_url) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.errorText}>{error || 'Failed to load video'}</Text>
        <TouchableOpacity style={styles.backButton} onPress={handleBack}>
          <Text style={styles.backButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    )
  }

  // No player yet
  if (!player) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#e50914" />
        <Text style={styles.loadingText}>Preparing player...</Text>
      </View>
    )
  }

  const progressPercent = duration > 0 ? (currentTime / duration) * 100 : 0

  // Find next episode info
  const currentEpisodeIndex = episodes?.findIndex((ep) => ep.id === mediaId) ?? -1
  const nextEpisode = currentEpisodeIndex >= 0 && episodes && currentEpisodeIndex < episodes.length - 1
    ? episodes[currentEpisodeIndex + 1]
    : null

  return (
    // @ts-ignore - GestureHandlerRootView children prop type issue
    <GestureHandlerRootView style={{ flex: 1 }}>
      <GestureDetector gesture={panGesture}>
        <Pressable style={styles.container} onPress={handleTapScreen}>
          {/* Casting Overlay — replaces local video when casting */}
          {chromecast.casting ? (
            <View style={castStyles.overlay}>
              {currentEpisode?.still_path ? (
                <Image
                  source={{ uri: mediaImage(currentEpisode.still_path, 'w500') || '' }}
                  style={castStyles.poster}
                  resizeMode="contain"
                />
              ) : (
                <View style={castStyles.posterPlaceholder} />
              )}
              <Text style={castStyles.deviceText}>
                Casting to {chromecast.deviceName || 'Chromecast'}
              </Text>
              {mediaTitle ? (
                <Text style={castStyles.mediaTitle} numberOfLines={1}>
                  {mediaTitle}
                </Text>
              ) : null}
              {/* Remote playback controls */}
              <View style={castStyles.controls}>
                <TouchableOpacity
                  style={castStyles.controlButton}
                  onPress={() => chromecast.seekBy(-10)}
                >
                  <Text style={{ color: '#fff', fontSize: 16, fontWeight: '600' }}>-10</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={castStyles.playButton}
                  onPress={chromecast.togglePlayPause}
                >
                  {chromecast.isPlaying ? (
                    <Pause size={32} color="#fff" />
                  ) : (
                    <Play size={32} color="#fff" />
                  )}
                </TouchableOpacity>
                <TouchableOpacity
                  style={castStyles.controlButton}
                  onPress={() => chromecast.seekBy(10)}
                >
                  <Text style={{ color: '#fff', fontSize: 16, fontWeight: '600' }}>+10</Text>
                </TouchableOpacity>
              </View>
              {/* Remote progress bar */}
              <View style={castStyles.progress}>
                <View style={castStyles.progressTrack}>
                  <View
                    style={[
                      castStyles.progressFill,
                      {
                        width: chromecast.duration > 0
                          ? `${(chromecast.currentTime / chromecast.duration) * 100}%`
                          : '0%',
                      },
                    ]}
                  />
                </View>
                <View style={castStyles.timeRow}>
                  <Text style={castStyles.timeText}>
                    {formatTime(chromecast.currentTime)}
                  </Text>
                  <Text style={castStyles.timeText}>
                    {chromecast.duration > 0 ? formatTime(chromecast.duration) : '--:--'}
                  </Text>
                </View>
              </View>
              {/* Stop casting */}
              <TouchableOpacity style={castStyles.stopButton} onPress={chromecast.stopCasting}>
                <Text style={castStyles.stopText}>Stop Casting</Text>
              </TouchableOpacity>
            </View>
          ) : (
            /* Video Player */
            // @ts-ignore -- expo-video VideoView JSX type issue with strict mode
            <VideoView
              ref={videoViewRef as any}
              style={styles.video}
              player={player}
              nativeControls={false}
              contentFit={aspectRatio}
              allowsFullscreen={true}
              allowsPictureInPicture={true}
              startsPictureInPictureAutomatically={false}
              onFullscreenEnter={() => setIsFullscreen(true)}
              onFullscreenExit={() => setIsFullscreen(false)}
            />
          )}

          {/* Dual Subtitle Overlay */}
          <DualSubtitleOverlay
            primaryUrl={primarySubtitleVttUrl}
            secondaryUrl={secondarySubtitleVttUrl}
            currentTime={currentTime}
            visible={!!selectedSubtitle || !!secondarySubLang}
            offsetSeconds={subtitleDelay}
            appearance={{
              size: subtitleSize,
              color: subtitleColor,
              background: subtitleBackground,
            }}
          />

          {/* Gesture feedback overlays */}
          {gestureSeek !== null && (
            <View style={styles.gestureFeedback}>
              <Text style={[styles.gestureFeedbackText, { fontSize: scaledFont(24, layout.fontScale) }]}>
                {gestureSeek > 0 ? '+' : ''}{gestureSeek}s
              </Text>
            </View>
          )}
          {gestureVolume !== null && (
            <View style={styles.volumeFeedback}>
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                <Volume2 size={scaledFont(20, layout.fontScale)} color="#fff" />
                <Text style={[styles.volumeFeedbackText, { fontSize: scaledFont(20, layout.fontScale) }]}>
                  {Math.round(gestureVolume * 100)}%
                </Text>
              </View>
            </View>
          )}
          {gestureBrightness !== null && (
            <View style={styles.brightnessFeedback}>
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                <Sun size={scaledFont(20, layout.fontScale)} color="#fff" />
                <Text style={[styles.brightnessFeedbackText, { fontSize: scaledFont(20, layout.fontScale) }]}>
                  {gestureBrightness >= 0 ? '+' : ''}{Math.round(gestureBrightness * 100)}%
                </Text>
              </View>
            </View>
          )}

          {/* Buffering spinner - centered on screen */}
          {isBuffering && (
            <View style={styles.pauseIndicatorContainer}>
              <ActivityIndicator size="large" color="#fff" />
            </View>
          )}

          {/* Pause indicator - only when user actually paused (not buffering) */}
          {showControls && !isPlaying && !isBuffering && !isLocked && (
            <TouchableOpacity style={styles.pauseIndicatorContainer} activeOpacity={0.8} onPress={togglePlayPause}>
              <View style={styles.pauseIndicator}>
                <Play size={44} color="#fff" fill="#fff" style={{ marginLeft: 4 }} />
              </View>
            </TouchableOpacity>
          )}

          {/* Skip Intro/Outro Buttons */}
          {showSkipIntro && (
            <TouchableOpacity style={[styles.skipButton, layout.device !== 'phone' && { paddingHorizontal: 24, paddingVertical: 14 }]} onPress={handleSkipIntro}>
              <Text style={[styles.skipButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>Skip Intro ⏭</Text>
            </TouchableOpacity>
          )}
          {showSkipOutro && (
            <TouchableOpacity style={[styles.skipButton, layout.device !== 'phone' && { paddingHorizontal: 24, paddingVertical: 14 }]} onPress={handleSkipOutro}>
              <Text style={[styles.skipButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>Skip Outro ⏭</Text>
            </TouchableOpacity>
          )}

          {/* Video Stats Overlay */}
          {showStats && (
            <View style={[
              styles.statsOverlay,
              layout.device !== 'phone' && { top: 100, left: 24, minWidth: 280 },
            ]}>
              <Text style={[styles.statsTitle, { fontSize: scaledFont(12, layout.fontScale) }]}>Video Stats</Text>
              <View style={styles.statsDivider} />
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Bitrate:      '}</Text>
                <Text>{videoStats.bitrate > 0
                  ? videoStats.bitrate >= 1_000_000
                    ? `${(videoStats.bitrate / 1_000_000).toFixed(1)} Mbps`
                    : `${Math.round(videoStats.bitrate / 1_000)} Kbps`
                  : 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Resolution:   '}</Text>
                <Text>{videoStats.resolution || 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Frame Rate:   '}</Text>
                <Text>{videoStats.frameRate > 0
                  ? `${videoStats.frameRate % 1 === 0 ? videoStats.frameRate : videoStats.frameRate.toFixed(3)} fps`
                  : 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Codec:        '}</Text>
                <Text>{videoStats.codec || 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Buffer:       '}</Text>
                <Text>{videoStats.bufferHealth > 0 ? `${videoStats.bufferHealth.toFixed(1)}s` : 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Connection:   '}</Text>
                <Text>{videoStats.connectionQuality}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Qualities:    '}</Text>
                <Text>{videoStats.qualityLevels > 0 ? `${videoStats.qualityLevels} levels` : 'N/A'}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'ABR Switches: '}</Text>
                <Text>{videoStats.abrSwitches}</Text>
              </Text>
              <Text style={[styles.statsText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                <Text style={styles.statsLabel}>{'Audio:        '}</Text>
                <Text>{videoStats.audioInfo}</Text>
              </Text>
            </View>
          )}

          {/* Controls Overlay */}
          {showControls && !isLocked && (
            <View style={styles.controlsOverlay}>
              {/* Top bar - minimal */}
              <View style={styles.topBar}>
                <TouchableOpacity style={styles.topBackButton} onPress={handleBack}>
                  <ChevronLeft size={22} color="rgba(255,255,255,0.8)" />
                  <Text style={styles.backText}>Back</Text>
                </TouchableOpacity>
                <View style={styles.topBarRight}>
                  <CastButton />
                </View>
              </View>

              {/* Spacer for tap area */}
              <View style={styles.centerArea} />

              {/* Bottom panel */}
              <View style={styles.bottomPanel}>
                {/* Row 1: Title */}
                <Text style={[styles.bottomTitle, { fontSize: scaledFont(15, layout.fontScale) }]} numberOfLines={1}>
                  {mediaTitle}
                </Text>

                {/* Row 2: Action buttons */}
                <View style={styles.actionButtonsRow}>
                  {/* Subtitles */}
                  <TouchableOpacity
                    style={[styles.actionButton, selectedSubtitle && styles.actionButtonActive]}
                    onPress={() => { setShowSubtitleMenu(true); scheduleHideControls() }}
                  >
                    <Captions size={18} color={selectedSubtitle ? '#fff' : 'rgba(255,255,255,0.7)'} />
                  </TouchableOpacity>
                  {/* Audio */}
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={() => { setShowAudioMenu(true); scheduleHideControls() }}
                  >
                    <Music size={17} color="rgba(255,255,255,0.7)" />
                  </TouchableOpacity>
                  {/* Speed */}
                  <TouchableOpacity
                    style={[styles.actionButton, playbackSpeed !== 1 && styles.actionButtonActive]}
                    onPress={() => { setShowSpeedMenu(true); scheduleHideControls() }}
                  >
                    <Text style={[styles.speedText, playbackSpeed !== 1 && { color: '#fff' }]}>
                      {playbackSpeed === 1 ? '1×' : `${playbackSpeed}×`}
                    </Text>
                  </TouchableOpacity>
                  {/* Settings */}
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={() => { setShowAspectRatioMenu(true); scheduleHideControls() }}
                  >
                    <Settings size={17} color="rgba(255,255,255,0.7)" />
                  </TouchableOpacity>
                  {/* Next Episode */}
                  {nextEpisode && (
                    <TouchableOpacity
                      style={styles.actionButton}
                      onPress={() => handlePlayNextEpisode(nextEpisode.id)}
                    >
                      <SkipForward size={17} color="rgba(255,255,255,0.7)" />
                    </TouchableOpacity>
                  )}
                  {/* Lock */}
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={toggleLock}
                  >
                    <Lock size={17} color="rgba(255,255,255,0.7)" />
                  </TouchableOpacity>
                  {/* Fullscreen */}
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={toggleFullscreen}
                  >
                    {isFullscreen ? <Minimize2 size={17} color="rgba(255,255,255,0.7)" /> : <Maximize2 size={17} color="rgba(255,255,255,0.7)" />}
                  </TouchableOpacity>
                </View>

                {/* Row 2: Progress bar */}
                <Pressable
                  style={styles.progressBar}
                  onPress={(e) => {
                    const { locationX } = e.nativeEvent
                    const barWidth = SCREEN_WIDTH - 32
                    const seekTime = (locationX / barWidth) * duration
                    handleSeekTo(seekTime)
                  }}
                >
                  <View style={styles.progressTrack}>
                    <View style={[styles.progressFill, { width: `${progressPercent}%` }]} />
                    <View style={[styles.progressThumb, { left: `${progressPercent}%` }]} />
                  </View>
                </Pressable>

                {/* Row 3: Transport */}
                <View style={styles.transportRow}>
                  <View style={styles.transportLeft}>
                    <Text style={[styles.timeText, { fontSize: scaledFont(13, layout.fontScale) }]}>
                      {formatTime(currentTime)}
                    </Text>
                    <View style={styles.transportControls}>
                      <TouchableOpacity style={styles.seekSmallButton} onPress={() => handleSeek(-5)}>
                        <RotateCcw size={20} color="rgba(255,255,255,0.75)" />
                        <Text style={styles.seekSmallText}>5</Text>
                      </TouchableOpacity>
                      <TouchableOpacity style={styles.playSmallButton} onPress={togglePlayPause}>
                        {isPlaying || isBuffering ? (
                          <Pause size={26} color="#fff" fill="#fff" />
                        ) : (
                          <Play size={26} color="#fff" fill="#fff" style={{ marginLeft: 2 }} />
                        )}
                      </TouchableOpacity>
                      <TouchableOpacity style={styles.seekSmallButton} onPress={() => handleSeek(5)}>
                        <Text style={styles.seekSmallText}>5</Text>
                        <RotateCw size={20} color="rgba(255,255,255,0.75)" />
                      </TouchableOpacity>
                    </View>
                  </View>
                  <Text style={[styles.remainingText, { fontSize: scaledFont(13, layout.fontScale) }]}>
                    {duration > 0 ? `-${formatTime(duration - currentTime)}` : ''}
                  </Text>
                </View>
              </View>
            </View>
          )}

          {/* Lock indicator */}
          {isLocked && (
            <TouchableOpacity style={styles.lockIndicator} onPress={toggleLock}>
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                <Lock size={16} color="#fff" />
                <Text style={styles.lockIndicatorText}>Tap to unlock</Text>
              </View>
            </TouchableOpacity>
          )}

          {/* Next Episode Preview */}
          {showNextEpisode && nextEpisode && (
            <View style={styles.nextEpisodeOverlay}>
              <View style={[styles.nextEpisodeCard, layout.device !== 'phone' && { maxWidth: 500 }]}>
                <Text style={[styles.nextEpisodeLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Up Next</Text>
                <Text style={[styles.nextEpisodeTitle, { fontSize: scaledFont(18, layout.fontScale) }]} numberOfLines={1}>
                  {nextEpisode.title}
                </Text>
                <Text style={[styles.nextEpisodeCountdown, { fontSize: scaledFont(14, layout.fontScale) }]}>
                  Playing in {nextEpisodeCountdown}s
                </Text>
                <View style={styles.nextEpisodeButtons}>
                  <TouchableOpacity style={styles.nextEpisodeCancel} onPress={handleCancelNextEpisode}>
                    <Text style={[styles.nextEpisodeCancelText, { fontSize: scaledFont(16, layout.fontScale) }]}>Cancel</Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    style={styles.nextEpisodePlay}
                    onPress={() => handlePlayNextEpisode(nextEpisode.id)}
                  >
                    <Text style={[styles.nextEpisodePlayText, { fontSize: scaledFont(16, layout.fontScale) }]}>Play Now</Text>
                  </TouchableOpacity>
                </View>
              </View>
            </View>
          )}
        </Pressable>
      </GestureDetector>

      {/* Subtitle Picker Modal — matches web SubtitlePicker */}
      <Modal
        visible={showSubtitleMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowSubtitleMenu(false)}
      >
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1} onPress={() => setShowSubtitleMenu(false)}>
          <View style={[styles.subPickerContainer, { width: SCREEN_WIDTH * (layout.device === 'phone' ? 0.85 : 0.55) }]} onStartShouldSetResponder={() => true}>
            {/* Header */}
            <View style={styles.subPickerHeader}>
              <Text style={styles.subPickerHeaderText}>Subtitles</Text>
            </View>

            <ScrollView style={{ maxHeight: SCREEN_HEIGHT * 0.6 }} showsVerticalScrollIndicator={false}>
              {/* Primary subtitle list */}
              <TouchableOpacity
                style={styles.subRow}
                onPress={() => { handleSubtitleSelect(null); setShowSubtitleMenu(false) }}
              >
                <Captions size={18} color={!selectedSubtitle ? '#fff' : 'rgba(255,255,255,0.4)'} />
                <View style={styles.subRowContent}>
                  <Text style={[styles.subRowName, !selectedSubtitle && { color: '#fff' }]}>Off</Text>
                </View>
                {!selectedSubtitle && <Text style={styles.subRowCheck}>✓</Text>}
              </TouchableOpacity>

              {buildVisibleSubtitles(backendSubtitleTracks, false).map((track) => {
                const { name, fmt } = parseSubtitleLabel(track.label, track.language)
                const isSelected = languageMatches(primaryBackendSub?.language ?? null, track.language)
                return (
                  <TouchableOpacity
                    key={track.id}
                    style={styles.subRow}
                    onPress={() => {
                      const nativeTrack = availableSubtitles.find((t) => t.language === track.language)
                      if (nativeTrack) handleSubtitleSelect(nativeTrack)
                      setShowSubtitleMenu(false)
                    }}
                  >
                    <Captions size={18} color={isSelected ? '#fff' : 'rgba(255,255,255,0.4)'} />
                    <View style={styles.subRowContent}>
                      <Text style={[styles.subRowName, isSelected && { color: '#fff' }]}>{name}</Text>
                      {(fmt || track.format) ? <Text style={styles.subRowFmt}>{fmt || track.format}</Text> : null}
                    </View>
                    {isSelected && <Text style={styles.subRowCheck}>✓</Text>}
                  </TouchableOpacity>
                )
              })}

              {/* Secondary subtitle section */}
              <View style={styles.subSectionDivider}>
                <Text style={styles.subSectionLabel}>Secondary subtitle</Text>
              </View>

              <TouchableOpacity
                style={styles.subRow}
                onPress={() => { setSecondarySubtitleLanguage(null); setShowSubtitleMenu(false) }}
              >
                <Captions size={18} color={!secondarySubLang ? '#fff' : 'rgba(255,255,255,0.4)'} />
                <View style={styles.subRowContent}>
                  <Text style={[styles.subRowName, !secondarySubLang && { color: '#fff' }]}>Off</Text>
                </View>
                {!secondarySubLang && <Text style={styles.subRowCheck}>✓</Text>}
              </TouchableOpacity>

              {buildVisibleSubtitles(backendSubtitleTracks, false)
                .filter((s) => !s.is_image)
                .map((track) => {
                  const { name, fmt } = parseSubtitleLabel(track.label, track.language)
                  const isSelected = languageMatches(secondarySubLang, track.language)
                  return (
                    <TouchableOpacity
                      key={`sec-${track.id}`}
                      style={styles.subRow}
                      onPress={() => { setSecondarySubtitleLanguage(track.language, track.id); setShowSubtitleMenu(false) }}
                    >
                      <Captions size={18} color={isSelected ? '#fff' : 'rgba(255,255,255,0.4)'} />
                      <View style={styles.subRowContent}>
                        <Text style={[styles.subRowName, isSelected && { color: '#fff' }]}>{name}</Text>
                        {(fmt || track.format) ? <Text style={styles.subRowFmt}>{fmt || track.format}</Text> : null}
                      </View>
                      {isSelected && <Text style={styles.subRowCheck}>✓</Text>}
                    </TouchableOpacity>
                  )
                })}

              {backendSubtitleTracks.length === 0 && (
                <View style={{ padding: 20, alignItems: 'center' }}>
                  <Text style={styles.noSubtitlesText}>No subtitles available</Text>
                </View>
              )}
            </ScrollView>
          </View>
        </TouchableOpacity>
      </Modal>

      {/* Speed Selection Modal */}
      <Modal
        visible={showSpeedMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowSpeedMenu(false)}
      >
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1} onPress={() => setShowSpeedMenu(false)}>
          <View style={[styles.menuContainer, { width: SCREEN_WIDTH * (layout.device === 'phone' ? 0.7 : 0.45), maxHeight: SCREEN_HEIGHT * 0.5 }]}>
            <Text style={styles.menuTitle}>Playback Speed</Text>
            <ScrollView style={[styles.menuList, { maxHeight: SCREEN_HEIGHT * 0.35 }]}>
              {speedOptions.map((option) => (
                <TouchableOpacity
                  key={option.value}
                  style={[styles.menuItem, playbackSpeed === option.value && styles.menuItemSelected]}
                  onPress={() => handleSpeedSelect(option.value)}
                >
                  <Text style={[styles.menuItemText, playbackSpeed === option.value && styles.menuItemTextSelected]}>
                    {option.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>
        </TouchableOpacity>
      </Modal>

      {/* Audio Track Selection Modal — uses backend tracks (like web) */}
      <Modal
        visible={showAudioMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowAudioMenu(false)}
      >
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1} onPress={() => setShowAudioMenu(false)}>
          <View style={[styles.menuContainer, { width: SCREEN_WIDTH * (layout.device === 'phone' ? 0.8 : 0.5), maxHeight: SCREEN_HEIGHT * 0.6 }]}>
            <Text style={styles.menuTitle}>Audio Track</Text>
            <ScrollView style={[styles.menuList, { maxHeight: SCREEN_HEIGHT * 0.45 }]}>
              {(playbackInfo?.audio_tracks ?? []).map((track) => (
                <TouchableOpacity
                  key={track.id}
                  style={[styles.menuItem, track.selected && styles.menuItemSelected]}
                  onPress={() => {
                    handleAudioSelect(String(track.id))
                    setShowAudioMenu(false)
                  }}
                >
                  <Text style={[styles.menuItemText, track.selected && styles.menuItemTextSelected]}>
                    {track.label || track.language}
                    {track.channels ? ` · ${track.channels}ch` : ''}
                    {track.codec ? ` · ${track.codec}` : ''}
                  </Text>
                </TouchableOpacity>
              ))}
              {(playbackInfo?.audio_tracks ?? []).length === 0 && (
                <Text style={styles.noSubtitlesText}>No audio tracks available</Text>
              )}
            </ScrollView>
          </View>
        </TouchableOpacity>
      </Modal>

      {/* Unified Settings Modal (like web Settings gear) */}
      <Modal
        visible={showAspectRatioMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowAspectRatioMenu(false)}
      >
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1} onPress={() => setShowAspectRatioMenu(false)}>
          <View style={[styles.settingsContainer, { width: SCREEN_WIDTH * (layout.device === 'phone' ? 0.75 : 0.5) }]} onStartShouldSetResponder={() => true}>
            <ScrollView showsVerticalScrollIndicator={false} style={{ maxHeight: SCREEN_HEIGHT * 0.7 }}>
              {/* Aspect Ratio */}
              <Text style={styles.settingsSectionTitle}>Aspect Ratio</Text>
              <View style={styles.settingsGrid}>
                {(['contain', 'cover', 'fill'] as const).map((r) => (
                  <TouchableOpacity
                    key={r}
                    style={[styles.settingsGridItem, aspectRatio === r && styles.settingsGridItemActive]}
                    onPress={() => setAspectRatio(r)}
                  >
                    <Text style={[styles.settingsGridText, aspectRatio === r && styles.settingsGridTextActive]}>
                      {r === 'contain' ? 'Auto' : r.charAt(0).toUpperCase() + r.slice(1)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>

              {/* Subtitle Appearance */}
              <View style={styles.settingsDivider} />
              <Text style={styles.settingsSectionTitle}>Subtitles</Text>
              <View style={styles.settingsGrid}>
                {(['small', 'medium', 'large'] as const).map((s) => (
                  <TouchableOpacity
                    key={s}
                    style={[styles.settingsGridItem, subtitleSize === s && styles.settingsGridItemActive]}
                    onPress={() => setSubtitleSize(s)}
                  >
                    <Text style={[styles.settingsGridText, subtitleSize === s && styles.settingsGridTextActive]}>
                      {s === 'small' ? 'S' : s === 'medium' ? 'M' : 'L'}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
              <View style={[styles.settingsGrid, { marginTop: 6 }]}>
                {([
                  { value: 'none' as const, label: 'None' },
                  { value: 'semi' as const, label: 'Semi' },
                  { value: 'solid' as const, label: 'Solid' },
                ]).map(({ value, label }) => (
                  <TouchableOpacity
                    key={value}
                    style={[styles.settingsGridItem, subtitleBackground === value && styles.settingsGridItemActive]}
                    onPress={() => setSubtitleBackground(value)}
                  >
                    <Text style={[styles.settingsGridText, subtitleBackground === value && styles.settingsGridTextActive]}>
                      {label}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
              {/* Color picker */}
              <View style={styles.colorRow}>
                {['#ffffff', '#fde047', '#4ade80', '#60a5fa'].map((c) => (
                  <TouchableOpacity
                    key={c}
                    onPress={() => setSubtitleColor(c)}
                    style={[styles.colorDot, { backgroundColor: c }, subtitleColor === c && styles.colorDotActive]}
                  />
                ))}
              </View>
              {/* Subtitle Delay */}
              <View style={styles.settingsDelayRow}>
                <Text style={styles.settingsDelayLabel}>Delay</Text>
                <Text style={styles.settingsDelayValue}>
                  {subtitleDelay > 0 ? '+' : ''}{subtitleDelay.toFixed(2)}s
                </Text>
              </View>
              <View style={styles.settingsGrid}>
                <TouchableOpacity
                  style={styles.settingsGridItem}
                  onPress={() => handleSubtitleDelayAdjust(-0.25)}
                >
                  <Text style={styles.settingsGridText}>-0.25s</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.settingsGridItem}
                  onPress={handleSubtitleDelayReset}
                >
                  <Text style={styles.settingsGridText}>Reset</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.settingsGridItem}
                  onPress={() => handleSubtitleDelayAdjust(0.25)}
                >
                  <Text style={styles.settingsGridText}>+0.25s</Text>
                </TouchableOpacity>
              </View>

              {/* Quality (HLS only) */}
              {streamData?.hls_url && (
                <>
                  <View style={styles.settingsDivider} />
                  <Text style={styles.settingsSectionTitle}>Quality</Text>
                  <View style={{ gap: 4 }}>
                    {qualityOptions.map((q) => (
                      <TouchableOpacity
                        key={q.label}
                        style={[styles.settingsRow, selectedQuality === q.label && styles.settingsRowActive]}
                        onPress={() => handleQualitySelect(q.label)}
                      >
                        <Text style={[styles.settingsRowText, selectedQuality === q.label && { color: '#fff', fontWeight: '600' }]}>
                          {q.label}
                        </Text>
                      </TouchableOpacity>
                    ))}
                  </View>
                </>
              )}

              {/* Repeat Mode */}
              <View style={styles.settingsDivider} />
              <Text style={styles.settingsSectionTitle}>Repeat</Text>
              <View style={styles.settingsGrid}>
                {(['off', 'one', 'all'] as const).map((m) => (
                  <TouchableOpacity
                    key={m}
                    style={[styles.settingsGridItem, repeatMode === m && styles.settingsGridItemActive]}
                    onPress={() => setRepeatMode(m)}
                  >
                    <Text style={[styles.settingsGridText, repeatMode === m && styles.settingsGridTextActive]}>
                      {m.charAt(0).toUpperCase() + m.slice(1)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>

              {/* Playback Info */}
              <View style={styles.settingsDivider} />
              <TouchableOpacity
                style={styles.settingsRow}
                onPress={() => { setShowAspectRatioMenu(false); toggleStats() }}
              >
                <Text style={styles.settingsRowText}>Playback Info</Text>
              </TouchableOpacity>
            </ScrollView>
          </View>
        </TouchableOpacity>
      </Modal>
    </GestureHandlerRootView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#000',
  },
  loadingText: {
    color: '#fff',
    marginTop: 16,
    fontSize: 16,
  },
  errorText: {
    color: '#e50914',
    fontSize: 16,
    textAlign: 'center',
    padding: 20,
  },
  backButton: {
    marginTop: 20,
    paddingHorizontal: 20,
    paddingVertical: 10,
    backgroundColor: '#333',
    borderRadius: 8,
  },
  backButtonText: {
    color: '#fff',
    fontSize: 14,
  },
  video: {
    flex: 1,
  },
  controlsOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'space-between',
  },
  topBar: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 48,
    paddingBottom: 16,
    backgroundColor: 'rgba(0,0,0,0.5)',
  },
  topBackButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  backText: {
    color: 'rgba(255,255,255,0.8)',
    fontSize: 14,
    fontWeight: '500',
  },
  topBarRight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  centerArea: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  pauseIndicatorContainer: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
  },
  pauseIndicator: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: 'rgba(0,0,0,0.3)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  bottomPanel: {
    paddingHorizontal: 16,
    paddingBottom: 34,
    paddingTop: 16,
    backgroundColor: 'rgba(0,0,0,0.85)',
  },
  bottomTitle: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '700',
    marginBottom: 10,
  },
  actionButtonsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'flex-end',
    gap: 8,
    marginBottom: 10,
  },
  actionButton: {
    width: 36,
    height: 36,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.3)',
    backgroundColor: 'rgba(255,255,255,0.05)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  actionButtonActive: {
    borderColor: '#fff',
    backgroundColor: 'rgba(255,255,255,0.2)',
  },
  speedText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 12,
    fontWeight: '700',
    fontVariant: ['tabular-nums'],
  },
  progressBar: {
    height: 32,
    justifyContent: 'center',
    marginBottom: 4,
  },
  progressTrack: {
    height: 3,
    backgroundColor: 'rgba(255,255,255,0.2)',
    borderRadius: 2,
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#fff',
    borderRadius: 2,
  },
  progressThumb: {
    position: 'absolute',
    top: -5,
    width: 13,
    height: 13,
    borderRadius: 7,
    backgroundColor: '#fff',
    marginLeft: -6,
  },
  transportRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  transportLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  transportControls: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  seekSmallButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 2,
    padding: 8,
  },
  seekSmallText: {
    color: 'rgba(255,255,255,0.75)',
    fontSize: 11,
    fontWeight: '700',
  },
  playSmallButton: {
    width: 44,
    height: 44,
    justifyContent: 'center',
    alignItems: 'center',
  },
  timeText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 13,
    fontVariant: ['tabular-nums'],
  },
  remainingText: {
    color: 'rgba(255,255,255,0.5)',
    fontSize: 13,
    fontVariant: ['tabular-nums'],
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  menuContainer: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 20,
    maxWidth: 500,
  },
  menuTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 16,
    textAlign: 'center',
  },
  menuList: {
    maxHeight: 300,
  },
  menuItem: {
    paddingVertical: 14,
    paddingHorizontal: 16,
    borderRadius: 8,
    marginBottom: 4,
  },
  menuItemSelected: {
    backgroundColor: '#e50914',
  },
  menuItemText: {
    color: '#fff',
    fontSize: 16,
    textAlign: 'center',
  },
  menuItemTextSelected: {
    fontWeight: '600',
  },
  noSubtitlesText: {
    color: '#888',
    fontSize: 14,
    textAlign: 'center',
    padding: 20,
  },
  gestureFeedback: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: [{ translateX: -50 }, { translateY: -50 }],
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  gestureFeedbackText: {
    color: '#fff',
    fontSize: 24,
    fontWeight: '600',
  },
  volumeFeedback: {
    position: 'absolute',
    right: 20,
    top: '40%',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
  },
  volumeFeedbackText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
  },
  brightnessFeedback: {
    position: 'absolute',
    left: 20,
    top: '40%',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
  },
  brightnessFeedbackText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
  },
  skipButton: {
    position: 'absolute',
    bottom: 120,
    right: 20,
    backgroundColor: 'rgba(229, 9, 20, 0.9)',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
  },
  skipButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  statsOverlay: {
    position: 'absolute',
    top: 80,
    left: 16,
    backgroundColor: 'rgba(0, 0, 0, 0.75)',
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderRadius: 8,
    minWidth: 220,
  },
  statsTitle: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '700',
    fontFamily: 'monospace',
    marginBottom: 4,
  },
  statsDivider: {
    height: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.25)',
    marginBottom: 6,
  },
  statsLabel: {
    color: 'rgba(255, 255, 255, 0.6)',
    fontFamily: 'monospace',
  },
  statsText: {
    color: '#fff',
    fontSize: 11,
    fontFamily: 'monospace',
    marginBottom: 2,
    lineHeight: 16,
  },
  lockIndicator: {
    position: 'absolute',
    bottom: 100,
    alignSelf: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
  },
  lockIndicatorText: {
    color: '#fff',
    fontSize: 14,
  },
  nextEpisodeOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.9)',
    padding: 20,
    paddingBottom: 40,
  },
  nextEpisodeCard: {
    alignItems: 'center',
  },
  nextEpisodeLabel: {
    color: '#e50914',
    fontSize: 12,
    fontWeight: '600',
    marginBottom: 4,
  },
  nextEpisodeTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 8,
    textAlign: 'center',
  },
  nextEpisodeCountdown: {
    color: '#888',
    fontSize: 14,
    marginBottom: 16,
  },
  nextEpisodeButtons: {
    flexDirection: 'row',
    gap: 16,
  },
  nextEpisodeCancel: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: '#333',
  },
  nextEpisodeCancelText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  nextEpisodePlay: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: '#e50914',
  },
  nextEpisodePlayText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  delayControls: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    gap: 16,
    marginBottom: 16,
  },
  delayButton: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    backgroundColor: '#333',
    borderRadius: 8,
    minWidth: 80,
  },
  delayButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
    textAlign: 'center',
  },
  delayValueContainer: {
    minWidth: 80,
    paddingVertical: 10,
    alignItems: 'center',
  },
  delayValueText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  // Subtitle picker styles (matches web SubtitlePicker)
  subPickerContainer: {
    backgroundColor: '#242424',
    borderRadius: 16,
    overflow: 'hidden',
    maxWidth: 500,
  },
  subPickerHeader: {
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.1)',
    paddingHorizontal: 16,
    paddingVertical: 12,
    alignItems: 'center',
  },
  subPickerHeaderText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  subRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingHorizontal: 16,
    paddingVertical: 10,
  },
  subRowContent: {
    flex: 1,
    minWidth: 0,
  },
  subRowName: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 14,
    fontWeight: '500',
  },
  subRowFmt: {
    color: 'rgba(255,255,255,0.4)',
    fontSize: 12,
    marginTop: 1,
  },
  subRowCheck: {
    color: '#fff',
    fontSize: 14,
    width: 16,
    textAlign: 'center',
  },
  subSectionDivider: {
    borderTopWidth: 1,
    borderTopColor: 'rgba(255,255,255,0.1)',
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  subSectionLabel: {
    color: 'rgba(255,255,255,0.4)',
    fontSize: 10,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  // Unified settings modal styles
  settingsContainer: {
    backgroundColor: '#1e1e1e',
    borderRadius: 16,
    padding: 16,
    maxWidth: 500,
  },
  settingsSectionTitle: {
    color: 'rgba(255,255,255,0.4)',
    fontSize: 10,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: 8,
    paddingHorizontal: 4,
  },
  settingsGrid: {
    flexDirection: 'row',
    gap: 6,
  },
  settingsGridItem: {
    flex: 1,
    paddingVertical: 8,
    borderRadius: 8,
    backgroundColor: 'rgba(255,255,255,0.05)',
    alignItems: 'center',
  },
  settingsGridItemActive: {
    backgroundColor: 'rgba(255,255,255,0.2)',
  },
  settingsGridText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 12,
    fontWeight: '500',
  },
  settingsGridTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  settingsDivider: {
    height: 1,
    backgroundColor: 'rgba(255,255,255,0.1)',
    marginVertical: 12,
  },
  settingsRow: {
    paddingVertical: 8,
    paddingHorizontal: 12,
    borderRadius: 8,
  },
  settingsRowActive: {
    backgroundColor: 'rgba(255,255,255,0.1)',
  },
  settingsRowText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 13,
  },
  colorRow: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 8,
    paddingHorizontal: 4,
  },
  colorDot: {
    width: 22,
    height: 22,
    borderRadius: 11,
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.2)',
  },
  colorDotActive: {
    borderColor: '#fff',
  },
  settingsDelayRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 10,
    marginBottom: 6,
    paddingHorizontal: 4,
  },
  settingsDelayLabel: {
    color: 'rgba(255,255,255,0.65)',
    fontSize: 11,
  },
  settingsDelayValue: {
    color: 'rgba(255,255,255,0.85)',
    fontSize: 11,
    fontWeight: '600',
    fontVariant: ['tabular-nums'],
  },
})

// Chromecast casting overlay styles
const castStyles = StyleSheet.create({
  overlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: '#141414',
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  poster: {
    width: 200,
    height: 300,
    borderRadius: 8,
    marginBottom: 24,
  },
  posterPlaceholder: {
    width: 200,
    height: 300,
    borderRadius: 8,
    marginBottom: 24,
    backgroundColor: '#333',
  },
  deviceText: {
    color: '#4a9eff',
    fontSize: 14,
    fontWeight: '500',
    marginBottom: 4,
  },
  mediaTitle: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 32,
    textAlign: 'center',
  },
  controls: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    gap: 36,
    marginBottom: 32,
  },
  controlButton: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: 'rgba(255, 255, 255, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  playButton: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: 'rgba(229, 9, 20, 0.9)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  progress: {
    width: '100%',
    maxWidth: 400,
    marginBottom: 24,
  },
  progressTrack: {
    height: 4,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 2,
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#e50914',
    borderRadius: 2,
  },
  timeRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 6,
  },
  timeText: {
    color: '#aaa',
    fontSize: 12,
  },
  stopButton: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: '#333',
  },
  stopText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
})
