# Phase 03: Scanner & Routing Logic
Status: ⬜ Pending
Dependencies: phase-02-anilist-client.md

## Objective
Sửa thiết kế thiết kế routing nhận diện ở đúng service layer để tránh gãy cấu trúc `Matcher`.

## Requirements
### Functional
- [ ] Trong module scanner (`backend/internal/scanner/pipeline.go`), check `Library.Type == "anime"`. Nếu đúng, chuyển tiếp sang func routing `metadataService.MatchAnime()` thay vì gọi `metadataSuite`.
- [ ] Trong `backend/internal/service/metadata_editor.go` & `metadata.go`: Code logic xử lý `MatchAnime(ctx, media, series)` gọi qua package `pkg/anilist`. Ghi dữ liệu trực tiếp vào Entity `Series` (cho Series level) và `Media` (cho Episodic level).
- [ ] Mapping ảnh Poster của Anilist thẳng vào ImageResource để Download.
- [ ] Tạo Fallback: Nếu không ra kết quả ở AniList, rớt về TMDB bình thường.

### Non-Functional
- [ ] Tránh nhét biến phụ `LibraryType` vào Media nếu nó ko có sẵn DB. Xử lý routing lúc scan là an toàn nhất.

## Files to Create/Modify
- `backend/internal/scanner/pipeline.go`
- `backend/internal/service/metadata.go`
- `backend/internal/service/metadata_editor.go`

---
Next Phase: `phase-04-frontend-ui.md`
