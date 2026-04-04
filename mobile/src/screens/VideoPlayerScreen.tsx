import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
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
  Animated as RNAnimated,
} from 'react-native'
import { Volume2, Sun, X, ChevronLeft, ChevronRight, Settings, Pause, Play, Lock, RotateCcw, RotateCw, SkipForward, Music, Maximize2, Minimize2, Captions } from 'lucide-react-native'
import * as ScreenOrientation from 'expo-screen-orientation'
import { CastButton } from '../components/CastButton'
import { useChromecast } from '../hooks/useChromecast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'
import { useRoute, RouteProp, useNavigation } from '@react-navigation/native'
import { useVeloxPlayer, VeloxPlayerView } from '../../modules/velox-player'
import type { VideoPlayerStatus } from '../../modules/velox-player'

import { GestureDetector, Gesture, GestureHandlerRootView } from 'react-native-gesture-handler'

import type { SubtitleTrack } from '../../modules/velox-player'
import { useStreamUrls, useEpisodes, usePlaybackInfo, useMediaWithFiles } from '@velox/shared/hooks'
import { useTranslateSubtitle, useSubtitleSearch, useDownloadSubtitle } from '@velox/shared/hooks/media/useSubtitleOps'
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
type RepeatMode = 'none' | 'one' | 'all'

const CONTROLS_HIDE_TIMEOUT = 3000
const STATE_POLL_INTERVAL = 500

/** Current wall clock as HH:MM string. */
function getWallClock(): string {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

import type { QualityOption } from '@velox/shared/types/playback'

function SeekFeedbackOverlay({ side, amount }: { side: 'left' | 'right'; amount: number }) {
  const anim1 = useRef(new RNAnimated.Value(0)).current
  const anim2 = useRef(new RNAnimated.Value(0)).current
  const anim3 = useRef(new RNAnimated.Value(0)).current

  useEffect(() => {
    // Staggered arrow animation — direction matches seek direction
    const anims = side === 'right' ? [anim1, anim2, anim3] : [anim3, anim2, anim1]
    anims.forEach((a) => a.setValue(0))
    RNAnimated.stagger(100, anims.map((a) =>
      RNAnimated.sequence([
        RNAnimated.timing(a, { toValue: 1, duration: 250, useNativeDriver: true }),
        RNAnimated.timing(a, { toValue: 0.3, duration: 250, useNativeDriver: true }),
      ]),
    )).start()
  }, [amount, side]) // Re-animate on each tap

  const Arrow = side === 'right' ? ChevronRight : ChevronLeft

  return (
    <View style={[
      seekStyles.container,
      side === 'left' ? { left: 80 } : { right: 80 },
    ]}>
      <View style={seekStyles.arrowRow}>
        <RNAnimated.View style={{ opacity: anim1 }}>
          <Arrow size={24} color="#fff" />
        </RNAnimated.View>
        <RNAnimated.View style={{ opacity: anim2, marginLeft: -10 }}>
          <Arrow size={24} color="#fff" />
        </RNAnimated.View>
        <RNAnimated.View style={{ opacity: anim3, marginLeft: -10 }}>
          <Arrow size={24} color="#fff" />
        </RNAnimated.View>
      </View>
      <Text style={seekStyles.text}>{side === 'left' ? '-' : '+'}{amount}s</Text>
    </View>
  )
}

const seekStyles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    justifyContent: 'center',
    alignItems: 'center',
    pointerEvents: 'none',
  },
  arrowRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 2,
  },
  text: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '700',
    textShadowColor: 'rgba(0,0,0,0.7)',
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 4,
  },
})

const TRANSLATE_LANGS = [
  { code: 'vi', label: 'Vietnamese' },
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'French' },
  { code: 'de', label: 'German' },
  { code: 'es', label: 'Spanish' },
  { code: 'ja', label: 'Japanese' },
  { code: 'ko', label: 'Korean' },
  { code: 'zh', label: 'Chinese' },
  { code: 'pt', label: 'Portuguese' },
  { code: 'ru', label: 'Russian' },
  { code: 'th', label: 'Thai' },
]

function SubtitleTranslateSection({ subtitles, onTranslated }: {
  subtitles: { id: number; label: string; language: string; format: string }[]
  onTranslated: () => void
}) {
  const [open, setOpen] = useState(false)
  const [sourceId, setSourceId] = useState<number | null>(null)
  const [targetLang, setTargetLang] = useState<string | null>(null)
  const translateMutation = useTranslateSubtitle()

  if (!open) {
    return (
      <TouchableOpacity style={stl.subBottomRow} onPress={() => setOpen(true)}>
        <Text style={stl.subBottomIcon}>✦</Text>
        <Text style={stl.subBottomText}>Translate Subtitle</Text>
      </TouchableOpacity>
    )
  }

  const effectiveSourceId = sourceId ?? subtitles[0]?.id
  return (
    <View style={stl.translateSection}>
      <Text style={stl.translateLabel}>Translate subtitle</Text>

      {subtitles.length > 1 && (
        <View style={stl.translatePickerRow}>
          {subtitles.map((s) => (
            <TouchableOpacity
              key={s.id}
              style={[stl.translateChip, effectiveSourceId === s.id && stl.translateChipActive]}
              onPress={() => setSourceId(s.id)}
            >
              <Text style={[stl.translateChipText, effectiveSourceId === s.id && { color: '#fff' }]} numberOfLines={1}>
                {s.label || s.language} ({s.format})
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      )}

      <ScrollView horizontal showsHorizontalScrollIndicator={false} style={{ marginTop: 8 }}>
        {TRANSLATE_LANGS.map((l) => (
          <TouchableOpacity
            key={l.code}
            style={[stl.translateChip, targetLang === l.code && stl.translateChipActive]}
            onPress={() => setTargetLang(l.code)}
          >
            <Text style={[stl.translateChipText, targetLang === l.code && { color: '#fff' }]}>{l.label}</Text>
          </TouchableOpacity>
        ))}
      </ScrollView>

      <View style={stl.translateActions}>
        <TouchableOpacity
          style={[stl.translateBtn, !targetLang && { opacity: 0.5 }]}
          disabled={!targetLang || translateMutation.isPending}
          onPress={() => {
            if (!effectiveSourceId || !targetLang) return
            translateMutation.mutate(
              { subtitleId: effectiveSourceId, targetLanguage: targetLang },
              { onSuccess: () => { onTranslated(); setOpen(false); setTargetLang(null) } },
            )
          }}
        >
          <Text style={stl.translateBtnText}>
            {translateMutation.isPending ? 'Translating...' : 'Translate'}
          </Text>
        </TouchableOpacity>
        <TouchableOpacity style={stl.translateCancelBtn} onPress={() => setOpen(false)}>
          <Text style={stl.translateCancelText}>Cancel</Text>
        </TouchableOpacity>
      </View>

      {translateMutation.isError && (
        <Text style={{ color: '#f87171', fontSize: 12, marginTop: 4 }}>
          {(translateMutation.error as Error)?.message || 'Translation failed'}
        </Text>
      )}
    </View>
  )
}

const SEARCH_LANGS = [
  { code: 'en', name: 'English' },
  { code: 'vi', name: 'Vietnamese' },
  { code: 'fr', name: 'French' },
  { code: 'de', name: 'German' },
  { code: 'es', name: 'Spanish' },
  { code: 'ja', name: 'Japanese' },
  { code: 'ko', name: 'Korean' },
  { code: 'zh', name: 'Chinese' },
]

function SubtitleSearchSection({ mediaId, defaultLang, onDownloaded }: {
  mediaId: number
  defaultLang: string | null
  onDownloaded: () => void
}) {
  const [open, setOpen] = useState(false)
  const [lang, setLang] = useState(defaultLang || 'en')
  const { data: results, isLoading } = useSubtitleSearch(mediaId, lang, open)
  const downloadMutation = useDownloadSubtitle(mediaId)

  if (!open) {
    return (
      <TouchableOpacity style={stl.subBottomRow} onPress={() => setOpen(true)}>
        <Text style={stl.subBottomIcon}>⌕</Text>
        <Text style={stl.subBottomText}>Search for Subtitles</Text>
      </TouchableOpacity>
    )
  }

  return (
    <View style={stl.translateSection}>
      <Text style={stl.translateLabel}>Search for Subtitles</Text>

      <ScrollView horizontal showsHorizontalScrollIndicator={false} style={{ marginBottom: 8 }}>
        {SEARCH_LANGS.map((l) => (
          <TouchableOpacity
            key={l.code}
            style={[stl.translateChip, lang === l.code && stl.translateChipActive]}
            onPress={() => setLang(l.code)}
          >
            <Text style={[stl.translateChipText, lang === l.code && { color: '#fff' }]}>{l.name}</Text>
          </TouchableOpacity>
        ))}
      </ScrollView>

      {isLoading && <ActivityIndicator size="small" color="#e50914" style={{ marginVertical: 12 }} />}

      {results && results.length > 0 && (
        <ScrollView style={{ maxHeight: 200 }}>
          {results.map((r, i) => (
            <TouchableOpacity
              key={i}
              style={stl.searchResultRow}
              onPress={() => {
                downloadMutation.mutate(
                  { provider: r.provider, external_id: r.external_id, language: r.language },
                  { onSuccess: () => onDownloaded() },
                )
              }}
            >
              <View style={{ flex: 1 }}>
                <Text style={{ color: '#fff', fontSize: 13 }} numberOfLines={1}>{r.title}</Text>
                <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 11 }}>{r.provider} • {r.language} • {r.format}</Text>
              </View>
              {downloadMutation.isPending ? (
                <ActivityIndicator size="small" color="#fff" />
              ) : (
                <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 16 }}>↓</Text>
              )}
            </TouchableOpacity>
          ))}
        </ScrollView>
      )}

      {results && results.length === 0 && !isLoading && (
        <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 13, textAlign: 'center', paddingVertical: 12 }}>
          No subtitles found
        </Text>
      )}

      <TouchableOpacity style={[stl.translateCancelBtn, { marginTop: 8 }]} onPress={() => setOpen(false)}>
        <Text style={stl.translateCancelText}>Close</Text>
      </TouchableOpacity>
    </View>
  )
}

// Inline styles for translate/search sections
const stl = StyleSheet.create({
  subBottomRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  subBottomIcon: {
    fontSize: 18,
    color: 'rgba(255,255,255,0.5)',
    width: 18,
    textAlign: 'center',
  },
  subBottomText: {
    fontSize: 14,
    color: 'rgba(255,255,255,0.5)',
  },
  translateSection: {
    borderTopWidth: 1,
    borderTopColor: 'rgba(255,255,255,0.1)',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  translateLabel: {
    color: 'rgba(255,255,255,0.4)',
    fontSize: 10,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: 8,
  },
  translatePickerRow: {
    gap: 6,
  },
  translateChip: {
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 8,
    marginRight: 6,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  translateChipActive: {
    borderColor: 'rgba(255,255,255,0.3)',
    backgroundColor: 'rgba(255,255,255,0.12)',
  },
  translateChipText: {
    fontSize: 12,
    color: 'rgba(255,255,255,0.6)',
  },
  translateActions: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 10,
  },
  translateBtn: {
    flex: 1,
    backgroundColor: '#2563eb',
    borderRadius: 8,
    paddingVertical: 10,
    alignItems: 'center',
  },
  translateBtnText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  translateCancelBtn: {
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 8,
    paddingVertical: 10,
    paddingHorizontal: 16,
    alignItems: 'center',
  },
  translateCancelText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 13,
  },
  searchResultRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.06)',
  },
})

function SubRow({ name, fmt, selected, onPress }: {
  name: string
  fmt?: string
  selected: boolean
  onPress: () => void
}) {
  return (
    <TouchableOpacity
      style={{ flexDirection: 'row', alignItems: 'center', gap: 12, paddingHorizontal: 16, paddingVertical: 10 }}
      onPress={onPress}
    >
      <Captions size={18} color={selected ? '#fff' : 'rgba(255,255,255,0.4)'} />
      <View style={{ flex: 1, minWidth: 0 }}>
        <Text style={{ fontSize: 14, fontWeight: '500', color: selected ? '#fff' : 'rgba(255,255,255,0.7)' }} numberOfLines={1}>
          {name}
        </Text>
        {fmt ? <Text style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginTop: 1 }}>{fmt}</Text> : null}
      </View>
      <Text style={{ width: 16, textAlign: 'center', fontSize: 14, color: '#fff' }}>{selected ? '✓' : ''}</Text>
    </TouchableOpacity>
  )
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

  // Quality state (must be before playbackRequest)
  const [maxQuality, setMaxQuality] = useState<number | 'auto'>('auto')

  // Stream URL state
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  // Mobile capabilities: ExoPlayer (Android) supports h264/hevc + aac/mp3/opus/ac3/eac3
  // Cap at 1080p by default — 4K HEVC Main 10 can fail on many mobile devices
  const playbackRequest = useMemo(() => ({
    video_codecs: ['h264', 'hevc'],
    audio_codecs: ['aac', 'mp3', 'opus', 'ac3', 'eac3'],
    containers: ['mp4', 'hls'],
    ...(maxQuality !== 'auto' ? { max_height: maxQuality } : { max_height: 1080 }),
  }), [maxQuality])
  const { data: streamUrls, isLoading: streamLoading, isError: streamError } = useStreamUrls(mediaId, playbackRequest)
  const { data: progress } = useProgress(mediaId)
  const updateProgress = useUpdateProgress()

  // Media detail for title + season/series info
  const { data: mediaDetail } = useMediaWithFiles(mediaId)
  const seasonId = mediaDetail?.season_id ?? 0

  // Episode data for next episode (need correct seasonId)
  const { data: episodes } = useEpisodes(seriesId || 0, seasonId)

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


  // Video source — use server's URL directly (VeloxPlayer handles most codecs)
  // Fallback: if direct play fails (H.264 10-bit etc), auto-retry with HLS
  const fallbackRef = useRef<'direct' | 'hls'>('direct')
  const loadingStartRef = useRef<number>(0)
  const videoSource = useMemo(() => {
    if (!streamUrls) return null
    const serverUrl = getServerUrl() ?? ''
    let url = streamUrls.hls || streamUrls.direct
    if (!url) return null
    // HLS fallback: convert direct URL to HLS transcode
    if (fallbackRef.current === 'hls' && !url.includes('/hls/')) {
      const mh = maxQuality !== 'auto' ? maxQuality : 1080
      url = url.replace(/\/api\/stream\/(\d+)/, '/api/stream/$1/hls/master.m3u8')
      url += (url.includes('?') ? '&' : '?') + `mh=${mh}`
    }
    const fullUrl = url.startsWith('http') ? url : `${serverUrl}${url}`
    return accessToken
      ? fullUrl + (fullUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(accessToken)
      : fullUrl
  }, [streamUrls, accessToken, maxQuality, fallbackRef.current])

  // Player state
  const playerRef = useRef<ReturnType<typeof useVeloxPlayer> | null>(null)
  const [playerStatus, setPlayerStatus] = useState<VideoPlayerStatus>('idle')
  const [isPlaying, setIsPlaying] = useState(false)
  const [isBuffering, setIsBuffering] = useState(false)
  const userWantsToPlayRef = useRef(true) // user intent: true = wants to play
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [isLandscape, setIsLandscape] = useState(false)
  const [showControls, setShowControls] = useState(true)
  const [showSubtitleMenu, setShowSubtitleMenu] = useState(false)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)
  const [showAudioMenu, setShowAudioMenu] = useState(false)
  const [showNextEpisode, setShowNextEpisode] = useState(false)
  const [nextEpisodeCountdown, setNextEpisodeCountdown] = useState(15)
  const [isLocked, setIsLocked] = useState(false)
  const [showStats, setShowStats] = useState(false)
  const storedSubLang = usePlayerStore((s) => s.subtitleLanguage)
  const setStoredSubLang = usePlayerStore((s) => s.setSubtitleLanguage)
  const [_selectedSubtitle, _setSelectedSubtitle] = useState<string | null>(storedSubLang)
  const selectedSubtitle = _selectedSubtitle
  const setSelectedSubtitle = useCallback((lang: string | null) => {
    _setSelectedSubtitle(lang)
    setStoredSubLang(lang)
  }, [setStoredSubLang])
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
  // selectedSubtitle stores the backend subtitle language directly
  const primaryBackendSub = selectedSubtitle
    ? backendSubtitleTracks.find(
        (t) => languageMatches(t.language, selectedSubtitle) && !t.is_image,
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
  const currentEpisode = episodes?.find((ep) => ep.media_id === mediaId)
  const mediaTitle = currentEpisode?.title || mediaDetail?.media?.title || ''

  // Skip markers from playback info (backend fingerprint detection)
  const skipSegments = (playbackInfo as any)?.skip_segments ?? []
  const introSegment = skipSegments.find((s: any) => s.type === 'intro')
  const outroSegment = skipSegments.find((s: any) => s.type === 'credits')
  const introEnd = introSegment?.end ?? null
  const outroStart = outroSegment?.start ?? null
  const [showSkipIntro, setShowSkipIntro] = useState(false)
  const [showSkipOutro, setShowSkipOutro] = useState(false)


  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const progressSaveTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hasSeekedRef = useRef(false)
  const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Quality options from server (dynamic, matches web)
  const qualityOptions: QualityOption[] = playbackInfo?.available_qualities ?? []
  const currentQualityLabel = maxQuality === 'auto'
    ? 'Auto'
    : (qualityOptions.find((q) => q.height === maxQuality)?.label ?? `${maxQuality}p`)

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
    { label: 'None', value: 'none', icon: '🔁' },
    { label: 'One', value: 'one', icon: '🔂' },
    { label: 'All', value: 'all', icon: '🔁' },
  ]
  const [repeatMode, setRepeatMode] = useState<RepeatMode>('none')

  // Subtitle delay offset (stored per mediaId, in seconds)
  const [subtitleDelay, setSubtitleDelay] = useState(0)


  // Aspect ratio modal
  const [showAspectRatioMenu, setShowAspectRatioMenu] = useState(false)
  const [settingsView, setSettingsView] = useState<'main' | 'quality'>('main')

  // Double/multi tap detection (YouTube-style accumulating seek)
  const seekAccumulatorRef = useRef<number>(0)
  const seekResetTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [seekFeedback, setSeekFeedback] = useState<{ side: 'left' | 'right'; amount: number } | null>(null)

  // No auto-select needed — subtitle language is initialized from player store preference
  // User explicitly selects via subtitle picker, same as web

  // Sync loading/error state from stream query
  useEffect(() => {
    if (streamLoading) {
      setIsLoading(true)
      setError(null)
    } else if (streamError) {
      setError('Failed to load video')
      setIsLoading(false)
    } else if (streamUrls) {
      setIsLoading(false)
      console.log('[VideoPlayer] method:', playbackInfo?.method, '| reason:', playbackInfo?.decision_reason)
      console.log('[VideoPlayer] videoSource:', videoSource?.substring(0, 150))
    }
  }, [streamLoading, streamError, streamUrls, videoSource])

  // Create player with setup callback
  const player = useVeloxPlayer(videoSource, (p) => {
    playerRef.current = p
    p.timeUpdateEventInterval = 1
    p.volume = 1
    p.muted = false
  })

  // When videoSource changes (initial load or episode switch), replace and play
  const prevSourceRef = useRef<string | null>(null)
  useEffect(() => {
    if (player && videoSource && prevSourceRef.current !== videoSource) {
      player.replaceAsync({ uri: videoSource })
      player.volume = 1
      player.muted = false
      player.play()
      prevSourceRef.current = videoSource
      hasSeekedRef.current = false // allow resume seek for new episode
    }
  }, [player, videoSource])

  // Poll player state
  useEffect(() => {
    if (!player) return

    const pollState = () => {
      try {
        const status = player.status
        const playing = player.playing
        setPlayerStatus(status)
        setIsPlaying(playing)


        // Buffering = user wants to play but player is not playing (stalled/loading)
        const buffering = userWantsToPlayRef.current && !playing && status !== 'idle'
        setIsBuffering(buffering)

        // Auto-fallback: error OR stuck not playing for 5s → switch to HLS
        if ((status === 'error' || (status !== 'idle' && status !== 'readyToPlay' && !playing)) &&
            fallbackRef.current === 'direct') {
          if (!loadingStartRef.current) loadingStartRef.current = Date.now()
          if (status === 'error' || (Date.now() - loadingStartRef.current) > 5000) {
            fallbackRef.current = 'hls'
            loadingStartRef.current = 0
            const mh = maxQuality !== 'auto' ? maxQuality : 1080
            const hlsUrl = (videoSource || '').replace(/\/api\/stream\/(\d+)/, '/api/stream/$1/hls/master.m3u8')
              + (videoSource?.includes('?') ? '&' : '?') + `mh=${mh}`
            console.log('[VideoPlayer] Direct play failed, falling back to HLS')
            try {
              player.replaceAsync({ uri: hlsUrl })
              player.play()
            } catch {}
            return
          }
        } else {
          loadingStartRef.current = 0
        }

        setCurrentTime(player.currentTime)
        // Always prefer server duration (accurate) — HLS player.duration grows as segments load
        const serverDuration = playbackInfo?.duration ?? 0
        setDuration(serverDuration > 0 ? serverDuration : player.duration)
        setVolume(player.volume)
        setIsMuted(player.muted)
        // @ts-ignore - availableSubtitleTracks type mismatch with local SubtitleTrack
        setAvailableSubtitles(player.availableSubtitleTracks)
        // @ts-ignore - availableAudioTracks not in types but exists on player
        setAvailableAudioTracks(player.availableAudioTracks || [])
        // @ts-ignore - playbackRate not in types but exists
        setPlaybackSpeed(player.playbackRate || 1)

        if (showStats) {
          const track = player.videoTrack
          const bitrateRaw = track?.bitrate ?? 0
          const bufferSec =
            player.bufferedPosition >= 0 && player.currentTime >= 0
              ? Math.max(0, player.bufferedPosition - player.currentTime)
              : 0

          let connectionQuality = 'N/A'
          if (player.bufferedPosition >= 0) {
            if (bufferSec >= 10) connectionQuality = 'Excellent'
            else if (bufferSec >= 5) connectionQuality = 'Good'
            else if (bufferSec >= 2) connectionQuality = 'Poor'
            else connectionQuality = 'Buffering...'
          }

          const audio = player.audioTrack
          const audioInfo = audio ? (audio.label || audio.language || 'Unknown') : 'N/A'

          setVideoStats({
            bitrate: bitrateRaw,
            resolution:
              track ? `${track.width}\u00D7${track.height}` : '',
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
      } catch {
        // Player was released during fallback transition — skip poll
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

  // Check for next episode when in credits region or near end
  useEffect(() => {
    if (!seriesId || !episodes || episodes.length === 0) return
    if (!isPlaying || duration <= 0) return

    // Trigger "Up Next" when entering credits region, or last 30s if no credits marker
    const inCredits = outroStart && currentTime >= outroStart
    const timeRemaining = duration - currentTime
    const isNearEnd = timeRemaining > 0 && timeRemaining < 30

    if ((inCredits || isNearEnd) && !showNextEpisode) {
      const currentIndex = episodes.findIndex((ep) => ep.media_id === mediaId)
      if (currentIndex >= 0 && currentIndex < episodes.length - 1) {
        setShowNextEpisode(true)
        setNextEpisodeCountdown(15)
      }
    }
  }, [currentTime, duration, isPlaying, seriesId, episodes, mediaId, outroStart])

  // Check for intro/outro skip markers
  useEffect(() => {
    if (!isPlaying || !currentTime) return

    // Check if in intro region (use segment start/end, not just end)
    const introStart = introSegment?.start ?? 0
    if (introEnd && currentTime >= introStart && currentTime < introEnd) {
      setShowSkipIntro(true)
    } else {
      setShowSkipIntro(false)
    }

    // Check if in outro/credits region — for series, show "Up Next" instead
    // Skip Credits only for movies (no seriesId)
    if (outroStart && duration > 0 && currentTime >= outroStart && !seriesId) {
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
            handlePlayNextEpisode(nextEp.media_id)
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

  // Refs for gesture callbacks (RNGH callbacks capture stale closures)
  const showControlsRef = useRef(showControls)
  showControlsRef.current = showControls
  const isPlayingRef = useRef(isPlaying)
  isPlayingRef.current = isPlaying
  const isLockedRef = useRef(isLocked)
  isLockedRef.current = isLocked
  // Guard ref to block tap gestures while stats overlay is open (or just closed)
  const statsGuardRef = useRef(false)
  useEffect(() => {
    if (showStats) {
      statsGuardRef.current = true
    } else {
      // Keep guard up briefly after closing so gesture callback doesn't sneak through
      setTimeout(() => { statsGuardRef.current = false }, 300)
    }
  }, [showStats])

  // YouTube-style tap gestures via RNGH (no Pressable conflicts)
  // Double tap on left/right third → seek ±5s (accumulate on rapid taps)
  const doubleTapGesture = Gesture.Tap()
    .numberOfTaps(2)
    .maxDuration(250)
    .onEnd((event) => {
      if (isLockedRef.current) return
      if (statsGuardRef.current) return

      // Ignore double taps in controls area
      if (showControlsRef.current) {
        if (event.y < 80 || event.y > SCREEN_HEIGHT - 160) return
      }

      const tapX = event.x
      const zoneWidth = SCREEN_WIDTH / 3
      const isLeftZone = tapX < zoneWidth
      const isRightZone = tapX > SCREEN_WIDTH - zoneWidth

      if (isLeftZone || isRightZone) {
        const direction = isRightZone ? 5 : -5
        seekAccumulatorRef.current += direction
        handleSeek(direction)

        setSeekFeedback({
          side: isRightZone ? 'right' : 'left',
          amount: Math.abs(seekAccumulatorRef.current),
        })

        if (seekResetTimerRef.current) clearTimeout(seekResetTimerRef.current)
        seekResetTimerRef.current = setTimeout(() => {
          seekAccumulatorRef.current = 0
          setSeekFeedback(null)
        }, 800)

        // Show controls briefly during seek feedback
        setShowControls(true)
        scheduleHideControls()
      }
    })

  // Single tap behavior:
  // - Center: controls hidden → show controls; controls visible → toggle play/pause with animation
  // - Edges: toggle controls visibility
  // - Top bar / bottom panel: ignored (let TouchableOpacity buttons handle)
  const singleTapGesture = Gesture.Tap()
    .maxDuration(250)
    .onEnd((event) => {
      if (isLockedRef.current) return
      if (statsGuardRef.current) return

      const tapY = event.y
      const tapX = event.x
      const zoneWidth = SCREEN_WIDTH / 3
      const isCenterZone = tapX >= zoneWidth && tapX <= SCREEN_WIDTH - zoneWidth

      // When controls visible, ignore taps on top bar / bottom panel areas
      if (showControlsRef.current) {
        if (tapY < 80 || tapY > SCREEN_HEIGHT - 160) return
      }

      if (isCenterZone) {
        if (!showControlsRef.current) {
          // Center tap, controls hidden → show controls
          setShowControls(true)
          scheduleHideControls()
        } else {
          // Center tap, controls visible → toggle play/pause with animated icon
          togglePlayPause()
        }
      } else {
        // Edge tap → toggle controls
        if (showControlsRef.current) {
          setShowControls(false)
        } else {
          setShowControls(true)
          scheduleHideControls()
        }
      }
    })

  // Double tap takes priority over single tap
  const tapGestures = Gesture.Exclusive(doubleTapGesture, singleTapGesture)

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
  // Animated feedback for play/pause (YouTube-style scale+fade)
  const playPauseAnim = useRef(new RNAnimated.Value(0)).current
  // Lock which icon to show during animation (true = show Play icon, false = show Pause icon)
  const [animIcon, setAnimIcon] = useState<'play' | 'pause' | null>(null)

  const togglePlayPause = () => {
    if (!player) return

    const wasPlaying = player.playing || userWantsToPlayRef.current

    // Lock the icon BEFORE state changes: show what action happened
    // Was playing → now pausing → show Pause icon; was paused → now playing → show Play icon
    setAnimIcon(wasPlaying ? 'pause' : 'play')

    // Trigger scale+fade animation
    playPauseAnim.setValue(1)
    RNAnimated.timing(playPauseAnim, {
      toValue: 0,
      duration: 400,
      useNativeDriver: true,
    }).start(() => setAnimIcon(null))

    if (wasPlaying) {
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


  // Subtitle handling
  const handleSubtitleSelect = (track: SubtitleTrack | null) => {
    if (!player) return
    player.subtitleTrack = track
    setSelectedSubtitle(track?.id || null)
    setShowSubtitleMenu(false)
    scheduleHideControls()
  }

  // Quality handling — changes max_height → re-fetches stream URL → player reloads
  const handleQualitySelect = (height: number | 'auto') => {
    if (height === maxQuality) return
    setMaxQuality(height)
    // Always start from normal — fallback chain handles failures:
    // server decision (direct play) → pretranscode → HLS transcode
    scheduleHideControls()
    console.log('[VideoPlayer] Quality changed to:', height === 'auto' ? 'auto (1080p)' : `${height}p`)
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

  // Toggle orientation (landscape/portrait)
  const toggleFullscreen = async () => {
    if (isLandscape) {
      await ScreenOrientation.lockAsync(ScreenOrientation.OrientationLock.PORTRAIT_UP)
      setIsLandscape(false)
    } else {
      await ScreenOrientation.lockAsync(ScreenOrientation.OrientationLock.LANDSCAPE_RIGHT)
      setIsLandscape(true)
    }
    scheduleHideControls()
  }

  // Unlock orientation on unmount
  useEffect(() => {
    return () => {
      ScreenOrientation.unlockAsync()
    }
  }, [])

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
    const modes: RepeatMode[] = ['none', 'one', 'all']
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
        handlePlayNextEpisode(nextEpisode?.media_id ?? 0)
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
  if (error || (!streamLoading && !videoSource)) {
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
  const currentEpisodeIndex = episodes?.findIndex((ep) => ep.media_id === mediaId) ?? -1
  const nextEpisode = currentEpisodeIndex >= 0 && episodes && currentEpisodeIndex < episodes.length - 1
    ? episodes[currentEpisodeIndex + 1]
    : null

  // Compose all gestures: taps are exclusive (double wins), race with pan
  const composedGesture = Gesture.Race(panGesture, tapGestures)

  return (
    // @ts-ignore - GestureHandlerRootView children prop type issue
    <GestureHandlerRootView style={{ flex: 1 }}>
      <GestureDetector gesture={composedGesture}>
        <View style={styles.container}>
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
            <VeloxPlayerView
              style={styles.video}
              player={player}
              nativeControls={false}
              contentFit={aspectRatio}
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

          {/* Double-tap seek feedback (YouTube-style) */}
          {seekFeedback && (
            <SeekFeedbackOverlay side={seekFeedback.side} amount={seekFeedback.amount} />
          )}

          {/* Gesture feedback overlays */}
          {gestureSeek !== null && (
            <View style={styles.gestureFeedback}>
              <Text style={[styles.gestureFeedbackText, { fontSize: scaledFont(24, layout.fontScale) }]}>
                {gestureSeek > 0 ? '+' : ''}{gestureSeek}s
              </Text>
            </View>
          )}
          {gestureVolume !== null && (
            <View style={styles.volumeFeedback} pointerEvents="none">
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: 'rgba(0,0,0,0.7)', paddingHorizontal: 20, paddingVertical: 12, borderRadius: 8 }}>
                <Volume2 size={scaledFont(20, layout.fontScale)} color="#fff" />
                <Text style={[styles.volumeFeedbackText, { fontSize: scaledFont(20, layout.fontScale) }]}>
                  {Math.round(gestureVolume * 100)}%
                </Text>
              </View>
            </View>
          )}
          {gestureBrightness !== null && (
            <View style={styles.brightnessFeedback} pointerEvents="none">
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: 'rgba(0,0,0,0.7)', paddingHorizontal: 20, paddingVertical: 12, borderRadius: 8 }}>
                <Sun size={scaledFont(20, layout.fontScale)} color="#fff" />
                <Text style={[styles.brightnessFeedbackText, { fontSize: scaledFont(20, layout.fontScale) }]}>
                  {gestureBrightness >= 0 ? '+' : ''}{Math.round(gestureBrightness * 100)}%
                </Text>
              </View>
            </View>
          )}

          {/* Buffering spinner */}
          {isBuffering && (
            <View style={styles.pauseIndicatorContainer}>
              <ActivityIndicator size="large" color="#fff" />
            </View>
          )}

          {/* Skip Intro/Outro Buttons */}
          {showSkipIntro && (
            <TouchableOpacity style={[styles.skipButton, layout.device !== 'phone' && { paddingHorizontal: 24, paddingVertical: 14 }]} onPress={handleSkipIntro}>
              <Text style={[styles.skipButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>Skip Intro</Text>
              <SkipForward size={16} color="#000" />
            </TouchableOpacity>
          )}
          {showSkipOutro && (
            <TouchableOpacity style={[styles.skipButton, layout.device !== 'phone' && { paddingHorizontal: 24, paddingVertical: 14 }]} onPress={handleSkipOutro}>
              <Text style={[styles.skipButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>Skip Credits</Text>
              <SkipForward size={16} color="#000" />
            </TouchableOpacity>
          )}

          {/* Playback Info Overlay — matches web WatchPlaybackStatsOverlay */}
          {showStats && playbackInfo && (() => {
            const pi = playbackInfo as any
            const selectedAudio = pi.audio_tracks?.find((t: any) => t.selected) ?? pi.audio_tracks?.find((t: any) => t.is_default) ?? pi.audio_tracks?.[0]
            const isTranscoding = pi.method === 'FullTranscode' || pi.method === 'TranscodeAudio'
            const methodColors: Record<string, string> = { DirectPlay: '#4ade80', DirectStream: '#60a5fa', TranscodeAudio: '#facc15', FullTranscode: '#f87171' }
            const methodLabels: Record<string, string> = { DirectPlay: 'Direct Play', DirectStream: 'Direct Stream', TranscodeAudio: 'Transcode Audio', FullTranscode: 'Full Transcode' }
            const methodColor = methodColors[pi.method] ?? '#fff'
            const fmtBitrate = (b: number) => b >= 1000 ? `${(b / 1000).toFixed(1)} Mbps` : `${b} Kbps`
            const fmtChannels = (ch: number) => ch === 6 ? '5.1' : ch === 8 ? '7.1' : ch === 2 ? 'Stereo' : ch === 1 ? 'Mono' : `${ch}ch`
            return (
              <View style={[styles.statsOverlay, layout.device !== 'phone' && { top: 100, left: 24, minWidth: 280 }]}>
                {/* Close button */}
                <TouchableOpacity style={styles.statsCloseButton} onPress={() => setShowStats(false)}>
                  <X size={16} color="rgba(255,255,255,0.5)" />
                </TouchableOpacity>

                {/* Playback Method */}
                <View style={styles.statsSection}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <Text style={styles.statsTitle}>Playback</Text>
                    <View style={{ backgroundColor: methodColor + '33', paddingHorizontal: 6, paddingVertical: 2, borderRadius: 4 }}>
                      <Text style={{ color: methodColor, fontSize: 10, fontWeight: '700' }}>{methodLabels[pi.method] ?? pi.method}</Text>
                    </View>
                  </View>
                  {pi.decision_reason ? <Text style={styles.statsMonoMuted}>{pi.decision_reason}</Text> : null}
                </View>

                {/* Video */}
                <View style={styles.statsSection}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <Text style={styles.statsTitle}>Video</Text>
                    <View style={{ backgroundColor: pi.method === 'FullTranscode' ? '#f8717133' : '#4ade8033', paddingHorizontal: 6, paddingVertical: 2, borderRadius: 4 }}>
                      <Text style={{ color: pi.method === 'FullTranscode' ? '#f87171' : '#4ade80', fontSize: 10, fontWeight: '700' }}>
                        {pi.method === 'FullTranscode' ? 'Transcoding' : 'Direct'}
                      </Text>
                    </View>
                  </View>
                  <Text style={styles.statsMono}>
                    {pi.video_codec?.toUpperCase() || '—'} {pi.height > 0 ? `${pi.width}×${pi.height}` : ''}
                    {pi.video_profile ? ` ${pi.video_profile}` : ''}{pi.video_level > 0 ? ` L${pi.video_level}` : ''}
                  </Text>
                  <Text style={styles.statsMonoMuted}>
                    {pi.bitrate > 0 ? fmtBitrate(pi.bitrate) : ''}
                    {pi.video_fps > 0 ? ` · ${Number.isInteger(pi.video_fps) ? pi.video_fps : pi.video_fps.toFixed(2)} fps` : ''}
                  </Text>
                </View>

                {/* Audio */}
                {selectedAudio && (
                  <View style={styles.statsSection}>
                    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                      <Text style={styles.statsTitle}>Audio</Text>
                      <View style={{ backgroundColor: isTranscoding ? '#facc1533' : '#4ade8033', paddingHorizontal: 6, paddingVertical: 2, borderRadius: 4 }}>
                        <Text style={{ color: isTranscoding ? '#facc15' : '#4ade80', fontSize: 10, fontWeight: '700' }}>
                          {isTranscoding ? 'Transcoding' : 'Direct'}
                        </Text>
                      </View>
                    </View>
                    <Text style={styles.statsMono}>
                      {selectedAudio.codec?.toUpperCase() || '—'} {fmtChannels(selectedAudio.channels)}
                      {selectedAudio.language ? ` · ${selectedAudio.language}` : ''}
                      {selectedAudio.is_default ? ' (Default)' : ''}
                    </Text>
                    <Text style={styles.statsMonoMuted}>
                      {selectedAudio.bitrate > 0 ? `${selectedAudio.bitrate >= 1000 ? `${Math.round(selectedAudio.bitrate / 1000)} Kbps` : `${selectedAudio.bitrate} bps`}` : ''}
                      {selectedAudio.sample_rate > 0 ? ` · ${selectedAudio.sample_rate} Hz` : ''}
                    </Text>
                  </View>
                )}

                {/* Stream */}
                <View style={[styles.statsSection, { borderBottomWidth: 0 }]}>
                  <Text style={[styles.statsTitle, { marginBottom: 4 }]}>Stream</Text>
                  <Text style={styles.statsMono}>
                    {pi.method === 'DirectPlay' ? 'HTTP Range' : 'HLS'}
                    {' · '}{pi.container?.toUpperCase() || '—'}
                    {pi.file_size > 0 ? ` · ${(pi.file_size / (1024 * 1024 * 1024)).toFixed(1)} GB` : ''}
                  </Text>
                  {pi.estimated_bitrate > 0 && isTranscoding && (
                    <Text style={styles.statsMonoMuted}>Estimated: {fmtBitrate(pi.estimated_bitrate)}</Text>
                  )}
                </View>
              </View>
            )
          })()}

          {/* Controls Overlay */}
          {showControls && !isLocked && (
            <View style={styles.controlsOverlay}>
              {/* Top bar */}
              <View style={styles.topBar}>
                <TouchableOpacity style={styles.topBackButton} onPress={handleBack}>
                  <ChevronLeft size={22} color="rgba(255,255,255,0.8)" />
                  <Text style={styles.backText}>Back</Text>
                </TouchableOpacity>
                <View style={styles.topBarRight}>
                  <CastButton />
                </View>
              </View>

              {/* Center — animated play/pause icon, only visible during animation */}
              <View style={styles.centerArea} pointerEvents="none">
                {animIcon && (
                  <RNAnimated.View
                    style={[
                      styles.centerPlayButton,
                      {
                        opacity: playPauseAnim,
                        transform: [{
                          scale: playPauseAnim.interpolate({
                            inputRange: [0, 1],
                            outputRange: [1.6, 1],
                          }),
                        }],
                      },
                    ]}
                  >
                    {animIcon === 'play' ? (
                      <Play size={playPauseIconSize} color="#fff" fill="#fff" style={{ marginLeft: 4 }} />
                    ) : (
                      <Pause size={playPauseIconSize} color="#fff" fill="#fff" />
                    )}
                  </RNAnimated.View>
                )}
              </View>

              {/* Bottom panel */}
              <View style={styles.bottomPanel}>
                {/* Title + Action buttons in one row */}
                <View style={styles.titleActionRow}>
                  <Text
                    style={[
                      styles.bottomTitle,
                      { fontSize: scaledFont(14, layout.fontScale), flex: layout.device === 'tablet' ? 5 : 3 },
                    ]}
                    numberOfLines={1}
                  >
                    {mediaTitle}
                  </Text>
                  <View style={[styles.actionButtonsRow, { flex: layout.device === 'tablet' ? 5 : 7 }]}>
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
                    onPress={() => { setShowAspectRatioMenu(true); setSettingsView('main'); scheduleHideControls() }}
                  >
                    <Settings size={17} color="rgba(255,255,255,0.7)" />
                  </TouchableOpacity>
                  {/* Next Episode */}
                  {nextEpisode && (
                    <TouchableOpacity
                      style={styles.actionButton}
                      onPress={() => handlePlayNextEpisode(nextEpisode.media_id)}
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
                  {/* Fullscreen — rotate orientation */}
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={toggleFullscreen}
                  >
                    {isLandscape ? <Minimize2 size={17} color="rgba(255,255,255,0.7)" /> : <Maximize2 size={17} color="rgba(255,255,255,0.7)" />}
                  </TouchableOpacity>
                  </View>
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
                      <TouchableOpacity style={styles.seekSmallButton} hitSlop={{ top: 12, bottom: 12, left: 12, right: 4 }} onPress={() => handleSeek(-5)}>
                        <RotateCcw size={20} color="rgba(255,255,255,0.75)" />
                        <Text style={styles.seekSmallText}>5</Text>
                      </TouchableOpacity>
                      <TouchableOpacity style={styles.playSmallButton} hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }} onPress={togglePlayPause}>
                        {isPlaying || isBuffering ? (
                          <Pause size={26} color="#fff" fill="#fff" />
                        ) : (
                          <Play size={26} color="#fff" fill="#fff" style={{ marginLeft: 2 }} />
                        )}
                      </TouchableOpacity>
                      <TouchableOpacity style={styles.seekSmallButton} hitSlop={{ top: 12, bottom: 12, left: 4, right: 12 }} onPress={() => handleSeek(5)}>
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

          {/* Up Next card — floating bottom-right, matching web */}
          {showNextEpisode && nextEpisode && (
            <View style={styles.upNextCard} onStartShouldSetResponder={() => true}>
              <Text style={styles.upNextLabel}>Up next</Text>
              <Text style={styles.upNextTitle} numberOfLines={2}>{nextEpisode.title}</Text>
              <View style={styles.upNextButtons}>
                <TouchableOpacity
                  style={styles.upNextPlayBtn}
                  onPress={() => handlePlayNextEpisode(nextEpisode.media_id)}
                >
                  <Play size={13} color="#fff" fill="#fff" />
                  <Text style={styles.upNextPlayText}>Play Next</Text>
                </TouchableOpacity>
                <TouchableOpacity style={styles.upNextDismissBtn} onPress={handleCancelNextEpisode}>
                  <Text style={styles.upNextDismissText}>Dismiss</Text>
                </TouchableOpacity>
              </View>
            </View>
          )}
        </View>
      </GestureDetector>

      {/* Subtitle Picker Modal — matches web SubtitlePicker */}
      <Modal
        visible={showSubtitleMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowSubtitleMenu(false)}
      >
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1} onPress={() => setShowSubtitleMenu(false)}>
          <View style={[styles.subPickerContainer, { width: layout.device === 'phone' ? 280 : 320 }]} onStartShouldSetResponder={() => true}>
            {/* Header */}
            <View style={styles.subPickerHeader}>
              <Text style={styles.subPickerHeaderText}>Subtitles</Text>
            </View>

            <ScrollView style={{ maxHeight: SCREEN_HEIGHT * 0.6 }} showsVerticalScrollIndicator={false}>
              {/* ── Primary subtitle list ── */}
              <SubRow
                name="Off"
                selected={!selectedSubtitle}
                onPress={() => { setSelectedSubtitle(null); setShowSubtitleMenu(false) }}
              />

              {buildVisibleSubtitles(backendSubtitleTracks, false).map((track) => {
                const { name, fmt } = parseSubtitleLabel(track.label, track.language)
                const isSelected = languageMatches(primaryBackendSub?.language ?? null, track.language)
                return (
                  <SubRow
                    key={track.id}
                    name={name}
                    fmt={fmt || track.format}
                    selected={isSelected}
                    onPress={() => { setSelectedSubtitle(track.language); setShowSubtitleMenu(false) }}
                  />
                )
              })}

              {/* ── Secondary subtitle section ── */}
              <View style={styles.subSectionDivider}>
                <Text style={styles.subSectionLabel}>Secondary subtitle</Text>
              </View>

              <SubRow
                name="Off"
                selected={!secondarySubLang}
                onPress={() => { setSecondarySubtitleLanguage(null); setShowSubtitleMenu(false) }}
              />

              {buildVisibleSubtitles(backendSubtitleTracks, false)
                .filter((s) => !s.is_image)
                .map((track) => {
                  const { name, fmt } = parseSubtitleLabel(track.label, track.language)
                  const isSelected = languageMatches(secondarySubLang, track.language)
                  return (
                    <SubRow
                      key={`sec-${track.id}`}
                      name={name}
                      fmt={fmt || track.format}
                      selected={isSelected}
                      onPress={() => { setSecondarySubtitleLanguage(track.language, track.id); setShowSubtitleMenu(false) }}
                    />
                  )
                })}

              {backendSubtitleTracks.length === 0 && (
                <View style={{ padding: 20, alignItems: 'center' }}>
                  <Text style={styles.noSubtitlesText}>No subtitles available</Text>
                </View>
              )}

              {/* ── Secondary source selector (expandable) ── */}
              {secondarySubLang && (() => {
                const sources = backendSubtitleTracks.filter(
                  (t) => !t.is_image && languageMatches(t.language, secondarySubLang)
                )
                if (sources.length <= 1) return null
                return (
                  <View style={styles.subSourceSection}>
                    <Text style={styles.subSectionLabel}>Secondary Source</Text>
                    {sources.map((s) => {
                      const { name: srcName } = parseSubtitleLabel(s.label, s.language)
                      const label = `${srcName} (#${s.id} • ${s.format?.toUpperCase()}${s.is_default ? ' • Default' : ''})`
                      const isActive = secondaryBackendSub?.id === s.id
                      return (
                        <TouchableOpacity
                          key={s.id}
                          style={[styles.subSourceOption, isActive && styles.subSourceOptionActive]}
                          onPress={() => setSecondarySubtitleLanguage(s.language, s.id)}
                        >
                          <Text style={[styles.subSourceOptionText, isActive && { color: '#fff' }]} numberOfLines={1}>{label}</Text>
                          {isActive && <Text style={{ color: '#fff', fontSize: 14 }}>✓</Text>}
                        </TouchableOpacity>
                      )
                    })}
                  </View>
                )
              })()}

              {/* ── Translate Subtitle ── */}
              {backendSubtitleTracks.filter((s) => !s.is_image).length > 0 && (
                <SubtitleTranslateSection
                  subtitles={backendSubtitleTracks.filter((s) => !s.is_image)}
                  onTranslated={() => {
                    // Refresh playback info to get new subtitle
                    setShowSubtitleMenu(false)
                  }}
                />
              )}

              {/* ── Search for Subtitles ── */}
              <SubtitleSearchSection
                mediaId={mediaId}
                defaultLang={selectedSubtitle}
                onDownloaded={() => {
                  setShowSubtitleMenu(false)
                }}
              />
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

              {/* Quality — clickable row opens submenu (like web) */}
              {qualityOptions.length > 0 && (
                <>
                  <View style={styles.settingsDivider} />
                  <TouchableOpacity
                    style={styles.settingsRow}
                    onPress={() => setSettingsView((prev) => prev === 'quality' ? 'main' : 'quality')}
                  >
                    <Text style={styles.settingsRowText}>Quality</Text>
                    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
                      <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>
                        {currentQualityLabel}
                      </Text>
                      <ChevronRight size={14} color="rgba(255,255,255,0.5)" />
                    </View>
                  </TouchableOpacity>
                  {settingsView === 'quality' && (
                    <View style={{ gap: 2, marginTop: 4 }}>
                      {qualityOptions.map((q) => {
                        const isSelected = maxQuality !== 'auto' && maxQuality === q.height
                        return (
                          <TouchableOpacity
                            key={`${q.source}-${q.height}`}
                            style={[styles.settingsRow, isSelected && styles.settingsRowActive]}
                            onPress={() => { handleQualitySelect(q.height); setSettingsView('main') }}
                          >
                            <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                              <Text style={[styles.settingsRowText, isSelected && { color: '#fff', fontWeight: '600' }]}>
                                {q.label}
                              </Text>
                              {q.instant && q.source !== 'original' && (
                                <Text style={{ fontSize: 11, color: '#facc15' }}>⚡</Text>
                              )}
                            </View>
                          </TouchableOpacity>
                        )
                      })}
                      {/* Auto option at bottom with separator */}
                      <View style={{ borderTopWidth: 1, borderTopColor: 'rgba(255,255,255,0.1)', marginTop: 2, paddingTop: 2 }}>
                        <TouchableOpacity
                          style={[styles.settingsRow, maxQuality === 'auto' && styles.settingsRowActive]}
                          onPress={() => { handleQualitySelect('auto'); setSettingsView('main') }}
                        >
                          <Text style={[styles.settingsRowText, maxQuality === 'auto' && { color: '#fff', fontWeight: '600' }]}>
                            Auto
                          </Text>
                        </TouchableOpacity>
                      </View>
                    </View>
                  )}
                </>
              )}

              {/* Repeat Mode */}
              <View style={styles.settingsDivider} />
              <Text style={styles.settingsSectionTitle}>Repeat</Text>
              <View style={styles.settingsGrid}>
                {(['none', 'one', 'all'] as const).map((m) => (
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
    backgroundColor: 'transparent',
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
  volumeRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 14,
  },
  volumeSlider: {
    width: 120,
    height: 32,
    justifyContent: 'center',
  },
  volumeTrack: {
    height: 4,
    borderRadius: 2,
    backgroundColor: 'rgba(255,255,255,0.25)',
  },
  volumeFill: {
    height: '100%',
    borderRadius: 2,
    backgroundColor: '#fff',
  },
  volumeThumb: {
    position: 'absolute',
    top: -5,
    marginLeft: -7,
    width: 14,
    height: 14,
    borderRadius: 7,
    backgroundColor: '#fff',
  },
  centerArea: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  centerPlayButton: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: 'rgba(0,0,0,0.45)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  pauseIndicatorContainer: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
  },
  pauseIndicator: {
    width: 88,
    height: 88,
    borderRadius: 44,
    backgroundColor: 'rgba(0,0,0,0.3)',
    justifyContent: 'center',
    alignItems: 'center',
    overflow: 'hidden',
  },
  bottomPanel: {
    paddingHorizontal: 16,
    paddingBottom: 34,
    paddingTop: 16,
    backgroundColor: 'transparent',
  },
  titleActionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 10,
    gap: 8,
  },
  bottomTitle: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '700',
  },
  actionButtonsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'flex-end',
    gap: 8,
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
    padding: 12,
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
  // doubleTapFeedback styles moved to SeekFeedbackOverlay component
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
    top: 0,
    bottom: 0,
    justifyContent: 'center',
  },
  volumeFeedbackText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
  },
  brightnessFeedback: {
    position: 'absolute',
    left: 20,
    top: 0,
    bottom: 0,
    justifyContent: 'center',
  },
  brightnessFeedbackText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
  },
  skipButton: {
    position: 'absolute',
    bottom: 240,
    right: 20,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
    zIndex: 50,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  skipButtonText: {
    color: '#000',
    fontSize: 14,
    fontWeight: '700',
  },
  statsOverlay: {
    position: 'absolute',
    top: 80,
    left: 16,
    backgroundColor: 'rgba(0, 0, 0, 0.75)',
    borderRadius: 12,
    width: 240,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  statsCloseButton: {
    position: 'absolute',
    right: 8,
    top: 8,
    padding: 4,
    borderRadius: 8,
    zIndex: 1,
  },
  statsSection: {
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.1)',
  },
  statsTitle: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '700',
  },
  statsMono: {
    color: 'rgba(255,255,255,0.8)',
    fontSize: 11,
    fontFamily: 'monospace',
    lineHeight: 18,
  },
  statsMonoMuted: {
    color: 'rgba(255,255,255,0.5)',
    fontSize: 11,
    fontFamily: 'monospace',
    lineHeight: 18,
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
  upNextCard: {
    position: 'absolute',
    bottom: 240,
    right: 20,
    width: 240,
    backgroundColor: '#1e1e1e',
    borderRadius: 12,
    padding: 16,
    zIndex: 50,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.6,
    shadowRadius: 16,
    elevation: 20,
  },
  upNextLabel: {
    color: 'rgba(255,255,255,0.5)',
    fontSize: 11,
    marginBottom: 4,
  },
  upNextTitle: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 12,
  },
  upNextButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  upNextPlayBtn: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 10,
  },
  upNextPlayText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  upNextDismissBtn: {
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 8,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  upNextDismissText: {
    color: 'rgba(255,255,255,0.7)',
    fontSize: 13,
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
    borderRadius: 12,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.6,
    shadowRadius: 24,
    elevation: 20,
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
  subSourceSection: {
    borderTopWidth: 1,
    borderTopColor: 'rgba(255,255,255,0.1)',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  subSourceOption: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 6,
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  subSourceOptionActive: {
    borderColor: 'rgba(255,255,255,0.25)',
    backgroundColor: 'rgba(255,255,255,0.1)',
  },
  subSourceOptionText: {
    flex: 1,
    fontSize: 13,
    color: 'rgba(255,255,255,0.6)',
  },
  _subBottomRowRemoved: {
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
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
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
