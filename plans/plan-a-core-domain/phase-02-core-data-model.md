# Phase 02: Core Data Model
Status: ⬜ Pending
Plan: A - Core Domain & Ingestion
Dependencies: Phase 01

## Mục tiêu
Data model phân tách rõ: media item (logical) vs media file (physical) vs metadata.
Hỗ trợ Movie + TV Series + multi-version files.

## Tasks

### 1. Media Item (logical identity)
- [ ] Refactor `media` table → đại diện cho 1 "content" (movie hoặc episode)
- [ ] Fields: id, library_id, media_type (movie|episode), title, sort_title, tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path, created_at, updated_at
- [ ] `media_type` quyết định đây là movie hay episode
- [ ] `tmdb_id` + `imdb_id` indexed trực tiếp (primary providers)
- [ ] Nếu episode: có FK tới `episodes` table
- **Migration:** `002_refactor_media_item.go`

### 2. Media File (physical file)
- [ ] NEW table `media_files` - đại diện cho 1 file vật lý trên disk
- [ ] Fields: id, media_id (FK), file_path, size, duration, width, height, video_codec, audio_codec, container, bitrate, file_hash (size-based fingerprint), added_at, last_verified_at, is_primary (boolean)
- [ ] 1 media item có thể có nhiều files (720p + 1080p + 4K versions)
- [ ] `is_primary` = version mặc định khi play
- **Migration:** `002_refactor_media_item.go`

### 3. Series Model
- [ ] Table `series`: id, title, sort_title, tmdb_id, imdb_id, overview, status (continuing|ended), first_air_date, poster_path, backdrop_path, created_at
- [ ] Table `seasons`: id, series_id, season_number, title, overview, poster_path, episode_count
- [ ] Table `episodes`: id, season_id, series_id, episode_number, title, overview, still_path, air_date, media_id (FK → media)
- **Migration:** `003_series_model.go`

### 4. Genre & People
- [ ] Table `genres`: id, name, tmdb_id
- [ ] Table `media_genres`: media_id, genre_id (also series_id for series)
- [ ] Table `people`: id, name, tmdb_id, profile_path
- [ ] Table `credits`: id, media_id/series_id, person_id, character, role (cast|director|writer), display_order
- **Migration:** `004_genres_people.go`

### 5. Go Structs
- [ ] `model/media.go` - refactor Media, add MediaFile
- [ ] `model/series.go` - Series, Season, Episode
- [ ] `model/genre.go` - Genre
- [ ] `model/person.go` - Person, Credit

### 6. Repository: Media
- [ ] Refactor `repository/media.go` - tách media item vs media file queries
- [ ] `MediaRepo.Upsert(item)` - upsert media item
- [ ] `MediaFileRepo.Upsert(file)` - upsert file record
- [ ] `MediaFileRepo.FindByHash(hash)` - detect renamed files
- [ ] `MediaFileRepo.MarkMissing(id)` - khi file không còn trên disk

### 7. Repository: Series
- [ ] `repository/series.go` - CRUD Series, Season, Episode
- [ ] `SeriesRepo.GetWithSeasons(id)` - series + all seasons
- [ ] `SeasonRepo.GetWithEpisodes(seriesID, seasonNum)` - season + episodes
- [ ] `EpisodeRepo.GetBySeriesAndNumber(seriesID, season, episode)` - lookup

### 8. Repository: Genre & People
- [ ] `repository/genre.go` - CRUD, link/unlink media
- [ ] `repository/person.go` - CRUD, credits

### 9. Service Layer Updates
- [ ] `service/media.go` - update to use new model
- [ ] `service/series.go` - NEW
- [ ] Update existing handlers to work with new model (backward compat)

## Files to Create/Modify
- `internal/database/migrate/migrations/002_refactor_media_item.go` - NEW
- `internal/database/migrate/migrations/003_series_model.go` - NEW
- `internal/database/migrate/migrations/004_genres_people.go` - NEW
- `internal/model/media.go` - Refactor (was model.go)
- `internal/model/series.go` - NEW
- `internal/model/genre.go` - NEW
- `internal/model/person.go` - NEW
- `internal/repository/media.go` - Refactor
- `internal/repository/media_file.go` - NEW
- `internal/repository/series.go` - NEW
- `internal/repository/genre.go` - NEW
- `internal/repository/person.go` - NEW
- `internal/service/media.go` - Update
- `internal/service/series.go` - NEW

---
Next: phase-03-scan-pipeline.md
