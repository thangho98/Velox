# Phase 02: Backend Engine
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Xây dựng engine encode nền: scheduler quét media thiếu pre-transcode, queue jobs, chạy FFmpeg, quản lý lifecycle.

## Implementation Steps

### Pre-transcode Service
1. [ ] Tạo `service/pretranscode.go` — orchestrator chính
   - `Start()` — khởi động scheduler goroutine
   - `Stop()` — graceful shutdown, đợi job đang chạy xong
   - `EnqueueLibrary(ctx, libraryID)` — scan library, thêm media chưa encode vào queue
   - `EnqueueMedia(ctx, mediaFileID)` — thêm 1 file cụ thể
   - `CancelAll()` — hủy tất cả jobs đang chờ
   - `Pause()` / `Resume()` — tạm dừng/tiếp tục scheduler
   - `GetProgress()` — trả về stats (total, done, encoding, failed)

### Scheduler
2. [ ] Scheduler loop:
   ```
   for {
     if paused → sleep 10s, continue
     check schedule (night/always/idle)
     if not in schedule → sleep 60s, continue
     pick next job from queue (ORDER BY priority DESC, created_at ASC)
     if no job → sleep 30s, continue
     run encode job
   }
   ```

3. [ ] Schedule modes:
   - `always` — chạy liên tục
   - `night` — chỉ chạy 00:00-06:00 (configurable)
   - `idle` — chạy khi không có active transcode sessions

### FFmpeg Encode
4. [ ] Tạo function `encodeFile(mediaFilePath, outputPath, profile)`:
   ```
   ffmpeg -i input.mkv \
     -vf scale=-2:{height} \
     -c:v h264_vaapi (or libx264) \
     -b:v {bitrate}k \
     -c:a aac -b:a {audio_bitrate}k -ac 2 \
     -movflags +faststart \
     -y output.mp4
   ```
   - Detect VAAPI/software automatically
   - Fallback software nếu HW fail
   - Progress tracking qua FFmpeg stderr parsing

5. [ ] Output path: `{VELOX_DATA_DIR}/pretranscode/{media_file_id}/{profile_name}.mp4`

6. [ ] Skip logic:
   - Skip nếu source đã là H.264+AAC ở cùng hoặc thấp hơn resolution
   - Skip nếu source resolution < profile height (không upscale)
   - Skip nếu file đã có bản pre-transcode ready

### Progress & Status
7. [ ] Update DB status: queued → encoding → ready/failed
8. [ ] Parse FFmpeg progress: `frame=`, `time=`, `speed=`
9. [ ] WebSocket notification khi job complete/fail
10. [ ] Log activity vào `activity_log`

### Cleanup
11. [ ] Auto-delete pre-transcode files khi:
    - Source media_file bị xóa (CASCADE trong DB + file cleanup)
    - Profile bị disable
    - Admin chạy cleanup manually

### API Endpoints
12. [ ] Handler endpoints (admin only):
    - `GET /api/admin/pretranscode/status` — overall progress
    - `POST /api/admin/pretranscode/start` — enqueue all libraries
    - `POST /api/admin/pretranscode/stop` — cancel + pause
    - `POST /api/admin/pretranscode/resume` — resume
    - `GET /api/admin/pretranscode/estimate` — storage estimation
    - `DELETE /api/admin/pretranscode/files` — cleanup all

## Files to Create/Modify
- `backend/internal/service/pretranscode.go` — new (main service)
- `backend/internal/handler/pretranscode.go` — new (API handlers)
- `backend/cmd/server/main.go` — wire up service + routes

## Test Criteria
- [ ] Encode single file: H.264+AAC MP4 output
- [ ] Skip logic: no upscale, no re-encode same codec
- [ ] Queue ordering: priority DESC, created_at ASC
- [ ] Pause/Resume works
- [ ] VAAPI fallback to software
- [ ] Cleanup removes files on disk + DB rows

---
Next Phase: [phase-03-playback-integration.md](phase-03-playback-integration.md)
