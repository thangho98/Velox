# Phase 01: Backend — PATCH API + Metadata Lock
Status: ⬜ Pending
Dependencies: None (first phase)

## Objective
Tạo PATCH API cho phép admin chỉnh sửa metadata của media và series.
Thêm `metadata_locked` flag để bảo vệ edit thủ công khỏi bị rescan ghi đè.

## Requirements

### Functional
- [ ] Migration 021: thêm `metadata_locked` + `tagline` columns
- [ ] PATCH `/api/media/{id}/metadata` — partial update media fields
- [ ] PATCH `/api/series/{id}/metadata` — partial update series fields
- [ ] DELETE `/api/media/{id}/metadata/lock` — unlock metadata
- [ ] DELETE `/api/series/{id}/metadata/lock` — unlock metadata
- [ ] Genre sync: nhận array genre names → clear + re-link
- [ ] Credits sync: nhận array credits → clear + re-create people + link
- [ ] Auto-set `metadata_locked = true` khi edit (trừ khi explicit `false`)
- [ ] Scanner pipeline check `metadata_locked` trước khi call MetadataMatcher
- [ ] Identify (re-match TMDb) auto-unlock metadata

### Non-Functional
- [ ] Admin-only access (check `user.IsAdmin` trong middleware/handler)
- [ ] Partial update: chỉ update field có trong request body, field missing giữ nguyên
- [ ] Transaction: genre + credits sync trong cùng tx với metadata update
- [ ] Validation: title không rỗng, release_date format YYYY-MM-DD (nếu có)

## Implementation Steps

### 1. Migration 021
- [ ] Thêm migration `{Version: 21, Name: "metadata_lock"}` vào `registry.go`
- [ ] SQL Up:
  ```sql
  ALTER TABLE media ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
  ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT '';
  ALTER TABLE series ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
  ```
- [ ] SQL Down:
  ```sql
  -- SQLite không hỗ trợ DROP COLUMN trước 3.35.0
  -- Sử dụng recreate pattern nếu cần
  ```

### 2. Model Updates
- [ ] `model/media.go`: thêm `MetadataLocked bool` + `Tagline string` với json tags
- [ ] `model/series.go`: thêm `MetadataLocked bool`
- [ ] Tạo DTO struct `MetadataEditRequest` cho PATCH request body:
  ```go
  type MetadataEditRequest struct {
      Title       *string  `json:"title"`
      SortTitle   *string  `json:"sort_title"`
      Overview    *string  `json:"overview"`
      Tagline     *string  `json:"tagline"`
      ReleaseDate *string  `json:"release_date"`
      Rating      *float64 `json:"rating"`
      Genres      []string `json:"genres"`
      Credits     []CreditInput `json:"credits"`
      SaveNFO     bool     `json:"save_nfo"`
      MetadataLocked *bool `json:"metadata_locked"`
  }
  ```
  - Pointer fields (`*string`, `*float64`) cho partial update: nil = không thay đổi
- [ ] Tạo `SeriesMetadataEditRequest` tương tự (thêm Status, Network, FirstAirDate)

### 3. Repository Updates
- [ ] `repository/media.go`: thêm `UpdateMetadata(ctx, id, fields)` method
  - Build dynamic UPDATE query chỉ SET fields có giá trị
  - Luôn SET `updated_at = CURRENT_TIMESTAMP`
  - Return updated media
- [ ] `repository/media.go`: update `Update()` method include `metadata_locked`, `tagline`
- [ ] `repository/series.go`: thêm `UpdateMetadata(ctx, id, fields)` tương tự
- [ ] `repository/media.go`: thêm `GetMetadataLocked(ctx, id)` — quick check flag

### 4. Service Layer
- [ ] `service/metadata.go`: thêm `EditMediaMetadata(ctx, mediaID, req MetadataEditRequest) error`
  - Validate input
  - Begin transaction
  - Update media fields (repo.UpdateMetadata)
  - If genres provided: clear + sync genres (trong tx)
  - If credits provided: clear + sync credits (trong tx)
  - Commit transaction
  - If `save_nfo`: ghi NFO (Phase 04, skip for now)
- [ ] `service/metadata.go`: thêm `EditSeriesMetadata(ctx, seriesID, req)` tương tự
- [ ] `service/metadata.go`: thêm `UnlockMetadata(ctx, mediaType, id)` — set locked=false
- [ ] `service/metadata.go`: sửa `MatchAndPersistMovie()` — check metadata_locked trước khi update
- [ ] `service/metadata.go`: sửa `MatchAndPersistEpisode()` — check metadata_locked
- [ ] `service/metadata.go`: sửa `IdentifyByTmdbID()` — auto set metadata_locked=false

### 5. Handler Layer
- [ ] `handler/metadata.go`: thêm `EditMediaMetadata(w, r)` — PATCH handler
  - Parse `{id}` from URL
  - Decode JSON body → MetadataEditRequest
  - Check user is admin
  - Call service.EditMediaMetadata()
  - respondJSON with updated media
- [ ] `handler/metadata.go`: thêm `EditSeriesMetadata(w, r)` — PATCH handler
- [ ] `handler/metadata.go`: thêm `UnlockMediaMetadata(w, r)` — DELETE handler
- [ ] `handler/metadata.go`: thêm `UnlockSeriesMetadata(w, r)` — DELETE handler

### 6. Route Wiring
- [ ] `cmd/server/main.go`: thêm routes
  ```go
  mux.Handle("PATCH /api/media/{id}/metadata", adminOnly(metadataHandler.EditMediaMetadata))
  mux.Handle("PATCH /api/series/{id}/metadata", adminOnly(metadataHandler.EditSeriesMetadata))
  mux.Handle("DELETE /api/media/{id}/metadata/lock", adminOnly(metadataHandler.UnlockMediaMetadata))
  mux.Handle("DELETE /api/series/{id}/metadata/lock", adminOnly(metadataHandler.UnlockSeriesMetadata))
  ```

## Files to Create/Modify
- `backend/internal/database/migrate/registry.go` — migration 021
- `backend/internal/model/media.go` — MetadataLocked, Tagline, MetadataEditRequest
- `backend/internal/model/series.go` — MetadataLocked, SeriesMetadataEditRequest
- `backend/internal/repository/media.go` — UpdateMetadata()
- `backend/internal/repository/series.go` — UpdateMetadata()
- `backend/internal/service/metadata.go` — EditMediaMetadata(), EditSeriesMetadata(), UnlockMetadata()
- `backend/internal/handler/metadata.go` — PATCH/DELETE handlers
- `backend/cmd/server/main.go` — new routes

## Test Criteria
- [ ] PATCH media metadata → fields updated in DB, other fields unchanged
- [ ] PATCH with genres → old genres cleared, new genres linked
- [ ] PATCH with credits → old credits cleared, new people created + linked
- [ ] PATCH auto-sets metadata_locked=true
- [ ] Rescan skips metadata refresh for locked media
- [ ] Identify (re-match TMDb) auto-unlocks
- [ ] DELETE lock → metadata_locked=false
- [ ] Non-admin user → 403 Forbidden
- [ ] Invalid input (empty title, bad date) → 400 Bad Request

## Notes
- Dùng pointer fields (`*string`) cho partial update pattern — nil means "don't change"
- Genre sync: tạo genre nếu chưa có (FindOrCreate pattern)
- Credits sync: tạo person nếu chưa có (FindOrCreate pattern)
- NFO write sẽ được implement ở Phase 04, Phase 01 chỉ cần placeholder `if req.SaveNFO { /* TODO Phase 04 */ }`

---
Next Phase: [Phase 02 — Image Upload](phase-02-image-upload.md)
