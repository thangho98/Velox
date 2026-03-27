# Plan P: Pre-transcode (Offline Encoding)
Created: 2026-03-26
Status: ✅ Done

## Overview
Netflix-style pre-encoding: hệ thống tự encode sẵn media thành H.264+AAC ở nhiều chất lượng (480p, 720p, 1080p) trong background. User bấm play → serve file có sẵn, instant playback. Fallback realtime transcode nếu chưa encode xong.

**Điểm khác biệt:** Emby/Jellyfin KHÔNG có tính năng này. Chỉ Plex có "Optimize". Velox sẽ là open-source media server đầu tiên có pre-transcode.

## Tech Stack
- Backend: Go (scheduler, FFmpeg orchestration, storage estimation)
- Frontend: React (Settings UI, progress dashboard)
- Database: SQLite (new migration for pretranscode tables)
- FFmpeg: VAAPI/software encoding

## Phases

| Phase | Name | Status | Tasks |
|-------|------|--------|-------|
| 01 | Database + Settings | ✅ Done | 8 |
| 02 | Backend Engine | ✅ Done | 12 |
| 03 | Playback Integration | ✅ Done | 6 |
| 04 | Settings UI + Dashboard | ✅ Done | 10 |
| 05 | Testing & Polish | ✅ Done | 7 |

**Total:** 43 tasks

## Key Decisions
- Pre-transcode files stored in `{VELOX_DATA_DIR}/pretranscode/{media_id}/`
- Output format: MP4 (H.264 + AAC) — universally browser-compatible
- Admin-only feature, opt-in per library
- Scheduler runs with configurable concurrency (default: 1 job)
- Storage estimation shown BEFORE enabling
- Auto-cleanup when source file removed or feature disabled

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
