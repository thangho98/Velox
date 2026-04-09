# Lessons Learned: Live Transcode & HLS Subtitle Synchronization

*Ngày viết: 09/04/2026*

Tài liệu này ghi lại các kinh nghiệm xương máu trong việc xử lý đồng bộ mốc thời gian (timeline sync) giữa Backend (FFmpeg HLS) và Frontend (hls.js), đặc biệt là tình trạng Subtitle chạy không khớp hình ảnh khi Resume.

---

## 1. Vấn đề "Phụ đề chạy nhanh hơn phim" (Subtitle Desync)

### Triệu chứng
Khi user bấm Resume một tập phim ở mốc thời gian lẻ (VD: `150.5s`), HLS Player tải phim lên thành công nhưng phụ đề luôn xuất hiện sớm hơn miệng nhân vật khoảng tầm `0.5s` đến `1.0s`.

### Nguyên nhân cốt lõi (Sự lầm tưởng của Frontend)
Theo thói quen, Frontend (FE) cho rằng: "Tôi gửi yêu cầu xem từ giây `150.5`, thì Backend (BE) sẽ cắt đúng frame tại `150.5s` và trả về cho tôi". Từ đó FE gán mốc `video.currentTime = 0` tương ứng trực tiếp với `150.5s` để lấy phụ đề.

Trực tế **không hoạt động như vậy**:
- Để tối ưu tài nguyên và ít bị lỗi khi decode bằng Card màn hình (HW Accel), BE dùng cơ chế **hls_time chunking**.
- BE tự động **làm tròn** thời gian yêu cầu xuống số nguyên lẻ của một bội số đoạn (Segment Length, mặc định là 6.0s).
- Công thức nội bộ của BE là: `⌊150.5 / 6.0⌋ * 6.0 = 150.0s`.
- Do đó gốc nội dung (frame đầu tiên có PTS=0) trả về thực chất nằm ở mốc `150.0s`. Chênh lệch `0.5s` sinh ra từ đây.

### Giải pháp "Separation of Concerns" (Phân tán trách nhiệm)
**1. Tại Backend:**
- Vẫn giữ nguyên cơ cấu làm tròn chunk `6.0s`. Dứt khoát không để BE dùng phép tính phức tạp mổ xẻ frame hình tránh quá tải FFmpeg.

**2. Tại Frontend (WatchPage.tsx):**
- FE cần phải bắt chiếc lại thuật toán làm tròn của BE mỗi khi load segment đầu tiên:
  ```javascript
  const segLength = 6.0;
  const trueStartOffset = Math.floor(sessionOffset / segLength) * segLength;
  ```
- Lấy `trueStartOffset` (150.0s) làm Time Base cho toàn bộ Player thay vì dùng `sessionOffset`. Điều này sửa dứt điểm lỗi lệch sub. Cuối cùng, FE chủ động tua tới đoạn lố (gap) để Resume khớp khung hình:
  ```javascript
  const gap = sessionOffset - trueStartOffset; // 0.5s
  v.currentTime = gap; // Tiến hình lên đúng 150.5s
  ```

---

## 2. Tuyệt đối không dùng Kiểu số Nguyên (Int) cho Progress
- Thời gian tiến trình (progress) phải **luôn là dạng Float (Double precision)** thay vì kiểu số tròn (giây).
- Vì tốc độ khung hình (như 23.976 fps) hay các file Subtitle (SRT/VTT) thường hoạt động ở đơn vị Mili-giây (`xx.625s`).
- Lưu thành Int làm tích tụ dồn sai số qua các lần Encode/Decode khiến hệ thập phân nhảy sai frame hình, làm mất chữ và gây đứt đoạn âm thanh khi Seek/Jump.

---

## 3. Kiến trúc Tự Cứu (Recovery Automation) cho hls.js

Trong luồng Live Transcode, đường truyền mạng đôi khi làm vỡ frame HLS (buffer hỏng). Để giao diện mượt nhất mà không bị xoay vòng tải vô tận:

*   **Lỗi Tràn Buffer (`BUFFER_FULL_ERROR`)**:
    Các trình duyệt web quy định dung lượng RAM tối đa cho thẻ `<video>`. Thay vì throw lỗi ngầm, nay FE chủ động gọi `hls.flushBuffer()` để xả bộ đệm dĩ vãng, giữ cho phim phía trước tiếp tục chạy mà không sập player.
*   **Chết ngầm `[v.readyState < 3]`**:
    Khi hình bị khựng do thiếu 1 frame I-Frame nhưng hls.js vẫn báo "playing", code tự động bơm `v.currentTime += 0.1` một vài minisecond để nhảy qua khung hình xấu và ép luồng video play tiếp.
*   **Reset Tổng Thể (Hard Fallback)**:
    Nếu tải 5 segment liên tục đều cho ra data rỗng (`0 byte`), Player lập tức hủy object HLS và tải vòng lặp bắt ép Full Transcode.
