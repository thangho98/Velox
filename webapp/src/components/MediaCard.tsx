import { Link } from 'react-router'
import { useState } from 'react'
import { useToggleFavorite, useProgress } from '@/hooks/stores/useMedia'

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

  // Use external progress if provided, otherwise fetch it
  const rawProgress = externalProgress ?? fetchedProgress
  // Normalize progress data - UserData has media_duration, but we need duration
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

  const aspectClass = aspectRatio === 'poster' ? 'aspect-[2/3]' : 'aspect-video'
  const sizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  }

  return (
    <div
      className="group relative cursor-pointer"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Link to={`/media/${id}`} className="block">
        {/* Image Container */}
        <div
          className={`${aspectClass} relative overflow-hidden rounded-lg bg-netflix-dark transition-transform duration-300 group-hover:scale-105 ${sizeClasses[size]}`}
        >
          {posterPath ? (
            <img
              src={posterPath}
              alt={title}
              className="h-full w-full object-cover"
              loading="lazy"
            />
          ) : (
            <div className="flex h-full w-full flex-col items-center justify-center bg-netflix-gray p-4">
              <svg
                className="mb-2 h-10 w-10 text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                />
              </svg>
              <span className="line-clamp-2 text-center text-gray-500">{title}</span>
            </div>
          )}

          {/* Hover Overlay */}
          <div
            className={`absolute inset-0 flex flex-col items-center justify-center bg-black/60 transition-opacity duration-200 ${
              isHovered ? 'opacity-100' : 'opacity-0'
            }`}
          >
            {/* Play Button */}
            <div className="mb-4 rounded-full bg-netflix-red p-3 shadow-lg transition-transform hover:scale-110">
              <svg className="h-8 w-8 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" />
              </svg>
            </div>

            {/* Rating Badge */}
            {rating && rating > 0 && (
              <div className="flex items-center gap-1 rounded bg-yellow-500/90 px-2 py-0.5 text-xs font-medium text-black">
                <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
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
              <svg className="h-4 w-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
          )}
        </div>
      </Link>

      {/* Title and Year */}
      <div className="mt-2">
        <h3 className="truncate font-medium text-white group-hover:text-netflix-red transition-colors">
          {title}
        </h3>
        {year && <p className="text-xs text-gray-500">{year}</p>}
      </div>

      {/* Favorite Button */}
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
          <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
            <path
              fillRule="evenodd"
              d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      )}
    </div>
  )
}
