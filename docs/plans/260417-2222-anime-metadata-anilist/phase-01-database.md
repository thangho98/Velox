# Phase 01: Database & Models
Status: ✅ Complete
Dependencies: N/A

## Objective
Mở rộng Schema ở mức Database và Model để hỗ trợ Anime context cho cả Series lẫn Media.

## Requirements
### Functional
- [x] Bổ sung Type `anime` vào Enums Type của Library (hiện tại đang là `movies|tvshows|mixed`).
- [x] Thêm các trường `AnilistID`, `RomajiTitle`, `Studio` vào cả **model `Series`** lẫn **model `Media`**.
- [x] Viết migration SQL vào `backend/internal/database/migrate/registry.go` để add column an toàn.
- [x] Map các trường mới này bên trong `backend/internal/model/series.go` và `backend/internal/model/media.go`.

### Non-Functional
- [ ] Flow database cho `tvshows` và `movies` hiện tại không bị gián đoạn.

## Files to Create/Modify
- `backend/internal/database/migrate/registry.go`
- `backend/internal/model/model.go` (cập nhật type ENUM nếu có hằng số)
- `backend/internal/model/media.go`
- `backend/internal/model/series.go`

---
Next Phase: `phase-02-anilist-client.md`
