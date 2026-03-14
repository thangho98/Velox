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
  Series,
  SeriesListParams,
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
  SubtitleSearchResult,
  SubtitleDownloadRequest,
  ContinueWatchingItem,
  NextUpItem,
  ContinueWatchingParams,
  NextUpParams,
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
  scan: (id: number, force = false) =>
    api.post<void>(`/libraries/${id}/scan${force ? '?force=true' : ''}`, {}),
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
    mutationFn: ({ id, force = false }: { id: number; force?: boolean }) =>
      libraryApi.scan(id, force),
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
      // Invalidate continue-watching and next-up when progress updates
      queryClient.invalidateQueries({
        queryKey: continueWatchingKeys.all,
      })
      queryClient.invalidateQueries({
        queryKey: nextUpKeys.all,
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

// Continue Watching / Next Up API Functions
const continueWatchingApi = {
  list: (params: ContinueWatchingParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<ContinueWatchingItem[]>(`/profile/continue-watching${query ? `?${query}` : ''}`)
  },
  dismiss: (mediaId: number) => api.delete(`/profile/progress/${mediaId}/dismiss`),
}

const nextUpApi = {
  list: (params: NextUpParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<NextUpItem[]>(`/profile/next-up${query ? `?${query}` : ''}`)
  },
}

// Query Keys
export const continueWatchingKeys = {
  all: ['continueWatching'] as const,
  list: (params: ContinueWatchingParams) => [...continueWatchingKeys.all, 'list', params] as const,
}

export const nextUpKeys = {
  all: ['nextUp'] as const,
  list: (params: NextUpParams) => [...nextUpKeys.all, 'list', params] as const,
}

// React Query Hooks
export function useContinueWatching(params: ContinueWatchingParams = { limit: 20 }) {
  return useQuery({
    queryKey: continueWatchingKeys.list(params),
    queryFn: () => continueWatchingApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useNextUp(params: NextUpParams = { limit: 20 }) {
  return useQuery({
    queryKey: nextUpKeys.list(params),
    queryFn: () => nextUpApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useDismissProgress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (mediaId: number) => continueWatchingApi.dismiss(mediaId),
    onSuccess: (_, mediaId) => {
      queryClient.invalidateQueries({ queryKey: userDataKeys.progress(mediaId) })
      queryClient.invalidateQueries({ queryKey: userDataKeys.recentlyWatched({}) })
      queryClient.invalidateQueries({ queryKey: continueWatchingKeys.all })
      queryClient.invalidateQueries({ queryKey: nextUpKeys.all })
    },
  })
}

// Season/Episode API Functions
const seriesApi = {
  list: (params: SeriesListParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    const query = searchParams.toString()
    return api.get<Series[]>(`/series${query ? `?${query}` : ''}`)
  },
  get: (id: number) => api.get<Series>(`/series/${id}`),
  search: (query: string, limit = 20) =>
    api.get<Series[]>(`/series/search?q=${encodeURIComponent(query)}&limit=${limit}`),
  getSeasons: (seriesId: number) => api.get<Season[]>(`/series/${seriesId}/seasons`),
  getEpisodes: (seriesId: number, seasonId: number) =>
    api.get<Episode[]>(`/series/${seriesId}/seasons/${seasonId}/episodes`),
  getEpisode: (episodeId: number) => api.get<Episode>(`/episodes/${episodeId}`),
}

// Query Keys
export const seriesKeys = {
  all: ['series'] as const,
  list: (params: SeriesListParams) => [...seriesKeys.all, 'list', params] as const,
  detail: (id: number) => [...seriesKeys.all, 'detail', id] as const,
  search: (query: string) => [...seriesKeys.all, 'search', query] as const,
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

export function usePlaybackInfo(mediaId: number, request: PlaybackInfoRequest = {}) {
  return useQuery({
    queryKey: streamingKeys.playbackInfo(mediaId, request),
    queryFn: () => streamingApi.getPlaybackInfo(mediaId, request),
    staleTime: 5 * 60 * 1000,
    enabled: mediaId > 0,
  })
}
export function useSeriesList(params: SeriesListParams = {}) {
  return useQuery({
    queryKey: seriesKeys.list(params),
    queryFn: () => seriesApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useSeriesDetail(id: number) {
  return useQuery({
    queryKey: seriesKeys.detail(id),
    queryFn: () => seriesApi.get(id),
    staleTime: 5 * 60 * 1000,
    enabled: id > 0,
  })
}

export function useSeriesSearch(query: string, limit = 20) {
  return useQuery({
    queryKey: seriesKeys.search(query),
    queryFn: () => seriesApi.search(query, limit),
    staleTime: 60 * 1000,
    enabled: query.length > 0,
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

// Subtitle Search API Functions (external providers)
const subtitleSearchApi = {
  search: (mediaId: number, lang: string) =>
    api.get<SubtitleSearchResult[]>(
      `/media/${mediaId}/subtitles/search?lang=${encodeURIComponent(lang)}`,
    ),
  download: (mediaId: number, body: SubtitleDownloadRequest) =>
    api.post<unknown>(`/media/${mediaId}/subtitles/download`, body),
}

export const subtitleSearchKeys = {
  all: ['subtitleSearch'] as const,
  search: (mediaId: number, lang: string) => [...subtitleSearchKeys.all, mediaId, lang] as const,
}

export function useSubtitleSearch(mediaId: number, lang: string, enabled = true) {
  return useQuery({
    queryKey: subtitleSearchKeys.search(mediaId, lang),
    queryFn: () => subtitleSearchApi.search(mediaId, lang),
    staleTime: 2 * 60 * 1000,
    enabled: enabled && mediaId > 0 && lang !== '',
  })
}

export function useDownloadSubtitle(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: SubtitleDownloadRequest) => subtitleSearchApi.download(mediaId, body),
    onSuccess: () => {
      // Invalidate playback info so subtitle list refreshes
      queryClient.invalidateQueries({ queryKey: streamingKeys.all })
    },
  })
}

export function useRefreshMetadata(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<Media>(`/media/${mediaId}/refresh`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      queryClient.invalidateQueries({ queryKey: mediaKeys.detail(mediaId) })
    },
  })
}
