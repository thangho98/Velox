# Phase 03: Playback Integration
Status: ⬜ Pending
Dependencies: Phase 02

## Objective
Sửa playback engine: khi user bấm play, check có bản pre-transcode sẵn không → serve luôn (instant play). Không có → fallback realtime transcode như bình thường.

## Implementation Steps

### Playback Decision
1. [ ] Sửa `playback/engine.go` — thêm check pre-transcode:
   ```
   func Decide(media, profile) Decision:
     // Step 0: Check pre-transcode (NEW)
     if pretranscodeFile := findBestPretranscode(media, profile):
       return Decision{
         Method: MethodPreTranscode,
         FilePath: pretranscodeFile.Path,
         Reason: "Pre-transcoded file available"
       }
     // Step 1: Direct Play check (existing)
     // Step 2: Direct Stream check (existing)
     // Step 3: Full Transcode (existing)
   ```

2. [ ] `findBestPretranscode(mediaFileID, clientProfile)`:
   - Query `pretranscode_files` WHERE status='ready' AND media_file_id=X
   - Chọn quality cao nhất mà client hỗ trợ
   - VD: client max 720p → chọn 720p pre-transcode, không lấy 1080p

### Stream Handler
3. [ ] Sửa `handler/stream.go` — thêm case MethodPreTranscode:
   - Serve file MP4 trực tiếp (http.ServeFile hoặc ServeContent)
   - Hỗ trợ Range requests (seeking)
   - Không cần HLS, không cần FFmpeg

4. [ ] Playback info response thêm field:
   ```json
   {
     "method": "PreTranscode",
     "reason": "Pre-transcoded 720p available",
     "pretranscode_quality": "720p"
   }
   ```

### Subtitle Handling
5. [ ] Pre-transcode files KHÔNG burn-in subtitle
   - Subtitle vẫn serve riêng (SRT/VTT client-side)
   - Lý do: burn-in = phải encode lại mỗi khi đổi sub

### ABR (Adaptive Bitrate) Support
6. [ ] Nếu có nhiều pre-transcode profiles (480p + 720p + 1080p):
   - Tạo HLS master playlist on-the-fly pointing to pre-transcode MP4s
   - Client tự switch quality dựa trên bandwidth
   - Hoặc đơn giản: serve quality cao nhất client hỗ trợ

## Files to Create/Modify
- `backend/internal/playback/engine.go` — add pre-transcode check
- `backend/internal/playback/engine.go` — add MethodPreTranscode constant
- `backend/internal/handler/stream.go` — serve pre-transcode files
- `backend/internal/handler/playback.go` — include pretranscode in info response

## Test Criteria
- [ ] Play pre-transcoded file: instant start, no FFmpeg
- [ ] Seeking works (Range headers)
- [ ] Fallback to realtime when no pre-transcode available
- [ ] Quality selection: picks best quality for client
- [ ] Subtitles work alongside pre-transcode playback

---
Next Phase: [phase-04-settings-ui.md](phase-04-settings-ui.md)
