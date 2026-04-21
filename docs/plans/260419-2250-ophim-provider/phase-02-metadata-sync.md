# Phase 02: Metadata Sync
Status: Complete 🟢
Dependencies: phase-01-backend-pkg.md

## Objective
Xây dựng cơ chế "import" dữ liệu vào Velox DB từ dữ liệu parse được của `pkg/ophim`. Map các struct của API thành `domain.MediaItem` của Velox và lưu vào database một cách mượt mà.

## Requirements
### Functional
- [ ] Xác định định danh phim đến từ Ophim để không bị nhầm lẫn với phim quét local, ví dụ sinh ID tạm thời hoặc dựa trên trường `slug`.
- [ ] Khởi tạo tính năng `OphimScanner` có khả năng fetch list phim phân trang từ API và cấy vào SQLite (Bảng `media_items` và `episodes`).
- [ ] Gỡ hình nền của OPhim CDN (domain `img.ophim.live`) dùng thẳng để không phải tải về NAS.

## Implementation Steps
1. [ ] Định nghĩa Interface `OphimScanner` trong thư mục `backend/internal/scanner`.
2. [ ] Code phần mapping logic: Convert struct Ophim Movie thành struct Velox (`Model.MediaItem`).
3. [ ] Code hàm kiểm tra xem phim đã có sẵn chưa để tránh insert trùng lặp theo `slug`.
4. [ ] Viết API `/api/admin/ophim/sync` trên backend để người dùng (Admin) có thể manually trigger việc lấy 2-3 trang phim mới cho vào thư viện.

## Files to Create/Modify
- `backend/internal/scanner/ophim_scanner.go` - Logic quét và import nội dung OPhim.
- `backend/internal/handler/admin_handler.go` - (Cập nhật) nếu thêm Web API để gọi quét.

## Test Criteria
- [ ] Mở App Velox hoặc Web -> Thấy xuất hiện thêm một bộ phim mới (với đầy đủ poster và mô tả) bắt nguồn từ OPhim trong danh sách thư viện.

---
Next Phase: phase-03-playback-routing.md
