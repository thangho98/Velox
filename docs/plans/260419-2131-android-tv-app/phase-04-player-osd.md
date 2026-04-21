# Phase 04: Player OSD & Actions
Status: ⬜ Pending

## Objective
Đóng gói trải nghiệm xem phim bằng cách tuỳ biến giao diện Điều khiển Video (On-Screen Display - OSD) của ExoPlayer chuyên biệt cho Remote Controller.

## Requirements
### Functional
- [x] Sử dụng chung instance `PlaybackManager` để không phải code lại engine xử lí HLS/Direct.
- [x] Giao diện tự thiết kế chồng lên (Overlay) để tua nhanh tiến/lùi (vd: nhấn mũi tên phải tua tới 15s).
- [x] Pop-up menu nhỏ để chọn Audio Track & Subtitle trên màn hình đang xem.
- [x] Tự động ẩn thanh OSD sau 3 giây không tương tác.

### Non-Functional
- [x] Đảm bảo cơ chế lưu tiến độ (Progress update) vẫn chạy ngầm như cũ.
- [x] Tối ưu UI không giật lag.

## Implementation Steps
1. [x] Viết bộ controller `TvPlayerOSD.kt` nhận event key codes từ D-Pad.
2. [x] Tích hợp ExoPlayer surface view vào `TvPlayerScreen.kt`.

## Files to Create/Modify
- `android/app/src/main/java/com/velox/app/presentation/tv/screens/TvPlayerScreen.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/components/TvPlayerOSD.kt`

## Test Criteria
- [ ] Mở phim chạy êm, có support HDR.
- [ ] Bấm D-Pad sang phải tua tới, bấm Lên/Xuống mở menu OSD.

---
All Phases Complete!
