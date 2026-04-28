# Phase 02: Backend API & Sync Task
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Hoàn thiện luồng Back-end. Tự động đồng bộ các Playlist đã cấu hình theo chu kỳ và mở ra các Endpoints chuẩn RESTful để Frontend có thể lấy danh sách kênh mượt mà.

## Requirements
### Functional
- [ ] Task Scheduler: Background worker tải danh sách update từ các internet playlists (VD: github iptv-org) 1 lần/ngày.
- [ ] API Get Playlists: Cấu hình / Thêm / Sửa / Xóa nguồn M3U.
- [ ] API Get Channels: `GET /api/livetv/channels` có phân trang, lọc theo Group (Thể thao, VTV...).
- [ ] API Search Channels: Tìm kiếm Kênh theo tên nhanh.
- [ ] Logic "Đảm bảo ID kênh" (Nếu link m3u8 đổi nhưng tên kênh không đổi, phải ưu tiên update link chứ không tạo mới Kênh rác trong DB để tránh mất dữ liệu "Kênh Yêu Thích" về sau).

## Implementation Steps
1. [ ] Cài đặt Background Sync Ticker trong `internal/scanner/`.
2. [ ] Tạo `internal/handler/livetv_handler.go`.
3. [ ] Đăng ký Route `/api/livetv/...` trong App Router Go 1.22+.
4. [ ] Inject service vào DI, setup Log theo format.

---
Next Phase: Phase 03 (Frontend Web UI)
