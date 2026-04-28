# Phase 03: Frontend Web UI
Status: 🟩 Complete
Dependencies: Phase 02

## Objective
Giao diện người dùng trên Webapp. Thêm luồng Live TV vào thanh điều hướng, thiết kế Grid layout đẹp mắt mảng truyền hình, kết nối với trình phát video để có thể xem ngay lập tức.

## Requirements
### Functional
- [x] Thêm icon "Live TV" vào Sidebar/Navbar chính.
- [x] Màn hình `LiveTvPage`: Split View (Trái hiển thị danh sách Categories/Groups, Phải hiển thị các Kênh trong Group đó bằng Grid).
- [x] Xử lý tải Logo lỗi (Fallbacks).
- [x] Bấm vào Kênh -> Gọi Modal Player hoặc Push Route qua `/play/livetv/{id}`.
- [x] Trình chơi Video (HLS) kết nối trơn tru, không bị kẹt bộ đệm.

## Implementation Steps
1. [x] Cấu hình Route mới trong `App.tsx` hoặc `routes`.
2. [x] Viết API Hook (React Query) call `GET /api/livetv/channels`.
3. [x] Tạo các Component: `LiveTvPage`, `ChannelCard`, `CategorySidebar`, `LiveTVPlayerModal`.
4. [x] Tinh chỉnh Tailwind 4 cho giao diện theo chuẩn thiết kế Vercel/Linear (Glassmorphism / Glow nhẹ).

---
Next Phase: Phase 04 (Android Mobile & TV UI)
