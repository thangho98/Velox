# Phase 02: AniList GraphQL Client
Status: ⬜ Pending
Dependencies: phase-01-database.md

## Objective
Tạo bộ kết nối độc lập tới AniList qua GraphQL API với format Data Struct đầy đủ để thay thế flow episodic của TheTVDB/TMDB.

## Requirements
### Functional
- [ ] Khởi tạo package `backend/pkg/anilist/client.go`. Wrap HTTP request hỗ trợ limit rate.
- [ ] Viết GraphQL queries lấy CHUẨN bộ data mapping cần thiết: `id`, `title (romaji, english, native)`, `coverImage (extraLarge)`, `bannerImage`, `description`, `studios`, `status`, `episodes` (Tổng số tập).
- [ ] Cần viết lấy data từng episodes lẻ nếu có (Anilist không lưu meta từng tập nhiều như TMDb, có thể phải parse qua `nextAiringEpisode` hoặc `streamingEpisodes`). Đảm bảo trả về models cung cấp đủ data cho `backend/internal/service/metadata.go`.

### Non-Functional
- [ ] Xử lý Rate Limit 90 request/min của AniList bằng channels cẩn thận giống package tmdb.

## Files to Create/Modify
- `backend/pkg/anilist/client.go`
- `backend/pkg/anilist/types.go`

---
Next Phase: `phase-03-matcher-strategy.md`
