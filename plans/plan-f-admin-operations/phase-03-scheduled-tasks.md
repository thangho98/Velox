# Phase 03: Scheduled Tasks
Status: ✅ Done

## Tasks
### 1. Task Scheduler
- [ ] Simple cron scheduler (goroutine + time.Ticker)
- [ ] Register tasks with interval, track last/next run

### 2. Built-in Tasks
- [ ] LibraryScan (daily 3AM), TranscodeCleanup (daily, 7d old), MissingFileCheck (weekly), MetadataRefresh (weekly)

### 3. Task API
- [ ] `GET /api/admin/tasks` - list tasks + status
- [ ] `POST /api/admin/tasks/{name}/run` - trigger manually
- [ ] `GET /api/admin/tasks/{name}/history` - last 100 runs

### 4. Startup Tasks
- [ ] Check missing files, cleanup stale transcode sessions, resume interrupted scans

---
✅ End of Plan F
Next Plan: plan-g-nice-to-have
