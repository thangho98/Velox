# Phase 01: Activity Logging & Stats
Status: ⬜ Pending

## Tasks
### 1. Activity Log Table
- [ ] `activity_log`: id, user_id, action, media_id, details_json, ip, created_at
- [ ] Actions: play_start, play_stop, login, library_scan, media_added

### 2. Async Logger
- [ ] Fire-and-forget logging (don't block requests)
- [ ] Buffer writes, batch insert every 5s

### 3. Log Play Events
- [ ] Log on stream start/stop with duration watched

### 4. Activity API (Admin)
- [ ] `GET /api/admin/activity?limit=50&user_id=&action=&from=&to=`

### 5. Playback Statistics
- [ ] Most watched items, most active users, playback by day/week
- [ ] `GET /api/admin/stats/playback`
