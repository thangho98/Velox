# Phase 01: Backend — Enhanced Search & Filter API
Status: ⬜ Pending
Dependencies: None

## Objective
Mở rộng `/api/media`, `/api/series`, `/api/genres` thêm server-side search/filter/sort params.
Tạo unified search endpoint `/api/search`. Tạo folder browse endpoint `/api/browse`.

## Hiện trạng

### Repository layer
- `MediaRepo.List(ctx, libraryID, mediaType, limit, offset)` → `[]model.Media`
  - SQL: SELECT all media columns, optional WHERE library_id/media_type, ORDER BY sort_title
  - Không JOIN genres, không search, không sort dynamic
- `MediaRepo.ListWithGenres(ctx, libraryID, mediaType)` → `[]model.MediaListItem`
  - SQL: JOIN media_genres + genres, GROUP_CONCAT genre names
  - Không pagination, không search/year filter
- `MediaRepo.Search(ctx, query, limit)` → `[]model.Media`
  - SQL: WHERE title LIKE ? OR sort_title LIKE ?
- `SeriesRepo.List(ctx, libraryID, limit, offset)` → `[]model.Series`
  - Không JOIN genres, không search
- `SeriesRepo.Search(ctx, query, limit)` → `[]model.Series`
  - title + sort_title LIKE only

### Handler layer
- `GET /api/media` → parses library_id, type, limit, offset
- `GET /api/series` → parses library_id, limit, offset
- `GET /api/series/search?q=` → separate search endpoint
- `GET /api/genres` → returns all genres (no type filter)
- `GET /api/admin/fs/browse` → admin-only filesystem browse

### Service layer
- `MediaService.List()` → delegates to repo.List()
- `MediaService.Search()` → delegates to repo.Search()
- `SeriesService.List()` → delegates to repo.List()
- `SeriesService.Search()` → delegates to repo.Search()

## Implementation Steps

### Task 1: MediaListFilter struct + enhanced MediaRepo

**File:** `backend/internal/repository/media.go`

```go
// New filter struct
type MediaListFilter struct {
    LibraryID int64
    MediaType string // "movie" | "episode" | ""
    Search    string // LIKE on title + sort_title
    Genre     string // exact match genre name
    Year      string // 4-digit year string
    Sort      string // "newest"|"oldest"|"rating"|"title" (default "title")
    Limit     int
    Offset    int
}

// Replace List() + ListWithGenres() + Search() with single method:
func (r *MediaRepo) ListFiltered(ctx context.Context, f MediaListFilter) ([]model.MediaListItem, error)
```

**SQL cho ListFiltered:**
```sql
-- Genre filter via EXISTS (exact match, no substring false positives)
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
ORDER BY {sort clause}
LIMIT ? OFFSET ?
```

**Sort clause mapping:**
- `"title"` (default) → `m.sort_title ASC`
- `"newest"` → `m.release_date DESC, m.sort_title ASC`
- `"oldest"` → `m.release_date ASC, m.sort_title ASC`
- `"rating"` → `m.rating DESC, m.sort_title ASC`

**Dynamic SQL construction:** Build WHERE/HAVING/ORDER clauses and args slice dynamically (like existing patterns in codebase). Do NOT use string interpolation for values.

**MediaListItem model update** (`model/media.go`):
- Replace current `MediaListItem` with canonical shape including ALL fields needed by any consumer:
  `ID, Title, SortTitle, PosterPath, MediaType, Genres, SeriesID, ReleaseDate, Rating, Overview`
- This is the ONLY response type for `/api/media` list endpoints from now on.

- [ ] Replace `MediaListItem` model — add ReleaseDate, Rating, Overview fields
- [ ] Tạo `MediaListFilter` struct
- [ ] Implement `ListFiltered()` method with EXISTS genre filter
- [ ] Giữ `List()` + `Search()` methods cũ (backward compat cho internal callers)

### Task 2: SeriesListItem model + enhanced SeriesRepo

**File:** `backend/internal/model/series.go`

```go
// Superset of Series — all existing fields + genres. No fields removed.
type SeriesListItem struct {
    ID              int64    `json:"id"`
    LibraryID       int64    `json:"library_id"`
    Title           string   `json:"title"`
    SortTitle       string   `json:"sort_title"`
    TmdbID          *int64   `json:"tmdb_id,omitempty"`
    ImdbID          *string  `json:"imdb_id,omitempty"`
    TvdbID          *int64   `json:"tvdb_id,omitempty"`
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
    Genres          []string `json:"genres"`  // only addition vs Series
}
```

**File:** `backend/internal/repository/series.go`

```go
type SeriesListFilter struct {
    LibraryID int64
    Search    string
    Genre     string
    Year      string
    Sort      string // "newest"|"oldest"|"title" (default "title")
    Limit     int
    Offset    int
}

func (r *SeriesRepo) ListFiltered(ctx context.Context, f SeriesListFilter) ([]model.SeriesListItem, error)
```

**SQL:**
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
ORDER BY {sort clause}
LIMIT ? OFFSET ?
```

**Sort clause:**
- `"title"` → `s.sort_title ASC`
- `"newest"` → `s.first_air_date DESC, s.sort_title ASC`
- `"oldest"` → `s.first_air_date ASC, s.sort_title ASC`

- [ ] Tạo `SeriesListItem` model
- [ ] Tạo `SeriesListFilter` struct
- [ ] Implement `SeriesRepo.ListFiltered()`
- [ ] Parse `genre_names` → `[]string` (reuse pattern từ `MediaRepo.ListWithGenres`)

### Task 3: Enhanced handlers

**File:** `backend/internal/handler/media.go`

Update `List()` handler:
```go
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
    // Existing params (keep)
    libraryID := parseOptionalInt64(r, "library_id")
    mediaType := r.URL.Query().Get("type")
    limit := parseOptionalInt(r, "limit", 50)
    offset := parseOptionalInt(r, "offset", 0)

    // New params
    search := r.URL.Query().Get("search")
    genre := r.URL.Query().Get("genre")
    year := r.URL.Query().Get("year")
    sort := r.URL.Query().Get("sort")

    filter := repository.MediaListFilter{
        LibraryID: libraryID,
        MediaType: mediaType,
        Search:    search,
        Genre:     genre,
        Year:      year,
        Sort:      sort,
        Limit:     limit,
        Offset:    offset,
    }
    items, err := h.service.ListFiltered(r.Context(), filter)
    // ...respondJSON
}
```

**File:** `backend/internal/handler/series.go`

Update `ListSeries()` handler similarly. Merge `SearchSeries()` vào `ListSeries()` (search via `?search=` param). Keep `GET /api/series/search` route nhưng redirect internally.

- [ ] Update `MediaHandler.List()` — parse search/genre/year/sort params
- [ ] Update `SeriesHandler.ListSeries()` — parse search/genre/year/sort params
- [ ] `SeriesHandler.SearchSeries()` — delegate to `ListSeries()` (or keep as alias)

### Task 4: Enhanced genres endpoint

**File:** `backend/internal/repository/genre.go`

```go
func (r *GenreRepo) ListByType(ctx context.Context, mediaType string) ([]model.Genre, error)
```

**SQL khi type = "movie":**
```sql
SELECT DISTINCT g.id, g.name, g.tmdb_id
FROM genres g
JOIN media_genres mg ON mg.genre_id = g.id
WHERE mg.media_id IS NOT NULL
ORDER BY g.name
```

**SQL khi type = "series":**
```sql
SELECT DISTINCT g.id, g.name, g.tmdb_id
FROM genres g
JOIN media_genres mg ON mg.genre_id = g.id
WHERE mg.series_id IS NOT NULL
ORDER BY g.name
```

**File:** `backend/internal/handler/genre.go` (hoặc nơi handle GET /api/genres)

Thêm `type` param:
```go
// GET /api/genres?type=movie|series
genreType := r.URL.Query().Get("type")
if genreType != "" {
    genres, err = genreRepo.ListByType(ctx, genreType)
} else {
    genres, err = genreRepo.List(ctx)
}
```

- [ ] Implement `GenreRepo.ListByType(ctx, mediaType)`
- [ ] Update genre handler thêm `type` query param

### Task 5: Unified search endpoint

**File:** `backend/internal/handler/search.go` — NEW

```go
// GET /api/search?q=xxx&limit=10
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        respondError(w, http.StatusBadRequest, "query parameter 'q' is required")
        return
    }
    limit := parseOptionalInt(r, "limit", 10)

    // Parallel search: movies + series
    movies, err := h.mediaService.ListFiltered(ctx, repository.MediaListFilter{
        Search:    query,
        MediaType: "movie",  // only movies, not episodes
        Limit:     limit,
    })
    series, err := h.seriesService.ListFiltered(ctx, repository.SeriesListFilter{
        Search: query,
        Limit:  limit,
    })

    result := model.SearchResult{
        Movies: movies,
        Series: series,
    }
    respondJSON(w, result)
}
```

**File:** `backend/internal/model/search.go` — NEW
```go
type SearchResult struct {
    Movies []MediaListItem  `json:"movies"`
    Series []SeriesListItem `json:"series"`
}
```

- [ ] Tạo `model/search.go` với `SearchResult` struct
- [ ] Tạo `handler/search.go` với `SearchHandler`
- [ ] Inject `mediaService` + `seriesService` dependencies

### Task 6: Folder browse endpoint

**File:** `backend/internal/handler/browse.go` — NEW

```go
// GET /api/browse?library_id=1&path=
func (h *BrowseHandler) Browse(w http.ResponseWriter, r *http.Request) {
    libraryID := parseRequiredInt64(r, "library_id")
    if libraryID == 0 {
        respondError(w, http.StatusBadRequest, "library_id is required")
        return
    }
    relativePath := r.URL.Query().Get("path") // "" = root

    // Security: reject path traversal
    if strings.Contains(relativePath, "..") {
        respondError(w, http.StatusBadRequest, "invalid path")
        return
    }

    // Check library exists + user has access
    library, err := h.libraryService.Get(ctx, libraryID)
    // Check user access (from JWT claims)

    // Multi-root resolution
    rootPath, subPath, err := resolveRootPath(library.Paths, relativePath)
    // resolveRootPath logic:
    //   Single root: rootPath = paths[0], subPath = relativePath
    //   Multi root + path="": return sentinel (browse shows roots as folders)
    //   Multi root + path="root:N/...": rootPath = paths[N], subPath = rest

    absPath := filepath.Join(rootPath, subPath)

    result, err := h.browseService.Browse(ctx, libraryID, rootPath, absPath, relativePath)
    respondJSON(w, result)
}
```

**File:** `backend/internal/service/browse.go` — NEW

```go
func (s *BrowseService) Browse(ctx context.Context, libraryID int64, paths []string, relativePath string) (*model.BrowseResult, error) {
    // Multi-root: path="" with multiple roots → show roots as top-level folders
    // Media at root of each path intentionally NOT shown here (root picker only).
    // User clicks "root:0" → then sees subfolders + media directly in that root.
    if relativePath == "" && len(paths) > 1 {
        folders := make([]model.BrowseFolder, 0, len(paths))
        for i, p := range paths {
            name := filepath.Base(p)
            count, _ := s.mediaFileRepo.CountByPathPrefix(ctx, p, libraryID)
            folders = append(folders, model.BrowseFolder{
                Name:       name,
                Path:       fmt.Sprintf("root:%d", i),
                MediaCount: count,
            })
        }
        return &model.BrowseResult{
            LibraryID: libraryID, Path: "", Parent: "",
            Folders: folders, Media: nil,
        }, nil
    }

    // Resolve to absolute path
    rootPath, subPath := resolveRootPath(paths, relativePath)
    absPath := filepath.Join(rootPath, subPath)

    // 1. Get all media_files under absPath
    filePaths, err := s.mediaFileRepo.ListByPathPrefix(ctx, absPath, libraryID)

    // 2. Extract immediate subdirectories
    folders := extractSubdirs(absPath, relativePath, filePaths)

    // 3. Get media directly in this folder (not subdirectories)
    media, err := s.mediaFileRepo.ListMediaInFolder(ctx, absPath, libraryID)

    // 4. Compute parent path
    parent := ""
    if relativePath != "" {
        parent = filepath.Dir(relativePath)
        if parent == "." { parent = "" }
    }

    return &model.BrowseResult{
        LibraryID: libraryID, Path: relativePath, Parent: parent,
        Folders: folders, Media: media,
    }, nil
}

// resolveRootPath: single root → (paths[0], relativePath)
// multi root + "root:N/sub" → (paths[N], "sub")
func resolveRootPath(paths []string, relativePath string) (string, string) {
    if len(paths) == 1 {
        return paths[0], relativePath
    }
    // Parse "root:N" prefix
    if strings.HasPrefix(relativePath, "root:") {
        rest := relativePath[5:] // after "root:"
        idx := strings.IndexByte(rest, '/')
        var nStr, subPath string
        if idx == -1 {
            nStr, subPath = rest, ""
        } else {
            nStr, subPath = rest[:idx], rest[idx+1:]
        }
        n, _ := strconv.Atoi(nStr)
        if n >= 0 && n < len(paths) {
            return paths[n], subPath
        }
    }
    return paths[0], relativePath // fallback
}
```

**File:** `backend/internal/repository/media_file.go` (hoặc media.go)

```go
// Get all file paths under a directory prefix
func (r *MediaFileRepo) ListByPathPrefix(ctx context.Context, dirPath string) ([]string, error) {
    query := `SELECT DISTINCT file_path FROM media_files WHERE file_path LIKE ? || '/%'`
    // Returns list of absolute file paths
}

// Get media items whose files are directly in this folder
func (r *MediaFileRepo) ListMediaInFolder(ctx context.Context, dirPath string, libraryID int64) ([]model.MediaListItem, error) {
    // file_path LIKE dirPath || '/%'
    // AND file_path NOT LIKE dirPath || '/%/%'  ← no deeper subdirectories
    // JOIN media for metadata
    query := `
        SELECT DISTINCT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
               GROUP_CONCAT(DISTINCT g.name) as genre_names
        FROM media_files mf
        JOIN media m ON m.id = mf.media_id
        LEFT JOIN media_genres mg ON mg.media_id = m.id
        LEFT JOIN genres g ON g.id = mg.genre_id
        WHERE mf.file_path LIKE ? || '/%'
          AND mf.file_path NOT LIKE ? || '/%/%'
          AND m.library_id = ?
        GROUP BY m.id
        ORDER BY m.sort_title
    `
}
```

**extractSubdirs helper** (in service or handler):
```go
func extractSubdirs(absPath string, filePaths []string) []BrowseFolder {
    // For each filePath, strip absPath prefix, take first path segment
    // Count unique immediate subdirectory names + media count per subdir
    dirCounts := map[string]int{}
    prefix := absPath + "/"
    for _, fp := range filePaths {
        rel := strings.TrimPrefix(fp, prefix)
        parts := strings.SplitN(rel, string(os.PathSeparator), 2)
        if len(parts) > 1 { // has subdirectory
            dirCounts[parts[0]]++
        }
    }
    // Sort alphabetically, build []BrowseFolder with library-relative paths
}
```

- [ ] Tạo `model/browse.go` — BrowseFolder, BrowseResult structs
- [ ] Tạo `handler/browse.go` — BrowseHandler with path validation + ACL
- [ ] Tạo `service/browse.go` — BrowseService with folder extraction logic
- [ ] Thêm `MediaFileRepo.ListByPathPrefix()` + `ListMediaInFolder()`
- [ ] `extractSubdirs()` helper function

### Task 7: Service layer updates

**File:** `backend/internal/service/media.go`
```go
func (s *MediaService) ListFiltered(ctx context.Context, f repository.MediaListFilter) ([]model.MediaListItem, error) {
    if f.Limit == 0 {
        f.Limit = 50
    }
    return s.repo.ListFiltered(ctx, f)
}
```

**File:** `backend/internal/service/series.go`
```go
func (s *SeriesService) ListFiltered(ctx context.Context, f repository.SeriesListFilter) ([]model.SeriesListItem, error) {
    if f.Limit == 0 {
        f.Limit = 50
    }
    return s.repo.ListFiltered(ctx, f)
}
```

- [ ] Add `MediaService.ListFiltered()`
- [ ] Add `SeriesService.ListFiltered()`

### Task 8: Route wiring

**File:** `backend/cmd/server/main.go`

```go
// New: unified search
searchHandler := handler.NewSearchHandler(mediaSvc, seriesSvc)
mux.HandleFunc("GET /api/search", authMw(searchHandler.Search))

// New: folder browse (authenticated, not admin-only)
browseHandler := handler.NewBrowseHandler(librarySvc, browseSvc)
mux.HandleFunc("GET /api/browse", authMw(browseHandler.Browse))

// Existing routes unchanged — handlers updated internally
```

- [ ] Wire SearchHandler to `GET /api/search`
- [ ] Wire BrowseHandler to `GET /api/browse`
- [ ] Verify existing routes still work

## Files to Create/Modify

### Create
- `backend/internal/model/search.go` — SearchResult struct
- `backend/internal/model/browse.go` — BrowseFolder, BrowseResult structs
- `backend/internal/handler/search.go` — unified search handler
- `backend/internal/handler/browse.go` — folder browse handler
- `backend/internal/service/browse.go` — browse logic

### Modify
- `backend/internal/model/media.go` — update MediaListItem (add ReleaseDate, Rating, Overview)
- `backend/internal/model/series.go` — add SeriesListItem struct
- `backend/internal/repository/media.go` — add ListFiltered(), keep List()/Search()
- `backend/internal/repository/series.go` — add ListFiltered()
- `backend/internal/repository/genre.go` — add ListByType()
- `backend/internal/handler/media.go` — parse new query params, call ListFiltered
- `backend/internal/handler/series.go` — parse new query params, call ListFiltered
- `backend/internal/handler/genre.go` (hoặc route handler) — add type param
- `backend/internal/service/media.go` — add ListFiltered()
- `backend/internal/service/series.go` — add ListFiltered()
- `backend/cmd/server/main.go` — wire new routes

## Verification

### Manual test cases (curl)
```bash
# Enhanced media list
curl "localhost:8080/api/media?type=movie&genre=Action&year=1999&sort=rating&limit=10"
# Expected: MediaListItem[] with genres, filtered + sorted

# Enhanced series list
curl "localhost:8080/api/series?genre=Drama&sort=newest&limit=10"
# Expected: SeriesListItem[] with genres array populated

# Unified search
curl "localhost:8080/api/search?q=matrix&limit=5"
# Expected: {"data": {"movies": [...], "series": [...]}}

# Genre list by type
curl "localhost:8080/api/genres?type=series"
# Expected: only genres that have series linked

# Folder browse
curl "localhost:8080/api/browse?library_id=1&path="
# Expected: {"data": {"folders": [...], "media": [...], "path": "", "parent": ""}}

# Folder browse subdirectory
curl "localhost:8080/api/browse?library_id=1&path=Action"
# Expected: subfolders + media in Action/ folder

# Path traversal blocked
curl "localhost:8080/api/browse?library_id=1&path=../../etc"
# Expected: 400 Bad Request

# Backward compat: existing calls still work
curl "localhost:8080/api/media?library_id=1&type=movie&limit=50&offset=0"
# Expected: same behavior as before (now returns MediaListItem[] instead of Media[])
```

## Risks
- **EXISTS subquery + GROUP_CONCAT** in same query: overhead acceptable cho < 10K items. If slow, add index on `genres(name)`.
- **Dynamic SQL construction:** phải careful tránh SQL injection. Dùng parameterized queries + args slice. NEVER string-interpolate values.
- **MediaListItem superset of Media[]:** response thêm fields (genres, release_date, rating, overview). Runtime backward-compatible — Phase 02 updates TS types.
- **Multi-root browse:** `root:N` prefix convention. Media ở root level chỉ hiện khi user click vào root folder (intentional — root picker only). Edge case: library paths change after scan → `root:1` may point to different path. Mitigation: re-browse from root on library change.

---
Next Phase: [Phase 02 — Frontend Filter UI](phase-02-frontend-filter.md)
