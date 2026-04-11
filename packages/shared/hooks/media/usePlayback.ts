/**
 * Playback / streaming hooks — stream URLs, subtitles, audio tracks
 */

import { useQuery, useMutation } from '@tanstack/react-query'
import type { UseMutationOptions } from '@tanstack/react-query'
import { api } from '../../api'
import type {
  StreamUrls,
  PlaybackInfo,
  PlaybackInfoRequest,
  StreamURLResponse,
  PlaybackSubtitleTrack,
  PlaybackAudioTrack,
} from '../../types'

// Streaming API Functions
// POST /api/playback/{id}/info returns stream_url, audio_tracks, subtitle_tracks in one call
const streamingApi = {
  getPlaybackInfo: (mediaId: number, request: PlaybackInfoRequest = {}) =>
    api.post<PlaybackInfo>(`/playback/${mediaId}/info`, request),
}

// Query Keys for Streaming
export const streamingKeys = {
  all: ['streaming'] as const,
  playbackInfo: (mediaId: number, request: PlaybackInfoRequest = {}) =>
    [...streamingKeys.all, 'info', mediaId, request] as const,
}

// React Query Hooks for Streaming

export function useStreamUrls(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    select: (info: PlaybackInfo): StreamUrls => {
      // Backend now provides three explicit URLs (direct/pretranscode/hls) plus
      // a `prefer` hint. Client owns the actual fallback chain — see WatchPage.
      // Legacy fallback: derive the v1 HLS URL from stream_url if backend didn't include hls_url.
      let hlsUrl = info.hls_url
      if (!hlsUrl) {
        const isHLS = info.stream_url.includes('/hls/')
        if (isHLS) {
          hlsUrl = info.stream_url
        } else {
          const [path, qs] = info.stream_url.split('?')
          const params = new URLSearchParams(qs || '')
          params.delete('pm')
          if (request.max_height && request.max_height > 0) {
            params.set('mh', String(request.max_height))
          }
          hlsUrl =
            path + '/hls/master.m3u8' + (params.toString() ? '?' + params.toString() : '')
        }
      }

      // direct: always the file-gốc with pm=direct (forces bypass of pretranscode).
      // Falls back to stream_url for older backends.
      const directUrl = info.direct_url || info.stream_url

      return {
        direct: directUrl,
        pretranscode: info.pretranscode_url,
        hls: hlsUrl,
        prefer: info.prefer,
        primary_file_id: info.primary_file_id,
        stream_session_id: info.stream_session_id,
      }
    },
    staleTime: 0,
    refetchOnMount: 'always',
    placeholderData: (prev) => prev,
    structuralSharing: (prev, next) => {
      const p = prev as StreamUrls | undefined
      const n = next as StreamUrls
      if (
        p &&
        p.direct === n.direct &&
        p.pretranscode === n.pretranscode &&
        p.hls === n.hls &&
        p.prefer === n.prefer &&
        p.primary_file_id === n.primary_file_id &&
        p.stream_session_id === n.stream_session_id
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
    placeholderData: (prev) => prev,
    enabled: mediaId > 0,
  })
}

type StreamUrlMutationOptions = Omit<
  UseMutationOptions<StreamURLResponse, Error, void, unknown>,
  'mutationFn'
>

export function useStreamUrl(mediaId: number, options?: StreamUrlMutationOptions) {
  return useMutation({
    mutationFn: () => api.post<StreamURLResponse>(`/stream/${mediaId}/url`, {}),
    ...options,
  })
}
