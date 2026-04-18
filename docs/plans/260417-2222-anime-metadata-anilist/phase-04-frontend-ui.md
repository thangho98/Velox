# Phase 04: Settings & FrontEnd
Status: ⬜ Pending
Dependencies: phase-03-matcher-strategy.md

## Objective
Hoàn thiện UI cấu hình AniList kết nối với luồng Cài đặt chuẩn từ DB lên API và Frontend Hooks.

## Requirements
### Functional
- [ ] Bổ sung Keys cho AniList trong `backend/internal/model/app_settings.go`. (Mặc định Anilist k cần Key, nhưng nếu tích hợp OAuth Token private thì cần).
- [ ] Code Service, Handlers ở Backend để Fetch & Cập nhật AniList Settings.
- [ ] Ở Layer Frontend, bổ sung store properties trong `useMetadataSettings.ts` để đọc/ghi Data.
- [ ] Xây dựng UI trong `MetadataSection.tsx` cho phần Settings Providers.
- [ ] Trong `CreateLibrary.tsx`, bổ sung enum Type `anime` vào Dropdown lựa chọn Content Type.

## Files to Create/Modify
- `backend/internal/model/app_settings.go`
- `backend/internal/service/settings.go`
- `backend/cmd/server/server_app.go`
- `packages/shared/hooks/settings/useMetadataSettings.ts`
- `webapp/src/pages/settings/components/MetadataSection.tsx`
- `webapp/src/pages/admin/libraries/CreateLibrary.tsx`

---
Next Phase: `phase-05-testing.md`
