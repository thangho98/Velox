import { Link } from 'react-router'
import { useState, useRef, useEffect } from 'react'
import { LuFilm, LuPlay, LuStar, LuCheck, LuHeart } from 'react-icons/lu'
import { useToggleFavorite, useProgress } from '@/hooks/stores/useMedia'
import { useCinemaSettings } from '@/hooks/stores/useSettings'
import { tmdbImage } from '@/lib/image'
import { api } from '@/lib/fetch'

interface MediaCardProps {
  id: number
  title: string
  posterPath?: string | null
  type?: 'movie' | 'series'
  seriesId?: number // ADD — series.id for routing
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
  directPlay?: boolean // browse mode: skip detail page, go straight to /watch/{id}
}

export function MediaCard({
  id,
  title,
  posterPath,
  type = 'movie',
  seriesId,
  year,
  rating,
  progress: externalProgress,
  showProgress = true,
  showFavorite = true,
  aspectRatio = 'poster',
  size = 'md',
  directPlay = false,
}: MediaCardProps) {
  const [isHovered, setIsHovered] = useState(false)
  const [trailerKey, setTrailerKey] = useState<string | null>(null)
  const [showTrailer, setShowTrailer] = useState(false)
  const hoverTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { mutate: toggleFavorite } = useToggleFavorite()
  const { data: cinemaSettings } = useCinemaSettings()
  const cinemaEnabled = cinemaSettings?.enabled ?? false

  useEffect(() => {
    if (!isHovered || !cinemaEnabled) {
      if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current)
      setShowTrailer(false)
      return
    }
    // After 2s hover, fetch trailer
    hoverTimerRef.current = setTimeout(async () => {
      if (!trailerKey) {
        try {
          const endpoint = isSeries ? `/series/${seriesId || id}/cinema` : `/media/${id}/cinema`
          const data = await api.get<{ items: Array<{ type: string; url: string }> }>(endpoint)
          const trailer = data.items?.find((i) => i.type === 'trailer')
          if (trailer) {
            const match = trailer.url.match(/embed\/([^?]+)/)
            if (match) setTrailerKey(match[1])
          }
        } catch {
          // No trailers available
        }
      }
      setShowTrailer(true)
    }, 2000)
    return () => {
      if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current)
    }
  }, [isHovered, id, trailerKey, cinemaEnabled])

  // Series cards: id is series.id, NOT media_id — skip progress/favorite
  const isSeries = type === 'series'
  const { data: fetchedProgress } = useProgress(!isSeries && showProgress ? id : 0)

  // Force no progress/favorite for series cards
  const effectiveShowProgress = isSeries ? false : showProgress
  const effectiveShowFavorite = isSeries ? false : showFavorite

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

  // Type-aware link — directPlay skips detail page and goes straight to watch
  const linkTo = directPlay
    ? `/watch/${id}`
    : isSeries
      ? `/series/${seriesId || id}`
      : `/movies/${id}`

  return (
    <div
      className="group relative cursor-pointer"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Link to={linkTo} className="block">
        <div
          className={`${aspectClass} relative overflow-hidden rounded-lg bg-netflix-dark transition-transform duration-300 group-hover:scale-105 ${sizeClasses[size]}`}
        >
          {/* Trailer preview or poster */}
          {showTrailer && trailerKey ? (
            <div className="absolute inset-0 overflow-hidden">
              <iframe
                src={`https://www.youtube.com/embed/${trailerKey}?autoplay=1&mute=1&controls=0&showinfo=0&rel=0&modestbranding=1&iv_load_policy=3&disablekb=1&fs=0`}
                allow="autoplay; encrypted-media"
                style={{
                  border: 'none',
                  position: 'absolute',
                  top: '50%',
                  left: '50%',
                  transform: 'translate(-50%, -50%)',
                  width: '200%',
                  height: '200%',
                  pointerEvents: 'none',
                }}
              />
            </div>
          ) : posterSrc ? (
            <img
              src={posterSrc}
              alt={title}
              className="h-full w-full object-cover"
              loading="lazy"
              width={500}
              height={750}
            />
          ) : (
            <div className="flex h-full w-full flex-col items-center justify-center bg-netflix-gray p-4">
              <LuFilm size={40} className="mb-2 text-gray-600" />
              <span className="line-clamp-2 text-center text-gray-500">{title}</span>
            </div>
          )}

          {/* Hover Overlay (hidden during trailer) */}
          {!(showTrailer && trailerKey) && (
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
          )}

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
          {effectiveShowProgress && hasProgress && (
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
          {effectiveShowProgress && progress?.completed && (
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

      {effectiveShowFavorite && (
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
