import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type {
  StreamUrls,
  PlaybackInfo,
  PlaybackInfoRequest,
  PlaybackSubtitleTrack,
  PlaybackAudioTrack,
} from '@/types/api'

// Streaming API Functions
// POST /api/playback/{id}/info returns stream_url, audio_tracks, subtitle_tracks in one call
const streamingApi = {
  getPlaybackInfo: (mediaId: number, request: PlaybackInfoRequest = {}) =>
    api.post<PlaybackInfo>(`/playback/${mediaId}/info`, request),
}

// Query Keys for Streaming — request params included so selection changes trigger refetch
export const streamingKeys = {
  all: ['streaming'] as const,
  playbackInfo: (mediaId: number, request: PlaybackInfoRequest = {}) =>
    [...streamingKeys.all, 'info', mediaId, request] as const,
}

// React Query Hooks for Streaming — all derived from a single POST /api/playback/{id}/info call

export function useStreamUrls(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    select: (info: PlaybackInfo): StreamUrls => {
      const isHLS = info.method === 'TranscodeAudio' || info.method === 'FullTranscode'
      return {
        direct: info.stream_url,
        hls: isHLS ? info.stream_url : undefined,
        abr: info.abr_url || undefined,
        primary_file_id: info.primary_file_id,
      }
    },
    staleTime: 5 * 60 * 1000,
    placeholderData: (prev) => prev, // Keep previous data while refetching to prevent video remount
    // Preserve object identity when stream URLs haven't changed — prevents
    // the HLS init effect from restarting the video on subtitle selection changes.
    structuralSharing: (prev, next) => {
      const p = prev as StreamUrls | undefined
      const n = next as StreamUrls
      if (
        p &&
        p.direct === n.direct &&
        p.hls === n.hls &&
        p.abr === n.abr &&
        p.primary_file_id === n.primary_file_id
      ) {
        return p
      }
      return n
    },
    enabled: mediaId > 0,
  })
}

export function useSubtitles(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    select: (info: PlaybackInfo): PlaybackSubtitleTrack[] => info.subtitle_tracks ?? [],
    staleTime: 5 * 60 * 1000,
    enabled: mediaId > 0,
  })
}

export function useAudioTracks(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    select: (info: PlaybackInfo): PlaybackAudioTrack[] => info.audio_tracks ?? [],
    staleTime: 5 * 60 * 1000,
    enabled: mediaId > 0,
  })
}

export function usePlaybackInfo(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    staleTime: 5 * 60 * 1000,
    placeholderData: (prev) => prev, // Keep previous data while refetching to prevent video remount
    enabled: mediaId > 0,
  })
}

export function useStreamUrl(mediaId: number) {
  return useMutation({
    mutationFn: () =>
      api.post<{ direct_url: string; hls_url: string; token: string; expires_in: number }>(
        `/stream/${mediaId}/url`,
        {},
      ),
  })
}
