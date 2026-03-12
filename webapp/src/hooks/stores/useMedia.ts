import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type {
  FsBrowseResponse,
  Library,
  CreateLibraryRequest,
  Media,
  MediaWithFiles,
  MediaListItem,
  MediaListParams,
  Season,
  Episode,
  StreamUrls,
  PlaybackInfo,
  PlaybackInfoRequest,
  PlaybackSubtitleTrack,
  PlaybackAudioTrack,
  UserData,
  UpdateProgressRequest,
  ToggleFavoriteResponse,
  FavoritesListParams,
  RecentlyWatchedParams,
} from '@/types/api'

// Filesystem browser API (admin only)
export function useFsBrowse(path: string) {
  return useQuery({
    queryKey: ['fs-browse', path],
    queryFn: () => api.get<FsBrowseResponse>(`/admin/fs/browse?path=${encodeURIComponent(path)}`),
    staleTime: 0, // always fresh — filesystem can change
  })
}

// Library API Functions
const libraryApi = {
  list: () => api.get<Library[]>('/libraries'),
  create: (data: CreateLibraryRequest) => api.post<Library>('/libraries', data),
  delete: (id: number) => api.delete(`/libraries/${id}`),
  scan: (id: number) => api.post<void>(`/libraries/${id}/scan`, {}),
}

// Query Keys
export const libraryKeys = {
  all: ['libraries'] as const,
  list: () => [...libraryKeys.all, 'list'] as const,
  detail: (id: number) => [...libraryKeys.all, 'detail', id] as const,
}

// React Query Hooks
export function useLibraries() {
  return useQuery({
    queryKey: libraryKeys.list(),
    queryFn: libraryApi.list,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useCreateLibrary() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: libraryApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: libraryKeys.list() })
    },
  })
}

export function useDeleteLibrary() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: libraryApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: libraryKeys.list() })
    },
  })
}

export function useScanLibrary() {
  return useMutation({
    mutationFn: libraryApi.scan,
  })
}

// Media API Functions
const mediaApi = {
  list: (params: MediaListParams) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.type) searchParams.append('type', params.type)
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))

    const query = searchParams.toString()
    return api.get<MediaListItem[]>(`/media${query ? `?${query}` : ''}`)
  },
  get: (id: number) => api.get<Media>(`/media/${id}`),
  getWithFiles: (id: number) => api.get<MediaWithFiles>(`/media/${id}/files`),
}

// Streaming API Functions
// POST /api/playback/{id}/info returns stream_url, audio_tracks, subtitle_tracks in one call
const streamingApi = {
  getPlaybackInfo: (mediaId: number, request: PlaybackInfoRequest = {}) =>
    api.post<PlaybackInfo>(`/playback/${mediaId}/info`, request),
}

// Query Keys
export const mediaKeys = {
  all: ['media'] as const,
  list: (params: MediaListParams) => [...mediaKeys.all, 'list', params] as const,
  detail: (id: number) => [...mediaKeys.all, 'detail', id] as const,
  withFiles: (id: number) => [...mediaKeys.all, 'withFiles', id] as const,
}

// React Query Hooks
export function useMediaList(params: MediaListParams = {}) {
  return useQuery({
    queryKey: mediaKeys.list(params),
    queryFn: () => mediaApi.list(params),
    staleTime: 60 * 1000, // 1 minute
  })
}

export function useMedia(id: number) {
  return useQuery({
    queryKey: mediaKeys.detail(id),
    queryFn: () => mediaApi.get(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: id > 0,
  })
}

export function useMediaWithFiles(id: number) {
  return useQuery({
    queryKey: mediaKeys.withFiles(id),
    queryFn: () => mediaApi.getWithFiles(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: id > 0,
  })
}

// User Data API Functions (Progress, Favorites)
const userDataApi = {
  getProgress: (mediaId: number) => api.get<UserData | null>(`/profile/progress/${mediaId}`),
  updateProgress: (mediaId: number, data: UpdateProgressRequest) =>
    api.put<void>(`/profile/progress/${mediaId}`, data),
  listFavorites: (params: FavoritesListParams) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/favorites${query ? `?${query}` : ''}`)
  },
  toggleFavorite: (mediaId: number) =>
    api.post<ToggleFavoriteResponse>(`/profile/favorites/${mediaId}`, {}),
  listRecentlyWatched: (params: RecentlyWatchedParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/recently-watched${query ? `?${query}` : ''}`)
  },
}

// Query Keys
export const userDataKeys = {
  all: ['userData'] as const,
  progress: (mediaId: number) => [...userDataKeys.all, 'progress', mediaId] as const,
  favorites: (params: FavoritesListParams) => [...userDataKeys.all, 'favorites', params] as const,
  recentlyWatched: (params: RecentlyWatchedParams) =>
    [...userDataKeys.all, 'recentlyWatched', params] as const,
}

// React Query Hooks
export function useProgress(mediaId: number) {
  return useQuery({
    queryKey: userDataKeys.progress(mediaId),
    queryFn: () => userDataApi.getProgress(mediaId),
    staleTime: 0, // Always fresh
  })
}

export function useUpdateProgress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ mediaId, data }: { mediaId: number; data: UpdateProgressRequest }) =>
      userDataApi.updateProgress(mediaId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: userDataKeys.progress(variables.mediaId),
      })
      queryClient.invalidateQueries({
        queryKey: userDataKeys.recentlyWatched({}),
      })
    },
  })
}

export function useFavorites(params: FavoritesListParams = {}) {
  return useQuery({
    queryKey: userDataKeys.favorites(params),
    queryFn: () => userDataApi.listFavorites(params),
    staleTime: 60 * 1000,
  })
}

export function useToggleFavorite() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: userDataApi.toggleFavorite,
    onSuccess: () => {
      // Invalidate favorites list and any media item that might show favorite status
      queryClient.invalidateQueries({ queryKey: userDataKeys.all })
    },
  })
}

export function useRecentlyWatched(params: RecentlyWatchedParams = { limit: 20 }) {
  return useQuery({
    queryKey: userDataKeys.recentlyWatched(params),
    queryFn: () => userDataApi.listRecentlyWatched(params),
    staleTime: 60 * 1000,
  })
}

// Season/Episode API Functions
const seriesApi = {
  getSeasons: (seriesId: number) => api.get<Season[]>(`/series/${seriesId}/seasons`),
  getEpisodes: (seriesId: number, seasonId: number) =>
    api.get<Episode[]>(`/series/${seriesId}/seasons/${seasonId}/episodes`),
  getEpisode: (episodeId: number) => api.get<Episode>(`/episodes/${episodeId}`),
}

// Query Keys
export const seriesKeys = {
  all: ['series'] as const,
  seasons: (seriesId: number) => [...seriesKeys.all, 'seasons', seriesId] as const,
  episodes: (seriesId: number, seasonId: number) =>
    [...seriesKeys.all, 'episodes', seriesId, seasonId] as const,
  episode: (episodeId: number) => [...seriesKeys.all, 'episode', episodeId] as const,
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
export function useSeasons(seriesId: number) {
  return useQuery({
    queryKey: seriesKeys.seasons(seriesId),
    queryFn: () => seriesApi.getSeasons(seriesId),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: seriesId > 0,
  })
}

export function useEpisodes(seriesId: number, seasonId: number) {
  return useQuery({
    queryKey: seriesKeys.episodes(seriesId, seasonId),
    queryFn: () => seriesApi.getEpisodes(seriesId, seasonId),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: seriesId > 0 && seasonId > 0,
  })
}

export function useEpisode(episodeId: number) {
  return useQuery({
    queryKey: seriesKeys.episode(episodeId),
    queryFn: () => seriesApi.getEpisode(episodeId),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: episodeId > 0,
  })
}
