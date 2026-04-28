# Phase 01: Database & M3U Parser
Status: ✅ Complete
Dependencies: None

## Objective
Xây dựng Schema cơ sở dữ liệu để lưu trữ cả nguồn danh sách (Playlists) và từng Kênh cụ thể (Channels). Phát triển module parse file M3U mạnh mẽ, hỗ trợ đọc mở rộng nhiều biến (tvg-logo, group-title...).

## Requirements
### Functional
- [x] Bảng `live_playlists`: Lưu thông tin URL (quốc tế) hoặc File path (nội địa), chu kỳ update (24h...).
- [x] Bảng `live_channels`: Lưu ID, Tên, Logo, Thể loại (Group), Quốc gia (nếu có rút ra từ iptv-org), URL Stream, FK `playlist_id`.
- [x] Logic Parser: Đọc file M3U, bóc tách `EXTINF`, hỗ trợ ignore các tag rác, format lại đúng link m3u8.

### Non-Functional
- [x] Khả năng xử lý M3U cực lớn (hơn 10,000 dòng của iptv-org) nhanh chóng không làm nghẽn RAM do loop lâu.

## Implementation Steps
1. [x] Thêm file schema migration mới (`040_live_tv.go`).
2. [x] Tạo struct models cho `LivePlaylist` và `LiveChannel` trong `internal/model/livetv.go`.
3. [x] Khởi tạo `internal/repository/livetv_repo.go` (CRUD chuẩn SQLite WAL với batch insert tránh limit).
4. [x] Khởi tạo `internal/service/m3uparser/parser.go`.
5. [x] Viết unit tests cho parser.

---
Next Phase: Phase 02 (Backend API & Sync Task)
