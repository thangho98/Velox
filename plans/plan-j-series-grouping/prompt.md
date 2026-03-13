# Task: Netflix-style Series Grouping + URL Routing (All 4 Phases)

## Goal
Restructure Velox so TV series are grouped like Netflix — one card per series instead of one card per episode. New URL routes: `/movies/:id` for movies, `/series/:seriesId` for series overview. Delete `/media/:id` route entirely (no backwards compat — self-hosted app).

## Critical Design Decisions (READ FIRST)

### ID Domain Mismatch
- `series.id` is NOT a `media_id`. They are different tables, different ID spaces.
- Series cards MUST NOT call `useProgress(id)` or `toggleFavorite(id)` — these expect media_id and would 404.
- Series cards: `showProgress={false}`, `showFavorite={false}`.

### Route Migration (no backwards compat)
| Location | Current | New |
|----------|---------|-----|
| Router.tsx | `/media/:id` → MediaDetailPage | DELETE this route |
| Router.tsx | — | ADD `/movies/:id` → MediaDetailPage |
| Router.tsx | — | ADD `/series/:seriesId` → SeriesDetailPage (NEW) |
| MediaCard.tsx | `/media/${id}` | type-aware: `/movies/${id}` or `/series/${seriesId}` |
| WatchPage.tsx:1317 | `navigate('/media/${mediaId}')` | `navigate(-1)` |
| LibraryListPage.tsx:160 | `to={/media/${item.id}}` | type-aware routing |

### Series API Fields
Available: id, title, poster_path, backdrop_path, overview, status, network, first_air_date, logo_path, thumb_path
NOT available: genres, rating, progress, favorites

### EpisodeCard Bug
Current code uses `episode.id` in play links (`/watch/${episode.id}`). But `episode.id` is the `episodes` table PK, NOT the playable `media_id`. Must use `episode.media_id` instead.

---

## Phase 1: Backend Series API

### Files to modify:
1. `backend/internal/handler/series.go`
2. `backend/cmd/server/main.go`

### Current SeriesHandler (handler/series.go):
```go
type SeriesHandler struct {
	seasonRepo  *repository.SeasonRepo
	episodeRepo *repository.EpisodeRepo
}

func NewSeriesHandler(seasonRepo *repository.SeasonRepo, episodeRepo *repository.EpisodeRepo) *SeriesHandler {
	return &SeriesHandler{seasonRepo: seasonRepo, episodeRepo: episodeRepo}
}

// Existing: ListSeasons (GET /api/series/{id}/seasons)
// Existing: ListEpisodes (GET /api/series/{id}/seasons/{seasonId}/episodes)
```

### Existing SeriesRepo methods (repository/series.go — DO NOT MODIFY):
```go
func (r *SeriesRepo) List(ctx context.Context, libraryID int64, limit, offset int) ([]model.Series, error)
func (r *SeriesRepo) GetByID(ctx context.Context, id int64) (*model.Series, error)
func (r *SeriesRepo) Search(ctx context.Context, query string, limit int) ([]model.Series, error)
```

### Available helpers (handler/respond.go — DO NOT MODIFY):
```go
func respondJSON(w http.ResponseWriter, status int, data any)      // wraps in {"data": ...}
func respondError(w http.ResponseWriter, status int, msg string)   // wraps in {"error": "..."}
func parseID(r *http.Request, param string) (int64, error)         // r.PathValue(param) → int64
func parseIntQuery(r *http.Request, key string, fallback int) int  // query param with default
```

### Changes:

**1. Add seriesRepo to SeriesHandler:**
```go
type SeriesHandler struct {
	seriesRepo  *repository.SeriesRepo  // ADD
	seasonRepo  *repository.SeasonRepo
	episodeRepo *repository.EpisodeRepo
}

func NewSeriesHandler(seriesRepo *repository.SeriesRepo, seasonRepo *repository.SeasonRepo, episodeRepo *repository.EpisodeRepo) *SeriesHandler {
	return &SeriesHandler{seriesRepo: seriesRepo, seasonRepo: seasonRepo, episodeRepo: episodeRepo}
}
```

**2. Add 3 new handlers:**

`ListSeries` — `GET /api/series?library_id=&limit=&offset=`
- Parse: library_id (default 0), limit (default 50), offset (default 0)
- Call `h.seriesRepo.List(ctx, libraryID, limit, offset)`
- respondJSON 200

`GetSeries` — `GET /api/series/{id}`
- parseID(r, "id")
- Call `h.seriesRepo.GetByID(ctx, id)`
- If `errors.Is(err, sql.ErrNoRows)` → respondError 404 "series not found"
- respondJSON 200

`SearchSeries` — `GET /api/series/search?q=&limit=`
- Get `q` from r.URL.Query().Get("q"), if empty → respondError 400 "query required"
- Parse limit (default 20)
- Call `h.seriesRepo.Search(ctx, q, limit)`
- respondJSON 200

**3. Wire in main.go:**

Line 269, change:
```go
// FROM:
seriesHandler := handler.NewSeriesHandler(seasonRepo, episodeRepo)
// TO:
seriesHandler := handler.NewSeriesHandler(seriesRepo, seasonRepo, episodeRepo)
```

Replace the series routes section (lines 358-360) with:
```go
// API routes - Series
mux.HandleFunc("GET /api/series", seriesHandler.ListSeries)
mux.HandleFunc("GET /api/series/search", seriesHandler.SearchSeries)  // BEFORE {id}!
mux.HandleFunc("GET /api/series/{id}", seriesHandler.GetSeries)
mux.HandleFunc("GET /api/series/{id}/seasons", seriesHandler.ListSeasons)
mux.HandleFunc("GET /api/series/{id}/seasons/{seasonId}/episodes", seriesHandler.ListEpisodes)
```

**IMPORTANT:** `GET /api/series/search` MUST be registered BEFORE `GET /api/series/{id}` — Go 1.22+ routing matches literal segments before wildcards only if registered first.

### Verify: `cd backend && go build ./...`

---

## Phase 2: Frontend Types + Hooks

### Files to modify:
1. `webapp/src/types/api.ts`
2. `webapp/src/hooks/stores/useMedia.ts`

### Changes to api.ts:

Add after the `MediaListItem` interface:

```typescript
// Series Types (from GET /api/series)
export interface Series {
  id: number
  library_id: number
  title: string
  sort_title: string
  tmdb_id?: number
  imdb_id?: string
  tvdb_id?: number
  overview: string
  status: string       // "Returning Series" | "Ended" | "Canceled"
  network: string
  first_air_date: string
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  created_at: string
  updated_at: string
}

export interface SeriesWithSeasons {
  series: Series
  seasons: Season[]
}

export interface SeriesListParams {
  library_id?: number
  limit?: number
  offset?: number
}
```

### Changes to useMedia.ts:

Add imports for new types:
```typescript
import type { Series, SeriesListParams } from '@/types/api'
```

Add after the existing `seriesApi` object (which currently only has getSeasons/getEpisodes/getEpisode):

```typescript
// Extend existing seriesApi with new endpoints
// Add these methods to the existing seriesApi object:
list: (params: SeriesListParams = {}) => {
  const searchParams = new URLSearchParams()
  if (params.library_id) searchParams.append('library_id', String(params.library_id))
  if (params.limit) searchParams.append('limit', String(params.limit))
  if (params.offset) searchParams.append('offset', String(params.offset))
  const query = searchParams.toString()
  return api.get<Series[]>(`/series${query ? `?${query}` : ''}`)
},
get: (id: number) => api.get<Series>(`/series/${id}`),
search: (query: string, limit = 20) =>
  api.get<Series[]>(`/series/search?q=${encodeURIComponent(query)}&limit=${limit}`),
```

Add query keys (extend existing `seriesKeys`):
```typescript
// Add to existing seriesKeys:
list: (params: SeriesListParams) => [...seriesKeys.all, 'list', params] as const,
detail: (id: number) => [...seriesKeys.all, 'detail', id] as const,
search: (query: string) => [...seriesKeys.all, 'search', query] as const,
```

Add hooks:
```typescript
export function useSeriesList(params: SeriesListParams = {}) {
  return useQuery({
    queryKey: seriesKeys.list(params),
    queryFn: () => seriesApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useSeriesDetail(id: number) {
  return useQuery({
    queryKey: seriesKeys.detail(id),
    queryFn: () => seriesApi.get(id),
    staleTime: 5 * 60 * 1000,
    enabled: id > 0,
  })
}

export function useSeriesSearch(query: string, limit = 20) {
  return useQuery({
    queryKey: seriesKeys.search(query),
    queryFn: () => seriesApi.search(query, limit),
    staleTime: 60 * 1000,
    enabled: query.length > 0,
  })
}
```

### Verify: `cd webapp && npx tsc --noEmit`

---

## Phase 3: Routes + SeriesDetailPage

### Files to create:
1. `webapp/src/components/EpisodeCard.tsx`
2. `webapp/src/pages/SeriesDetailPage.tsx`

### Files to modify:
3. `webapp/src/pages/MediaDetailPage.tsx`
4. `webapp/src/providers/Router.tsx`

### 3a. Extract EpisodeCard component

Create `webapp/src/components/EpisodeCard.tsx`.

Move the `EpisodeCard` function and `formatDuration` helper from MediaDetailPage.tsx into this new file.

**BUG FIX:** Change ALL `episode.id` references in play links to `episode.media_id`:
```typescript
// WRONG (current):
<Link to={`/watch/${episode.id}`} ...>

// CORRECT:
<Link to={`/watch/${episode.media_id}`} ...>
```

There are TWO places in EpisodeCard where this link appears — fix both.

The component should export:
```typescript
import { Link } from 'react-router'
import { LuFilm, LuPlay } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'
import type { Episode } from '@/types/api'

export function EpisodeCard({ episode }: { episode: Episode }) { ... }

// Keep formatDuration as a non-exported function in this file
function formatDuration(seconds: number): string { ... }
```

### 3b. Create SeriesDetailPage

Create `webapp/src/pages/SeriesDetailPage.tsx`.

This is a Netflix-style series overview page:
- Route param: `useParams<{ seriesId: string }>()`
- Fetch series: `useSeriesDetail(Number(seriesId))`
- Fetch seasons: `useSeasons(Number(seriesId))`
- Fetch episodes: `useEpisodes(Number(seriesId), selectedSeasonId)`
- Layout: backdrop image + poster + title + overview + status/network badges
- Season selector tabs (buttons, same pattern as current MediaDetailPage seasons)
- Episode list using the new `EpisodeCard` component
- Play button: link to `/watch/${firstEpisode?.media_id}` (use `media_id`, NOT `id`)

Key imports:
```typescript
import { useSeriesDetail, useSeasons, useEpisodes } from '@/hooks/stores/useMedia'
import { EpisodeCard } from '@/components/EpisodeCard'
```

Style: reuse the same dark Netflix-style layout as MediaDetailPage (backdrop with gradients, etc.)

### 3c. Simplify MediaDetailPage (movie-only)

This page now only handles movies (route: `/movies/:id`).

Remove:
- The entire `isSeries` branch and all series-related code
- `seriesMedia` fetch (`useMediaWithFiles(isSeries ? seriesId : 0)`)
- `useSeasons`, `useEpisodes` imports and calls
- `displayTitle` logic — just use `media.media.title`
- `seriesId`, `selectedSeasonId` state
- The `EpisodeCard` function (moved to its own file)
- The seasons/episodes section (the entire `{isSeries && (...)}` block)
- The `seriesTitle`, `seriesYear` variables

Keep:
- Movie detail display, poster, backdrop
- Progress bar, play button, favorite, refresh metadata
- `formatDuration`, `getEndTime`, `formatFileSize` helper functions

### 3d. Update Router

In `webapp/src/providers/Router.tsx`:

Add import:
```typescript
import { SeriesDetailPage } from '@/pages/SeriesDetailPage'
```

Change routes:
```typescript
// REMOVE:
<Route path="/media/:id" element={<MediaDetailPage />} />

// ADD:
<Route path="/movies/:id" element={<MediaDetailPage />} />
<Route path="/series/:seriesId" element={<SeriesDetailPage />} />
```

### Verify: `cd webapp && npx tsc --noEmit`

---

## Phase 4: Update Listing Pages + Components

### Files to modify:
1. `webapp/src/components/MediaCard.tsx`
2. `webapp/src/components/MediaRow.tsx`
3. `webapp/src/pages/SeriesPage.tsx`
4. `webapp/src/pages/HomePage.tsx`
5. `webapp/src/pages/SearchPage.tsx`
6. `webapp/src/pages/WatchPage.tsx`
7. `webapp/src/pages/LibraryListPage.tsx`

### 4a. MediaCard — type-aware routing + series card semantics

Current MediaCardProps already has `type?: 'movie' | 'series'`. Add `seriesId`:

```typescript
interface MediaCardProps {
  id: number
  title: string
  posterPath?: string | null
  type?: 'movie' | 'series'
  seriesId?: number          // ADD — series.id for routing
  year?: number
  rating?: number
  progress?: { ... } | null
  showProgress?: boolean
  showFavorite?: boolean
  aspectRatio?: 'poster' | 'wide'
  size?: 'sm' | 'md' | 'lg'
}
```

Changes inside the component:

1. **Skip useProgress for series cards** (series.id is NOT a media_id):
```typescript
// CURRENT:
const { data: fetchedProgress } = useProgress(showProgress ? id : 0)

// NEW:
const isSeries = type === 'series'
const { data: fetchedProgress } = useProgress(!isSeries && showProgress ? id : 0)
```

2. **Force no progress/favorite for series cards:**
```typescript
const effectiveShowProgress = isSeries ? false : showProgress
const effectiveShowFavorite = isSeries ? false : showFavorite
```
Use `effectiveShowProgress` and `effectiveShowFavorite` everywhere instead of `showProgress` and `showFavorite`.

3. **Type-aware link:**
```typescript
// CURRENT:
<Link to={`/media/${id}`} className="block">

// NEW:
<Link to={isSeries ? `/series/${seriesId}` : `/movies/${id}`} className="block">
```

4. **Fix favorite button** — use `effectiveShowFavorite`:
```typescript
{effectiveShowFavorite && (
  <button onClick={...}>...</button>
)}
```

### 4b. MediaRow — pass seriesId

In `webapp/src/components/MediaRow.tsx`, the non-UserData branch (line 113-125):

```typescript
// ADD series_id pass-through:
<MediaCard
  id={item.id}
  title={item.title}
  posterPath={item.poster_path}
  type={item.type ?? (item.media_type === 'episode' ? 'series' : 'movie')}
  seriesId={'series_id' in item ? (item as any).series_id : undefined}  // ADD
  year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
  rating={item.rating}
  showProgress={showProgress}
/>
```

### 4c. SeriesPage — use real series data

Complete rewrite of `webapp/src/pages/SeriesPage.tsx`:

```typescript
import { useState } from 'react'
import { useSeriesList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { LuX, LuTv } from 'react-icons/lu'
import type { Series } from '@/types/api'

export function SeriesPage() {
  const [filters, setFilters] = useState({ year: '', sortBy: 'newest' })
  const { data: series, isLoading } = useSeriesList({ limit: 100 })

  // Filter
  const filteredSeries = series?.filter((s) => {
    if (filters.year) {
      const year = s.first_air_date ? new Date(s.first_air_date).getFullYear() : null
      if (year !== Number(filters.year)) return false
    }
    return true
  })

  // Sort
  const sortedSeries = filteredSeries?.sort((a, b) => {
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
  const years = [...new Set(
    series?.map((s) => s.first_air_date ? new Date(s.first_air_date).getFullYear() : null)
      .filter((y): y is number => y !== null && !Number.isNaN(y)) || []
  )].sort((a, b) => b - a)

  return (
    // Same layout structure as current SeriesPage but:
    // - NO genre filter (series table has no genres)
    // - NO rating sort option
    // - Year uses first_air_date instead of release_date
    // - MediaCard gets seriesId={s.id} and type="series"
    // In the grid:
    <MediaCard
      key={s.id}
      id={s.id}
      title={s.title}
      posterPath={s.poster_path}
      type="series"
      seriesId={s.id}
      year={s.first_air_date ? new Date(s.first_air_date).getFullYear() : undefined}
    />
  )
}
```

Remove: genre filter, genre state, genre extraction, "Highest Rated" sort option.

### 4d. HomePage — series row uses real series

In `webapp/src/pages/HomePage.tsx`:

```typescript
// REPLACE:
import { useLibraries, useContinueWatching, useNextUp, useMediaList } from '@/hooks/stores/useMedia'

// WITH:
import { useLibraries, useContinueWatching, useNextUp, useMediaList, useSeriesList } from '@/hooks/stores/useMedia'

// REPLACE the series fetch (lines 20-23):
// FROM:
const { data: recentSeries, isLoading: seriesLoading } = useMediaList({
  type: 'episode',
  limit: 20,
})

// TO:
const { data: rawSeries, isLoading: seriesLoading } = useSeriesList({ limit: 20 })

// Map Series[] to MediaRow-compatible shape:
const recentSeries = rawSeries?.map((s) => ({
  id: s.id,
  title: s.title,
  sort_title: s.sort_title,
  poster_path: s.poster_path,
  media_type: 'episode' as const,  // MediaRow uses this for type detection
  type: 'series' as const,
  genres: [] as string[],
  release_date: s.first_air_date,
  series_id: s.id,   // for MediaRow → MediaCard routing
}))
```

Remove the `// TODO: replace with GET /api/series` comment.

### 4e. SearchPage — merge movies + series results

Complete rewrite of `webapp/src/pages/SearchPage.tsx`:

Key changes:
- Remove `useMediaList({ limit: 500 })` — no more client-side full-list fetch
- Add `useSeriesSearch(debouncedQuery)` for series results
- Add `useMediaList({ type: 'movie', limit: 100 })` for browsing (no query), or use media search endpoint
- Type filter: "Movies" → show only media results, "Series" → show only series results
- Genre filter: only show when viewing movies (series have no genres)
- Series cards use `type='series'`, `seriesId={s.id}`, link to `/series/${s.id}`
- Default view (no search): show Movies section (from media list) + Series section (from series list)
- Search view: merge movie results + series results
- Fix the "View all" button for Series section: `onClick={() => setFilters({ ...filters, type: 'series' })}`

For the default view Series section:
```typescript
import { useSeriesList, useSeriesSearch } from '@/hooks/stores/useMedia'

// For browsing (no search query):
const { data: allSeries } = useSeriesList({ limit: 100 })

// For search:
const { data: seriesResults } = useSeriesSearch(debouncedQuery)
```

When rendering series results:
```typescript
<MediaCard
  key={`series-${s.id}`}
  id={s.id}
  title={s.title}
  posterPath={s.poster_path}
  type="series"
  seriesId={s.id}
  year={s.first_air_date ? new Date(s.first_air_date).getFullYear() : undefined}
/>
```

### 4f. WatchPage — fix back button

In `webapp/src/pages/WatchPage.tsx` line 1317:

```typescript
// REPLACE:
onClick={() => navigate(`/media/${mediaId}`)}

// WITH:
onClick={() => navigate(-1)}
```

### 4g. LibraryListPage — type-aware routing

In `webapp/src/pages/LibraryListPage.tsx` line 160:

```typescript
// REPLACE:
to={`/media/${item.id}`}

// WITH:
to={item.media_type === 'episode' ? `/series/${item.id}` : `/movies/${item.id}`}
```

Note: This is imperfect for episodes (item.id is media_id, not series_id), but acceptable for now — episodes in library list view should link to the movie detail page or could link to watch directly. A better approach: link movies to `/movies/${item.id}`, link episodes to `/watch/${item.id}`.

### Verify: `cd webapp && npx tsc --noEmit`

---

## Verification Checklist

After ALL phases are complete, verify:

1. `cd backend && go build ./...` — zero errors
2. `cd webapp && npx tsc --noEmit` — zero errors
3. No `/media/` route links remain in frontend `.tsx` files (except API endpoints like `/api/media/`)
4. `grep -r "to=.*\/media\/" webapp/src/` should return zero results
5. `grep -r "navigate.*\/media\/" webapp/src/` should return zero results

## Conventions
- Go: error wrapping with `fmt.Errorf("doing X: %w", err)`, receiver `func (h *SeriesHandler)`
- TypeScript: no `any` (use `as const` for type narrowing), functional components, named exports for components
- Imports: use `@/` path alias for `src/`
- Styling: TailwindCSS 4, Netflix dark theme
- API responses: `{"data": ...}` for success, `{"error": "..."}` for errors
