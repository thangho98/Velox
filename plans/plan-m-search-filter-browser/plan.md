# Plan M: Search, Filter & Folder Browser
Created: 2026-03-17
Status: ⬜ Pending
Priority: 🔴 High
Dependencies: Plans A-F done (media model, scan pipeline, web app)

## Overview
Cải thiện search + filter hiện tại (đang client-side, thiếu genre filter cho series) và thêm
tính năng browse media theo folder structure — cho phép user tìm phim theo cách tổ chức thư mục
trên disk, không chỉ theo library/genre.

## Vấn đề hiện tại

| # | Vấn đề | Root Cause | Impact |
|---|--------|------------|--------|
| 1 | Movie search client-side | `MoviesPage` load 100 items, `Array.filter()` by genre/year | Library > 200 items sẽ lag |
| 2 | Series page thiếu genre filter | `SeriesRepo.List()` không JOIN genres, `Series` type không có genres field | UX không nhất quán vs MoviesPage |
| 3 | Search không thống nhất | Movies = client-side `title.includes()`, Series = server `LIKE`. SearchPage gọi 2 hook riêng | Mixed results quality + extra payload |
| 4 | Không browse theo folder | Chỉ có admin-only `GET /api/admin/fs/browse` đọc filesystem trực tiếp | User không thấy cấu trúc thư mục |

## Hiện trạng Code

### Backend
- `MediaRepo.List(ctx, libraryID, mediaType, limit, offset)` — no search/genre/year/sort
- `MediaRepo.ListWithGenres(ctx, libraryID, mediaType)` — JOIN genres nhưng **không pagination, không filter** (chỉ dùng internal)
- `MediaRepo.Search(ctx, query, limit)` — LIKE trên title + sort_title, không genre/year
- `SeriesRepo.List(ctx, libraryID, limit, offset)` — no search/genre/sort
- `SeriesRepo.Search(ctx, query, limit)` — LIKE, no genre
- `GenreRepo.ListBySeriesID(ctx, id)` — exists, genre→series link qua `media_genres.series_id`
- `FSBrowse` handler — admin-only, reads filesystem, returns absolute paths

### Frontend
- `MediaListItem` type đã có `genres: string[]` (từ `ListWithGenres`)
- `Series` type **không có** `genres` field
- `useMediaList({library_id, type, limit, offset})` — no search/genre/year/sort params
- `useSeriesList({library_id, limit, offset})` — no search/genre params
- `MoviesPage` — client-side filter/sort bằng `Array.filter()` + `Array.sort()`
- `SeriesPage` — client-side year filter + sort, **no genre filter**
- `SearchPage` — loads all movies + all series, then client-side filter + `useSeriesSearch()`

## Giải pháp

### Quyết định kiến trúc (Locked)

1. **Backward-compatible enhancement**: Mở rộng `/api/media` và `/api/series` thêm query params mới. Params cũ vẫn hoạt động, params mới là optional. Response shape thay đổi từ `Media[]` → `MediaListItem[]` và `Series[]` → `SeriesListItem[]` — đây là **superset** (thêm fields, không xóa/rename fields cũ). Frontend callers cần update TypeScript types nhưng runtime behavior không break vì JSON chỉ thêm fields mới.

2. **Unified search = new endpoint**: `GET /api/search?q=xxx` trả `{movies: [], series: []}`. SearchPage chuyển sang dùng endpoint này. Endpoints cũ `/api/media` + `/api/series` giữ nguyên.

3. **Series genre**: Thêm `genres` field vào response của `GET /api/series` (giống `MediaListItem` đã có `genres`). Tạo `SeriesListItem` model mới với GROUP_CONCAT genres.

4. **Folder browser = DB-based, library-scoped**: Query `media_files.file_path` với LIKE prefix match trên library paths. Trả library-relative paths. Check `user_library_access` cho ACL. **KHÔNG** đọc filesystem trực tiếp — chỉ browse media đã scan.

5. **Genre dropdown data**: `GET /api/genres?type=movie|series` — filter genres by type. Reuse existing `/api/genres` endpoint, thêm optional `type` param.

## Proposed Data Model Changes

### Không cần migration mới
Tất cả data đã có trong schema hiện tại:
- `media_genres` table đã có cả `media_id` và `series_id` columns
- `media_files.file_path` chứa absolute paths
- `libraries.paths` (JSON array) chứa root folders

### New Go Models

```go
// model/media.go — REPLACE existing MediaListItem
// Canonical shape for ALL /api/media list responses.
// Includes every field used by MoviesPage, SearchPage, MediaCard, HomePage.
type MediaListItem struct {
    ID          int64    `json:"id"`
    Title       string   `json:"title"`
    SortTitle   string   `json:"sort_title"`
    PosterPath  string   `json:"poster_path"`
    MediaType   string   `json:"media_type"`
    Genres      []string `json:"genres"`
    SeriesID    int64    `json:"series_id,omitempty"`
    ReleaseDate string   `json:"release_date"` // needed by MoviesPage sort + year filter
    Rating      float64  `json:"rating"`       // needed by MoviesPage sort + card display
    Overview    string   `json:"overview"`     // needed by SearchPage highlight
}

// model/series.go — NEW type
// Canonical shape for ALL /api/series list responses.
// Superset of current Series — all existing fields kept, genres added.
type SeriesListItem struct {
    ID              int64    `json:"id"`
    LibraryID       int64    `json:"library_id"`
    Title           string   `json:"title"`
    SortTitle       string   `json:"sort_title"`
    TmdbID          *int64   `json:"tmdb_id,omitempty"`   // nullable — matches Series
    ImdbID          *string  `json:"imdb_id,omitempty"`   // nullable — matches Series
    TvdbID          *int64   `json:"tvdb_id,omitempty"`   // nullable — matches Series
    Overview        string   `json:"overview"`
    Status          string   `json:"status"`
    Network         string   `json:"network"`
    FirstAirDate    string   `json:"first_air_date"`
    PosterPath      string   `json:"poster_path"`
    BackdropPath    string   `json:"backdrop_path"`
    LogoPath        string   `json:"logo_path"`
    ThumbPath       string   `json:"thumb_path"`
    MetadataLocked  bool     `json:"metadata_locked"`
    CreatedAt       string   `json:"created_at"`
    UpdatedAt       string   `json:"updated_at"`
    Genres          []string `json:"genres"`    // NEW — only addition vs Series
}

// model/search.go — NEW
type SearchResult struct {
    Movies []MediaListItem  `json:"movies"`
    Series []SeriesListItem `json:"series"`
}

// model/browse.go — NEW
type BrowseFolder struct {
    Name       string `json:"name"`
    Path       string `json:"path"`        // library-relative
    MediaCount int    `json:"media_count"`
}

type BrowseResult struct {
    LibraryID int64            `json:"library_id"`
    Path      string           `json:"path"`       // current library-relative path
    Parent    string           `json:"parent"`      // parent relative path, "" if root
    Folders   []BrowseFolder   `json:"folders"`
    Media     []MediaListItem  `json:"media"`
}
```

### New TypeScript Types

```typescript
// types/api.ts — MediaListItem REPLACES current definition
interface MediaListItem {
  id: number
  title: string
  sort_title: string
  poster_path: string
  media_type: 'movie' | 'episode'
  genres: string[]
  series_id?: number
  release_date: string   // needed by MoviesPage
  rating: number         // needed by MoviesPage
  overview: string       // needed by SearchPage
}

// SeriesListItem = Series + genres. Superset — no fields removed.
interface SeriesListItem {
  id: number
  library_id: number
  title: string
  sort_title: string
  tmdb_id?: number    // nullable — matches Series
  imdb_id?: string    // nullable — matches Series
  tvdb_id?: number    // nullable — matches Series
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
  genres: string[]  // only addition vs Series
}

interface SearchResult {
  movies: MediaListItem[]
  series: SeriesListItem[]
}

interface BrowseFolder {
  name: string
  path: string
  media_count: number
}

interface BrowseResult {
  library_id: number
  path: string
  parent: string
  folders: BrowseFolder[]
  media: MediaListItem[]
}
```

## Proposed API Contract

### Enhanced: GET /api/media (backward compatible)
```
Existing params (unchanged):
  library_id  int64    optional
  type        string   optional ("movie"|"episode")
  limit       int      optional (default 50)
  offset      int      optional (default 0)

New params:
  search      string   optional  — LIKE on title + sort_title
  genre       string   optional  — exact match genre name (JOIN media_genres → genres)
  year        string   optional  — SUBSTR(release_date, 1, 4) = ?
  sort        string   optional  — "newest"|"oldest"|"rating"|"title" (default "title")

Response: {"data": MediaListItem[]}
  Response shape thay đổi từ Media[] → MediaListItem[]. Đây là superset — tất cả
  fields cũ (id, title, sort_title, poster_path, media_type) vẫn có, thêm genres,
  release_date, rating, overview. Existing callers nhận thêm fields nhưng không mất
  fields cũ → runtime backward-compatible. Phase 02 cập nhật TypeScript types.
```

### Enhanced: GET /api/series (backward compatible)
```
Existing params (unchanged):
  library_id  int64    optional
  limit       int      optional (default 50)
  offset      int      optional (default 0)

New params:
  search      string   optional
  genre       string   optional
  year        string   optional  — SUBSTR(first_air_date, 1, 4) = ?
  sort        string   optional  — "newest"|"oldest"|"title" (default "title")

Response: {"data": SeriesListItem[]}
  ⚠️ CHANGE: trả SeriesListItem[] (có genres) thay vì Series[]
```

### New: GET /api/search?q=xxx
```
Params:
  q           string   required
  limit       int      optional (default 10 per type)

Response: {"data": SearchResult}
  {
    "movies": MediaListItem[],    // max {limit}
    "series": SeriesListItem[]    // max {limit}
  }
```

### Enhanced: GET /api/genres
```
Existing: (no change)
  → returns all genres

New param:
  type        string   optional ("movie"|"series")
  → filter: only genres that have at least 1 media/series linked

Response: {"data": Genre[]}
```

### New: GET /api/browse?library_id=1&path=
```
Params:
  library_id  int64    required
  path        string   optional (default "" = library root)

Security:
  - Validate library_id exists
  - Check user has access to library (user_library_access)
  - path must not contain ".." (prevent traversal)
  - path is library-relative (prepend matching root path before querying)

Response: {"data": BrowseResult}

Multi-root handling:
  Library có thể có nhiều root paths (e.g., ["/nas1/movies", "/nas2/movies"]).
  Khi path="" (root level):
    → Mỗi root path trở thành một top-level folder.
    → Folder name = basename of root path (e.g., "movies" from "/nas1/movies").
    → Nếu trùng tên → append index: "movies", "movies-2".
    → path trong response = "root:0", "root:1" (index-based, không expose absolute path).
    → ⚠️ Media nằm trực tiếp ở root paths KHÔNG hiện ở level này (intentional).
      User phải click vào root folder để thấy. Lý do: root level chỉ là
      directory chooser, không aggregate media từ nhiều roots lẫn lộn.
  Khi path = "root:N" (inside a specific root, no subpath):
    → rootPath = library.Paths[N]
    → Show subfolders + media trực tiếp trong rootPath (bao gồm files ở root!)
  Khi path bắt đầu bằng "root:N/..." (e.g., "root:0/Action"):
    → Parse root index N → get library.Paths[N] as rootPath
    → absolutePath = rootPath + "/" + rest-of-path
    → Query bình thường.
  Khi library chỉ có 1 root (phổ biến nhất):
    → path="" shows folders + media directly under root (skip root:0 prefix)
    → path="Action" → absolutePath = singleRoot + "/Action"
    → Transparent — user không thấy "root:0" prefix.

Logic:
  1. Get library by ID → get root paths[]
  2. If path="" AND len(paths) > 1:
     → Return folders = [{name: basename(paths[0]), path: "root:0"}, ...]
     → No media at this level (intentional — root picker only, media shows inside each root)
  3. If path="root:N" (no subpath) AND len(paths) > 1:
     → rootPath = paths[N], query subfolders + media directly under rootPath
  4. If path="" AND len(paths) == 1:
     → rootPath = paths[0], query subfolders + media under rootPath
  5. If path starts with "root:N/...":
     → Parse N, rootPath = paths[N], subPath = rest after "root:N/"
     → absolutePath = rootPath + "/" + subPath
  6. Else (single root, no prefix):
     → absolutePath = paths[0] + "/" + path
  7. Query media_files WHERE file_path LIKE absolutePath + '/%'
     → Extract unique immediate subdirectory names
     → Count media per subdirectory
  8. Query media WHERE id IN (media_files directly in this folder)
  9. Return folders[] + media[] with library-relative paths
```

## SQL Query Sketches

### Enhanced media list (repo)
```sql
-- Genre filter uses EXISTS subquery (not HAVING LIKE) to avoid substring false matches
-- e.g., "Action" must not match "Live Action"
SELECT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
       m.release_date, m.rating, m.overview,
       GROUP_CONCAT(DISTINCT g.name) as genre_names,
       COALESCE(e.series_id, 0) as series_id
FROM media m
LEFT JOIN media_genres mg ON mg.media_id = m.id
LEFT JOIN genres g ON g.id = mg.genre_id
LEFT JOIN episodes e ON e.media_id = m.id
WHERE 1=1
  {AND m.library_id = ?}                                   -- if libraryID > 0
  {AND m.media_type = ?}                                   -- if mediaType != ""
  {AND (m.title LIKE ? OR m.sort_title LIKE ?)}            -- if search != ""
  {AND SUBSTR(m.release_date, 1, 4) = ?}                   -- if year != ""
  {AND EXISTS (                                             -- if genre != ""
    SELECT 1 FROM media_genres mg2
    JOIN genres g2 ON g2.id = mg2.genre_id
    WHERE mg2.media_id = m.id AND g2.name = ?
  )}
GROUP BY m.id
ORDER BY {dynamic}
LIMIT ? OFFSET ?
```

**Why EXISTS instead of HAVING LIKE:**
- `HAVING genre_names LIKE '%Action%'` would match "Live Action" → false positive
- `EXISTS` does exact `g2.name = ?` match → correct
- EXISTS can use index on `media_genres(media_id)` + `genres(name)` → faster
- No dependency on GROUP_CONCAT delimiter/order

### New series list with genres (repo)
```sql
SELECT s.id, s.library_id, s.title, s.sort_title,
       s.tmdb_id, s.imdb_id, s.tvdb_id,
       s.overview, s.status, s.network, s.first_air_date,
       s.poster_path, s.backdrop_path, s.logo_path, s.thumb_path,
       s.metadata_locked, s.created_at, s.updated_at,
       GROUP_CONCAT(DISTINCT g.name) as genre_names
FROM series s
LEFT JOIN media_genres mg ON mg.series_id = s.id
LEFT JOIN genres g ON g.id = mg.genre_id
WHERE 1=1
  {AND s.library_id = ?}
  {AND (s.title LIKE ? OR s.sort_title LIKE ?)}
  {AND SUBSTR(s.first_air_date, 1, 4) = ?}
  {AND EXISTS (
    SELECT 1 FROM media_genres mg2
    JOIN genres g2 ON g2.id = mg2.genre_id
    WHERE mg2.series_id = s.id AND g2.name = ?
  )}
GROUP BY s.id
ORDER BY {dynamic}
LIMIT ? OFFSET ?
```

### Folder browse — get subfolders from media_files
```sql
-- Step 1: Get all file paths under this directory
SELECT DISTINCT
  SUBSTR(mf.file_path, LENGTH(?) + 2) as relative_path  -- strip library root + /
FROM media_files mf
JOIN media m ON m.id = mf.media_id
WHERE mf.file_path LIKE ? || '/%'                        -- under absolute path
  AND m.library_id = ?

-- In Go: parse relative_path to extract immediate subdirectory names
-- Count media per subdirectory
```

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Backend — Enhanced Search & Filter API | 8 tasks | ⬜ Pending |
| 02 | Frontend — Improved Filter UI | 7 tasks | ⬜ Pending |
| 03 | Folder Browser | 6 tasks | ⬜ Pending |

## Delivery Strategy

### Ship 1: Server-side Filter (Phase 01 + 02)
- Enhanced /api/media + /api/series with search/genre/year/sort
- SeriesListItem with genres
- Unified /api/search
- Frontend refactor: server-side filtering
→ MoviesPage + SeriesPage + SearchPage đều dùng server-side filter

### Ship 2: Folder Browser (Phase 03)
- /api/browse endpoint (DB-based, library-scoped)
- BrowsePage UI
→ User có thể navigate folder structure

## Success Criteria
- MoviesPage: select genre "Action" → server returns filtered results (verify network tab: 1 request)
- SeriesPage: select genre "Drama" → server returns filtered series with genres in response
- SearchPage: type "matrix" → unified results with movies + series sections
- BrowsePage: select library → see folder tree → click folder → see media inside
- All filters sync to URL params → refresh page → same results
- Path traversal attempt on /api/browse → 400 Bad Request
- Existing callers (HomePage, LibraryPage) still work without changes

## Risks
- **Genre filter performance:** EXISTS subquery + GROUP_CONCAT in same query adds overhead. SQLite OK for < 10K media. If slow, add index on `genres(name)`.
- **MediaListItem superset change:** `/api/media` trả MediaListItem[] thay Media[]. JSON thêm fields mới (genres, release_date, rating, overview) — runtime OK vì JS bỏ qua extra fields. Phase 02 cập nhật TS types + verify consumers.
- **Multi-root browse UX:** `root:0/Action` prefix is ugly for multi-root libraries. Acceptable tradeoff — most libraries have 1 root path. Single-root case has no prefix.
- **Folder browse from media_files paths:** requires consistent path format from scanner. Current scanner stores absolute paths — confirmed OK.

## Decisions Locked
- **Backward-compatible API** — params mới optional, response shape là superset (thêm fields, không xóa)
- **New unified `/api/search`** — SearchPage uses this, not enhanced /api/media + /api/series
- **`SeriesListItem` with genres** — new model, list endpoint returns this instead of Series
- **DB-based folder browse** — query media_files paths, NOT filesystem. Only shows scanned media.
- **Library-relative paths** in browse response — never expose absolute server paths
- **User library access check** on browse endpoint — non-admin users only see their accessible libraries
- **Genre type filter** on existing `/api/genres` — add `?type=movie|series` param

## Quick Commands
- Start: `/code phase-01`
- Check: `/next`
- Save context: `/save-brain`
