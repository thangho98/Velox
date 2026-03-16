# Phase 04: NFO Write + Sync
Status: ⬜ Pending
Dependencies: Phase 01 (PATCH API), Phase 03 (Editor UI save_nfo checkbox)

## Objective
Khi admin edit metadata và chọn "Save NFO", ghi file NFO ra disk cạnh video file.
Format chuẩn Kodi/Emby XML để portable giữa các media server.
Rescan pipeline respect `metadata_locked` flag — không ghi đè edit thủ công.

## Requirements

### Functional
- [ ] NFO Writer: generate XML từ DB metadata cho Movie, TVShow, Episode
- [ ] Movie NFO: ghi `movie.nfo` hoặc `{basename}.nfo` cạnh video file
- [ ] TVShow NFO: ghi `tvshow.nfo` trong series folder
- [ ] Episode NFO: ghi `{basename}.nfo` cạnh video file
- [ ] NFO chứa: title, sort_title, overview (plot), tagline, rating, genres, cast/crew, IDs (tmdb, imdb, tvdb), poster/fanart paths, release date
- [ ] Khi save_nfo=true trong PATCH request → auto-write NFO
- [ ] Admin endpoint: POST `/api/media/{id}/nfo` — force write NFO cho single item
- [ ] Admin endpoint: POST `/api/admin/nfo/export` — bulk export NFO cho toàn bộ library
- [ ] Scanner pipeline: skip MetadataMatcher nếu `metadata_locked = true`

### Non-Functional
- [ ] UTF-8 encoding với XML declaration
- [ ] Pretty-print XML (indented, readable)
- [ ] Backup: rename existing NFO → `.nfo.bak` trước khi overwrite
- [ ] Handle Unicode characters (Vietnamese, CJK, etc.)
- [ ] File permissions: match parent directory

## Implementation Steps

### 1. NFO Writer Package
- [ ] Tạo `backend/pkg/nfo/writer.go` — NEW (cùng package với parser):
  ```go
  // WriteMovie generates a movie.nfo XML file
  func WriteMovie(m *MovieNFO, path string) error

  // WriteTVShow generates a tvshow.nfo XML file
  func WriteTVShow(s *TVShowNFO, path string) error

  // WriteEpisode generates an episode .nfo XML file
  func WriteEpisode(e *EpisodeNFO, path string) error
  ```
- [ ] Tạo NFO structs cho write (reuse existing parser structs nếu có xml tags):
  ```go
  // MovieNFO — output struct for movie.nfo
  type MovieNFO struct {
      XMLName     xml.Name `xml:"movie"`
      Title       string   `xml:"title"`
      SortTitle   string   `xml:"sorttitle,omitempty"`
      Overview    string   `xml:"plot"`
      Tagline     string   `xml:"tagline,omitempty"`
      Rating      float64  `xml:"rating"`
      Year        int      `xml:"year,omitempty"`
      Premiered   string   `xml:"premiered,omitempty"`
      MPAA        string   `xml:"mpaa,omitempty"`
      UniqueIDs   []UniqueID `xml:"uniqueid"`
      Genres      []string `xml:"genre"`
      Actors      []Actor  `xml:"actor"`
      Directors   []string `xml:"director"`
      Credits     []string `xml:"credits"` // writers
      Poster      string   `xml:"thumb,omitempty"`
      Fanart      *Fanart  `xml:"fanart,omitempty"`
  }
  ```
- [ ] XML output format phải match Kodi/Emby NFO spec:
  ```xml
  <?xml version="1.0" encoding="UTF-8" standalone="yes"?>
  <movie>
    <title>The Matrix</title>
    <sorttitle>Matrix, The</sorttitle>
    <plot>A computer hacker learns...</plot>
    <tagline>Welcome to the Real World</tagline>
    <rating>8.7</rating>
    <year>1999</year>
    <premiered>1999-03-31</premiered>
    <uniqueid type="tmdb" default="true">603</uniqueid>
    <uniqueid type="imdb">tt0133093</uniqueid>
    <genre>Action</genre>
    <genre>Science Fiction</genre>
    <actor>
      <name>Keanu Reeves</name>
      <role>Neo</role>
      <order>0</order>
    </actor>
    <director>Lana Wachowski</director>
  </movie>
  ```

### 2. Converter Functions
- [ ] `backend/pkg/nfo/converter.go` — NEW:
  ```go
  // FromMedia converts DB model + genres + credits to MovieNFO
  func FromMedia(media model.Media, genres []model.Genre, credits []model.CreditWithPerson) *MovieNFO

  // FromSeries converts DB model + genres + credits to TVShowNFO
  func FromSeries(series model.Series, genres []model.Genre, credits []model.CreditWithPerson) *TVShowNFO

  // FromEpisode converts DB media (episode type) to EpisodeNFO
  func FromEpisode(media model.Media, series model.Series) *EpisodeNFO
  ```

### 3. NFO Path Resolution
- [ ] `backend/pkg/nfo/writer.go` — thêm path resolution:
  ```go
  // MovieNFOPath returns the path where movie.nfo should be written
  // Strategy: prefer {basename}.nfo, fallback to movie.nfo in same dir
  func MovieNFOPath(videoPath string) string

  // TVShowNFOPath returns tvshow.nfo path in series directory
  func TVShowNFOPath(seriesDir string) string

  // EpisodeNFOPath returns {basename}.nfo for episode
  func EpisodeNFOPath(videoPath string) string
  ```
- [ ] Resolve video file path: lấy từ `media_files` table (primary file)

### 4. Service Integration
- [ ] `service/metadata.go` — implement NFO write trong `EditMediaMetadata()`:
  ```go
  if req.SaveNFO {
      // Get primary media file path
      // Get genres + credits
      // Convert to NFO struct
      // Write to disk
      s.writeMediaNFO(ctx, media)
  }
  ```
- [ ] `service/metadata.go` — thêm `WriteMediaNFO(ctx, mediaID)` public method
- [ ] `service/metadata.go` — thêm `WriteSeriesNFO(ctx, seriesID)` public method
- [ ] `service/metadata.go` — thêm `BulkExportNFO(ctx, libraryID)`:
  - List all media in library
  - Write NFO for each (skip failures, log errors)
  - Return count of success/failures

### 5. Handler + Routes
- [ ] `handler/metadata.go` — thêm `WriteMediaNFO(w, r)`:
  - POST `/api/media/{id}/nfo` — force write NFO cho single item
  - Admin only
- [ ] `handler/metadata.go` — thêm `WriteSeriesNFO(w, r)`:
  - POST `/api/series/{id}/nfo` — force write NFO cho series
- [ ] `handler/metadata.go` — thêm `BulkExportNFO(w, r)`:
  - POST `/api/admin/nfo/export` — bulk export cho library
  - Optional query: `?library_id=1`
  - Returns: `{"data": {"total": 150, "success": 148, "failed": 2}}`
- [ ] `cmd/server/main.go` — wire routes

## Files to Create/Modify
- `backend/pkg/nfo/writer.go` — NEW: XML writer functions
- `backend/pkg/nfo/converter.go` — NEW: model → NFO struct converters
- `backend/internal/service/metadata.go` — NFO write integration + BulkExport
- `backend/internal/handler/metadata.go` — WriteNFO + BulkExport handlers
- `backend/cmd/server/main.go` — new routes

## Test Criteria
- [ ] WriteMovie → generates valid XML with correct structure
- [ ] WriteMovie → UTF-8 encoding, XML declaration present
- [ ] WriteMovie → genres as separate `<genre>` elements
- [ ] WriteMovie → actors with name, role, order
- [ ] WriteMovie → uniqueid for tmdb + imdb
- [ ] WriteTVShow → tvshow.nfo with series fields + seasons
- [ ] WriteEpisode → {basename}.nfo with episode fields
- [ ] PATCH with save_nfo=true → NFO file created on disk
- [ ] NFO file content matches DB metadata
- [ ] Existing NFO → backup to .nfo.bak before overwrite
- [ ] Bulk export → processes all media, returns success/fail count
- [ ] Unicode characters (Vietnamese) → correctly encoded in XML
- [ ] Round-trip test: write NFO → parse NFO → compare fields

## Notes
- Reuse existing parser structs từ `pkg/nfo/parser.go` nếu có `xml` tags phù hợp cho marshal
- Nếu parser structs chỉ có unmarshal tags, tạo separate write structs
- `encoding/xml` trong Go stdlib đủ mạnh cho NFO format
- Backup .nfo.bak chỉ giữ 1 version (overwrite backup cũ)
- Bulk export nên chạy async (goroutine) nếu library lớn — hoặc dùng existing task system

---
Previous Phase: [Phase 03 — Frontend Editor UI](phase-03-editor-ui.md)
