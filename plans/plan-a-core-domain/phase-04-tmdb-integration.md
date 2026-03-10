# Phase 04: TMDb Integration
Status: ⬜ Pending
Plan: A - Core Domain & Ingestion
Dependencies: Phase 03

## Mục tiêu
Fetch metadata từ TMDb: poster, backdrop, overview, cast, genres cho Movie + TV Series.
Integrate vào scan pipeline stage "Match".

## Tasks

### 1. TMDb Client
- [ ] `pkg/tmdb/client.go` - HTTP client với API key, rate limiter (40 req/10s)
- [ ] Config: `VELOX_TMDB_API_KEY` env var
- [ ] Retry on 429 (rate limited) với backoff

### 2. Search & Detail APIs
- [ ] `SearchMovie(title, year) ([]MovieResult, error)`
- [ ] `SearchTV(title, year) ([]TVResult, error)`
- [ ] `GetMovie(tmdbID) (*MovieDetail, error)` - kèm credits, genres
- [ ] `GetTVShow(tmdbID) (*TVDetail, error)` - kèm seasons overview
- [ ] `GetSeason(tmdbID, seasonNum) (*SeasonDetail, error)` - kèm episodes
- **Files:** `pkg/tmdb/search.go`, `pkg/tmdb/movie.go`, `pkg/tmdb/tv.go`

### 3. Image Handling
- [ ] `PosterURL(path, size) string` - construct TMDb image URL
- [ ] `BackdropURL(path, size) string`
- [ ] Download + cache images locally: `data/images/{type}/{tmdb_id}_{size}.jpg`
- [ ] Serve cached images: `GET /api/images/{type}/{filename}`
- [ ] Don't re-download if file exists
- **File:** `pkg/tmdb/image.go`, `internal/handler/image.go` - NEW

### 4. NFO & Local Metadata Fallback (Priority before TMDb)
- [ ] Check for `.nfo` file beside video: `movie.nfo`, `tvshow.nfo`
- [ ] Parse NFO (XML format): title, year, tmdb_id, overview, rating
- [ ] Check for local images: `poster.jpg`, `fanart.jpg`, `folder.jpg` in same directory
- [ ] If NFO has tmdb_id → use it directly (skip search)
- [ ] If NFO has metadata → use it, skip TMDb API call
- [ ] Priority: NFO > Local images > TMDb API (save network, respect user curation)
- **File:** `internal/scanner/nfo.go` - NEW

### 5. Metadata Matcher (Scan Pipeline Stage)
- [ ] Integrate into pipeline: **NFO check first** → then TMDb search if no NFO
  - Movie: check NFO → search TMDb → pick best match → GetMovie → save metadata + genres + cast
  - Series: check NFO → search TMDb → GetTVShow → create/update Series → GetSeason → create Episodes
- [ ] Confidence scoring: exact title+year match = high, fuzzy = low
- [ ] Skip TMDb if already has tmdb_id (re-scan shouldn't re-fetch)
- **File:** `internal/scanner/matcher.go` - NEW

### 6. Manual Identify
- [ ] `PUT /api/media/{id}/identify` - body: `{tmdb_id, media_type}`
- [ ] Override auto-match khi scanner chọn sai
- [ ] Re-fetch metadata từ TMDb với tmdb_id mới
- **File:** `internal/handler/media.go`

### 7. Metadata Refresh
- [ ] `POST /api/media/{id}/refresh` - re-fetch metadata từ TMDb
- [ ] `POST /api/libraries/{id}/refresh-metadata` - refresh all items in library
- [ ] Useful khi TMDb update info (new poster, updated rating)

### 8. Genres Sync
- [ ] Fetch genre list từ TMDb on first run: `GET /genre/movie/list`, `GET /genre/tv/list`
- [ ] Cache genre mapping (tmdb_genre_id → name)
- [ ] Link genres to media items during scan

## Files to Create/Modify
- `pkg/tmdb/client.go` - NEW
- `pkg/tmdb/search.go` - NEW
- `pkg/tmdb/movie.go` - NEW
- `pkg/tmdb/tv.go` - NEW
- `pkg/tmdb/image.go` - NEW
- `internal/scanner/nfo.go` - NEW
- `internal/scanner/matcher.go` - NEW
- `internal/handler/image.go` - NEW
- `internal/handler/media.go` - Add identify/refresh
- `internal/config/config.go` - Add TMDb key

---
Next: phase-05-subtitle-discovery.md
