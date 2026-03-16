# Phase 02: Image Upload
Status: ⬜ Pending
Dependencies: Phase 01 (PATCH API exists)

## Objective
Cho phép admin upload poster/backdrop thủ công thay vì chỉ dùng ảnh TMDb proxy.
Upload lưu local trên disk, serve qua endpoint riêng, DB lưu `local://` prefix để phân biệt.

## Requirements

### Functional
- [ ] POST `/api/media/{id}/images` — upload poster hoặc backdrop (multipart/form-data)
- [ ] POST `/api/series/{id}/images` — upload cho series
- [ ] GET `/api/images/local/{type}/{id}/{filename}` — serve local images
- [ ] DELETE `/api/media/{id}/images/{image_type}` — xóa custom image, revert về TMDb
- [ ] Resize image: poster max 1000x1500, backdrop max 1920x1080
- [ ] Accept formats: JPEG, PNG, WebP → save as JPEG (quality 90)
- [ ] Update DB poster_path/backdrop_path với `local://` prefix
- [ ] Auto-set `metadata_locked = true` khi upload image

### Non-Functional
- [ ] Max upload size: 10MB
- [ ] Admin-only access
- [ ] MIME type validation (không chỉ dựa vào extension)
- [ ] Local images cached 30 days (Cache-Control header)
- [ ] Path traversal protection trên serve endpoint

## Implementation Steps

### 1. Storage Helper
- [ ] Tạo `backend/internal/storage/image.go`:
  ```go
  type ImageStorage struct {
      dataDir string // VELOX_DATA_DIR
  }

  func (s *ImageStorage) Save(mediaType string, id int64, imageType string, data []byte) (string, error)
  // Returns: "local://{id}/poster.jpg"
  // Saves to: {dataDir}/images/{mediaType}/{id}/{imageType}.jpg

  func (s *ImageStorage) Delete(mediaType string, id int64, imageType string) error

  func (s *ImageStorage) Path(mediaType string, id int64, filename string) string
  // Returns absolute filesystem path

  func (s *ImageStorage) Exists(mediaType string, id int64, imageType string) bool
  ```

### 2. Image Processing
- [ ] Tạo `backend/internal/storage/resize.go`:
  - Dùng `golang.org/x/image` + standard `image/jpeg` (không cần CGO dependency)
  - Hoặc `github.com/disintegration/imaging` (popular, pure Go, resize + crop)
  - `ProcessImage(data []byte, maxW, maxH int) ([]byte, error)`:
    - Decode (JPEG/PNG/WebP)
    - Resize fit within maxW x maxH (maintain aspect ratio)
    - Encode JPEG quality 90
  - Validate MIME by reading magic bytes, not Content-Type header

### 3. Upload Handler
- [ ] `handler/metadata.go`: thêm `UploadMediaImage(w, r)`:
  - Parse multipart form (max 10MB)
  - Read `image_type` field ("poster" or "backdrop")
  - Read `file` field
  - Validate: size ≤ 10MB, MIME is image/jpeg|png|webp
  - Process image (resize)
  - Save via ImageStorage
  - Update media poster_path/backdrop_path in DB
  - Set metadata_locked = true
  - respondJSON with new path
- [ ] `handler/metadata.go`: thêm `UploadSeriesImage(w, r)` tương tự
- [ ] `handler/metadata.go`: thêm `DeleteMediaImage(w, r)`:
  - Delete local file
  - Set poster_path/backdrop_path = "" (or revert to TMDb path if known)
  - respondJSON success

### 4. Local Image Serve Handler
- [ ] `handler/image.go`: thêm route `ServeLocal(w, r)`:
  - `GET /api/images/local/{type}/{id}/{filename}`
  - type: "media" or "series"
  - Validate: id is numeric, filename has no path separators
  - Resolve absolute path via ImageStorage
  - `http.ServeFile()` with Cache-Control: 30 days
  - 404 if file doesn't exist

### 5. Frontend Image Helper Update
- [ ] `webapp/src/lib/image.ts`: update `tmdbImage()` hoặc tạo `mediaImage()`:
  ```typescript
  export function mediaImage(path: string, size: string): string {
    if (path.startsWith('local://')) {
      // local://42/poster.jpg → /api/images/local/media/42/poster.jpg
      return `/api/images/local/media/${path.slice(8)}`
    }
    return tmdbImage(path, size)
  }
  ```
  - Cần sửa tất cả nơi dùng `tmdbImage()` trong detail pages + browse cards

### 6. Route Wiring
- [ ] `cmd/server/main.go`:
  ```go
  mux.Handle("POST /api/media/{id}/images", adminOnly(metadataHandler.UploadMediaImage))
  mux.Handle("POST /api/series/{id}/images", adminOnly(metadataHandler.UploadSeriesImage))
  mux.Handle("DELETE /api/media/{id}/images/{imageType}", adminOnly(metadataHandler.DeleteMediaImage))
  mux.Handle("DELETE /api/series/{id}/images/{imageType}", adminOnly(metadataHandler.DeleteSeriesImage))
  mux.Handle("GET /api/images/local/{type}/{id}/{filename}", imageHandler.ServeLocal)
  ```

## Files to Create/Modify
- `backend/internal/storage/image.go` — NEW: image storage helper
- `backend/internal/storage/resize.go` — NEW: image processing
- `backend/internal/handler/metadata.go` — upload + delete handlers
- `backend/internal/handler/image.go` — ServeLocal handler
- `backend/cmd/server/main.go` — new routes
- `webapp/src/lib/image.ts` — mediaImage() helper
- `webapp/src/pages/MediaDetailPage.tsx` — use mediaImage()
- `webapp/src/pages/SeriesDetailPage.tsx` — use mediaImage()
- `webapp/src/components/browse/MediaCard.tsx` — use mediaImage()
- `go.mod` — thêm imaging dependency (nếu dùng)

## Test Criteria
- [ ] Upload valid JPEG → saved to correct path, DB updated, response contains local:// path
- [ ] Upload 15MB file → 400 Bad Request (too large)
- [ ] Upload text file with .jpg extension → 400 Bad Request (invalid MIME)
- [ ] Upload PNG → processed and saved as JPEG
- [ ] Upload 4000x6000 poster → resized to 1000x1500
- [ ] GET local image → 200 with correct content-type and cache headers
- [ ] GET non-existent image → 404
- [ ] Path traversal attempt (../../etc/passwd) → 400 Bad Request
- [ ] Delete custom image → file removed, DB path cleared
- [ ] Non-admin upload → 403 Forbidden
- [ ] Detail page renders local:// image correctly
- [ ] Browse card renders local:// poster correctly

## Notes
- Prefer `github.com/disintegration/imaging` — pure Go, no CGO, supports JPEG/PNG/GIF/TIFF/BMP
- WebP decode có thể cần `golang.org/x/image/webp`
- Khi delete custom image, không tự động revert về TMDb — admin phải Refresh Metadata nếu muốn
- Directory structure: `{dataDir}/images/media/{id}/poster.jpg` — flat per-entity, dễ cleanup

---
Next Phase: [Phase 03 — Frontend Editor UI](phase-03-editor-ui.md)
