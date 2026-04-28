# Phase 02: AniList Provider
Status: ⬜ Pending
Dependencies: phase-01-foundation.md

## Objective
Triển khai nhà cung cấp dữ liệu (provider) riêng cho AniList. Đây là cầu nối trung gian quan trọng để Velox ưu tiên ảnh HD và metadata ngon cho tất cả Anime, kể cả nếu người dùng không dùng Shoko.

## Requirements
### Functional
- [ ] Build được GraphQL API Client căn bản để gọi lên cổng `https://graphql.anilist.co`.
- [ ] Tìm kiếm TV Show bằng GraphQL thông qua title của file tên gốc. Đánh giá tính chính xác.
- [ ] Chuyển đổi (map) fields từ AniList sang `TVMatchResult` chuẩn của Velox.

## Implementation Steps
1. [ ] Cấu thành package thư viện con/client gọi GraphQL tới AniList.
2. [ ] Ánh xạ JSON Response của graphql ra Model struct Golang.
3. [ ] Viết `AniListProvider` kết nối Client vừa xây dựng.
4. [ ] Thiết kế logic nếu query AniList trống không, trả về `Found: false` để nhường cơ hội lại cho TMDb.

## Files to Create/Modify
- `backend/pkg/anilist/client.go` - Tạo client query GraphQL đơn giản.
- `backend/pkg/anilist/models.go` - Các query variables.
- `backend/internal/metadata/provider_anilist.go` - The Provider implementation.

---
Next Phase: `phase-03-database-settings.md`
