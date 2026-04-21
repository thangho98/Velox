# Phase 02: Login & Home Dashboard
Status: ⬜ Pending

## Objective
Xây dựng luồng đăng nhập thân thiện với Remote TV và trang chủ (Dashboard) sử dụng các component trượt ngang của Compose TV.

## Requirements
### Functional
- [x] Màn hình Login: Nhập Server URL và API Key qua mã PIN hoăc Text Input thân thiện với Remote. 
- [x] Màn hình Home: Có Hero Banner ở trên cùng hiển thị bộ phim gợi ý.
- [x] Màn hình Home: Dải `TvLazyRow` hiển thị *Continue Watching*.
- [x] Màn hình Home: Dải `TvLazyRow` hiển thị *Recently Added* (Media và Series).

### Non-Functional
- [x] Quản lý trạng thái focus chuẩn xác: Phải có viền phát sáng (Glow/Scale) khi card phim được Focus.
- [x] Tái sử dụng `MediaRepository` & `SeriesRepository` của mobile.

## Implementation Steps
1. [x] Xây dựng `TvLoginScreen.kt`.
2. [x] Xây dựng `TvHomeScreen.kt`.
3. [x] Xây dựng component `TvMediaCard.kt` có tích hợp Focus effect.

## Files to Create/Modify
- `android/app/src/main/java/com/velox/app/presentation/tv/screens/TvLoginScreen.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/screens/TvHomeScreen.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/components/TvMediaCard.kt`

## Test Criteria
- [ ] Dùng phím lên/xuống/trái/phải để di chuyển các lựa chọn mượt mà.
- [ ] Đăng nhập thành công và fetch được dữ liệu đồ họa (Poster phim).

---
Next Phase: Phase 03 (Browse, Search & Details)
