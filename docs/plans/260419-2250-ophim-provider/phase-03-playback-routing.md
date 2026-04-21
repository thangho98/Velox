# Phase 03: Playback Routing
Status: Complete 🟢
Dependencies: phase-02-metadata-sync.md

## Objective
Cho phép Backend của Velox trả về link M3U8 từ Ophim thay vì kích hoạt luồng Transcoder cục bộ khi người dùng mở bộ phim lấy từ Ophim.

## Requirements
### Functional
- [ ] Khi Frontend / App gọi API lấy Video Source (`GET /api/playback/info` hoặc tương tự tuỳ luồng Velox), nhận diện nguồn phim là `ophim`.
- [ ] Kích hoạt `GetDetails(slug)` để xin lại link M3U8 gốc nếu chưa được cache hoặc nó đã bị hết hạn.
- [ ] Trả đối tượng `Direct Stream` (HLS url direct provider) mà không ép buộc FFmpeg bật luồng.

## Implementation Steps
1. [ ] Cập nhật module `playback` tại `backend/internal/playback`.
2. [ ] Cài cắm cờ `provider` phân loại source (vd: `LOCAL`, `FSHARE`, `OPHIM`).
3. [ ] Route thẳng HTTP response trả về URL M3U8 dạng Direct-Play cho client nếu nguồn là OPHIM.
4. [ ] Kiểm tra xem Frontend (Web / TV App) có cần xử lý cross-domain luồng M3U8 (CORS fallback) nếu CDN bị chặn.

## Files to Create/Modify
- `backend/internal/playback/playback_service.go` - Cập nhật luồng trả Video Endpoint.
- `backend/internal/handler/playback_handler.go` - (Cập nhật)

## Test Criteria
- [ ] Play thử phim OPhim trên Android TV app vừa code, xác nhận luồng video là `vip.opstream90.com` và phát với ExoPlayer thành công, không thấy CPU của NAS chạy quá tải (vì FFmpeg không cần kích hoạt).

---
Next Phase: phase-04-testing-integration.md
