# Phase 04: Android Mobile & TV UI
Status: ⬜ Pending
Dependencies: Phase 02

## Objective
Hoàn thành tích hợp luồng Live TV trên ứng dụng Android bằng Jetpack Compose (hỗ trợ cả Mobile và Remote của Android TV).

## Requirements
### Functional
- [ ] Bổ sung mục "Live TV" vào Navigation Drawer của TV và Bottom Bar của Mobile.
- [ ] Flow: `LiveTvScreen` chia bố cục D-Pad thân thiện (Horizontal Rows cho Group, Cards cho kênh).
- [ ] Khi chọn kênh: truyền `streamUrl` thẳng vào màn hình `VideoPlayerScreen` (sử dụng ExoPlayer gốc vì thư viện này hỗ trợ Live HLS rất mạnh).

## Implementation Steps
1. [ ] Update Repository `LiveTvRepository.kt` & Retrofit Interface + Dịch vụ DI (Hilt).
2. [ ] Tạo `TvLiveTvScreen.kt` theo chuẩn thiết kế hiện tại của nhánh TV.
3. [ ] Tích hợp Paging3 nếu danh sách kênh quốc tế quá khổng lồ (e.g. 10.000 kênh) để tối ưu luồng LazyRow/LazyColumn.
4. [ ] Chạy thủ tục Test E2E trên Emulator TV.
