import { Link } from 'react-router'
import { useRecentlyWatched } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { useState } from 'react'

export function RecentlyWatchedPage() {
  const [limit, setLimit] = useState(20)
  const { data: recentlyWatched, isLoading } = useRecentlyWatched({ limit })

  // Filter options
  const [filter, setFilter] = useState<'all' | 'in_progress' | 'completed'>('all')

  const filteredItems = recentlyWatched?.filter((item) => {
    if (filter === 'in_progress') return !item.completed && item.position > 0
    if (filter === 'completed') return item.completed
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Continue Watching</h1>
          <p className="text-gray-400">
            {filteredItems?.length || 0} {filteredItems?.length === 1 ? 'item' : 'items'}
          </p>
        </div>

        {/* Filters */}
        <div className="flex gap-2">
          {(['all', 'in_progress', 'completed'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
                filter === f
                  ? 'bg-netflix-red text-white'
                  : 'bg-netflix-dark text-gray-300 hover:bg-netflix-gray'
              }`}
            >
              {f === 'all' && 'All'}
              {f === 'in_progress' && 'In Progress'}
              {f === 'completed' && 'Completed'}
            </button>
          ))}
        </div>
      </div>

      {/* Recently Watched Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : filteredItems?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <svg
            className="mb-4 h-12 w-12 text-gray-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p className="text-gray-400">
            {filter === 'all'
              ? 'No watch history yet'
              : filter === 'in_progress'
                ? 'No items in progress'
                : 'No completed items'}
          </p>
          <p className="text-sm text-gray-500">
            {filter === 'all' && 'Start watching something to see it here'}
          </p>
          <Link
            to="/"
            className="mt-4 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover"
          >
            Browse Content
          </Link>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {filteredItems?.map((item) => (
              <MediaCard
                key={item.media_id}
                id={item.media_id}
                title={item.media_title || 'Unknown'}
                posterPath={item.media_poster}
                progress={{
                  position: item.position,
                  duration: item.media_duration || 1,
                  completed: item.completed,
                  is_favorite: item.is_favorite,
                }}
                showProgress
              />
            ))}
          </div>

          {/* Load More */}
          {recentlyWatched && recentlyWatched.length >= limit && (
            <div className="flex justify-center pt-4">
              <button
                onClick={() => setLimit((prev) => prev + 20)}
                className="rounded-lg bg-netflix-dark px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-gray"
              >
                Load More
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
