# Phase 02: Admin Dashboard
Status: ✅ Done

## Tasks
### 1. Active Streams Monitor
- [ ] `GET /api/admin/streams` - active playback sessions (user, media, quality, transcode status)
- [ ] `DELETE /api/admin/streams/{id}` - kill stream

### 2. Server Info
- [ ] `GET /api/admin/server` - CPU, RAM, disk, FFmpeg version, uptime

### 3. Library Stats
- [ ] Per-library: item count, total size, last scanned, codec/resolution breakdown

### 4. Dashboard UI (Frontend)
- [ ] Route: `/admin/dashboard`
- [ ] Cards: active streams, total items, server load
- [ ] Recent activity feed

### 5. Webhooks
- [ ] `webhooks` table: id, url, events, active
- [ ] Fire on: media_added, playback_start, scan_complete, error
- [ ] CRUD API for webhook management
