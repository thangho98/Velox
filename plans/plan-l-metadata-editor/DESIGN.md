# 🎨 DESIGN: Plan L — Metadata Editor (Emby-style)

Ngày tạo: 2026-03-16
Dựa trên: `plans/plan-l-metadata-editor/plan.md`

---

## 1. Thay đổi Database (Migration 021)

### 1.1. Sơ đồ thay đổi

```
┌─────────────────────────────────────────────────────────┐
│  📦 media (EXISTING — thêm 2 cột mới)                  │
│  ├── ... (tất cả cột hiện tại giữ nguyên)              │
│  ├── + tagline TEXT DEFAULT ''          ← MỚI           │
│  └── + metadata_locked INTEGER DEFAULT 0 ← MỚI         │
│       (0 = không khóa, 1 = đã khóa)                    │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  📦 series (EXISTING — thêm 1 cột mới)                 │
│  ├── ... (tất cả cột hiện tại giữ nguyên)              │
│  └── + metadata_locked INTEGER DEFAULT 0 ← MỚI         │
└─────────────────────────────────────────────────────────┘
```

### 1.2. Migration SQL

```sql
-- Migration 021: metadata_lock
-- Up
ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT '';
ALTER TABLE media ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
ALTER TABLE series ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;

-- Down (SQLite 3.35.0+)
ALTER TABLE media DROP COLUMN tagline;
ALTER TABLE media DROP COLUMN metadata_locked;
ALTER TABLE series DROP COLUMN metadata_locked;
```

### 1.3. Impact trên code hiện tại

Các chỗ cần update sau khi thêm column:

| File | Thay đổi |
|------|----------|
| `model/media.go` — Media struct | Thêm `Tagline string`, `MetadataLocked bool` |
| `model/series.go` — Series struct | Thêm `MetadataLocked bool` |
| `repository/media.go` — `mediaColumns` | Thêm `tagline, metadata_locked` |
| `repository/media.go` — `scanMedia()` | Scan thêm 2 fields |
| `repository/media.go` — `Create()` | Insert thêm tagline, metadata_locked |
| `repository/media.go` — `Update()` | SET thêm tagline, metadata_locked |
| `repository/series.go` — tương tự | Columns + scan + create + update |

---

## 2. API Design Chi Tiết

### 2.1. PATCH /api/media/{id}/metadata

**Request body** (tất cả fields optional — chỉ gửi field cần sửa):

```json
{
  "title": "Ma Trận",
  "sort_title": "Ma Trận",
  "overview": "Một hacker máy tính phát hiện...",
  "tagline": "Chào mừng đến thế giới thực",
  "release_date": "1999-03-31",
  "rating": 8.7,
  "genres": ["Hành Động", "Khoa Học Viễn Tưởng"],
  "credits": [
    {"person_name": "Keanu Reeves", "character": "Neo", "role": "cast", "order": 0},
    {"person_name": "Lana Wachowski", "role": "director", "order": 0}
  ],
  "save_nfo": false,
  "metadata_locked": true
}
```

**Go struct cho partial update:**

```go
// MetadataEditRequest represents a partial metadata edit.
// Pointer fields: nil = don't change. Non-pointer fields: always applied if present.
type MetadataEditRequest struct {
    Title       *string       `json:"title"`
    SortTitle   *string       `json:"sort_title"`
    Overview    *string       `json:"overview"`
    Tagline     *string       `json:"tagline"`
    ReleaseDate *string       `json:"release_date"`
    Rating      *float64      `json:"rating"`
    Genres      []string      `json:"genres"`      // nil = don't change, empty = clear all
    Credits     []CreditInput `json:"credits"`      // nil = don't change, empty = clear all
    SaveNFO     bool          `json:"save_nfo"`
    MetadataLocked *bool      `json:"metadata_locked"` // nil = auto-set true
}

type CreditInput struct {
    PersonName string `json:"person_name"`
    Character  string `json:"character,omitempty"`
    Role       string `json:"role"` // "cast" | "director" | "writer" | "producer"
    Order      int    `json:"order"`
}
```

**Partial update logic:**

```go
func (r *MediaRepo) UpdateMetadata(ctx context.Context, id int64, req MetadataEditRequest) error {
    // Build dynamic SET clauses
    setClauses := []string{}
    args := []any{}

    if req.Title != nil {
        setClauses = append(setClauses, "title = ?")
        args = append(args, *req.Title)
    }
    if req.SortTitle != nil {
        setClauses = append(setClauses, "sort_title = ?")
        args = append(args, *req.SortTitle)
    }
    // ... repeat for each pointer field ...

    // Always update timestamp + lock
    setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

    if len(setClauses) == 0 {
        return nil // nothing to update
    }

    query := fmt.Sprintf("UPDATE media SET %s WHERE id = ?", strings.Join(setClauses, ", "))
    args = append(args, id)
    _, err := r.db.ExecContext(ctx, query, args...)
    return err
}
```

**Response:**

```json
{
  "data": {
    "id": 42,
    "title": "Ma Trận",
    "sort_title": "Ma Trận",
    "tagline": "Chào mừng đến thế giới thực",
    "metadata_locked": true,
    "...": "..."
  }
}
```

### 2.2. PATCH /api/series/{id}/metadata

Tương tự media, thêm fields riêng cho series:

```go
type SeriesMetadataEditRequest struct {
    Title        *string  `json:"title"`
    SortTitle    *string  `json:"sort_title"`
    Overview     *string  `json:"overview"`
    Status       *string  `json:"status"`        // "Returning Series" | "Ended" | "Canceled"
    Network      *string  `json:"network"`
    FirstAirDate *string  `json:"first_air_date"`
    Genres       []string `json:"genres"`
    Credits      []CreditInput `json:"credits"`
    SaveNFO      bool     `json:"save_nfo"`
    MetadataLocked *bool  `json:"metadata_locked"`
}
```

### 2.3. POST /api/media/{id}/images

**Request:** multipart/form-data

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image_type` | string | ✅ | "poster" or "backdrop" |
| `file` | binary | ✅ | Image file (JPEG/PNG/WebP, max 10MB) |

**Processing pipeline:**

```
Upload → Validate MIME (magic bytes) → Validate size (≤10MB)
  → Resize (poster: fit 1000x1500, backdrop: fit 1920x1080)
  → Encode JPEG (quality 90)
  → Save to {VELOX_DATA_DIR}/images/media/{id}/{image_type}.jpg
  → Update DB: poster_path = "local://{id}/poster.jpg"
  → Set metadata_locked = true
  → Return {"data": {"path": "local://42/poster.jpg"}}
```

**Response:**

```json
{
  "data": {
    "path": "local://42/poster.jpg",
    "image_type": "poster"
  }
}
```

### 2.4. GET /api/images/local/{type}/{id}/{filename}

Serve local uploaded images.

| Param | Example | Description |
|-------|---------|-------------|
| `type` | "media" or "series" | Entity type |
| `id` | "42" | Entity ID |
| `filename` | "poster.jpg" | File name |

**Headers:**
- `Content-Type: image/jpeg`
- `Cache-Control: public, max-age=2592000` (30 days)

**Security:**
- Validate `id` is numeric
- Validate `filename` has no path separators (`/`, `\`, `..`)
- Validate `type` is "media" or "series"

### 2.5. DELETE /api/media/{id}/metadata/lock

Unlock metadata → cho phép rescan ghi đè.

```json
// Response
{"data": {"metadata_locked": false}}
```

### 2.6. Tổng hợp Routes mới

```go
// Metadata editing (admin only)
PATCH  /api/media/{id}/metadata          → EditMediaMetadata
PATCH  /api/series/{id}/metadata         → EditSeriesMetadata
DELETE /api/media/{id}/metadata/lock     → UnlockMediaMetadata
DELETE /api/series/{id}/metadata/lock    → UnlockSeriesMetadata

// Image upload (admin only)
POST   /api/media/{id}/images            → UploadMediaImage
POST   /api/series/{id}/images           → UploadSeriesImage
DELETE /api/media/{id}/images/{imageType} → DeleteMediaImage
DELETE /api/series/{id}/images/{imageType}→ DeleteSeriesImage

// Local image serve (public — images don't need auth)
GET    /api/images/local/{type}/{id}/{filename} → ServeLocal

// NFO write (admin only) — Phase 04
POST   /api/media/{id}/nfo              → WriteMediaNFO
POST   /api/series/{id}/nfo             → WriteSeriesNFO
POST   /api/admin/nfo/export            → BulkExportNFO
```

---

## 3. Service Layer Logic

### 3.1. EditMediaMetadata Flow

```
Handler.EditMediaMetadata(w, r)
  │
  ├── Parse {id} from URL
  ├── Decode JSON body → MetadataEditRequest
  ├── Validate: title not empty (if provided), date format
  │
  ├── service.EditMediaMetadata(ctx, mediaID, req)
  │     │
  │     ├── mediaRepo.GetByID(ctx, id) → media (verify exists)
  │     │
  │     ├── BEGIN TRANSACTION
  │     │     ├── mediaRepo.UpdateMetadata(ctx, id, req) → update scalar fields
  │     │     ├── if req.Genres != nil:
  │     │     │     ├── genreRepo.ClearMediaGenres(ctx, id)
  │     │     │     └── for each genre name:
  │     │     │           ├── genreRepo.GetByName() or Create()
  │     │     │           └── genreRepo.LinkToMedia()
  │     │     ├── if req.Credits != nil:
  │     │     │     ├── personRepo.ClearMediaCredits(ctx, id)
  │     │     │     └── for each credit:
  │     │     │           ├── personRepo.GetByName() or Create()
  │     │     │           └── personRepo.AddCredit()
  │     │     └── Set metadata_locked = true (unless explicit false)
  │     ├── COMMIT
  │     │
  │     └── if req.SaveNFO → writeMediaNFO(ctx, media) [Phase 04]
  │
  ├── mediaRepo.GetByID(ctx, id) → re-fetch updated media
  └── respondJSON(w, 200, updated)
```

### 3.2. Scanner Pipeline — Metadata Lock Check

Hiện tại trong `service/metadata.go`:

```go
// BEFORE (current)
func (s *MetadataService) MatchAndPersistMovie(ctx, media, parsed, filePath, force) {
    if !force && media.TmdbID != nil {
        return nil
    }
    // ... fetch TMDb and overwrite ...
}

// AFTER (with lock check)
func (s *MetadataService) MatchAndPersistMovie(ctx, media, parsed, filePath, force) {
    if !force && media.TmdbID != nil {
        return nil
    }
    // NEW: Check metadata lock
    if !force && media.MetadataLocked {
        return nil // Respect manual edits
    }
    // ... fetch TMDb and overwrite ...
}
```

**Lock behavior matrix:**

| Action | metadata_locked=true | metadata_locked=false |
|--------|---------------------|----------------------|
| Rescan (auto) | ❌ Skip metadata refresh | ✅ Override from TMDb |
| Refresh (manual button) | ✅ Allow (admin chose this) | ✅ Allow |
| Identify (re-match TMDb) | ✅ Allow + auto-unlock | ✅ Allow |
| PATCH edit | ✅ Allow + keep locked | ✅ Allow + auto-lock |

### 3.3. Genre Sync — FindOrCreate by Name

Khác với TMDb sync (có tmdb_id), manual edit chỉ có genre name:

```go
// ensureGenreByName gets or creates a genre by name (no TMDb ID).
func (s *MetadataService) ensureGenreByName(ctx context.Context, name string) (int64, error) {
    existing, err := s.genreRepo.GetByName(ctx, name)
    if err == nil {
        return existing.ID, nil
    }
    // Create without TMDb ID
    genre := &model.Genre{Name: name}
    if err := s.genreRepo.Create(ctx, genre); err != nil {
        return 0, err
    }
    return genre.ID, nil
}
```

### 3.4. Credit Sync — FindOrCreate Person by Name

```go
// ensurePersonByName gets or creates a person by name (no TMDb ID).
func (s *MetadataService) ensurePersonByName(ctx context.Context, name string) (int64, error) {
    existing, err := s.personRepo.GetByName(ctx, name)
    if err == nil {
        return existing.ID, nil
    }
    person := &model.Person{Name: name}
    if err := s.personRepo.Create(ctx, person); err != nil {
        return 0, err
    }
    return person.ID, nil
}
```

**Note:** PersonRepo cần thêm `GetByName()` method (hiện chỉ có GetByTmdbID).

---

## 4. Image Storage Design

### 4.1. Directory Structure

```
{VELOX_DATA_DIR}/
  images/
    media/
      42/
        poster.jpg       ← uploaded poster
        backdrop.jpg     ← uploaded backdrop
      105/
        poster.jpg
    series/
      7/
        poster.jpg
        backdrop.jpg
```

### 4.2. Image Path Convention

DB stores 2 types of image paths:

| Type | Format | Example | Served via |
|------|--------|---------|------------|
| TMDb | `/{hash}.jpg` | `/kqjL17yufvn9OVLyXYpvtyrFfak.jpg` | `GET /api/images/tmdb/{size}/{path}` |
| Local | `local://{id}/{file}` | `local://42/poster.jpg` | `GET /api/images/local/media/42/poster.jpg` |

### 4.3. Frontend Image Helper

```typescript
// webapp/src/lib/image.ts

// BEFORE: only TMDb
export function tmdbImage(path: string, size: string): string {
  if (!path) return ''
  const cleaned = path.startsWith('/') ? path.slice(1) : path
  return `/api/images/tmdb/${size}/${cleaned}`
}

// AFTER: TMDb + Local
export function mediaImage(path: string, size: string = 'w500'): string {
  if (!path) return ''
  if (path.startsWith('local://')) {
    // local://42/poster.jpg → /api/images/local/media/42/poster.jpg
    return `/api/images/local/media/${path.slice(8)}`
  }
  return tmdbImage(path, size)
}

export function seriesImage(path: string, size: string = 'w500'): string {
  if (!path) return ''
  if (path.startsWith('local://')) {
    return `/api/images/local/series/${path.slice(8)}`
  }
  return tmdbImage(path, size)
}
```

### 4.4. Image Processing

```go
// backend/internal/storage/resize.go

func ProcessImage(data []byte, maxWidth, maxHeight int) ([]byte, error) {
    // 1. Detect format from magic bytes
    // 2. Decode (jpeg/png/webp)
    // 3. Resize fit (maintain aspect ratio)
    // 4. Encode as JPEG quality 90
    // 5. Return bytes
}

// Preset dimensions
var (
    PosterMaxSize   = [2]int{1000, 1500}
    BackdropMaxSize = [2]int{1920, 1080}
)
```

**Dependency:** `github.com/disintegration/imaging` (pure Go, no CGO)

---

## 5. NFO Write Design (Phase 04)

### 5.1. NFO XML Format

**Movie NFO (`movie.nfo`):**

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
  <title>The Matrix</title>
  <originaltitle>The Matrix</originaltitle>
  <sorttitle>Matrix, The</sorttitle>
  <rating>8.7</rating>
  <year>1999</year>
  <plot>A computer hacker learns about the true nature of his reality...</plot>
  <tagline>Welcome to the Real World</tagline>
  <runtime>136</runtime>
  <premiered>1999-03-31</premiered>
  <uniqueid type="tmdb" default="true">603</uniqueid>
  <uniqueid type="imdb">tt0133093</uniqueid>
  <genre>Action</genre>
  <genre>Science Fiction</genre>
  <actor>
    <name>Keanu Reeves</name>
    <role>Neo</role>
    <order>0</order>
    <thumb>/abc123.jpg</thumb>
  </actor>
  <director>Lana Wachowski</director>
  <credits>Lana Wachowski</credits>
</movie>
```

### 5.2. NFO File Location

| Type | Location | Filename |
|------|----------|----------|
| Movie | Same dir as video | `{video_basename}.nfo` |
| TVShow | Series root folder | `tvshow.nfo` |
| Episode | Same dir as video | `{video_basename}.nfo` |

**Resolution logic:**
1. Get primary `media_file` → `file_path`
2. Extract directory + basename
3. Replace extension with `.nfo`

### 5.3. Write Safety

- Backup existing: rename `movie.nfo` → `movie.nfo.bak` before write
- Atomic write: write to `.nfo.tmp` → rename to `.nfo`
- Encoding: UTF-8 with BOM (Kodi/Emby compatible)
- Pretty-print: 2-space indentation

---

## 6. Luồng Hoạt Động (User Journeys)

### 6.1. Admin sửa metadata phim

```
1️⃣ Admin vào trang chi tiết phim "The Matrix"
2️⃣ Thấy tiêu đề tiếng Anh, muốn đổi sang tiếng Việt
3️⃣ Click nút "Edit Metadata" (✏️)
4️⃣ Panel editor trượt ra từ bên phải
5️⃣ Sửa Title: "Ma Trận"
6️⃣ Sửa Overview: "Một hacker máy tính..."
7️⃣ Thêm genre "Viễn Tưởng" (gõ + Enter)
8️⃣ Xóa genre "Science Fiction" (click ✕)
9️⃣ Check "Lock Metadata" ☑ (mặc định đã check)
🔟 Click "Save" → API gọi → Panel đóng → Trang cập nhật
```

### 6.2. Admin upload poster riêng

```
1️⃣ Admin mở editor → mục "Images"
2️⃣ Thấy poster hiện tại từ TMDb
3️⃣ Kéo thả file poster mới vào drop zone
4️⃣ Preview hiện lên → Đẹp → Click "Upload"
5️⃣ API upload → resize → save → DB cập nhật
6️⃣ Trang chi tiết hiện poster mới
7️⃣ Browse page cũng hiện poster mới (vì cùng poster_path)
```

### 6.3. Rescan không ghi đè edit thủ công

```
1️⃣ Admin đã edit "The Matrix" → metadata_locked = true
2️⃣ Admin chạy Rescan Library
3️⃣ Scanner tìm file "The.Matrix.1999.mkv"
4️⃣ Check media.MetadataLocked == true
5️⃣ SKIP TMDb fetch → giữ nguyên metadata đã edit
6️⃣ Admin thở phào 😌
```

### 6.4. Admin muốn revert về TMDb

```
1️⃣ Admin vào detail page → thấy "🔒 Locked" badge
2️⃣ Click "Unlock Metadata"
3️⃣ Confirm: "Bạn chắc chắn? Rescan tiếp theo sẽ ghi đè metadata."
4️⃣ Click "OK" → DELETE /api/media/{id}/metadata/lock
5️⃣ metadata_locked = false
6️⃣ Click "Refresh Metadata" → TMDb fetch → metadata quay về bản gốc
```

---

## 7. Checklist Kiểm Tra (Test Cases)

### TC-01: Edit media metadata — Happy path
```
Given: Admin đăng nhập, media "The Matrix" (id=42) tồn tại
When:  PATCH /api/media/42/metadata {"title": "Ma Trận", "overview": "Mô tả mới"}
Then:  ✓ Response 200 với media đã cập nhật
       ✓ DB: title = "Ma Trận", overview = "Mô tả mới"
       ✓ DB: metadata_locked = 1 (auto-set)
       ✓ DB: updated_at đã thay đổi
       ✓ Các field khác (rating, poster_path...) không đổi
```

### TC-02: Edit genres — Replace
```
Given: Media 42 có genres ["Action", "Sci-Fi"]
When:  PATCH /api/media/42/metadata {"genres": ["Hành Động", "Viễn Tưởng"]}
Then:  ✓ Old genres cleared (media_genres WHERE media_id=42 deleted)
       ✓ New genres created if not exist
       ✓ New links created
       ✓ GET /api/media/42 shows genres ["Hành Động", "Viễn Tưởng"]
```

### TC-03: Edit credits — Add cast member
```
Given: Media 42 có cast [Keanu Reeves → Neo]
When:  PATCH /api/media/42/metadata {"credits": [
         {"person_name": "Keanu Reeves", "character": "Neo", "role": "cast", "order": 0},
         {"person_name": "Hugo Weaving", "character": "Agent Smith", "role": "cast", "order": 1}
       ]}
Then:  ✓ Old credits cleared
       ✓ Person "Hugo Weaving" created if not exist
       ✓ 2 credits linked
```

### TC-04: Partial update — Only title
```
Given: Media 42 has title="The Matrix", overview="Original overview", rating=8.7
When:  PATCH /api/media/42/metadata {"title": "Ma Trận"}
Then:  ✓ title = "Ma Trận"
       ✓ overview = "Original overview" (unchanged)
       ✓ rating = 8.7 (unchanged)
```

### TC-05: metadata_locked blocks rescan
```
Given: Media 42 has metadata_locked=1, title="Ma Trận"
When:  Scanner runs MatchAndPersistMovie(ctx, media42, ...)
Then:  ✓ Function returns nil immediately
       ✓ title still "Ma Trận" (NOT overwritten by TMDb)
```

### TC-06: Identify auto-unlocks
```
Given: Media 42 has metadata_locked=1
When:  PUT /api/media/42/identify {"tmdb_id": 603, "media_type": "movie"}
Then:  ✓ metadata_locked = 0
       ✓ Metadata refreshed from TMDb
```

### TC-07: Image upload — Valid JPEG
```
Given: Admin uploads poster.jpg (800x1200, 500KB)
When:  POST /api/media/42/images {image_type: "poster", file: poster.jpg}
Then:  ✓ File saved to {DATA_DIR}/images/media/42/poster.jpg
       ✓ Resized to fit 1000x1500 (aspect ratio maintained)
       ✓ DB: poster_path = "local://42/poster.jpg"
       ✓ DB: metadata_locked = 1
       ✓ Response: {"data": {"path": "local://42/poster.jpg"}}
```

### TC-08: Image upload — Too large
```
Given: Admin uploads 15MB file
When:  POST /api/media/42/images {file: huge.jpg}
Then:  ✓ Response 400 "file too large (max 10MB)"
       ✓ No file saved
```

### TC-09: Image upload — Invalid MIME
```
Given: Admin uploads text.txt renamed to text.jpg
When:  POST /api/media/42/images {file: text.jpg}
Then:  ✓ Response 400 "invalid image format"
       ✓ Detected by magic bytes, not extension
```

### TC-10: Local image serve
```
Given: File exists at {DATA_DIR}/images/media/42/poster.jpg
When:  GET /api/images/local/media/42/poster.jpg
Then:  ✓ Response 200 with image/jpeg content
       ✓ Cache-Control: public, max-age=2592000
```

### TC-11: Path traversal protection
```
When:  GET /api/images/local/media/../../etc/passwd
Then:  ✓ Response 400 "invalid path"
```

### TC-12: Non-admin forbidden
```
Given: User (not admin) đăng nhập
When:  PATCH /api/media/42/metadata {"title": "test"}
Then:  ✓ Response 403 Forbidden
```

### TC-13: Edit series metadata
```
Given: Series "Breaking Bad" (id=7)
When:  PATCH /api/series/7/metadata {"status": "Ended", "network": "AMC"}
Then:  ✓ status = "Ended", network = "AMC"
       ✓ metadata_locked = 1
```

### TC-14: Delete custom image — revert
```
Given: Media 42 has poster_path = "local://42/poster.jpg"
When:  DELETE /api/media/42/images/poster
Then:  ✓ File deleted from disk
       ✓ DB: poster_path = "" (cleared)
       ✓ Frontend falls back to placeholder icon
```

### TC-15: NFO Write — Movie (Phase 04)
```
Given: Media 42 edited, primary file at /media/movies/The.Matrix.1999.mkv
When:  PATCH /api/media/42/metadata {"save_nfo": true, "title": "Ma Trận"}
Then:  ✓ File created: /media/movies/The.Matrix.1999.nfo
       ✓ XML valid, UTF-8 encoded
       ✓ Contains <title>Ma Trận</title>
       ✓ Contains <uniqueid type="tmdb">603</uniqueid>
       ✓ Contains <genre> elements for each genre
```

### TC-16: Frontend — Edit button visibility
```
Given: Admin user opens MediaDetailPage
Then:  ✓ "Edit Metadata" button visible
       ✓ "Locked" badge visible if metadata_locked

Given: Non-admin user opens same page
Then:  ✓ "Edit Metadata" button NOT visible
       ✓ "Locked" badge still visible (informational)
```

---

## 8. Dependency mới

| Package | Purpose | Size | CGO? |
|---------|---------|------|------|
| `github.com/disintegration/imaging` | Image resize/encode (pure Go) | ~50KB | No |

Alternative nếu không muốn thêm dependency:
- `golang.org/x/image` + stdlib `image/jpeg` — cần viết resize manually
- Recommendation: dùng `imaging` — mature, popular (4k+ stars), pure Go

---

## 9. Files Tổng Hợp (Tất Cả Phases)

### Backend — Tạo mới
| File | Phase | Purpose |
|------|-------|---------|
| `backend/internal/storage/image.go` | 02 | Image save/delete/path helpers |
| `backend/internal/storage/resize.go` | 02 | Image decode/resize/encode |
| `backend/pkg/nfo/writer.go` | 04 | NFO XML generation |
| `backend/pkg/nfo/converter.go` | 04 | Model → NFO struct conversion |

### Backend — Sửa
| File | Phase | Changes |
|------|-------|---------|
| `internal/database/migrate/registry.go` | 01 | Migration 021 |
| `internal/model/media.go` | 01 | +Tagline, +MetadataLocked |
| `internal/model/series.go` | 01 | +MetadataLocked |
| `internal/repository/media.go` | 01 | +UpdateMetadata(), update columns/scan |
| `internal/repository/series.go` | 01 | +UpdateMetadata(), update columns/scan |
| `internal/repository/person.go` | 01 | +GetByName() |
| `internal/service/metadata.go` | 01 | +EditMedia/Series, lock checks in Match* |
| `internal/handler/metadata.go` | 01,02 | +PATCH/DELETE/POST handlers |
| `internal/handler/image.go` | 02 | +ServeLocal() |
| `cmd/server/main.go` | 01,02 | Wire new routes |

### Frontend — Tạo mới
| File | Phase | Purpose |
|------|-------|---------|
| `src/api/metadata.ts` | 03 | API client functions |
| `src/components/metadata/MetadataEditor.tsx` | 03 | Main edit panel |
| `src/components/metadata/GenreEditor.tsx` | 03 | Tag-style genre input |
| `src/components/metadata/CreditEditor.tsx` | 03 | Cast/crew list editor |
| `src/components/metadata/ImageUploader.tsx` | 03 | Drag-and-drop upload |

### Frontend — Sửa
| File | Phase | Changes |
|------|-------|---------|
| `src/types/api.ts` | 03 | +EditRequest types |
| `src/hooks/stores/useMedia.ts` | 03 | +edit/upload mutations |
| `src/lib/image.ts` | 02 | +mediaImage(), seriesImage() |
| `src/pages/MediaDetailPage.tsx` | 03 | +Edit button, editor integration |
| `src/pages/SeriesDetailPage.tsx` | 03 | +Edit button, editor integration |

---

*Tạo bởi /design — Plan L Metadata Editor*
