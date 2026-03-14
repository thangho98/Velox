import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface PlayerState {
  // Playback state
  volume: number
  isMuted: boolean
  playbackRate: number

  // Subtitle preferences
  subtitleLanguage: string | null // 'vi', 'en', null = off
  subtitleTrackId: number | null // exact subtitle track ID for source selection
  secondarySubtitleLanguage: string | null // dual-subtitle secondary track, null = off
  secondarySubtitleTrackId: number | null
  subtitleSize: 'small' | 'medium' | 'large'
  subtitleColor: string
  subtitleBackground: 'solid' | 'semi' | 'none' // solid=black box, semi=translucent, none=text-stroke only
  subtitleOffsets: Record<number, number> // mediaId -> offset in seconds

  // Audio preferences
  audioLanguage: string | null
  audioTrackId: number | null // backend track ID for the selected audio track

  // Quality preference (Emby-style: resolution + bitrate cap)
  maxQuality: { height: number; bitrateKbps: number } | 'auto'

  // Video display
  aspectRatio: 'contain' | 'cover' | 'fill'

  // Repeat mode
  repeatMode: 'none' | 'one' | 'all'

  // Last position (for resume)
  lastPositions: Record<number, number> // mediaId -> position in seconds

  // Actions
  setVolume: (volume: number) => void
  toggleMute: () => void
  setPlaybackRate: (rate: number) => void

  setSubtitleLanguage: (lang: string | null, trackId?: number | null) => void
  setSecondarySubtitleLanguage: (lang: string | null, trackId?: number | null) => void
  setSubtitleTrackId: (trackId: number | null) => void
  setSecondarySubtitleTrackId: (trackId: number | null) => void
  setSubtitleSize: (size: 'small' | 'medium' | 'large') => void
  setSubtitleColor: (color: string) => void
  setSubtitleBackground: (bg: 'solid' | 'semi' | 'none') => void
  setSubtitleOffset: (mediaId: number, seconds: number) => void
  getSubtitleOffset: (mediaId: number) => number
  resetSubtitleOffset: (mediaId: number) => void

  setAudioTrack: (lang: string | null, id: number | null) => void
  setMaxQuality: (quality: { height: number; bitrateKbps: number } | 'auto') => void
  setAspectRatio: (ratio: 'contain' | 'cover' | 'fill') => void
  setRepeatMode: (mode: 'none' | 'one' | 'all') => void

  setLastPosition: (mediaId: number, position: number) => void
  getLastPosition: (mediaId: number) => number
  clearLastPosition: (mediaId: number) => void
}

export const usePlayerStore = create<PlayerState>()(
  persist(
    (set, get) => ({
      // Defaults
      volume: 1.0,
      isMuted: false,
      playbackRate: 1.0,

      subtitleLanguage: null,
      subtitleTrackId: null,
      secondarySubtitleLanguage: null,
      secondarySubtitleTrackId: null,
      subtitleSize: 'large',
      subtitleColor: '#ffffff',
      subtitleBackground: 'none',
      subtitleOffsets: {},

      audioLanguage: null,
      audioTrackId: null,

      maxQuality: 'auto',
      aspectRatio: 'contain',
      repeatMode: 'none',

      lastPositions: {},

      // Actions
      setVolume: (volume) => set({ volume: Math.max(0, Math.min(1, volume)) }),

      toggleMute: () => set((state) => ({ isMuted: !state.isMuted })),

      setPlaybackRate: (rate) =>
        set({ playbackRate: [0.5, 0.75, 1.0, 1.25, 1.5, 2.0].includes(rate) ? rate : 1.0 }),

      setSubtitleLanguage: (lang, trackId = null) =>
        set({ subtitleLanguage: lang, subtitleTrackId: lang ? trackId : null }),

      setSecondarySubtitleLanguage: (lang, trackId = null) =>
        set({ secondarySubtitleLanguage: lang, secondarySubtitleTrackId: lang ? trackId : null }),

      setSubtitleTrackId: (trackId) => set({ subtitleTrackId: trackId }),

      setSecondarySubtitleTrackId: (trackId) => set({ secondarySubtitleTrackId: trackId }),

      setSubtitleSize: (size) => set({ subtitleSize: size }),

      setSubtitleColor: (color) => set({ subtitleColor: color }),

      setSubtitleBackground: (bg) => set({ subtitleBackground: bg }),

      setSubtitleOffset: (mediaId, seconds) =>
        set((state) => ({
          subtitleOffsets: {
            ...state.subtitleOffsets,
            [mediaId]: Math.max(-10, Math.min(10, seconds)),
          },
        })),

      getSubtitleOffset: (mediaId) => get().subtitleOffsets[mediaId] || 0,

      resetSubtitleOffset: (mediaId) =>
        set((state) => {
          const subtitleOffsets = { ...state.subtitleOffsets }
          delete subtitleOffsets[mediaId]
          return { subtitleOffsets }
        }),

      setAudioTrack: (lang: string | null, id: number | null) =>
        set({ audioLanguage: lang, audioTrackId: id }),

      setMaxQuality: (quality) => set({ maxQuality: quality }),
      setAspectRatio: (ratio) => set({ aspectRatio: ratio }),
      setRepeatMode: (mode) => set({ repeatMode: mode }),

      setLastPosition: (mediaId, position) =>
        set((state) => ({
          lastPositions: { ...state.lastPositions, [mediaId]: position },
        })),

      getLastPosition: (mediaId) => get().lastPositions[mediaId] || 0,

      clearLastPosition: (mediaId) =>
        set((state) => {
          const newPositions = { ...state.lastPositions }
          delete newPositions[mediaId]
          return { lastPositions: newPositions }
        }),
    }),
    {
      name: 'velox-player',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        volume: state.volume,
        isMuted: state.isMuted,
        playbackRate: state.playbackRate,
        subtitleLanguage: state.subtitleLanguage,
        subtitleTrackId: state.subtitleTrackId,
        secondarySubtitleLanguage: state.secondarySubtitleLanguage,
        secondarySubtitleTrackId: state.secondarySubtitleTrackId,
        subtitleSize: state.subtitleSize,
        subtitleColor: state.subtitleColor,
        subtitleBackground: state.subtitleBackground,
        subtitleOffsets: state.subtitleOffsets,
        audioLanguage: state.audioLanguage,
        maxQuality: state.maxQuality,
        aspectRatio: state.aspectRatio,
        repeatMode: state.repeatMode,
        lastPositions: state.lastPositions,
      }),
    },
  ),
)
