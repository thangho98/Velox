# Phase 01: Foundation (Cấu trúc lại Matcher)
Status: ⬜ Pending
Dependencies: None

## Objective
Biến file `backend/internal/metadata/matcher.go` đang chứa luồng match hỗn hợp thành một hệ thống linh hoạt kiểu interface (`IMetadataProvider`). Điều này giúp dễ dàng cắm rút thêm providers nào tuỳ ý.

## Requirements
### Functional
- [ ] Định nghĩa `IMetadataProvider` interface với hàm `MatchMovie` và `MatchTVShow`.
- [ ] Tách đoạn parse Local NFO thành `NFOProvider` riêng biệt.
- [ ] Tách luồng search/match gốc thành `TMDbProvider` chuẩn chỉnh.
- [ ] Cấu hình một `ChainMatcher` để nối các provider theo đúng thứ tự.

### Non-Functional
- [ ] Không làm thay đổi kết quả matching với các file test hiện tại.

## Implementation Steps
1. [ ] Tạo struct/interface `Provider` chuẩn trong `metadata/provider.go`.
2. [ ] Tạo `NFOProvider`.
3. [ ] Tạo `TMDbProvider`.
4. [ ] Viết lại hàm `MatchMovie/MatchTVShow` của `Matcher` để iterate qua danh sách các provider đã đăng ký, gặp provider nào trả về kết quả ưng ý trước sẽ chốt.

## Files to Create/Modify
- `backend/internal/metadata/matcher.go` - Phân tách code, thiết lập `ChainMatcher`.
- `backend/internal/metadata/provider.go` - Tạo file Interface định dạng trả về (MatchResult).
- `backend/internal/metadata/provider_nfo.go` - Xử lý NFO.
- `backend/internal/metadata/provider_tmdb.go` - Xử lý fallback cho TMDb.

---
Next Phase: `phase-02-anilist.md`
