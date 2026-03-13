import { useRef, useState, useCallback, useEffect } from 'react'
import { Link } from 'react-router'
import { MediaCard } from './MediaCard'
import { MediaRowSkeleton } from './Skeleton'
import type { MediaListItem, UserData } from '@/types/api'
import { LuChevronLeft, LuChevronRight } from 'react-icons/lu'

interface MediaRowProps {
  title: string
  seeAllLink?: string
  items?: MediaListItem[] | UserData[]
  isLoading?: boolean
  showProgress?: boolean
}

function isUserData(item: MediaListItem | UserData): item is UserData {
  return 'media_id' in item
}

export function MediaRow({ title, seeAllLink, items, isLoading, showProgress }: MediaRowProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(true)

  const updateScrollState = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    setCanScrollLeft(el.scrollLeft > 0)
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1)
  }, [])

  // Check scroll state on mount and whenever items change
  useEffect(() => {
    // rAF ensures the DOM has painted and scrollWidth is accurate
    const id = requestAnimationFrame(updateScrollState)
    return () => cancelAnimationFrame(id)
  }, [items, isLoading, updateScrollState])

  const scroll = useCallback((direction: 'left' | 'right') => {
    const el = scrollRef.current
    if (!el) return
    const amount = el.clientWidth * 0.75
    el.scrollBy({ left: direction === 'right' ? amount : -amount, behavior: 'smooth' })
  }, [])

  return (
    <section className="space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white lg:text-xl">{title}</h2>
        {seeAllLink && (
          <Link
            to={seeAllLink}
            className="text-sm text-netflix-light-gray hover:text-white transition-colors"
          >
            See all →
          </Link>
        )}
      </div>

      {/* Row */}
      <div className="group relative">
        {/* Left arrow */}
        {canScrollLeft && (
          <button
            onClick={() => scroll('left')}
            className="absolute left-0 top-0 z-10 flex h-full w-10 items-center justify-center bg-gradient-to-r from-netflix-black to-transparent opacity-0 transition-opacity group-hover:opacity-100"
            aria-label="Scroll left"
          >
            <LuChevronLeft size={24} className="text-white" />
          </button>
        )}

        {/* Right arrow */}
        {canScrollRight && (
          <button
            onClick={() => scroll('right')}
            className="absolute right-0 top-0 z-10 flex h-full w-10 items-center justify-center bg-gradient-to-l from-netflix-black to-transparent opacity-0 transition-opacity group-hover:opacity-100"
            aria-label="Scroll right"
          >
            <LuChevronRight size={24} className="text-white" />
          </button>
        )}

        {/* Scrollable content */}
        {isLoading ? (
          <MediaRowSkeleton />
        ) : (
          <div
            ref={scrollRef}
            onScroll={updateScrollState}
            className="hide-scrollbar flex gap-3 overflow-x-auto pb-2"
          >
            {items?.map((item) => {
              if (isUserData(item)) {
                return (
                  <div key={item.media_id} className="w-36 shrink-0 lg:w-40">
                    <MediaCard
                      id={item.media_id}
                      title={item.media_title || 'Unknown'}
                      posterPath={item.media_poster}
                      showProgress={showProgress}
                      progress={{
                        position: item.position,
                        duration: item.media_duration || 1,
                        completed: item.completed,
                        is_favorite: item.is_favorite,
                      }}
                    />
                  </div>
                )
              }
              return (
                <div key={item.id} className="w-36 shrink-0 lg:w-40">
                  <MediaCard
                    id={item.id}
                    title={item.title}
                    posterPath={item.poster_path}
                    type={item.type ?? (item.media_type === 'episode' ? 'series' : 'movie')}
                    seriesId={
                      'series_id' in item ? (item as { series_id?: number }).series_id : undefined
                    }
                    year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
                    rating={item.rating}
                    showProgress={showProgress}
                  />
                </div>
              )
            })}
          </div>
        )}
      </div>
    </section>
  )
}
