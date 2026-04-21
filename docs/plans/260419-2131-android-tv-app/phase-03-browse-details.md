# Phase 03: Browse, Search & Details
Status: ⬜ Pending

## Objective
Cho phép người dùng duyệt toàn bộ thư viện, tìm kiếm phim và xem chi tiết một thiết kế tối ưu cho TV.

## Requirements
### Functional
- [x] Màn hình Chi tiết (Details): Layout chia làm 2 - Trái (Poster & Nút Play, Audio, Subtitle), Phải (Nội dung tóm tắt, Điểm số, Cast).
- [x] Navigation Drawer dọc bên trái: Rút gọn để di chuyển qua lại giữa Home, Browse, Search.
- [x] Màn hình Search: Bàn phím ảo cho TV để tìm tên phim.

### Non-Functional
- [x] KHÔNG làm phần Settings Server / Admin (đã bị loại bỏ theo yêu cầu).
- [x] Giao diện trực quan, không cắt chữ, chữ phải to/rõ nhìn được từ khoảng cách 3 mét.

## Implementation Steps
1. [x] Tạo `TvNavigationDrawer.kt` cho sidebar TV.
2. [x] Tạo `TvMediaDetailScreen.kt` cho phim lẻ và phim bộ.
3. [x] Xâu chuỗi từ Home -> Details.

## Files to Create/Modify
- `android/app/src/main/java/com/velox/app/presentation/tv/screens/TvMediaDetailScreen.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/screens/TvSearchScreen.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/components/TvNavigationDrawer.kt`

## Test Criteria
- [ ] Chuyển qua lại từ Trang chủ -> Sidebar -> Các trang phim trơn tru.
- [ ] Bấm vào một phim hiện đúng thông tin tóm tắt và Nút Play to bự cho phép click.

---
Next Phase: Phase 04 (Player OSD)
