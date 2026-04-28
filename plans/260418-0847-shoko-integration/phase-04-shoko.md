# Phase 04: Shoko Provider
Status: ⬜ Pending
Dependencies: phase-03-database-settings.md

## Objective
Nguyên lý làm ra "Option A": Tạo `ShokoProvider`. Provider này sẽ đọc config endpoint. Nếu thấy enabled, nó sẽ gọi Shoko Server API v3 để truy tìm định danh thư mục/anime chuẩn xác. 

## Requirements
### Functional
- [ ] Dùng `ShokoProvider` như là ưu tiên số 1 (sau NFO) nếu được bật.
- [ ] Endpoint `/api/v3/Series/Search` hoặc `/api/v3/File/Path` tuỳ thuộc vào việc matcher đưa gì (file path thực hay chỉ parsed name).
- [ ] Biến đổi Output API của Shoko thành id map (Ví dụ: Shoko nói "Đây là Naruto mùa 2, id tmdb là xyz"). Chúng ta lấy "tvdb" hoặc "anilist id" để enrich thêm art.

## Implementation Steps
1. [ ] Bổ sung package `shoko` để liên lạc vs Http API của Shoko Server.
2. [ ] Viết `provider_shoko.go` tích hợp vào matcher.
3. [ ] Quản lý Fail/Circuit-Breaker: nếu gọi lên Shoko mất > 5s và timeout -> Bỏ cuộc và chạy AniList (nhằm tránh đơ tiến trình quyét thư viện).

## Files to Create/Modify
- `backend/pkg/shoko/client.go` - Web API Callers.
- `backend/pkg/shoko/models.go` - Struct model JSON.
- `backend/internal/metadata/provider_shoko.go` - Implementation khớp nối.
- `backend/internal/metadata/matcher.go` - Chèn ShokoProvider vào top đầu của chain list.

---
Next Phase: `phase-05-testing.md`
