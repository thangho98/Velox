import { Link } from 'react-router'
import { LuFilm, LuPlay } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'
import type { Episode } from '@/types/api'

export function EpisodeCard({ episode }: { episode: Episode }) {
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
          to={`/watch/${episode.media_id}`}
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
        to={`/watch/${episode.media_id}`}
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
