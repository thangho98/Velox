import { useState } from 'react'
import { useMediaList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'

export function MoviesPage() {
  const [filters, setFilters] = useState({
    genre: '',
    year: '',
    sortBy: 'newest',
  })

  const { data: movies, isLoading } = useMediaList({
    type: 'movie',
    limit: 100,
  })

  // Filter and sort movies
  const filteredMovies = movies?.filter((movie) => {
    if (filters.genre && !movie.genres.includes(filters.genre)) return false
    if (filters.year) {
      const movieYear = movie.release_date ? new Date(movie.release_date).getFullYear() : null
      if (movieYear !== Number(filters.year)) return false
    }
    return true
  })

  // Sort movies
  const sortedMovies = filteredMovies?.sort((a, b) => {
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
  const genres = [...new Set(movies?.flatMap((m) => m.genres ?? []) || [])].filter(Boolean).sort()
  const years = [
    ...new Set(
      movies
        ?.map((m) => (m.release_date ? new Date(m.release_date).getFullYear() : null))
        .filter((y): y is number => y !== null && !Number.isNaN(y)) || [],
    ),
  ].sort((a, b) => b - a)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Movies</h1>
          <p className="text-gray-400">
            {sortedMovies?.length || 0} {sortedMovies?.length === 1 ? 'movie' : 'movies'}
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
            <span className="flex items-center gap-1 rounded-full bg-netflix-red/20 px-3 py-1 text-sm text-netflix-red">
              {filters.genre}
              <button
                onClick={() => setFilters({ ...filters, genre: '' })}
                className="ml-1 hover:text-white"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </span>
          )}
          {filters.year && (
            <span className="flex items-center gap-1 rounded-full bg-netflix-red/20 px-3 py-1 text-sm text-netflix-red">
              {filters.year}
              <button
                onClick={() => setFilters({ ...filters, year: '' })}
                className="ml-1 hover:text-white"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
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

      {/* Movies Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : sortedMovies?.length === 0 ? (
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
              d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
            />
          </svg>
          <p className="text-gray-400">
            {movies?.length === 0
              ? 'No movies found in your libraries.'
              : 'No movies match your filters.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {sortedMovies?.map((movie) => (
            <MediaCard
              key={movie.id}
              id={movie.id}
              title={movie.title}
              posterPath={movie.poster_path}
              type={movie.media_type === 'episode' ? 'series' : 'movie'}
              year={movie.release_date ? new Date(movie.release_date).getFullYear() : undefined}
              rating={movie.rating}
            />
          ))}
        </div>
      )}
    </div>
  )
}
