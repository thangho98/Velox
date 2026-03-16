# Plan L: Metadata Editor (Emby-style)
Created: 2026-03-16
Status: ⬜ Pending
Priority: 🟡 Medium
Dependencies: Plans A-F done (media model, auth, UI, metadata providers)

## Overview
Cho phép admin chỉnh sửa metadata trực tiếp từ UI — giống tính năng "Edit Metadata" của Emby.
Hiện tại Velox chỉ hỗ trợ Identify (re-match TMDb ID) và Refresh (re-fetch TMDb), admin không có cách
sửa tay title, overview, poster, genre, cast... khi metadata tự động bị sai.

Feature này bổ sung:
1. **PATCH API** cho media + series metadata (title, overview, genres, cast, images...)
2. **Image upload** — upload poster/backdrop thủ công thay vì chỉ dùng TMDb proxy
3. **Editor UI** — form chỉnh sửa metadata trên detail page (admin only)
4. **NFO Write** — ghi metadata đã edit ra file NFO trên disk (giống Emby auto-save)
5. **Metadata Lock** — flag "đừng ghi đè khi rescan" để bảo vệ edit thủ công

## Hiện trạng

### Đã có
- **Media model** (`model/media.go`): title, sort_title, overview, release_date, rating, poster_path, backdrop_path, logo_path, thumb_path, tmdb_id, imdb_id, tvdb_id, imdb_rating, rt_score, metacritic_score
- **Series model** (`model/series.go`): title, sort_title, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path
- **Repository Update** (`repository/media.go`): `Update(ctx, media)` updates ALL editable fields at once
- **Genre repo** (`repository/genre.go`): `ClearMediaGenres()` + `LinkToMedia()` — clear-and-sync pattern
- **Credits repo** (`repository/person.go`): `ClearMediaCredits()` + `AddCredit()` — clear-and-sync pattern
- **Metadata handler** (`handler/metadata.go`): `PUT /api/media/{id}/identify`, `POST /api/media/{id}/refresh`
- **NFO parser** (`pkg/nfo/parser.go`): Read-only, 3 struct types (Movie, TVShow, Episode), file discovery functions
- **Image proxy** (`handler/image.go`): `GET /api/images/tmdb/{size}/{path}` — proxy TMDb images, 7-day cache
- **MetadataMatcher** trong scanner pipeline: gọi sau persist(), enriches từ TMDb/OMDb/Fanart.tv/TVDB/TVmaze
- **Frontend detail page** (`pages/MediaDetailPage.tsx`): read-only display, có nút Refresh Metadata

### Chưa có
- PATCH/PUT endpoint để edit individual fields
- Image upload endpoint + local storage
- Frontend editor UI
- NFO write capability
- Metadata lock flag (prevent rescan override)
- Series metadata edit endpoint

## Product Goal
1. Admin vào detail page → click "Edit Metadata" → sửa title, overview, genres, poster... → Save
2. Save ghi vào DB + optional ghi NFO file ra disk
3. Khi rescan, media đã edit thủ công KHÔNG bị TMDb ghi đè (metadata_locked flag)
4. Admin có thể upload poster/backdrop riêng thay vì chỉ dùng ảnh TMDb
5. Hỗ trợ cả Movie và Series metadata

## Success Criteria
- Admin edit title/overview/genres/cast → lưu thành công, detail page hiển thị giá trị mới
- Upload poster 1MB → hiển thị đúng trên detail page + browse card
- Rescan library → media có metadata_locked=true KHÔNG bị override
- Edit metadata → NFO file được tạo/cập nhật đúng cạnh video file
- Frontend editor chỉ hiện cho admin, không hiện cho user thường

## Proposed Data Model Changes

### Migration 021: metadata_lock
```sql
-- Add metadata_locked flag to media and series
ALTER TABLE media ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
ALTER TABLE series ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;

-- Add tagline to media (Emby has it, we parse from NFO but don't store)
ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT '';
```

### Local image storage
- Upload path: `{VELOX_DATA_DIR}/images/{media|series}/{id}/poster.jpg`
- Serve via: `GET /api/images/local/{type}/{id}/{filename}`
- DB stores: `poster_path = "local://{id}/poster.jpg"` (prefix `local://` phân biệt với TMDb path)

## Proposed API Contract

### PATCH /api/media/{id}/metadata (Admin only)
```json
{
  "title": "The Matrix",
  "sort_title": "Matrix, The",
  "overview": "A computer hacker learns...",
  "tagline": "Welcome to the Real World",
  "release_date": "1999-03-31",
  "rating": 8.7,
  "genres": ["Action", "Sci-Fi"],
  "credits": [
    {"person_name": "Keanu Reeves", "character": "Neo", "role": "cast", "order": 0},
    {"person_name": "Lana Wachowski", "role": "director", "order": 0}
  ],
  "save_nfo": true,
  "metadata_locked": true
}
```
- Partial update: chỉ gửi field cần sửa, field không gửi giữ nguyên
- `save_nfo: true` → ghi NFO file ra disk cạnh video file
- `metadata_locked: true` → rescan sẽ skip metadata refresh cho item này

### PATCH /api/series/{id}/metadata (Admin only)
```json
{
  "title": "Breaking Bad",
  "overview": "...",
  "status": "Ended",
  "network": "AMC",
  "first_air_date": "2008-01-20",
  "genres": ["Drama", "Crime"],
  "save_nfo": true,
  "metadata_locked": true
}
```

### POST /api/media/{id}/images (Admin only, multipart/form-data)
```
image_type: "poster" | "backdrop"
file: <binary>
```
Response: `{"data": {"path": "local://42/poster.jpg"}}`

### POST /api/series/{id}/images (Admin only, multipart/form-data)
Same as above.

### DELETE /api/media/{id}/metadata/lock (Admin only)
Unlock metadata → rescan sẽ override lại từ TMDb.

## Detection Rules

### Image upload
- Accept: JPEG, PNG, WebP
- Max size: 10MB
- Resize: poster → max 1000x1500, backdrop → max 1920x1080
- Format: save as JPEG (quality 90)
- Storage: `{VELOX_DATA_DIR}/images/media/{id}/poster.jpg`

### NFO Write
- Movie: ghi `movie.nfo` cạnh video file, hoặc `{basename}.nfo`
- Series: ghi `tvshow.nfo` trong folder series
- Episode: ghi `{basename}.nfo` cạnh video file
- Format: standard Kodi/Emby NFO XML
- Encoding: UTF-8 with XML declaration

### Metadata Lock behavior
- `metadata_locked = true`:
  - Rescan: skip TMDb/OMDb/Fanart.tv enrichment
  - Manual Refresh: vẫn cho phép (admin chủ động)
  - Identify (re-match TMDb): auto unlock + re-fetch
- `metadata_locked = false`:
  - Rescan: override metadata bình thường

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Backend — PATCH API + Lock | 6 tasks | ⬜ Pending |
| 02 | Image Upload | 5 tasks | ⬜ Pending |
| 03 | Frontend — Editor UI | 7 tasks | ⬜ Pending |
| 04 | NFO Write + Sync | 5 tasks | ⬜ Pending |

## Delivery Strategy

### Ship 1: Edit Metadata (Phase 01 + 03)
- PATCH API cho media + series
- metadata_locked flag + migration
- Frontend editor form (admin only)
- Genres/credits edit
→ Admin có thể sửa metadata từ UI

### Ship 2: Image Upload (Phase 02)
- Upload endpoint + local storage
- Image serve endpoint
- Frontend drag-and-drop upload
→ Admin có thể upload poster/backdrop riêng

### Ship 3: NFO Write (Phase 04)
- NFO writer (XML generation)
- Auto-save NFO option
- Rescan respect lock
→ Full Emby parity

## Files Expected to Change

### Backend
- `backend/internal/database/migrate/registry.go` — migration 021
- `backend/internal/model/media.go` — add MetadataLocked, Tagline fields
- `backend/internal/model/series.go` — add MetadataLocked field
- `backend/internal/repository/media.go` — UpdateMetadata(), UpdateImages()
- `backend/internal/repository/series.go` — UpdateMetadata()
- `backend/internal/service/metadata.go` — EditMediaMetadata(), EditSeriesMetadata()
- `backend/internal/handler/metadata.go` — PATCH endpoints, image upload
- `backend/internal/scanner/pipeline.go` — check metadata_locked in persist()
- `backend/pkg/nfo/writer.go` — NEW: NFO XML writer
- `backend/cmd/server/main.go` — wire new routes

### Frontend
- `webapp/src/types/api.ts` — MetadataEditRequest, ImageUploadResponse
- `webapp/src/api/media.ts` — patchMetadata(), uploadImage()
- `webapp/src/pages/MediaDetailPage.tsx` — Edit button + editor integration
- `webapp/src/pages/SeriesDetailPage.tsx` — Edit button + editor integration
- `webapp/src/components/metadata/MetadataEditor.tsx` — NEW: edit form
- `webapp/src/components/metadata/ImageUploader.tsx` — NEW: drag-and-drop upload
- `webapp/src/components/metadata/GenreEditor.tsx` — NEW: genre tag editor
- `webapp/src/components/metadata/CreditEditor.tsx` — NEW: cast/crew editor

## Risks
- Image upload cần resize library (Go: `imaging` hoặc `bimg`) — thêm dependency
- NFO write phải handle đúng encoding + edge cases (Unicode, special chars in XML)
- Partial PATCH update cần careful nil/zero-value handling trong Go
- metadata_locked phải được check ở đúng chỗ trong scan pipeline để không bị bypass

## Decisions Locked
- **Partial PATCH, not full PUT** — chỉ update field được gửi, không yêu cầu gửi toàn bộ struct
- **`local://` prefix** cho image path — phân biệt local upload vs TMDb proxy path
- **metadata_locked per item** — không phải global setting, mỗi media/series có flag riêng
- **NFO write là optional** — `save_nfo: true` trong request, không auto-save mặc định
- **Identify auto-unlocks** — re-match TMDb ID sẽ tự unlock metadata

## Notes
- Feature này hoàn thành mục "Media info editor" trong `plans/feature-gap-emby.md`
- NFO write cũng cover phần "NFO Export" trong gap list
- Nên ship Phase 01+03 trước để admin có editor UI ngay, Phase 02+04 iterate sau

## Quick Commands
- Start: `/code phase-01`
- Check: `/next`
- Save context: `/save-brain`
