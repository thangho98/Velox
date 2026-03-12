import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface PlayerState {
  // Playback state
  volume: number
  isMuted: boolean
  playbackRate: number

  // Subtitle preferences
  subtitleLanguage: string | null // 'vi', 'en', null = off
  secondarySubtitleLanguage: string | null // dual-subtitle secondary track, null = off
  subtitleSize: 'small' | 'medium' | 'large'
  subtitleColor: string

  // Audio preferences
  audioLanguage: string | null
  audioTrackId: number | null // backend track ID for the selected audio track

  // Quality preference (sent as max_height in playback info request)
  maxStreamingQuality: 'auto' | '1080p' | '720p' | '480p'

  // Last position (for resume)
  lastPositions: Record<number, number> // mediaId -> position in seconds

  // Actions
  setVolume: (volume: number) => void
  toggleMute: () => void
  setPlaybackRate: (rate: number) => void

  setSubtitleLanguage: (lang: string | null) => void
  setSecondarySubtitleLanguage: (lang: string | null) => void
  setSubtitleSize: (size: 'small' | 'medium' | 'large') => void
  setSubtitleColor: (color: string) => void

  setAudioTrack: (lang: string | null, id: number | null) => void
  setMaxStreamingQuality: (quality: 'auto' | '1080p' | '720p' | '480p') => void

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
      secondarySubtitleLanguage: null,
      subtitleSize: 'medium',
      subtitleColor: '#ffffff',

      audioLanguage: null,
      audioTrackId: null,

      maxStreamingQuality: 'auto',

      lastPositions: {},

      // Actions
      setVolume: (volume) => set({ volume: Math.max(0, Math.min(1, volume)) }),

      toggleMute: () => set((state) => ({ isMuted: !state.isMuted })),

      setPlaybackRate: (rate) =>
        set({ playbackRate: [0.5, 0.75, 1.0, 1.25, 1.5, 2.0].includes(rate) ? rate : 1.0 }),

      setSubtitleLanguage: (lang) => set({ subtitleLanguage: lang }),

      setSecondarySubtitleLanguage: (lang) => set({ secondarySubtitleLanguage: lang }),

      setSubtitleSize: (size) => set({ subtitleSize: size }),

      setSubtitleColor: (color) => set({ subtitleColor: color }),

      setAudioTrack: (lang: string | null, id: number | null) =>
        set({ audioLanguage: lang, audioTrackId: id }),

      setMaxStreamingQuality: (quality) => set({ maxStreamingQuality: quality }),

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
        secondarySubtitleLanguage: state.secondarySubtitleLanguage,
        subtitleSize: state.subtitleSize,
        subtitleColor: state.subtitleColor,
        audioLanguage: state.audioLanguage,
        audioTrackId: state.audioTrackId,
        maxStreamingQuality: state.maxStreamingQuality,
        lastPositions: state.lastPositions,
      }),
    },
  ),
)
