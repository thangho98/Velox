# Design: Plan M — Search, Filter & Folder Browser
Created: 2026-03-17
Based on: plan.md, phase-01/02/03

## 1. Consumer Impact Matrix

All current consumers of `/api/media` and `/api/series` list endpoints:

| Consumer | File | Hook | Params | Fields Used |
|----------|------|------|--------|-------------|
| HomePage (movies) | HomePage.tsx:20 | useMediaList | type:'movie', limit:20 | id, title, poster_path, media_type, type, genres, release_date, series_id |
| HomePage (series) | HomePage.tsx:25 | useSeriesList | limit:20 | id, title, sort_title, poster_path, first_air_date |
| MoviesPage | MoviesPage.tsx:13 | useMediaList | type:'movie', limit:100 | id, title, poster_path, media_type, genres, release_date, rating |
| SeriesPage | SeriesPage.tsx:12 | useSeriesList | limit:100 | id, title, poster_path, first_air_date |
| SearchPage (movies) | SearchPage.tsx:49 | useMediaList | type:'movie', limit:100 | id, title, poster_path, genres, release_date, rating, overview |
| SearchPage (series) | SearchPage.tsx:53 | useSeriesList | limit:100 | id, title, poster_path, first_air_date |
| SearchPage (search) | SearchPage.tsx:56 | useSeriesSearch | query, limit:20 | id, title, poster_path, first_air_date |
| LibraryContent | LibraryContent.tsx:18 | useMediaList | library_id, limit:50 | id, title, poster_path, media_type, series_id, release_date, rating, genres |

### Superset Verification

**MediaListItem** (plan) vs fields used by consumers:
- `id, title, sort_title, poster_path, media_type, genres, series_id` — all present in plan
- `release_date, rating, overview` — added to plan's MediaListItem
- `type` field (frontend alias) — derived from media_type in MediaRow mapping, not from API
- **Result: all consumers covered**

**SeriesListItem** (plan) vs fields used by consumers:
- `id, title, sort_title, poster_path, first_air_date` — all present in plan
- `library_id, tmdb_id, imdb_id, tvdb_id, overview, status, network, backdrop_path, logo_path, thumb_path, metadata_locked, created_at, updated_at` — all present (superset)
- `genres` — NEW field, not used by current consumers → safe addition
- **Result: all consumers covered, genres is additive-only**

### Consumers That Need NO Code Changes (Phase 01 only)
- HomePage — maps raw Series to MediaRow shape using id/title/poster_path/first_air_date. Extra fields (genres) ignored by JS.
- LibraryContent — uses MediaListItem fields that exist in both old and new shape.

### Consumers Changed in Phase 02
- MoviesPage — remove client-side filter, use server-side params
- SeriesPage — add genre filter, use server-side params
- SearchPage — replace dual-fetch with unified /api/search

---

## 2. Acceptance Criteria

### Feature 1: Server-side Media Filtering

```
AC-1.1: Genre filter
  Given: MoviesPage loaded with movies in library
  When:  User selects "Action" from genre dropdown
  Then:  GET /api/media?type=movie&genre=Action sent (verify network tab)
         Only movies with genre "Action" displayed
         Movies with "Live Action" NOT included (exact match)

AC-1.2: Year filter
  Given: MoviesPage with movies from 1999-2026
  When:  User selects "1999" from year dropdown
  Then:  Only movies with release_date starting with "1999" displayed

AC-1.3: Sort
  Given: MoviesPage with multiple movies
  When:  User selects "Highest Rated"
  Then:  Movies ordered by rating descending
         Movies with same rating ordered by title

AC-1.4: Search
  Given: MoviesPage loaded
  When:  User types "matrix" in search box (wait 300ms debounce)
  Then:  GET /api/media?type=movie&search=matrix sent
         Only movies with "matrix" in title or sort_title displayed

AC-1.5: Combined filters
  Given: MoviesPage loaded
  When:  User selects genre="Sci-Fi", year="1999", sort="rating"
  Then:  Single request: GET /api/media?type=movie&genre=Sci-Fi&year=1999&sort=rating
         Results match ALL criteria

AC-1.6: Clear filters
  Given: MoviesPage with active genre + year filters
  When:  User clicks "Clear" button
  Then:  All filters removed, URL params cleared, all movies shown

AC-1.7: No filters = all results
  Given: MoviesPage with no active filters
  When:  Page loads
  Then:  GET /api/media?type=movie&limit=50 (same as current behavior)
         All movies displayed
```

### Feature 2: Series Genre Filter

```
AC-2.1: Genre dropdown populated
  Given: SeriesPage loaded
  When:  User clicks genre dropdown
  Then:  Dropdown shows genres from GET /api/genres?type=series
         Only genres linked to at least 1 series shown

AC-2.2: Genre filter works
  Given: SeriesPage loaded
  When:  User selects "Drama"
  Then:  GET /api/series?genre=Drama sent
         Each series in response has genres array containing "Drama"
         Series cards display genres

AC-2.3: Year + genre combined
  Given: SeriesPage loaded
  When:  User selects genre="Comedy", year="2024"
  Then:  Only series matching both criteria displayed
```

### Feature 3: Unified Search

```
AC-3.1: Search returns both types
  Given: SearchPage loaded, user types "breaking"
  When:  Debounce completes (300ms)
  Then:  GET /api/search?q=breaking&limit=10 sent (single request)
         Response has "movies" and "series" sections
         Both sections displayed with counts

AC-3.2: Empty query = empty page
  Given: SearchPage loaded
  When:  No query entered
  Then:  No API call made
         Page shows search prompt (not all media)

AC-3.3: No results
  Given: SearchPage loaded
  When:  User types "xyznonexistent123"
  Then:  "No results for 'xyznonexistent123'" message displayed

AC-3.4: URL persistence
  Given: SearchPage with query "matrix" in URL (?q=matrix)
  When:  User refreshes page
  Then:  Search input pre-filled with "matrix"
         Results for "matrix" displayed
```

### Feature 4: Folder Browser

```
AC-4.1: Library selector
  Given: User has access to 2 libraries (Movies, TV Shows)
  When:  User opens /browse
  Then:  Library dropdown shows both libraries
         First library auto-selected
         Root folders displayed

AC-4.2: Single-root navigation
  Given: Library "Movies" has 1 root path with subfolders Action/, Comedy/, Drama/
  When:  User browses /browse?library_id=1
  Then:  Folders shown: Action (N items), Comedy (N items), Drama (N items)
         Media files directly in root also shown (if any)

AC-4.3: Multi-root navigation
  Given: Library "Movies" has 2 root paths: /nas1/movies, /nas2/movies
  When:  User browses /browse?library_id=1
  Then:  Top-level shows 2 folders: "movies" (root:0), "movies-2" (root:1)
         No media shown at this level (root picker only)
  When:  User clicks "movies" (root:0)
  Then:  Shows subfolders + media directly inside /nas1/movies

AC-4.4: Drill down
  Given: User browsing /browse?library_id=1&path=Action
  When:  User clicks subfolder "Marvel"
  Then:  URL updates to ?library_id=1&path=Action/Marvel
         Breadcrumb shows: Movies > Action > Marvel
         Subfolders of Marvel displayed + media in Marvel/

AC-4.5: Breadcrumb navigation
  Given: User at path Action/Marvel/MCU
  When:  User clicks "Action" in breadcrumb
  Then:  URL updates to ?path=Action
         Shows contents of Action/ folder

AC-4.6: Path traversal blocked
  Given: User attempts /api/browse?library_id=1&path=../../etc
  Then:  400 Bad Request returned
         No filesystem access

AC-4.7: ACL enforced
  Given: User has access to library 1 but NOT library 2
  When:  User requests /api/browse?library_id=2
  Then:  403 Forbidden returned

AC-4.8: Empty folder
  Given: User navigates to folder with no media files scanned
  Then:  "No media in this folder" message displayed
```

### Feature 5: URL Filter Persistence

```
AC-5.1: Filters saved to URL
  Given: MoviesPage with genre="Action", year="1999"
  Then:  URL shows ?genre=Action&year=1999
  When:  User copies URL and opens in new tab
  Then:  Same filters applied, same results shown

AC-5.2: Filter change resets offset
  Given: MoviesPage at offset=50
  When:  User changes genre filter
  Then:  offset removed from URL, results start from beginning
```

---

## 3. Test Cases

### Backend — Phase 01

```
TC-B01: MediaRepo.ListFiltered — no filters (backward compat)
  Given: 10 movies in DB, 5 series
  When:  ListFiltered({MediaType: "movie", Limit: 50})
  Then:  Returns 10 MediaListItem with genres populated
         Each item has ReleaseDate, Rating, Overview fields

TC-B02: MediaRepo.ListFiltered — genre exact match
  Given: Movie A (genres: Action, Sci-Fi), Movie B (genres: Live Action)
  When:  ListFiltered({Genre: "Action"})
  Then:  Returns [Movie A] only
         Movie B NOT included (substring "Action" in "Live Action" must not match)

TC-B03: MediaRepo.ListFiltered — search
  Given: Movies: "The Matrix", "Matrix Reloaded", "Batman"
  When:  ListFiltered({Search: "matrix"})
  Then:  Returns [The Matrix, Matrix Reloaded] (case-insensitive LIKE)

TC-B04: MediaRepo.ListFiltered — year
  Given: Movies from 1999, 2003, 2021
  When:  ListFiltered({Year: "1999"})
  Then:  Returns only 1999 movies

TC-B05: MediaRepo.ListFiltered — sort newest
  Given: Movies: A (2020), B (1999), C (2023)
  When:  ListFiltered({Sort: "newest"})
  Then:  Returns [C, A, B]

TC-B06: MediaRepo.ListFiltered — combined filters
  Given: Action movies from 1999 and 2023, Comedy from 1999
  When:  ListFiltered({Genre: "Action", Year: "1999", Sort: "rating"})
  Then:  Returns only Action movies from 1999, sorted by rating DESC

TC-B07: SeriesRepo.ListFiltered — genre
  Given: Series A (Drama, Crime), Series B (Comedy)
  When:  ListFiltered({Genre: "Drama"})
  Then:  Returns [Series A] with all Series fields + genres array

TC-B08: SeriesRepo.ListFiltered — all Series fields present
  Given: Series with tmdb_id=1234, imdb_id=null, tvdb_id=5678
  When:  ListFiltered({})
  Then:  tmdb_id=1234, imdb_id=nil (omitted in JSON), tvdb_id=5678
         library_id, logo_path, thumb_path, metadata_locked, timestamps all present

TC-B09: SearchHandler — unified search
  Given: Movie "Matrix" and Series "The Matrix: Ressurrections"
  When:  GET /api/search?q=matrix
  Then:  Response: {"data": {"movies": [...], "series": [...]}}
         Both sections populated

TC-B10: GenreRepo.ListByType — movie genres only
  Given: Genre "Action" linked to 3 movies + 0 series
         Genre "Drama" linked to 0 movies + 5 series
  When:  ListByType("movie")
  Then:  Returns [Action] (not Drama)

TC-B11: BrowseHandler — single root, root level
  Given: Library 1 with paths=["/media/movies"], media files in Action/ and root
  When:  GET /api/browse?library_id=1&path=
  Then:  folders: [{name:"Action", path:"Action", media_count:N}]
         media: [items directly in /media/movies/]

TC-B12: BrowseHandler — multi root, root level
  Given: Library 1 with paths=["/nas1/movies", "/nas2/movies"]
  When:  GET /api/browse?library_id=1&path=
  Then:  folders: [{name:"movies", path:"root:0"}, {name:"movies", path:"root:1"}]
         media: [] (root picker only)

TC-B13: BrowseHandler — multi root, inside root
  Given: Library 1 with paths=["/nas1/movies", "/nas2/movies"]
         /nas1/movies has Action/ subfolder and 2 files at root
  When:  GET /api/browse?library_id=1&path=root:0
  Then:  folders: [{name:"Action", path:"root:0/Action"}]
         media: [2 items directly in /nas1/movies/]

TC-B14: BrowseHandler — path traversal rejection
  When:  GET /api/browse?library_id=1&path=../../../etc/passwd
  Then:  400 Bad Request

TC-B15: BrowseHandler — unauthorized library
  Given: User has access to library 1 only
  When:  GET /api/browse?library_id=2
  Then:  403 Forbidden
```

### Frontend — Phase 02

```
TC-F01: MoviesPage — server-side genre filter
  Given: MoviesPage loaded
  When:  Select "Action" genre
  Then:  Network request includes &genre=Action
         No client-side Array.filter() calls
         Results update from server response

TC-F02: MoviesPage — genre dropdown options
  Given: MoviesPage loaded
  Then:  GET /api/genres?type=movie called
         Dropdown shows returned genres alphabetically
         "All Genres" as first option

TC-F03: SeriesPage — genre filter (NEW)
  Given: SeriesPage loaded
  When:  Select "Drama" genre
  Then:  GET /api/series?genre=Drama sent
         Each series card shows genres array
         Results filtered server-side

TC-F04: SearchPage — single unified request
  Given: SearchPage loaded
  When:  Type "matrix" (wait 300ms)
  Then:  Exactly 1 request: GET /api/search?q=matrix
         NOT: GET /api/media + GET /api/series (old behavior)

TC-F05: URL persistence roundtrip
  Given: MoviesPage with ?genre=Action&year=1999&sort=rating
  When:  Refresh page
  Then:  Genre dropdown shows "Action"
         Year dropdown shows "1999"
         Sort dropdown shows "Highest Rated"
         Results match server-filtered data

TC-F06: FilterBar — clear button
  Given: Active filters on MoviesPage
  When:  Click "Clear"
  Then:  URL params removed
         All dropdowns reset to default
         All movies shown

TC-F07: Empty results state
  Given: MoviesPage with genre="Documentary" (no docs in library)
  Then:  "No movies match your filters" message
         Clear filters button visible
```

### Frontend — Phase 03

```
TC-F08: BrowsePage — library auto-select
  Given: User has access to libraries [Movies, TV Shows]
  When:  Navigate to /browse
  Then:  First library auto-selected
         Root folders displayed

TC-F09: BrowsePage — folder click navigation
  Given: /browse?library_id=1
  When:  Click folder "Action"
  Then:  URL updates to ?library_id=1&path=Action
         Breadcrumb shows: Movies > Action
         Action's subfolders + media displayed

TC-F10: BrowsePage — breadcrumb back navigation
  Given: /browse?library_id=1&path=Action/Marvel
  When:  Click "Movies" in breadcrumb
  Then:  URL updates to ?library_id=1 (path removed)
         Root folders displayed

TC-F11: BrowsePage — library switch
  Given: Browsing library 1 at path=Action/Marvel
  When:  Switch library dropdown to library 2
  Then:  URL updates to ?library_id=2 (path reset)
         Library 2 root displayed
```

---

## 4. Edge Cases

| # | Case | Expected Behavior |
|---|------|-------------------|
| E1 | Movie with no genres | Shown in unfiltered list, excluded when any genre filter active |
| E2 | Series with no genres | Same as E1 — genres array empty in response |
| E3 | Genre name with comma | GROUP_CONCAT delimiter is comma — use DISTINCT to avoid dupes, parse carefully |
| E4 | Very long search query (>500 chars) | Truncate to 200 chars server-side before LIKE |
| E5 | Year filter "0000" or "9999" | Valid SQL, returns empty — no special handling needed |
| E6 | Sort param invalid value | Default to "title" if sort value not in allowed set |
| E7 | Folder with 500+ media files | Return all — no pagination on browse (acceptable for v1) |
| E8 | Unicode folder names (日本語, Tiếng Việt) | Must work — filepath.Join handles UTF-8, SQLite LIKE is byte-level |
| E9 | Symlink in media path | Scanner stores resolved absolute path — browse works on resolved paths |
| E10 | Library deleted while browsing | 404 from /api/browse — frontend shows error |
| E11 | Concurrent filter changes (rapid clicks) | TanStack Query cancels previous request via queryKey change |
| E12 | Multi-root with same basename | Append index: "movies", "movies-2" (handled in BrowseService) |

---

## 5. Component Design

### FilterBar (shared between MoviesPage, SeriesPage)

```
Props:
  genres: Genre[]           — from useGenres(type)
  sortOptions: Option[]     — page-specific
  filters: FilterState      — from useFilterParams
  onFilterChange: fn        — from useFilterParams
  onClear: fn               — from useFilterParams
  hasActiveFilters: boolean — from useFilterParams
  showSearch: boolean       — true on MoviesPage/SeriesPage

Renders:
  ┌──────────────────────────────────────────────────────┐
  │ [🔍 Search...   ] [Genre ▼] [Year ▼] [Sort ▼] [✕]  │
  └──────────────────────────────────────────────────────┘

Behavior:
  - Search input: 300ms debounce → setFilter('search', value)
  - Dropdowns: immediate → setFilter('genre'|'year'|'sort', value)
  - Clear (✕): only visible when hasActiveFilters
  - Mobile: horizontal scroll, search full width on top
```

### Breadcrumb (BrowsePage only)

```
Props:
  path: string           — "Action/Marvel/MCU"
  libraryName: string    — "Movies"
  onNavigate: (path) => void

Renders:
  Movies > Action > Marvel > MCU
  ^click   ^click   ^click   (current, not clickable)
```

### FolderCard (BrowsePage only)

```
Props:
  name: string
  mediaCount: number
  onClick: () => void

Renders:
  ┌──────────────┐
  │     📁       │  Same aspect ratio as MediaCard
  │   Action     │  Folder icon (Lucide)
  │   12 items   │  Name + count
  └──────────────┘
```

---

## 6. Data Flow Diagrams

### MoviesPage (After Phase 02)

```
URL ?genre=Action&year=1999&sort=rating
        │
        ▼
useFilterParams() ──→ { genre:"Action", year:"1999", sort:"rating" }
        │
        ├──→ useGenres('movie') ──→ GET /api/genres?type=movie ──→ dropdown options
        │
        └──→ useMediaList({ type:'movie', genre:'Action', year:'1999', sort:'rating' })
                │
                └──→ GET /api/media?type=movie&genre=Action&year=1999&sort=rating
                        │
                        ▼
                   MediaListItem[] ──→ MediaCard grid (no client-side filtering)
```

### SearchPage (After Phase 02)

```
Input "matrix" ──→ debounce 300ms ──→ debouncedQuery
        │
        ├──→ URL ?q=matrix
        │
        └──→ useSearch("matrix", 10)
                │
                └──→ GET /api/search?q=matrix&limit=10
                        │
                        ▼
                   SearchResult { movies: [...], series: [...] }
                        │
                        ├──→ Movies section (MediaCard grid)
                        └──→ Series section (MediaCard grid)
```

### BrowsePage (Phase 03)

```
URL ?library_id=1&path=Action
        │
        ├──→ useLibraries() ──→ library dropdown
        │
        └──→ useFolderBrowse({ libraryId:1, path:"Action" })
                │
                └──→ GET /api/browse?library_id=1&path=Action
                        │
                        ▼
                   BrowseResult { folders: [...], media: [...] }
                        │
                        ├──→ Breadcrumb: Movies > Action
                        ├──→ FolderCard grid (subfolders)
                        └──→ MediaCard grid (media in folder)
```
