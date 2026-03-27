# Phase 04: Settings UI + Dashboard
Status: ⬜ Pending
Dependencies: Phase 02, Phase 03

## Objective
Trang admin Settings cho pre-transcode: toggle bật/tắt, chọn quality, xem storage estimation, theo dõi tiến trình encode.

## Implementation Steps

### Settings Section
1. [ ] Thêm section "Pre-transcode" trong Settings page:
   ```
   ┌─────────────────────────────────────────────────────┐
   │ Pre-transcode (Offline Encoding)                     │
   │                                                      │
   │ [Toggle: OFF/ON]                                     │
   │                                                      │
   │ Encode phim sẵn để play tức thì, không cần chờ.      │
   │ Giống Netflix — mượt mà, không buffer.               │
   │                                                      │
   │ ⚠️ Tốn thêm dung lượng ổ cứng.                     │
   └─────────────────────────────────────────────────────┘
   ```

2. [ ] Quality Profile Selector (khi toggle ON):
   ```
   Chọn chất lượng encode:
   □ 480p  (SD)    ~0.6 GB/phim
   ☑ 720p  (HD)    ~1.4 GB/phim
   ☑ 1080p (Full HD) ~2.4 GB/phim
   ```

3. [ ] Schedule Selector:
   ```
   Thời gian encode:
   ○ Luôn luôn (nhanh nhất, NAS chạy liên tục)
   ● Ban đêm (00:00 - 06:00, không ảnh hưởng xem phim)
   ○ Khi rảnh (chỉ khi không ai đang xem)
   ```

4. [ ] Concurrency Selector:
   ```
   Số phim encode cùng lúc: [1 ▼]
   💡 NAS yếu nên để 1. NAS mạnh có thể tăng lên 2-3.
   ```

### Storage Estimation Panel
5. [ ] Hiển thị TRƯỚC khi bật:
   ```
   ┌─────────────────────────────────────────────────┐
   │ 📊 Ước tính dung lượng                          │
   │                                                  │
   │ Library: Movies (248 files)                      │
   │                                                  │
   │ 720p:  ≈ 350 GB                                  │
   │ 1080p: ≈ 600 GB                                  │
   │ ─────────────────                                │
   │ Tổng:  ≈ 950 GB                                  │
   │                                                  │
   │ Dung lượng trống: 2.1 TB ✅ Đủ                   │
   │                                                  │
   │ Thời gian ước tính: ~3 ngày (VAAPI)              │
   │                                                  │
   │ [Bắt đầu encode]  [Hủy]                         │
   └─────────────────────────────────────────────────┘
   ```

6. [ ] Warning khi không đủ dung lượng:
   ```
   ⚠️ Dung lượng trống: 200 GB — KHÔNG ĐỦ cho 950 GB!
   Gợi ý: Bỏ bớt profile hoặc giải phóng ổ cứng.
   ```

### Progress Dashboard
7. [ ] Hiển thị khi đang encode:
   ```
   ┌─────────────────────────────────────────────────┐
   │ 📊 Pre-transcode Progress                       │
   │                                                  │
   │ ████████░░░░░░░░░░░░ 45/248 (18%)               │
   │                                                  │
   │ ▶ Đang encode: Friends S01E05 (720p)             │
   │   Tốc độ: 2.5x realtime | ETA: 12 phút          │
   │                                                  │
   │ ✅ Xong: 44  ❌ Lỗi: 1  ⏳ Chờ: 203             │
   │                                                  │
   │ Dung lượng đã dùng: 62 GB / 350 GB               │
   │                                                  │
   │ [Tạm dừng]  [Dừng hẳn]                          │
   └─────────────────────────────────────────────────┘
   ```

8. [ ] Error details expandable:
   ```
   ❌ 1 file lỗi:
   └─ Malcolm.S01E03.mkv — FFmpeg error: ...
      [Thử lại]  [Bỏ qua]
   ```

### API Integration
9. [ ] Hook vào existing settings store (`useSettings.ts`)
10. [ ] WebSocket: listen cho pretranscode progress updates

## Files to Create/Modify
- `webapp/src/pages/SettingsPage.tsx` — add Pre-transcode section
- `webapp/src/components/settings/PretranscodeSettings.tsx` — new
- `webapp/src/components/settings/PretranscodeProgress.tsx` — new
- `webapp/src/components/settings/StorageEstimation.tsx` — new
- `webapp/src/api/pretranscode.ts` — new API client
- `webapp/src/hooks/stores/useSettings.ts` — add pretranscode mutations

## Test Criteria
- [ ] Toggle ON → shows estimation → confirms → starts encoding
- [ ] Toggle OFF → stops encoding, asks to delete files
- [ ] Progress bar updates in real-time (WebSocket)
- [ ] Storage warning khi không đủ dung lượng
- [ ] Pause/Resume from UI

---
Next Phase: [phase-05-testing.md](phase-05-testing.md)
