import { useParams, Link } from 'react-router'
import { useState } from 'react'
import { useSeriesDetail, useSeasons, useEpisodes } from '@/hooks/stores/useMedia'
import { EpisodeCard } from '@/components/EpisodeCard'
import { LuChevronLeft, LuFilm, LuPlay, LuTv } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'

export function SeriesDetailPage() {
  const { seriesId } = useParams<{ seriesId: string }>()
  const id = Number(seriesId)
  const [selectedSeasonId, setSelectedSeasonId] = useState<number | null>(null)

  const { data: series, isLoading: seriesLoading } = useSeriesDetail(id)
  const { data: seasons, isLoading: seasonsLoading } = useSeasons(id)
  const { data: episodes, isLoading: episodesLoading } = useEpisodes(
    id,
    selectedSeasonId || seasons?.[0]?.id || 0,
  )

  // Auto-select first season on load
  if (seasons && seasons.length > 0 && !selectedSeasonId) {
    setSelectedSeasonId(seasons[0].id)
  }

  if (seriesLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
      </div>
    )
  }

  if (!series) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <h1 className="mb-4 text-4xl font-bold text-white">404</h1>
        <p className="mb-8 text-xl text-gray-400">Series not found</p>
        <Link to="/" className="text-netflix-blue hover:underline">
          Go back home
        </Link>
      </div>
    )
  }

  const firstEpisode = episodes?.[0]
  const seriesYear = series.first_air_date ? new Date(series.first_air_date).getFullYear() : null

  return (
    <div className="min-h-screen bg-netflix-black">
      {/* Backdrop */}
      {series.backdrop_path && (
        <div className="fixed inset-0 h-screen">
          <img
            src={tmdbImage(series.backdrop_path, 'w1280')!}
            alt={series.title}
            className="h-full w-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-netflix-black via-netflix-black/80 to-netflix-black/30" />
          <div className="absolute inset-0 bg-gradient-to-r from-netflix-black via-netflix-black/50 to-transparent" />
        </div>
      )}

      {/* Content */}
      <div className="relative z-10 min-h-screen">
        {/* Back button */}
        <Link
          to="/"
          className="fixed left-4 top-20 z-20 flex items-center gap-2 rounded-full bg-black/50 p-3 text-white backdrop-blur-sm transition-colors hover:bg-black/70"
        >
          <LuChevronLeft size={20} />
        </Link>

        <div className="container mx-auto px-4 py-24 lg:px-8">
          <div className="flex flex-col gap-8 lg:flex-row">
            {/* Poster */}
            <div className="mx-auto flex-shrink-0 lg:mx-0">
              {series.poster_path ? (
                <img
                  src={tmdbImage(series.poster_path, 'w500')!}
                  alt={series.title}
                  className="w-64 rounded-lg shadow-2xl lg:w-80"
                />
              ) : (
                <div className="flex h-96 w-64 items-center justify-center rounded-lg bg-netflix-dark lg:w-80">
                  <LuFilm size={64} className="text-gray-600" />
                </div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1">
              <h1 className="mb-2 text-3xl font-bold text-white lg:text-5xl">{series.title}</h1>

              {/* Year · Status · Network */}
              <div className="mb-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-gray-400">
                {seriesYear && <span>{seriesYear}</span>}
                {series.status && (
                  <span className="rounded bg-purple-500/20 px-1.5 py-0.5 text-xs text-purple-400">
                    {series.status}
                  </span>
                )}
                {series.network && (
                  <span className="flex items-center gap-1">
                    <LuTv size={14} />
                    {series.network}
                  </span>
                )}
              </div>

              {series.overview && (
                <p className="mb-6 max-w-2xl text-base leading-relaxed text-gray-300">
                  {series.overview}
                </p>
              )}

              {/* Play Button */}
              {firstEpisode && (
                <div className="mb-6 flex flex-wrap items-center gap-3">
                  <Link
                    to={`/watch/${firstEpisode.media_id}`}
                    className="flex items-center gap-2 rounded bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-700"
                  >
                    <LuPlay size={18} className="fill-current" />
                    Play First Episode
                  </Link>
                </div>
              )}
            </div>
          </div>

          {/* Seasons and Episodes */}
          <div className="mt-12">
            <h2 className="mb-6 text-2xl font-bold text-white">Episodes</h2>

            {/* Season Selector */}
            {seasonsLoading ? (
              <div className="flex h-16 items-center justify-center">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
              </div>
            ) : seasons && seasons.length > 0 ? (
              <div className="mb-6">
                <div className="flex flex-wrap gap-2">
                  {seasons.map((season) => (
                    <button
                      key={season.id}
                      onClick={() => setSelectedSeasonId(season.id)}
                      className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
                        selectedSeasonId === season.id
                          ? 'bg-netflix-red text-white'
                          : 'bg-netflix-dark text-gray-300 hover:bg-netflix-gray'
                      }`}
                    >
                      Season {season.season_number}
                    </button>
                  ))}
                </div>
                {selectedSeasonId && seasons.find((s) => s.id === selectedSeasonId)?.title && (
                  <p className="mt-2 text-sm text-gray-400">
                    {seasons.find((s) => s.id === selectedSeasonId)?.title}
                  </p>
                )}
              </div>
            ) : null}

            {/* Episodes List */}
            {episodesLoading ? (
              <div className="flex h-32 items-center justify-center">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
              </div>
            ) : episodes && episodes.length > 0 ? (
              <div className="space-y-3">
                {episodes.map((episode) => (
                  <EpisodeCard key={episode.id} episode={episode} />
                ))}
              </div>
            ) : (
              <div className="flex h-32 flex-col items-center justify-center rounded-lg bg-netflix-dark">
                <p className="text-gray-400">No episodes found</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
