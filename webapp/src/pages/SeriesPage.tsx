import { useSeriesList, useGenres } from '@/hooks/stores/useMedia'
import { useFilterParams } from '@/hooks/useFilterParams'
import { MediaCard } from '@/components/MediaCard'
import { FilterBar } from '@/components/FilterBar'
import { AlphaIndex, useAlphaScroll } from '@/components/AlphaIndex'
import { LuTv } from 'react-icons/lu'
import type { SeriesListItem } from '@/types/api'

export function SeriesPage() {
  const { filters, setGenre, setYear, setSort, clearFilters, hasActiveFilters } = useFilterParams()

  const { data: series, isLoading } = useSeriesList({
    genre: filters.genre || undefined,
    year: filters.year || undefined,
    sort: filters.sort,
    limit: 500,
  })

  const { data: genreList } = useGenres('series')
  const genres = genreList?.map((g) => g.name) ?? []

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: currentYear - 1950 + 1 }, (_, i) => String(currentYear - i))

  const { activeLetters, currentLetter, scrollToLetter, getLetterForTitle } = useAlphaScroll(series)

  const showAlphaIndex = filters.sort === 'title' && (series?.length ?? 0) > 0

  const seenLetters = new Set<string>()

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Series</h1>
          <p className="text-gray-400">
            {series?.length || 0} {series?.length === 1 ? 'series' : 'series'}
          </p>
        </div>
      </div>

      {/* Filter Bar */}
      <FilterBar
        genre={filters.genre}
        year={filters.year}
        sort={filters.sort}
        genres={genres}
        years={years}
        sortOptions={[
          { value: 'newest', label: 'Newest' },
          { value: 'oldest', label: 'Oldest' },
          { value: 'title', label: 'Title A-Z' },
        ]}
        onGenreChange={setGenre}
        onYearChange={setYear}
        onSortChange={setSort}
        onClearFilters={clearFilters}
        hasActiveFilters={hasActiveFilters}
      />

      {showAlphaIndex && (
        <AlphaIndex
          activeLetters={activeLetters}
          currentLetter={currentLetter}
          onSelect={scrollToLetter}
        />
      )}

      {/* Series Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#e50914] border-t-transparent" />
        </div>
      ) : series?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-[#1a1a1a]">
          <LuTv size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">
            {hasActiveFilters ? 'No series match your filters.' : 'No series in your library yet.'}
          </p>
        </div>
      ) : (
        <div
          className={`grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 ${showAlphaIndex ? 'pr-8' : ''}`}
        >
          {series?.map((s: SeriesListItem) => {
            const letter = getLetterForTitle(s.sort_title || s.title)
            const isFirstOfLetter = showAlphaIndex && !seenLetters.has(letter)
            if (isFirstOfLetter) seenLetters.add(letter)
            return (
              <div key={s.id} {...(isFirstOfLetter ? { 'data-alpha-letter': letter } : {})}>
                <MediaCard
                  id={s.id}
                  title={s.title}
                  posterPath={s.poster_path}
                  type="series"
                  seriesId={s.id}
                  year={s.first_air_date ? new Date(s.first_air_date).getFullYear() : undefined}
                />
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
