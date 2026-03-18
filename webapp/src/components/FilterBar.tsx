import { LuX, LuSlidersHorizontal, LuCalendar, LuTag } from 'react-icons/lu'
import { Select } from '@/components/ui/Select'
import type { SortOption } from '@/hooks/useFilterParams'

interface FilterBarProps {
  genre: string
  year: string
  sort: SortOption
  genres: string[]
  years: string[]
  sortOptions?: { value: SortOption; label: string }[]
  onGenreChange: (value: string) => void
  onYearChange: (value: string) => void
  onSortChange: (value: SortOption) => void
  onClearFilters: () => void
  hasActiveFilters: boolean
}

const defaultSortOptions: { value: SortOption; label: string }[] = [
  { value: 'newest', label: 'Newest' },
  { value: 'oldest', label: 'Oldest' },
  { value: 'rating', label: 'Top Rated' },
  { value: 'title', label: 'Title A-Z' },
]

export function FilterBar({
  genre,
  year,
  sort,
  genres,
  years,
  sortOptions = defaultSortOptions,
  onGenreChange,
  onYearChange,
  onSortChange,
  onClearFilters,
  hasActiveFilters,
}: FilterBarProps) {
  return (
    <div className="flex flex-wrap items-center justify-end gap-3">
      {/* Sort */}
      <div className="flex items-center gap-2">
        <LuSlidersHorizontal className="w-4 h-4 text-gray-400" />
        <Select value={sort} onChange={(e) => onSortChange(e.target.value as SortOption)}>
          {sortOptions.map(({ value, label }) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </Select>
      </div>

      {/* Genre */}
      {genres.length > 0 && (
        <div className="flex items-center gap-2">
          <LuTag className="w-4 h-4 text-gray-400" />
          <Select value={genre} onChange={(e) => onGenreChange(e.target.value)}>
            <option value="">All Genres</option>
            {genres.map((g) => (
              <option key={g} value={g}>
                {g}
              </option>
            ))}
          </Select>
        </div>
      )}

      {/* Year */}
      {years.length > 0 && (
        <div className="flex items-center gap-2">
          <LuCalendar className="w-4 h-4 text-gray-400" />
          <Select value={year} onChange={(e) => onYearChange(e.target.value)}>
            <option value="">All Years</option>
            {years.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </Select>
        </div>
      )}

      {/* Clear */}
      {hasActiveFilters && (
        <button
          onClick={onClearFilters}
          className="flex items-center gap-1 px-3 py-2 text-sm text-gray-400 hover:text-white transition-colors"
        >
          <LuX className="w-4 h-4" />
          Clear
        </button>
      )}
    </div>
  )
}
