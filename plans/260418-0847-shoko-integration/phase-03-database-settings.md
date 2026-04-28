# Phase 03: Database & Settings UI
Status: ⬜ Pending
Dependencies: phase-02-anilist.md

## Objective
Chuẩn bị hệ thống lưu trữ nhận dạng AniDB ID vào Media Database và mở Cài đặt để quản trị viên bật tắt Shoko Integration.

## Requirements
### Functional
- [ ] Có thể lưu trữ `anidb_id` và `anilist_id` khi update metadata của MediaItem trong Database Schema.
- [ ] Lưu thêm các config về `Integrations` trong hệ thống Settings.
- [ ] Frontend có giao diện Bật/Tắt tích hợp Anime "Shoko Server", có trường nhập `Endpoint URL` và `Bật tắt`.

## Implementation Steps
1. [ ] Sửa lại GORM / DB Schema model cho bảng Media (MediaItem / Series), thêm field `ani_db_id`, `anilist_id`.
2. [ ] Sửa Model Config `Settings` trong Go, bổ sung block `Integrations`.
3. [ ] Cập nhật Web UI ở `/settings/metadata` -> Thêm mục "Integrations / External Sources".

## Files to Create/Modify
- `backend/internal/db/models/media.go` - Thêm columns.
- `backend/internal/config/settings.go` - Thêm config model.
- `webapp/src/pages/Settings/MetadataPage.tsx` - Cập nhật UI.
- `webapp/src/types/settings.ts` - Sync ts definitions.

---
Next Phase: `phase-04-shoko.md`
