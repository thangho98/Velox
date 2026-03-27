import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import { useCinemaSettings } from '@/hooks/stores/useSettings'

export interface CinemaItem {
  type: 'intro' | 'trailer' | 'main'
  title: string
  url: string
  duration: number
  skippable: boolean
}

interface CinemaPlaylist {
  items: CinemaItem[]
}

export function useSeriesTrailers(seriesId: number) {
  const { data: cinemaSettings } = useCinemaSettings()
  const cinemaEnabled = cinemaSettings?.enabled ?? false

  const { data } = useQuery({
    queryKey: ['cinema', 'series', seriesId],
    queryFn: () => api.get<CinemaPlaylist>(`/series/${seriesId}/cinema`),
    enabled: seriesId > 0 && cinemaEnabled,
    staleTime: 10 * 60 * 1000,
  })

  const trailers = (data?.items ?? []).filter((item) => item.type === 'trailer')

  return {
    trailers,
    youtubeKey: cinemaEnabled && trailers.length > 0 ? extractYouTubeKey(trailers[0].url) : null,
  }
}

export function useTrailers(mediaId: number) {
  const { data: cinemaSettings } = useCinemaSettings()
  const cinemaEnabled = cinemaSettings?.enabled ?? false

  const { data } = useQuery({
    queryKey: ['cinema', mediaId],
    queryFn: () => api.get<CinemaPlaylist>(`/media/${mediaId}/cinema`),
    enabled: mediaId > 0 && cinemaEnabled,
    staleTime: 10 * 60 * 1000,
  })

  const trailers = (data?.items ?? []).filter((item) => item.type === 'trailer')

  return {
    trailers,
    youtubeKey: cinemaEnabled && trailers.length > 0 ? extractYouTubeKey(trailers[0].url) : null,
    title: trailers.length > 0 ? trailers[0].title : null,
  }
}

function extractYouTubeKey(url: string): string | null {
  const match = url.match(/embed\/([^?]+)/)
  return match ? match[1] : null
}
