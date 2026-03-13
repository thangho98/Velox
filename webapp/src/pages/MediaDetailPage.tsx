import { useParams, Link } from 'react-router'
import { useState } from 'react'
import {
  useMediaWithFiles,
  useToggleFavorite,
  useProgress,
  useSeasons,
  useEpisodes,
  useRefreshMetadata,
} from '@/hooks/stores/useMedia'
import { useAuthStore } from '@/stores/auth'
import type { Episode } from '@/types/api'
import { LuChevronLeft, LuFilm, LuStar, LuPlay, LuHeart, LuRefreshCw } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'

export function MediaDetailPage() {
  const { id } = useParams<{ id: string }>()
  const mediaId = Number(id)
  const [selectedSeasonId, setSelectedSeasonId] = useState<number | null>(null)

  const { data: media, isLoading } = useMediaWithFiles(mediaId)
  const { data: progress } = useProgress(mediaId)
  const { mutate: toggleFavorite } = useToggleFavorite()
  const { mutate: refreshMetadata, isPending: isRefreshing } = useRefreshMetadata(mediaId)
  const { user } = useAuthStore()

  // Fetch seasons if this is a series (episode type with series_id)
  const isSeries = media?.media.media_type === 'episode' && media.media.series_id
  const seriesId = media?.media.series_id || mediaId

  const { data: seasons, isLoading: seasonsLoading } = useSeasons(seriesId)
  const { data: episodes, isLoading: episodesLoading } = useEpisodes(
    seriesId,
    selectedSeasonId || seasons?.[0]?.id || 0,
  )

  // Auto-select first season when seasons load
  if (seasons && seasons.length > 0 && !selectedSeasonId) {
    setSelectedSeasonId(seasons[0].id)
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
      </div>
    )
  }

  if (!media) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <h1 className="mb-4 text-4xl font-bold text-white">404</h1>
        <p className="mb-8 text-xl text-gray-400">Media not found</p>
        <Link to="/" className="text-netflix-blue hover:underline">
          Go back home
        </Link>
      </div>
    )
  }

  const primaryFile = media.files.find((f) => f.is_primary) || media.files[0]
  const progressPercent =
    progress && media?.media.duration
      ? Math.min(100, (progress.position / media.media.duration) * 100)
      : 0

  return (
    <div className="min-h-screen bg-netflix-black">
      {/* Backdrop */}
      {media.media.backdrop_path && (
        <div className="fixed inset-0 h-screen">
          <img
            src={tmdbImage(media.media.backdrop_path, 'w1280')!}
            alt={media.media.title}
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
              {media.media.poster_path ? (
                <img
                  src={tmdbImage(media.media.poster_path, 'w500')!}
                  alt={media.media.title}
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
              <h1 className="mb-2 text-3xl font-bold text-white lg:text-5xl">
                {media.media.title}
              </h1>

              <div className="mb-6 flex flex-wrap items-center gap-3 text-sm text-gray-400">
                {media.media.release_date && (
                  <span>{new Date(media.media.release_date).getFullYear()}</span>
                )}
                {media.media.rating > 0 && (
                  <>
                    <span className="text-gray-600">|</span>
                    <span className="flex items-center gap-1">
                      <LuStar size={16} className="text-yellow-500" />
                      {media.media.rating.toFixed(1)}
                    </span>
                  </>
                )}
                {media.media.imdb_rating > 0 && (
                  <>
                    <span className="text-gray-600">|</span>
                    <span className="rounded bg-yellow-500/20 px-2 py-0.5 text-xs font-medium text-yellow-400">
                      IMDb {media.media.imdb_rating.toFixed(1)}
                    </span>
                  </>
                )}
                {media.media.rt_score > 0 && (
                  <>
                    <span className="text-gray-600">|</span>
                    <span className="rounded bg-red-500/20 px-2 py-0.5 text-xs font-medium text-red-400">
                      RT {media.media.rt_score}%
                    </span>
                  </>
                )}
                {media.media.metacritic_score > 0 && (
                  <>
                    <span className="text-gray-600">|</span>
                    <span className="rounded bg-blue-500/20 px-2 py-0.5 text-xs font-medium text-blue-400">
                      MC {media.media.metacritic_score}
                    </span>
                  </>
                )}
                {isSeries && (
                  <>
                    <span className="text-gray-600">|</span>
                    <span className="rounded bg-purple-500/20 px-2 py-0.5 text-xs text-purple-400">
                      Series
                    </span>
                  </>
                )}
              </div>

              {media.media.overview && (
                <p className="mb-8 max-w-2xl text-lg leading-relaxed text-gray-300">
                  {media.media.overview}
                </p>
              )}

              {/* Actions */}
              <div className="mb-8 flex flex-wrap gap-4">
                {!isSeries ? (
                  // Movie play button
                  <Link
                    to={`/watch/${mediaId}`}
                    className="flex items-center gap-2 rounded bg-netflix-red px-8 py-3 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
                  >
                    <LuPlay size={20} />
                    {progress?.position && progress.position > 0 ? 'Resume' : 'Play'}
                  </Link>
                ) : (
                  // Series - Play latest episode or continue watching
                  <Link
                    to={`/watch/${episodes?.find((e) => !e.duration)?.id || episodes?.[0]?.id || mediaId}`}
                    className="flex items-center gap-2 rounded bg-netflix-red px-8 py-3 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
                  >
                    <LuPlay size={20} />
                    {progress?.position && progress.position > 0
                      ? 'Continue Watching'
                      : 'Play Latest'}
                  </Link>
                )}
                <button
                  onClick={() => toggleFavorite(mediaId)}
                  className={`flex items-center gap-2 rounded px-6 py-3 font-semibold transition-colors ${
                    progress?.is_favorite
                      ? 'bg-pink-600 text-white hover:bg-pink-700'
                      : 'bg-netflix-gray text-white hover:bg-gray-700'
                  }`}
                >
                  <LuHeart size={20} className={progress?.is_favorite ? 'fill-current' : ''} />
                  {progress?.is_favorite ? 'Favorited' : 'Favorite'}
                </button>
                {user?.is_admin && (
                  <button
                    onClick={() => refreshMetadata()}
                    disabled={isRefreshing}
                    className="flex items-center gap-2 rounded bg-netflix-gray px-6 py-3 font-semibold text-white transition-colors hover:bg-gray-700 disabled:opacity-50"
                  >
                    <LuRefreshCw size={20} className={isRefreshing ? 'animate-spin' : ''} />
                    {isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
                  </button>
                )}
              </div>

              {/* Progress bar */}
              {!isSeries && progress?.position && progress.position > 0 && (
                <div className="mb-8 max-w-md">
                  <div className="mb-2 flex justify-between text-sm text-gray-400">
                    <span>{progress.completed ? 'Completed' : 'Continue Watching'}</span>
                    <span>{Math.round(progressPercent)}%</span>
                  </div>
                  <div className="h-1 rounded-full bg-gray-700">
                    <div
                      className="h-1 rounded-full bg-netflix-red"
                      style={{ width: `${progressPercent}%` }}
                    />
                  </div>
                  <p className="mt-2 text-sm text-gray-400">
                    {formatTime(progress.position)} / {formatTime(media.media.duration || 0)}
                  </p>
                </div>
              )}

              {/* File info */}
              {!isSeries && primaryFile && (
                <div className="rounded-lg bg-netflix-dark/80 p-4 backdrop-blur-sm">
                  <h3 className="mb-3 font-semibold text-white">Media Details</h3>
                  <div className="grid grid-cols-2 gap-4 text-sm text-gray-400">
                    <div>
                      <span className="text-gray-500">Resolution:</span>{' '}
                      <span className="text-white">
                        {primaryFile.width}x{primaryFile.height}
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-500">Duration:</span>{' '}
                      <span className="text-white">{formatDuration(primaryFile.duration)}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Video:</span>{' '}
                      <span className="text-white">{primaryFile.video_codec}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Audio:</span>{' '}
                      <span className="text-white">{primaryFile.audio_codec}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Container:</span>{' '}
                      <span className="text-white">{primaryFile.container.toUpperCase()}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Size:</span>{' '}
                      <span className="text-white">{formatFileSize(primaryFile.file_size)}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Seasons and Episodes for Series */}
          {isSeries && (
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
          )}
        </div>
      </div>
    </div>
  )
}

function EpisodeCard({ episode }: { episode: Episode }) {
  return (
    <div className="group flex items-center gap-4 rounded-lg bg-netflix-dark/80 p-4 backdrop-blur-sm transition-colors hover:bg-netflix-gray">
      {/* Episode Number / Thumbnail */}
      <div className="relative flex h-20 w-32 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-netflix-black">
        {episode.still_path ? (
          <img
            src={tmdbImage(episode.still_path, 'w300')!}
            alt={episode.title}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <LuFilm size={32} className="text-gray-600" />
          </div>
        )}
        {/* Play overlay on hover */}
        <Link
          to={`/watch/${episode.id}`}
          className="absolute inset-0 flex items-center justify-center bg-black/60 opacity-0 transition-opacity group-hover:opacity-100"
        >
          <div className="rounded-full bg-netflix-red p-2">
            <LuPlay size={20} className="text-white" />
          </div>
        </Link>
      </div>

      {/* Episode Info */}
      <div className="flex-1">
        <div className="flex items-center gap-3">
          <span className="text-lg font-bold text-gray-500">{episode.episode_number}</span>
          <h3 className="font-semibold text-white">{episode.title}</h3>
        </div>
        {episode.overview && (
          <p className="mt-1 line-clamp-2 text-sm text-gray-400">{episode.overview}</p>
        )}
        {episode.duration && (
          <p className="mt-1 text-xs text-gray-500">{formatDuration(episode.duration)}</p>
        )}
      </div>

      {/* Play Button */}
      <Link
        to={`/watch/${episode.id}`}
        className="flex items-center gap-2 rounded-full bg-white/10 px-4 py-2 text-sm font-medium text-white opacity-0 transition-all group-hover:opacity-100 hover:bg-netflix-red"
      >
        <LuPlay size={16} />
        Play
      </Link>
    </div>
  )
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}h ${mins}m`
  }
  return `${mins}m`
}

function formatTime(seconds: number): string {
  if (!seconds) return '0:00'
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  if (hours > 0) {
    return `${hours}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function formatFileSize(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) {
    return `${gb.toFixed(2)} GB`
  }
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(2)} MB`
}
