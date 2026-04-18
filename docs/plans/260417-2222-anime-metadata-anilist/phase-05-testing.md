# Phase 05: Testing & Scenarios
Status: ⬜ Pending
Dependencies: phase-04-frontend-ui.md

## Objective
Đảm bảo toàn bộ luồng tích hợp mới hoạt động hoàn hảo và không tạo Regression Bug cho hệ thống Movies/TV truyền thống.

## Acceptance Criteria
- [ ] Tạo Test Library dạng `anime` và Scan thư mục thành công. Media + Series bóc tách được `AnilistID`, lấy `RomajiTitle` và Poster đúng từ Anilist.
- [ ] Chạy chức năng `Refresh All Metadata` cho 1 Library Movies (TMDB) -> Không bị crash schema, dữ liệu giữ nguyên.
- [ ] Khả năng Bypass TMDB diễn ra trơn tru. Fallback về TMDB nếu Anilist lỗi 404 (Ví dụ Live Actions Nhật Bản).
- [ ] Settings Page save AniList key (nếu cần thiết) và Restart Application thì Provider load được cấu hình.

## Notes
Hãy luôn giữ một thư mục Movie (TMDB) và TV Shows (TMDB) song song để Testing scan check chéo.
