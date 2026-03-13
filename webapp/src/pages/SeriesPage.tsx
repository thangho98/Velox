import { useState } from 'react'
import { useMediaList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { LuX, LuTv } from 'react-icons/lu'

export function SeriesPage() {
  const [filters, setFilters] = useState({
    genre: '',
    year: '',
    sortBy: 'newest',
  })

  const { data: series, isLoading } = useMediaList({
    type: 'episode',
    limit: 100,
  })

  // Filter and sort series
  const filteredSeries = series?.filter((s) => {
    if (filters.genre && !s.genres.includes(filters.genre)) return false
    if (filters.year) {
      const seriesYear = s.release_date ? new Date(s.release_date).getFullYear() : null
      if (seriesYear !== Number(filters.year)) return false
    }
    return true
  })

  // Sort series
  const sortedSeries = filteredSeries?.sort((a, b) => {
    switch (filters.sortBy) {
      case 'newest':
        return new Date(b.release_date || 0).getTime() - new Date(a.release_date || 0).getTime()
      case 'oldest':
        return new Date(a.release_date || 0).getTime() - new Date(b.release_date || 0).getTime()
      case 'rating':
        return (b.rating || 0) - (a.rating || 0)
      case 'title':
        return a.title.localeCompare(b.title)
      default:
        return 0
    }
  })

  // Get unique genres and years for filters
  const genres = [...new Set(series?.flatMap((s) => s.genres ?? []) || [])].filter(Boolean).sort()
  const years = [
    ...new Set(
      series
        ?.map((s) => (s.release_date ? new Date(s.release_date).getFullYear() : null))
        .filter((y): y is number => y !== null && !Number.isNaN(y)) || [],
    ),
  ].sort((a, b) => b - a)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Series</h1>
          <p className="text-gray-400">
            {sortedSeries?.length || 0} {sortedSeries?.length === 1 ? 'series' : 'series'}
          </p>
        </div>

        {/* Filters */}
        <div className="flex flex-wrap gap-3">
          {/* Genre Filter */}
          <select
            value={filters.genre}
            onChange={(e) => setFilters({ ...filters, genre: e.target.value })}
            className="rounded-lg bg-netflix-dark px-4 py-2 text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          >
            <option value="">All Genres</option>
            {genres.map((genre) => (
              <option key={genre} value={genre}>
                {genre}
              </option>
            ))}
          </select>

          {/* Year Filter */}
          <select
            value={filters.year}
            onChange={(e) => setFilters({ ...filters, year: e.target.value })}
            className="rounded-lg bg-netflix-dark px-4 py-2 text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          >
            <option value="">All Years</option>
            {years.map((year) => (
              <option key={year} value={year}>
                {year}
              </option>
            ))}
          </select>

          {/* Sort */}
          <select
            value={filters.sortBy}
            onChange={(e) => setFilters({ ...filters, sortBy: e.target.value })}
            className="rounded-lg bg-netflix-dark px-4 py-2 text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          >
            <option value="newest">Newest First</option>
            <option value="oldest">Oldest First</option>
            <option value="rating">Highest Rated</option>
            <option value="title">Title A-Z</option>
          </select>
        </div>
      </div>

      {/* Active Filters */}
      {(filters.genre || filters.year) && (
        <div className="flex flex-wrap gap-2">
          {filters.genre && (
            <span className="flex items-center gap-1 rounded-full bg-purple-500/20 px-3 py-1 text-sm text-purple-400">
              {filters.genre}
              <button
                onClick={() => setFilters({ ...filters, genre: '' })}
                className="ml-1 hover:text-white"
              >
                <LuX size={16} />
              </button>
            </span>
          )}
          {filters.year && (
            <span className="flex items-center gap-1 rounded-full bg-purple-500/20 px-3 py-1 text-sm text-purple-400">
              {filters.year}
              <button
                onClick={() => setFilters({ ...filters, year: '' })}
                className="ml-1 hover:text-white"
              >
                <LuX size={16} />
              </button>
            </span>
          )}
          <button
            onClick={() => setFilters({ genre: '', year: '', sortBy: 'newest' })}
            className="text-sm text-gray-400 hover:text-white"
          >
            Clear all
          </button>
        </div>
      )}

      {/* Series Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-purple-500 border-t-transparent" />
        </div>
      ) : sortedSeries?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuTv size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">
            {series?.length === 0
              ? 'No series found in your libraries.'
              : 'No series match your filters.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {sortedSeries?.map((s) => (
            <MediaCard
              key={s.id}
              id={s.id}
              title={s.title}
              posterPath={s.poster_path}
              type="series"
              year={s.release_date ? new Date(s.release_date).getFullYear() : undefined}
              rating={s.rating}
            />
          ))}
        </div>
      )}
    </div>
  )
}
