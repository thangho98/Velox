import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type { SubtitleSearchResult, SubtitleDownloadRequest } from '@/types/api'

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
  return useMutation({
    mutationFn: (body: SubtitleDownloadRequest) => subtitleSearchApi.download(mediaId, body),
    // NOTE: Do NOT invalidate streamingKeys here — it causes the video to reload.
    // The caller (SubtitleSearchModal) calls onSubtitleAdded which handles refresh.
  })
}

export function useTranslateSubtitle() {
  return useMutation({
    mutationFn: ({ subtitleId, targetLanguage }: { subtitleId: number; targetLanguage: string }) =>
      api.post<unknown>(`/subtitles/${subtitleId}/translate`, { target_language: targetLanguage }),
    // NOTE: Do NOT invalidate streamingKeys here — it causes the video to reload.
    // The caller (SubtitlePicker) calls onSubtitleAdded which handles refresh safely.
  })
}
