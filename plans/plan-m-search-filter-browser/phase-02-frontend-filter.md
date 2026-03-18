# Phase 02: Frontend — Improved Filter UI
Status: ⬜ Pending
Dependencies: Phase 01 (backend API ready)

## Objective
Refactor MoviesPage + SeriesPage để dùng server-side search/filter thay vì client-side.
Cải thiện SearchPage dùng unified search endpoint. Thêm genre filter cho SeriesPage.

## Hiện trạng

### MoviesPage (`pages/MoviesPage.tsx`)
- Calls `useMediaList({ type: 'movie', limit: 100 })`
- Client-side filtering: `movie.genres.includes(filters.genre)` + year from release_date
- Client-side sorting: `Array.sort()` by newest/oldest/rating/title
- Genre options derived: `[...new Set(movies?.flatMap(m => m.genres ?? []))]`
- Year options derived from all movies' release_date

### SeriesPage (`pages/SeriesPage.tsx`)
- Calls `useSeriesList({ limit: 100 })`
- Client-side year filter only (from first_air_date)
- Client-side sort: newest/oldest/title
- **No genre filter** — `Series` type has no `genres` field
- Year options derived from all series' first_air_date

### SearchPage (`pages/SearchPage.tsx`)
- Loads ALL movies via `useMediaList({ type: 'movie', limit: 100 })`
- Loads ALL series via `useSeriesList({ limit: 100 })`
- Client-side movie search: `title.includes()` + `overview.includes()` + `genres.includes()`
- Server-side series search: `useSeriesSearch(query)` (LIKE on title)
- Genre filter: client-side from all movies
- Type filter: Movie/Series selector
- Debounce: 300ms
- URL sync: ?q=, ?type=, ?genre=

### Hooks (`hooks/stores/useMedia.ts`)
```typescript
// Current params — NO search/genre/year/sort
interface MediaListParams {
  library_id?: number
  type?: 'movie' | 'episode'
  limit?: number
  offset?: number
}

interface SeriesListParams {
  library_id?: number
  limit?: number
  offset?: number
}
```

### Types (`types/api.ts`)
```typescript
// MediaListItem — already has genres
interface MediaListItem {
  id: number; title: string; sort_title: string; poster_path: string;
  media_type: string; genres: string[]; series_id?: number;
  // After Phase 01: + release_date, rating, overview
}

// Series — NO genres field
interface Series {
  id: number; title: string; /* ... */ overview: string; status: string;
  // NO genres
}
// After Phase 01: new SeriesListItem type with genres
```

## Implementation Steps

### Task 1: Update TypeScript types

**File:** `webapp/src/types/api.ts`

```typescript
// SeriesListItem = Series + genres. Superset — no fields removed from Series.
// Can replace Series type in list contexts (or just extend Series & { genres: string[] })
export interface SeriesListItem {
  id: number
  library_id: number
  title: string
  sort_title: string
  tmdb_id?: number
  imdb_id?: string
  tvdb_id?: number
  overview: string
  status: string
  network: string
  first_air_date: string
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  metadata_locked: boolean
  created_at: string
  updated_at: string
  genres: string[]  // only new field vs Series
}

// Update MediaListItem — add fields from Phase 01
export interface MediaListItem {
  id: number
  title: string
  sort_title: string
  poster_path: string
  media_type: 'movie' | 'episode'
  genres: string[]
  series_id?: number
  release_date?: string  // NEW
  rating?: number        // NEW
  overview?: string      // NEW
}

// NEW
export interface SearchResult {
  movies: MediaListItem[]
  series: SeriesListItem[]
}
```

- [ ] Add `SeriesListItem` interface
- [ ] Update `MediaListItem` — add release_date, rating, overview
- [ ] Add `SearchResult` interface

### Task 2: Update API hooks

**File:** `webapp/src/hooks/stores/useMedia.ts`

```typescript
// Enhanced params
interface MediaListParams {
  library_id?: number
  type?: 'movie' | 'episode'
  search?: string    // NEW
  genre?: string     // NEW
  year?: string      // NEW
  sort?: string      // NEW: 'newest'|'oldest'|'rating'|'title'
  limit?: number
  offset?: number
}

interface SeriesListParams {
  library_id?: number
  search?: string    // NEW
  genre?: string     // NEW
  year?: string      // NEW
  sort?: string      // NEW: 'newest'|'oldest'|'title'
  limit?: number
  offset?: number
}

// Update mediaApi.list to pass new params
const mediaApi = {
  list: (params: MediaListParams) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.set('library_id', String(params.library_id))
    if (params.type) searchParams.set('type', params.type)
    if (params.search) searchParams.set('search', params.search)
    if (params.genre) searchParams.set('genre', params.genre)
    if (params.year) searchParams.set('year', params.year)
    if (params.sort) searchParams.set('sort', params.sort)
    if (params.limit) searchParams.set('limit', String(params.limit))
    if (params.offset) searchParams.set('offset', String(params.offset))
    return api.get<MediaListItem[]>(`/media?${searchParams}`)
  },
}

// Update seriesApi.list similarly
// seriesApi.list returns SeriesListItem[] now (not Series[])

// NEW: unified search API
const searchApi = {
  search: (query: string, limit = 10) =>
    api.get<SearchResult>(`/search?q=${encodeURIComponent(query)}&limit=${limit}`),
}

// NEW: genre API with type filter
const genreApi = {
  list: (type?: 'movie' | 'series') => {
    const params = type ? `?type=${type}` : ''
    return api.get<Genre[]>(`/genres${params}`)
  },
}

// Update useSeriesList return type: SeriesListItem[] not Series[]
export function useSeriesList(params: SeriesListParams = {}) {
  return useQuery({
    queryKey: seriesKeys.list(params),
    queryFn: () => seriesApi.list(params),
    staleTime: 60_000,
  })
}

// NEW hook
export function useSearch(query: string, limit = 10) {
  return useQuery({
    queryKey: ['search', query, limit],
    queryFn: () => searchApi.search(query, limit),
    staleTime: 60_000,
    enabled: query.length > 0,
  })
}

// NEW hook
export function useGenres(type?: 'movie' | 'series') {
  return useQuery({
    queryKey: ['genres', type],
    queryFn: () => genreApi.list(type),
    staleTime: 5 * 60_000, // genres change rarely
  })
}
```

`useSeriesList` now returns `SeriesListItem[]` instead of `Series[]`.
`SeriesListItem` is a superset of `Series` (all fields + genres) — existing consumers work as-is.

- [ ] Update `MediaListParams` + `SeriesListParams` with new fields
- [ ] Update `mediaApi.list()` to pass new params
- [ ] Update `seriesApi.list()` to pass new params + return `SeriesListItem[]`
- [ ] Add `useSearch()` hook
- [ ] Add `useGenres()` hook with type filter
- [ ] Verify `useSeriesList` consumers handle `SeriesListItem` type

### Task 3: Shared filter hook

**File:** `webapp/src/hooks/useFilterParams.ts` — NEW

```typescript
import { useSearchParams } from 'react-router-dom'

interface FilterState {
  search: string
  genre: string
  year: string
  sort: string
}

export function useFilterParams(defaults: Partial<FilterState> = {}) {
  const [searchParams, setSearchParams] = useSearchParams()

  const filters: FilterState = {
    search: searchParams.get('search') ?? defaults.search ?? '',
    genre: searchParams.get('genre') ?? defaults.genre ?? '',
    year: searchParams.get('year') ?? defaults.year ?? '',
    sort: searchParams.get('sort') ?? defaults.sort ?? 'title',
  }

  const setFilter = (key: keyof FilterState, value: string) => {
    setSearchParams(prev => {
      if (value) {
        prev.set(key, value)
      } else {
        prev.delete(key)
      }
      // Reset offset when changing filters
      prev.delete('offset')
      return prev
    }, { replace: true })
  }

  const clearFilters = () => {
    setSearchParams({}, { replace: true })
  }

  const hasActiveFilters = filters.search || filters.genre || filters.year ||
    (filters.sort && filters.sort !== (defaults.sort ?? 'title'))

  return { filters, setFilter, clearFilters, hasActiveFilters }
}
```

- [ ] Create `useFilterParams` hook with URL sync
- [ ] Support defaults per page (MoviesPage default sort='title', etc.)
- [ ] Reset offset on filter change

### Task 4: FilterBar component

**File:** `webapp/src/components/FilterBar.tsx` — NEW

```typescript
interface FilterBarProps {
  genres?: Genre[]              // dropdown options
  years?: string[]              // dropdown options (derived or passed)
  sortOptions: { value: string; label: string }[]
  filters: FilterState
  onFilterChange: (key: string, value: string) => void
  onClear: () => void
  hasActiveFilters: boolean
  showSearch?: boolean          // true for pages with search bar
}
```

Layout:
```
┌─────────────────────────────────────────────────────────┐
│ [🔍 Search...    ] [Genre ▼] [Year ▼] [Sort ▼] [Clear] │
└─────────────────────────────────────────────────────────┘
```

- Search input: debounce 300ms, calls `setFilter('search', value)`
- Genre dropdown: options from `useGenres(type)`, "All Genres" as default
- Year dropdown: derived from data or recent 30 years
- Sort dropdown: page-specific options
- Clear button: only visible when hasActiveFilters
- Responsive: horizontal scroll on mobile

- [ ] Create `FilterBar` component
- [ ] Debounced search input (300ms)
- [ ] Genre/Year/Sort dropdowns
- [ ] Clear filters button
- [ ] Responsive design

### Task 5: Refactor MoviesPage

**File:** `webapp/src/pages/MoviesPage.tsx`

Before (client-side):
```typescript
const { data: movies } = useMediaList({ type: 'movie', limit: 100 })
// ... client-side filter + sort logic (50+ lines)
const filtered = movies?.filter(m => m.genres.includes(genre))
```

After (server-side):
```typescript
const { filters, setFilter, clearFilters, hasActiveFilters } = useFilterParams({ sort: 'title' })
const { data: genres } = useGenres('movie')
const { data: movies, isLoading } = useMediaList({
  type: 'movie',
  search: filters.search || undefined,
  genre: filters.genre || undefined,
  year: filters.year || undefined,
  sort: filters.sort || undefined,
  limit: 50,
})

return (
  <FilterBar
    genres={genres}
    sortOptions={[
      { value: 'title', label: 'Title A-Z' },
      { value: 'newest', label: 'Newest' },
      { value: 'oldest', label: 'Oldest' },
      { value: 'rating', label: 'Highest Rated' },
    ]}
    filters={filters}
    onFilterChange={setFilter}
    onClear={clearFilters}
    hasActiveFilters={hasActiveFilters}
    showSearch
  />
  // ... MediaCard grid (no client-side filtering needed)
)
```

Changes:
- Remove ALL client-side filter/sort logic
- Remove local `filters` state → use `useFilterParams`
- Remove genre/year derivation → fetch from API
- Remove `Array.filter()` + `Array.sort()` chains
- Add empty state: "No movies match your filters"

- [ ] Replace client-side filtering with server-side params
- [ ] Use `useFilterParams` + `FilterBar`
- [ ] Use `useGenres('movie')` for dropdown
- [ ] Remove client-side sort logic
- [ ] Add loading + empty states

### Task 6: Enhance SeriesPage

**File:** `webapp/src/pages/SeriesPage.tsx`

Similar refactor as MoviesPage:
```typescript
const { filters, setFilter, clearFilters, hasActiveFilters } = useFilterParams({ sort: 'title' })
const { data: genres } = useGenres('series')    // NEW: series genres
const { data: series, isLoading } = useSeriesList({
  search: filters.search || undefined,
  genre: filters.genre || undefined,             // NEW
  year: filters.year || undefined,
  sort: filters.sort || undefined,
  limit: 50,
})
```

⚠️ `useSeriesList` now returns `SeriesListItem[]` (has genres). Update card rendering:
- `series.genres` is now available for display
- Sort options: Title, Newest, Oldest (no rating for series)

- [ ] Add genre filter (NEW capability)
- [ ] Replace client-side filtering with server-side
- [ ] Use `useFilterParams` + `FilterBar`
- [ ] Update card rendering for `SeriesListItem` type
- [ ] Add loading + empty states

### Task 7: Refactor SearchPage

**File:** `webapp/src/pages/SearchPage.tsx`

Before:
```typescript
// Load ALL movies + ALL series, then client-side filter
const { data: allMovies } = useMediaList({ type: 'movie', limit: 100 })
const { data: allSeries } = useSeriesList({ limit: 100 })
const { data: seriesResults } = useSeriesSearch(query)
// ... complex client-side filtering
```

After:
```typescript
const [searchParams, setSearchParams] = useSearchParams()
const query = searchParams.get('q') ?? ''
const [inputValue, setInputValue] = useState(query)
const debouncedQuery = useDebounce(inputValue, 300)

const { data: results, isLoading } = useSearch(debouncedQuery, 20)

// Update URL on debounced query change
useEffect(() => {
  if (debouncedQuery) {
    setSearchParams({ q: debouncedQuery }, { replace: true })
  } else {
    setSearchParams({}, { replace: true })
  }
}, [debouncedQuery])

return (
  <div>
    <input value={inputValue} onChange={e => setInputValue(e.target.value)} />

    {results?.movies.length > 0 && (
      <section>
        <h2>Movies ({results.movies.length})</h2>
        <MediaCardGrid items={results.movies} />
      </section>
    )}

    {results?.series.length > 0 && (
      <section>
        <h2>Series ({results.series.length})</h2>
        <SeriesCardGrid items={results.series} />
      </section>
    )}

    {debouncedQuery && !isLoading && !results?.movies.length && !results?.series.length && (
      <EmptyState message={`No results for "${debouncedQuery}"`} />
    )}
  </div>
)
```

Changes:
- Remove loading ALL movies/series
- Use single `useSearch(query)` hook → unified `/api/search`
- Remove client-side filtering + genre filter (search page is for search, not browse)
- Keep URL sync (?q=)
- Show result counts per section
- Add proper empty state

- [ ] Replace dual-data-fetch with `useSearch()` hook
- [ ] Remove client-side filtering logic
- [ ] Show results grouped by type with counts
- [ ] URL sync: ?q= param
- [ ] Empty state + loading state

## Files to Create/Modify

### Create
- `webapp/src/hooks/useFilterParams.ts` — URL-synced filter state
- `webapp/src/components/FilterBar.tsx` — reusable filter bar

### Modify
- `webapp/src/types/api.ts` — add SeriesListItem, SearchResult, update MediaListItem
- `webapp/src/hooks/stores/useMedia.ts` — update hooks + add useSearch, useGenres
- `webapp/src/pages/MoviesPage.tsx` — server-side filtering
- `webapp/src/pages/SeriesPage.tsx` — add genre filter + server-side
- `webapp/src/pages/SearchPage.tsx` — unified search

## Verification

### Test Cases
- [ ] MoviesPage: select "Action" genre → network tab shows `GET /api/media?type=movie&genre=Action`
- [ ] MoviesPage: type "matrix" in search → debounce → server request with `&search=matrix`
- [ ] SeriesPage: select "Drama" genre → results update, each series card shows genres
- [ ] SeriesPage: genre dropdown populated from `GET /api/genres?type=series`
- [ ] SearchPage: type "breaking" → single `GET /api/search?q=breaking` → movies + series sections
- [ ] SearchPage: no query → empty state (no loading all data)
- [ ] URL sync: add ?genre=Action → refresh → filter preserved
- [ ] Clear filters → URL params removed → show all items
- [ ] No results → friendly "No movies match your filters" message
- [ ] HomePage + LibraryPage still work (backward compat of list hooks)

### Superset Verification
- [ ] `useSeriesList` returns `SeriesListItem[]` (superset of `Series`) — confirm no TS compile errors in:
  - HomePage, LibraryPage, SeriesPage (all use list hook)

---
Next Phase: [Phase 03 — Folder Browser](phase-03-folder-browser.md)
