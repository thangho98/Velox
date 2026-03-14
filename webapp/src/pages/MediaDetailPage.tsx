import { useParams, Link } from 'react-router'
import {
  useMediaWithFiles,
  useToggleFavorite,
  useProgress,
  useUpdateProgress,
  useDismissProgress,
  useRefreshMetadata,
  useSubtitles,
} from '@/hooks/stores/useMedia'
import { useAuthStore } from '@/stores/auth'
import { usePlayerStore } from '@/stores/player'
import {
  LuChevronLeft,
  LuFilm,
  LuStar,
  LuPlay,
  LuHeart,
  LuRefreshCw,
  LuCheck,
} from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'

export function MediaDetailPage() {
  const { id } = useParams<{ id: string }>()
  const mediaId = Number(id)

  const { data: media, isLoading } = useMediaWithFiles(mediaId)
  const { data: progress } = useProgress(mediaId)
  const { mutate: toggleFavorite } = useToggleFavorite()
  const { mutate: updateProgress } = useUpdateProgress()
  const { mutate: dismissProgress } = useDismissProgress()
  const { mutate: refreshMetadata, isPending: isRefreshing } = useRefreshMetadata(mediaId)
  const { user } = useAuthStore()
  const { subtitleLanguage, setSubtitleLanguage } = usePlayerStore()
  const { data: subtitles = [] } = useSubtitles(mediaId)

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
  const duration = primaryFile?.duration || media.media.duration || 0
  const progressPercent =
    progress && duration > 0 ? Math.min(100, (progress.position / duration) * 100) : 0
  const canResume = !!progress && progress.position > 0 && !progress.completed

  const handleToggleWatched = () => {
    if (progress?.completed) {
      dismissProgress(mediaId)
      return
    }

    updateProgress({
      mediaId,
      data: {
        position: duration > 0 ? duration : 0,
        completed: true,
      },
    })
  }

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

              {/* Year · Duration · Ends at · Ratings */}
              <div className="mb-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-gray-400">
                {media.media.release_date && (
                  <span>{new Date(media.media.release_date).getFullYear()}</span>
                )}
                {primaryFile && primaryFile.duration > 0 && (
                  <span>{formatDuration(primaryFile.duration)}</span>
                )}
                {duration > 0 && (
                  <span>Ends at {getEndTime(duration - (progress?.position || 0))}</span>
                )}
                {media.media.rating > 0 && (
                  <span className="flex items-center gap-1">
                    <LuStar size={14} className="text-yellow-500" />
                    {media.media.rating.toFixed(1)}
                  </span>
                )}
                {media.media.imdb_rating > 0 && (
                  <span className="rounded bg-yellow-500/20 px-1.5 py-0.5 text-xs font-medium text-yellow-400">
                    IMDb {media.media.imdb_rating.toFixed(1)}
                  </span>
                )}
                {media.media.rt_score > 0 && (
                  <span className="rounded bg-red-500/20 px-1.5 py-0.5 text-xs font-medium text-red-400">
                    RT {media.media.rt_score}%
                  </span>
                )}
                {media.media.metacritic_score > 0 && (
                  <span className="rounded bg-blue-500/20 px-1.5 py-0.5 text-xs font-medium text-blue-400">
                    MC {media.media.metacritic_score}
                  </span>
                )}
              </div>

              {/* Media info line */}
              {primaryFile &&
                (primaryFile.video_codec ||
                  primaryFile.audio_codec ||
                  primaryFile.file_size > 0) && (
                  <div className="mb-5 flex flex-wrap items-center gap-x-5 gap-y-1 text-sm">
                    {primaryFile.video_codec && (
                      <span>
                        <span className="text-gray-500">Video</span>{' '}
                        <span className="text-gray-300">
                          {primaryFile.height > 0 ? `${primaryFile.height}p ` : ''}
                          {primaryFile.video_codec.toUpperCase()}
                        </span>
                      </span>
                    )}
                    {primaryFile.audio_codec && (
                      <span>
                        <span className="text-gray-500">Audio</span>{' '}
                        <span className="text-gray-300">
                          {primaryFile.audio_codec.toUpperCase()}
                        </span>
                      </span>
                    )}
                    {primaryFile.container && (
                      <span>
                        <span className="text-gray-500">Container</span>{' '}
                        <span className="text-gray-300">{primaryFile.container.toUpperCase()}</span>
                      </span>
                    )}
                    {primaryFile.file_size > 0 && (
                      <span>
                        <span className="text-gray-500">Size</span>{' '}
                        <span className="text-gray-300">
                          {formatFileSize(primaryFile.file_size)}
                        </span>
                      </span>
                    )}
                  </div>
                )}

              {media.media.overview && (
                <p className="mb-6 max-w-2xl text-base leading-relaxed text-gray-300">
                  {media.media.overview}
                </p>
              )}

              {/* Actions */}
              <div className="mb-6 flex flex-wrap items-center gap-3">
                {/* Primary action */}
                {canResume ? (
                  <>
                    <Link
                      to={`/watch/${mediaId}`}
                      className="flex items-center gap-2 rounded bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-700"
                    >
                      <LuPlay size={18} className="fill-current" />
                      Resume
                    </Link>
                    <Link
                      to={`/watch/${mediaId}?t=0`}
                      className="flex items-center gap-2 rounded bg-white/10 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-white/20"
                    >
                      <LuPlay size={16} />
                      From Beginning
                    </Link>
                  </>
                ) : (
                  <Link
                    to={`/watch/${mediaId}`}
                    className="flex items-center gap-2 rounded bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-700"
                  >
                    <LuPlay size={18} className="fill-current" />
                    Play
                  </Link>
                )}

                {/* Icon actions */}
                <button
                  onClick={handleToggleWatched}
                  className={`p-2 transition-colors ${
                    progress?.completed ? 'text-green-500' : 'text-gray-400 hover:text-white'
                  }`}
                  title={progress?.completed ? 'Mark as unwatched' : 'Mark as watched'}
                >
                  <LuCheck size={22} />
                </button>

                <button
                  onClick={() => toggleFavorite(mediaId)}
                  className={`p-2 transition-colors ${
                    progress?.is_favorite ? 'text-pink-500' : 'text-gray-400 hover:text-white'
                  }`}
                  title={progress?.is_favorite ? 'Unfavorite' : 'Favorite'}
                >
                  <LuHeart size={22} className={progress?.is_favorite ? 'fill-current' : ''} />
                </button>

                {user?.is_admin && (
                  <button
                    onClick={() => refreshMetadata()}
                    disabled={isRefreshing}
                    className="p-2 text-gray-400 transition-colors hover:text-white disabled:opacity-50"
                    title="Refresh metadata"
                  >
                    <LuRefreshCw size={22} className={isRefreshing ? 'animate-spin' : ''} />
                  </button>
                )}

                {/* Subtitle selector */}
                {subtitles.length > 0 && (
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-gray-400">Subtitles</span>
                    <select
                      value={subtitleLanguage ?? ''}
                      onChange={(e) => setSubtitleLanguage(e.target.value || null, null)}
                      className="rounded-full bg-[#2a2a2a] px-4 py-2 pr-8 text-sm text-white outline-none appearance-none cursor-pointer hover:bg-[#333] transition-colors"
                      style={{
                        backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`,
                        backgroundRepeat: 'no-repeat',
                        backgroundPosition: 'right 12px center',
                      }}
                    >
                      <option value="">Off</option>
                      {subtitles
                        .filter((s) => !s.is_image)
                        .map((s) => (
                          <option key={s.id} value={s.language}>
                            {s.label || `${s.language} (${s.format.toUpperCase()})`}
                          </option>
                        ))}
                    </select>
                  </div>
                )}
              </div>

              {/* Progress bar */}
              {progress != null && progress.position > 0 && (
                <div className="mb-8 max-w-md">
                  <div className="flex items-center gap-3">
                    <div className="h-1 flex-1 rounded-full bg-gray-700">
                      <div
                        className="h-1 rounded-full bg-green-500"
                        style={{ width: `${progressPercent}%` }}
                      />
                    </div>
                    <span className="shrink-0 text-sm text-gray-400">
                      {progress.completed
                        ? 'Watched'
                        : `${formatDuration(Math.max(0, duration - progress.position))} remaining`}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
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

function getEndTime(remainingSeconds: number): string {
  const end = new Date(Date.now() + remainingSeconds * 1000)
  return end.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
}

function formatFileSize(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) {
    return `${gb.toFixed(2)} GB`
  }
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(2)} MB`
}
