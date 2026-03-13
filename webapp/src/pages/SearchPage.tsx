import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router'
import { useMediaList, useSeriesList, useSeriesSearch } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { LuSearch, LuX } from 'react-icons/lu'
import type { Series } from '@/types/api'

const DEBOUNCE_MS = 300

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const initialQuery = searchParams.get('q') || ''

  const [query, setQuery] = useState(initialQuery)
  const [debouncedQuery, setDebouncedQuery] = useState(initialQuery)
  const [filters, setFilters] = useState({
    type: searchParams.get('type') || '',
    genre: searchParams.get('genre') || '',
  })

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query)
      // Update URL params
      const newParams = new URLSearchParams(searchParams)
      if (query) {
        newParams.set('q', query)
      } else {
        newParams.delete('q')
      }
      if (filters.type) {
        newParams.set('type', filters.type)
      } else {
        newParams.delete('type')
      }
      if (filters.genre) {
        newParams.set('genre', filters.genre)
      } else {
        newParams.delete('genre')
      }
      setSearchParams(newParams, { replace: true })
    }, DEBOUNCE_MS)

    return () => clearTimeout(timer)
  }, [query, filters, searchParams, setSearchParams])

  // Fetch data based on view
  const { data: allMovies, isLoading: moviesLoading } = useMediaList({
    type: 'movie',
    limit: 100,
  })
  const { data: allSeries, isLoading: seriesLoading } = useSeriesList({ limit: 100 })

  // Search results
  const { data: seriesSearchResults, isLoading: seriesSearchLoading } = useSeriesSearch(
    debouncedQuery,
    20,
  )

  const isLoading = moviesLoading || seriesLoading || seriesSearchLoading

  // Filter movies based on search
  const filteredMovies = allMovies?.filter((item) => {
    if (debouncedQuery) {
      const searchLower = debouncedQuery.toLowerCase()
      const titleMatch = item.title.toLowerCase().includes(searchLower)
      const overviewMatch = item.overview?.toLowerCase().includes(searchLower)
      const genreMatch = item.genres.some((g) => g.toLowerCase().includes(searchLower))
      if (!titleMatch && !overviewMatch && !genreMatch) return false
    }
    // Genre filter (movies only)
    if (filters.genre && !item.genres.includes(filters.genre)) return false
    return true
  })

  // Filter series based on search (client-side when no server search)
  const filteredSeries = debouncedQuery ? seriesSearchResults || [] : allSeries || []

  // Get unique genres from movie results
  const genres = [...new Set(allMovies?.flatMap((m) => m.genres) || [])].sort()

  const clearSearch = useCallback(() => {
    setQuery('')
    setDebouncedQuery('')
    setFilters({ type: '', genre: '' })
  }, [])

  // Determine what to show based on filters
  const showMovies = !filters.type || filters.type === 'movie'
  const showSeries = !filters.type || filters.type === 'series'

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="space-y-4">
        <h1 className="text-3xl font-bold text-white">Search</h1>

        {/* Search Input */}
        <div className="relative max-w-2xl">
          <div className="absolute left-4 top-1/2 -translate-y-1/2">
            <LuSearch size={20} className="text-gray-500" />
          </div>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search movies, series, genres..."
            className="w-full rounded-lg bg-netflix-dark py-3 pl-12 pr-10 text-white placeholder-gray-500 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          />
          {query && (
            <button
              onClick={clearSearch}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
            >
              <LuX size={20} />
            </button>
          )}
        </div>

        {/* Filters */}
        <div className="flex flex-wrap gap-3">
          {/* Type Filter */}
          <select
            value={filters.type}
            onChange={(e) => setFilters({ ...filters, type: e.target.value })}
            className="rounded-lg bg-netflix-dark px-4 py-2 text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          >
            <option value="">All Types</option>
            <option value="movie">Movies</option>
            <option value="series">Series</option>
          </select>

          {/* Genre Filter - only show when viewing movies */}
          {showMovies && (
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
          )}
        </div>

        {/* Active Filters */}
        {(filters.type || filters.genre) && (
          <div className="flex flex-wrap gap-2">
            {filters.type && (
              <span className="flex items-center gap-1 rounded-full bg-netflix-red/20 px-3 py-1 text-sm text-netflix-red">
                {filters.type === 'movie' ? 'Movies' : 'Series'}
                <button
                  onClick={() => setFilters({ ...filters, type: '' })}
                  className="ml-1 hover:text-white"
                >
                  <LuX size={16} />
                </button>
              </span>
            )}
            {filters.genre && (
              <span className="flex items-center gap-1 rounded-full bg-netflix-red/20 px-3 py-1 text-sm text-netflix-red">
                {filters.genre}
                <button
                  onClick={() => setFilters({ ...filters, genre: '' })}
                  className="ml-1 hover:text-white"
                >
                  <LuX size={16} />
                </button>
              </span>
            )}
            <button
              onClick={() => setFilters({ type: '', genre: '' })}
              className="text-sm text-gray-400 hover:text-white"
            >
              Clear filters
            </button>
          </div>
        )}
      </div>

      {/* Results */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : debouncedQuery || filters.type || filters.genre ? (
        // Search results
        <div className="space-y-8">
          {/* Movies Results */}
          {showMovies && filteredMovies && filteredMovies.length > 0 && (
            <section>
              <h2 className="mb-4 text-lg font-semibold text-white">
                Movies ({filteredMovies.length})
              </h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {filteredMovies.map((item) => (
                  <MediaCard
                    key={item.id}
                    id={item.id}
                    title={item.title}
                    posterPath={item.poster_path}
                    type="movie"
                    year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
                    rating={item.rating}
                  />
                ))}
              </div>
            </section>
          )}

          {/* Series Results */}
          {showSeries && filteredSeries && filteredSeries.length > 0 && (
            <section>
              <h2 className="mb-4 text-lg font-semibold text-white">
                Series ({filteredSeries.length})
              </h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {filteredSeries.map((s: Series) => (
                  <MediaCard
                    key={`series-${s.id}`}
                    id={s.id}
                    title={s.title}
                    posterPath={s.poster_path}
                    type="series"
                    seriesId={s.id}
                    year={s.first_air_date ? new Date(s.first_air_date).getFullYear() : undefined}
                  />
                ))}
              </div>
            </section>
          )}

          {/* No results */}
          {((showMovies && !filteredMovies?.length) ||
            (showSeries && !filteredSeries?.length && filters.type === 'series')) &&
            !((showMovies && filteredMovies?.length) || (showSeries && filteredSeries?.length)) && (
              <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
                <LuSearch size={48} className="mb-4 text-gray-600" />
                <p className="text-gray-400">No results found for &quot;{debouncedQuery}&quot;</p>
                <button onClick={clearSearch} className="mt-2 text-netflix-red hover:underline">
                  Clear search
                </button>
              </div>
            )}
        </div>
      ) : (
        // Default view - show all content
        <div className="space-y-8">
          {/* Movies Section */}
          {allMovies && allMovies.length > 0 && (
            <section>
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-xl font-semibold text-white">Movies</h2>
                <button
                  onClick={() => setFilters({ ...filters, type: 'movie' })}
                  className="text-sm text-netflix-red hover:underline"
                >
                  View all
                </button>
              </div>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {allMovies.slice(0, 6).map((movie) => (
                  <MediaCard
                    key={movie.id}
                    id={movie.id}
                    title={movie.title}
                    posterPath={movie.poster_path}
                    type="movie"
                    year={
                      movie.release_date ? new Date(movie.release_date).getFullYear() : undefined
                    }
                    rating={movie.rating}
                  />
                ))}
              </div>
            </section>
          )}

          {/* Series Section */}
          {allSeries && allSeries.length > 0 && (
            <section>
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-xl font-semibold text-white">Series</h2>
                <button
                  onClick={() => setFilters({ ...filters, type: 'series' })}
                  className="text-sm text-netflix-red hover:underline"
                >
                  View all
                </button>
              </div>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {allSeries.slice(0, 6).map((s: Series) => (
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
            </section>
          )}

          {/* Empty State */}
          {allMovies?.length === 0 && allSeries?.length === 0 && (
            <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
              <LuSearch size={48} className="mb-4 text-gray-600" />
              <p className="text-gray-400">Start typing to search your media library</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
