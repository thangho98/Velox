import { useState } from 'react'
import { useSeriesList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { LuX, LuTv } from 'react-icons/lu'
import type { Series } from '@/types/api'

export function SeriesPage() {
  const [filters, setFilters] = useState({
    year: '',
    sortBy: 'newest',
  })
  const { data: series, isLoading } = useSeriesList({ limit: 100 })

  // Filter
  const filteredSeries = series?.filter((s: Series) => {
    if (filters.year) {
      const year = s.first_air_date ? new Date(s.first_air_date).getFullYear() : null
      if (year !== Number(filters.year)) return false
    }
    return true
  })

  // Sort
  const sortedSeries = filteredSeries?.sort((a: Series, b: Series) => {
    switch (filters.sortBy) {
      case 'newest':
        return new Date(b.first_air_date || 0).getTime() - new Date(a.first_air_date || 0).getTime()
      case 'oldest':
        return new Date(a.first_air_date || 0).getTime() - new Date(b.first_air_date || 0).getTime()
      case 'title':
        return a.title.localeCompare(b.title)
      default:
        return 0
    }
  })

  // Years for filter
  const years = [
    ...new Set(
      series
        ?.map((s: Series) => (s.first_air_date ? new Date(s.first_air_date).getFullYear() : null))
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
            <option value="title">Title A-Z</option>
          </select>
        </div>
      </div>

      {/* Active Filters */}
      {filters.year && (
        <div className="flex flex-wrap gap-2">
          <span className="flex items-center gap-1 rounded-full bg-purple-500/20 px-3 py-1 text-sm text-purple-400">
            {filters.year}
            <button
              onClick={() => setFilters({ ...filters, year: '' })}
              className="ml-1 hover:text-white"
            >
              <LuX size={16} />
            </button>
          </span>
          <button
            onClick={() => setFilters({ year: '', sortBy: 'newest' })}
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
          {sortedSeries?.map((s: Series) => (
            <MediaCard
              key={s.id}
              id={s.id}
              title={s.title}
              posterPath={s.poster_path}
              type="series"
              seriesId={s.id}
              year={s.first_air_date ? new Date(s.first_air_date).getFullYear() : undefined}
            />
          ))}
        </div>
      )}
    </div>
  )
}
