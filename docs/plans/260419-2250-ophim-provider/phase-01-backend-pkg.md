# Phase 01: OPhim Package API Client
Status: ✅ Complete
Dependencies: None

## Objective
Xây dựng một service HTTP Client thuần tuý trong package `pkg/ophim` để giao tiếp với public API của ophim1.com. Trách nhiệm duy nhất của module này là thực hiện GET request và parse kết quả JSON nguyên thuỷ về Go Struct.

## Requirements
### Functional
- [ ] Lấy danh sách phim mới/hot (`GET /v1/api/home`)
- [ ] Lấy danh sách phim theo phân trang (`GET /v1/api/danh-sach/phim-moi-cap-nhat?page={X}`)
- [ ] Lấy chi tiết thông tin và Stream M3U8 theo Slug (`GET /phim/{slug}`)

### Non-Functional
- [ ] Quản lý HTTP timeouts hợp lý (ví dụ: 10 giây).
- [ ] Không cần sử dụng bất kỳ access token hay API key nào.

## Implementation Steps
1. [ ] Tạo package mới liên quan đến ophim `backend/pkg/ophim/client.go`.
2. [ ] Khai báo các Go Structs tương ứng với cấu trúc JSON (đặc biệt chú ý trường `movie`, `episodes`, `server_data`, và `link_m3u8`).
3. [ ] Viết các hàm giao tiếp (`GetHome`, `GetDetails`, v.v.).

## Files to Create/Modify
- `backend/pkg/ophim/client.go` - HTTP client logic cơ bản.
- `backend/pkg/ophim/models.go` - JSON mapping structs từ `ophim1.com`.

## Test Criteria
- [ ] Chạy thành công hàm fetch thử một slug nào đó ví dụ: `cuoc-chien-trong-chung-ta` và log ra được url chứa `.m3u8`.

---
Next Phase: phase-02-metadata-sync.md
