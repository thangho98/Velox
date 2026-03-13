import { Link } from 'react-router'
import { useState } from 'react'
import { LuFilm, LuPlay, LuStar, LuCheck, LuHeart } from 'react-icons/lu'
import { useToggleFavorite, useProgress } from '@/hooks/stores/useMedia'
import { tmdbImage } from '@/lib/image'

interface MediaCardProps {
  id: number
  title: string
  posterPath?: string | null
  type?: 'movie' | 'series'
  year?: number
  rating?: number
  progress?: {
    position: number
    duration: number
    completed: boolean
    is_favorite: boolean
  } | null
  showProgress?: boolean
  showFavorite?: boolean
  aspectRatio?: 'poster' | 'wide'
  size?: 'sm' | 'md' | 'lg'
}

export function MediaCard({
  id,
  title,
  posterPath,
  type = 'movie',
  year,
  rating,
  progress: externalProgress,
  showProgress = true,
  showFavorite = true,
  aspectRatio = 'poster',
  size = 'md',
}: MediaCardProps) {
  const [isHovered, setIsHovered] = useState(false)
  const { mutate: toggleFavorite } = useToggleFavorite()
  const { data: fetchedProgress } = useProgress(showProgress ? id : 0)

  const rawProgress = externalProgress ?? fetchedProgress
  const progress = rawProgress
    ? {
        position: rawProgress.position,
        duration:
          'duration' in rawProgress
            ? rawProgress.duration
            : 'media_duration' in rawProgress
              ? (rawProgress as { media_duration?: number }).media_duration || 1
              : 1,
        completed: rawProgress.completed,
        is_favorite: rawProgress.is_favorite,
      }
    : null
  const hasProgress = progress && progress.position > 0 && !progress.completed
  const progressPercent = progress?.duration
    ? Math.min(100, (progress.position / progress.duration) * 100)
    : 0

  const posterSrc = tmdbImage(posterPath, aspectRatio === 'poster' ? 'w500' : 'w780')
  const aspectClass = aspectRatio === 'poster' ? 'aspect-[2/3]' : 'aspect-video'
  const sizeClasses = { sm: 'text-xs', md: 'text-sm', lg: 'text-base' }

  return (
    <div
      className="group relative cursor-pointer"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Link to={`/media/${id}`} className="block">
        <div
          className={`${aspectClass} relative overflow-hidden rounded-lg bg-netflix-dark transition-transform duration-300 group-hover:scale-105 ${sizeClasses[size]}`}
        >
          {posterSrc ? (
            <img
              src={posterSrc}
              alt={title}
              className="h-full w-full object-cover"
              loading="lazy"
            />
          ) : (
            <div className="flex h-full w-full flex-col items-center justify-center bg-netflix-gray p-4">
              <LuFilm size={40} className="mb-2 text-gray-600" />
              <span className="line-clamp-2 text-center text-gray-500">{title}</span>
            </div>
          )}

          {/* Hover Overlay */}
          <div
            className={`absolute inset-0 flex flex-col items-center justify-center bg-black/60 transition-opacity duration-200 ${
              isHovered ? 'opacity-100' : 'opacity-0'
            }`}
          >
            <div className="mb-4 rounded-full bg-netflix-red p-3 shadow-lg transition-transform hover:scale-110">
              <LuPlay size={32} className="text-white fill-white" />
            </div>
            {rating && rating > 0 && (
              <div className="flex items-center gap-1 rounded bg-yellow-500/90 px-2 py-0.5 text-xs font-medium text-black">
                <LuStar size={12} className="fill-black" />
                {rating.toFixed(1)}
              </div>
            )}
          </div>

          {/* Type Badge */}
          <div className="absolute left-2 top-2">
            <span
              className={`rounded px-1.5 py-0.5 text-xs font-medium ${
                type === 'movie' ? 'bg-blue-500/90 text-white' : 'bg-purple-500/90 text-white'
              }`}
            >
              {type === 'movie' ? 'Movie' : 'Series'}
            </span>
          </div>

          {/* Progress Bar */}
          {showProgress && hasProgress && (
            <div className="absolute bottom-0 left-0 right-0 bg-black/60 p-2">
              <div className="h-1 rounded-full bg-gray-600">
                <div
                  className="h-1 rounded-full bg-netflix-red transition-all"
                  style={{ width: `${progressPercent}%` }}
                />
              </div>
              <p className="mt-1 text-xs text-gray-300">{Math.round(progressPercent)}% watched</p>
            </div>
          )}

          {/* Completed Badge */}
          {showProgress && progress?.completed && (
            <div className="absolute bottom-2 right-2 rounded-full bg-green-500 p-1.5">
              <LuCheck size={16} className="text-white" />
            </div>
          )}
        </div>
      </Link>

      <div className="mt-2">
        <h3 className="truncate font-medium text-white group-hover:text-netflix-red transition-colors">
          {title}
        </h3>
        {year && <p className="text-xs text-gray-500">{year}</p>}
      </div>

      {showFavorite && (
        <button
          onClick={(e) => {
            e.preventDefault()
            e.stopPropagation()
            toggleFavorite(id)
          }}
          className={`absolute right-2 top-2 rounded-full p-1.5 transition-all ${
            progress?.is_favorite
              ? 'bg-pink-500 text-white opacity-100'
              : 'bg-black/50 text-gray-400 opacity-0 group-hover:opacity-100 hover:bg-pink-500/80 hover:text-white'
          }`}
        >
          <LuHeart size={16} className={progress?.is_favorite ? 'fill-white' : ''} />
        </button>
      )}
    </div>
  )
}
