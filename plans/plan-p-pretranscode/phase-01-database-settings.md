# Phase 01: Database + Settings
Status: ⬜ Pending
Dependencies: None

## Objective
Tạo schema lưu trữ pre-transcode jobs và files, thêm settings keys cho admin config.

## Implementation Steps

### Migration (026_pretranscode)
1. [ ] Tạo bảng `pretranscode_profiles` — quality profiles (480p, 720p, 1080p)
   ```sql
   CREATE TABLE pretranscode_profiles (
     id INTEGER PRIMARY KEY,
     name TEXT NOT NULL,           -- "720p", "1080p"
     height INTEGER NOT NULL,      -- 720, 1080
     video_bitrate INTEGER NOT NULL, -- kbps: 4000, 8000
     audio_bitrate INTEGER NOT NULL, -- kbps: 128, 192
     video_codec TEXT NOT NULL DEFAULT 'h264',
     audio_codec TEXT NOT NULL DEFAULT 'aac',
     enabled INTEGER NOT NULL DEFAULT 0,
     created_at TEXT NOT NULL DEFAULT (datetime('now'))
   );
   -- Seed default profiles
   INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate) VALUES
     ('480p', 480, 1500, 128),
     ('720p', 720, 4000, 128),
     ('1080p', 1080, 8000, 192);
   ```

2. [ ] Tạo bảng `pretranscode_files` — mỗi row = 1 file đã encode xong
   ```sql
   CREATE TABLE pretranscode_files (
     id INTEGER PRIMARY KEY,
     media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
     profile_id INTEGER NOT NULL REFERENCES pretranscode_profiles(id) ON DELETE CASCADE,
     file_path TEXT NOT NULL,       -- absolute path to encoded file
     file_size INTEGER NOT NULL,    -- bytes
     duration_secs REAL,
     status TEXT NOT NULL DEFAULT 'pending', -- pending, encoding, ready, failed
     error_message TEXT,
     started_at TEXT,
     completed_at TEXT,
     created_at TEXT NOT NULL DEFAULT (datetime('now')),
     UNIQUE(media_file_id, profile_id)
   );
   ```

3. [ ] Tạo bảng `pretranscode_queue` — job queue cho scheduler
   ```sql
   CREATE TABLE pretranscode_queue (
     id INTEGER PRIMARY KEY,
     media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
     profile_id INTEGER NOT NULL REFERENCES pretranscode_profiles(id) ON DELETE CASCADE,
     priority INTEGER NOT NULL DEFAULT 0, -- higher = first
     status TEXT NOT NULL DEFAULT 'queued', -- queued, encoding, done, failed, cancelled
     created_at TEXT NOT NULL DEFAULT (datetime('now')),
     UNIQUE(media_file_id, profile_id)
   );
   ```

### Settings Keys
4. [ ] Thêm constants vào `model/app_settings.go`:
   - `SettingPretranscodeEnabled` = "pretranscode_enabled" (true/false)
   - `SettingPretranscodeSchedule` = "pretranscode_schedule" (always/night/idle)
   - `SettingPretranscodeConcurrency` = "pretranscode_concurrency" (1-4)

### Model Structs
5. [ ] Tạo `model/pretranscode.go` — structs cho Profile, File, QueueItem

### Repository
6. [ ] Tạo `repository/pretranscode.go` — CRUD cho profiles, files, queue

### Storage Estimation
7. [ ] Thêm service method `EstimateStorage(ctx, libraryID, profileIDs)`:
   - Query tất cả media_files trong library
   - Tính: `file_count × avg_bitrate × avg_duration` cho mỗi profile
   - Trả về: estimated bytes per profile + total

8. [ ] Thêm service method `GetDiskFreeSpace(path)`:
   - Dùng `syscall.Statfs` để lấy dung lượng còn trống
   - So sánh với estimated → cảnh báo nếu không đủ

## Files to Create/Modify
- `backend/internal/database/migrate/registry.go` — migration 026
- `backend/internal/model/pretranscode.go` — new
- `backend/internal/model/app_settings.go` — add constants
- `backend/internal/repository/pretranscode.go` — new

## Test Criteria
- [ ] Migration up/down thành công
- [ ] CRUD operations cho profiles, files, queue
- [ ] Storage estimation trả về reasonable numbers
- [ ] Disk free space check hoạt động

---
Next Phase: [phase-02-backend-engine.md](phase-02-backend-engine.md)
