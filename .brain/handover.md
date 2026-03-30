
# HANDOVER DOCUMENT

**Date:** 2026-03-30

## Current State: Pretranscode scheduler fix + Playback UI improvements

### Done this session:

#### Pretranscode Scheduler (major fix):
1. **Auto-enqueue on startup** — scheduler now auto-enqueues audio-remux + video pretranscode jobs when container starts (previously required manual Start from UI)
2. **Dedicated background DB** — pretranscode scheduler uses separate `sql.DB` handle (`bgDB`) to avoid goroutine starvation from HTTP handler status polling on shared `MaxOpenConns(1)` connection
3. **GetStatusWith()** — status queries route through main DB (not bgDB), preventing status polling from starving the scheduler goroutine
4. **EnqueueJob UPSERT** — changed from `INSERT OR IGNORE` to `ON CONFLICT DO UPDATE SET status='queued' WHERE status IN ('cancelled','failed')`. Root cause: CancelAll set jobs to "cancelled", then restart re-enqueue silently ignored them → queue always empty
5. **TryActiveCount()** — non-blocking `TryLock` version of `ActiveCount()` to avoid mutex deadlock between transcoder and pretranscode scheduler
6. **nice -n 19** — all pretranscode FFmpeg processes run at lowest CPU priority via `niceFFmpeg()` helper. NAS stays responsive.
7. **Yield to realtime** — scheduler checks `TryActiveCount() > 0` and sleeps 5s, only blocks new job pickup (doesn't kill running FFmpeg)

#### Playback UI:
1. **Progress bar buffer fix** — buffer indicator was floating in the middle of the bar during server-side seeking. Fixed: buffer bar always starts from 0%, width = `bufferedRange.end / duration`

#### FullTranscode seeking fix (from previous context):
- Restricted HLS session reload to `TranscodeAudio` only
- FullTranscode seeks clamped to `hlsSeekableEndRef.current` instead of triggering new transcode (avoids VAAPI slot timeout)

### Files changed:
- `backend/internal/database/database.go` — `_busy_timeout=5000` added to SQLite DSN
- `backend/internal/database/migrate/028_audio_pretranscode.go` — audio-remux profile migration
- `backend/internal/database/migrate/registry.go` — register migration 028
- `backend/cmd/server/main.go` — separate bgDB for pretranscode, pretranscodeStatusRepo on main DB
- `backend/internal/service/pretranscode.go` — auto-enqueue, pickAudioRemuxJob, EnqueueAudioRemux, processJob audio-remux/universal-transcode, niceFFmpeg, TryActiveCount interface, GetStatusWith
- `backend/internal/handler/pretranscode.go` — statusRepo param, GetStatus uses GetStatusWith with main DB repos
- `backend/internal/handler/stream.go` — start offset / seeking support
- `backend/internal/handler/stream_test.go` — start offset tests
- `backend/internal/repository/pretranscode.go` — EnqueueJob UPSERT, PickNextJob excludes copy, PickNextJobForProfile, QueueStats excludes copy, GetAudioRemuxProfile, ListNonAACMediaFiles
- `backend/internal/transcoder/transcoder.go` — TryActiveCount() with TryLock, startOffset support
- `backend/internal/transcoder/transcoder_test.go` — offset tests
- `backend/internal/playback/engine.go` — HLS seeking startOffset
- `backend/internal/playback/profile.go` — profile changes
- `backend/internal/service/stream.go` — startOffset forwarding
- `backend/internal/service/library.go` — pretranscode integration
- `webapp/src/pages/WatchPage.tsx` — buffer bar fix, FullTranscode seek clamping, UI improvements (seek 5s, bigger buttons, iOS fullscreen)
- `webapp/src/components/DualSubtitleOverlay.tsx` — subtitle changes
- `Dockerfile` — pretranscode dir support

### NAS Docker Compose (canonical config):
Located at `/volume5/docker/velox/docker-compose.yml`:
```yaml
services:
  velox:
    image: velox:latest
    container_name: velox
    shm_size: '256m'
    group_add: ["937"]      # video group for VAAPI /dev/dri
    platform: linux/amd64
    environment:
      - UID=1026
      - GID=100
      - GIDLIST=937
      - VELOX_DATA_DIR=/data
      - VELOX_HW_ACCEL=auto
      - VELOX_MAX_TRANSCODES=2
      - VELOX_FILE_WATCHER=true
      - VELOX_TRICKPLAY_ENABLED=false
      - VELOX_PRETRANSCODE_DIR=/pretranscode
      - SUBSCENE_ENABLED=true
      - TZ=Asia/Ho_Chi_Minh
    volumes:
      - /volume5/docker/velox/data:/data:rw
      - /volume1/Data/pretranscode:/pretranscode:rw
      - /volume1/Media/Movie:/media:ro
      - /volume4/Downloads/qbittorrent:/media2:ro
      - /volume1/Media/Music:/media3:ro
      - /volume4/Downloads/pyload:/media4:ro
    devices: ["/dev/dri:/dev/dri"]
    ports: ["8098:80"]
    restart: on-failure
    tmpfs: ["/tmp:size=256m"]
```

### Deploy workflow:
```bash
docker build --platform linux/amd64 -t velox:latest .
docker save velox:latest | gzip | ssh thawnghoNas "cat > /tmp/velox.tar.gz"
ssh thawnghoNas "/usr/local/bin/docker load < /tmp/velox.tar.gz && rm /tmp/velox.tar.gz"
ssh thawnghoNas "cd /volume5/docker/velox && /usr/local/bin/docker compose up -d --force-recreate"
```
⚠️ NEVER use `docker restart` — it does NOT load new images!

### Pretranscode status:
- Audio-remux: 67 non-AAC files auto-enqueued (priority 100, runs first)
- Video pretranscode: 656 jobs auto-enqueued when enabled
- VAAPI h264_vaapi encode working (group_add 937 required)
- nice -n 19 keeps NAS responsive during encoding
- Scheduler yields to realtime transcode when users watching

### Pending:
1. Plan M: Search, Filter & Folder Browser
2. Plan N: i18n remaining pages
3. Frontend "Network error" UX — retry button when transcode slots busy
4. VAAPI sometimes fails on certain files → auto-fallback to libx264 (working, but slower)

### Important notes for next session:
- NAS deploy via `docker compose up -d --force-recreate` (NOT docker restart)
- SSH alias: `thawnghoNas` (user: thawngminh, key: id_ed25519)
- Docker path on NAS: `/usr/local/bin/docker`
- group_add 937 is critical for VAAPI — without it, HW encode fails silently and falls back to software
- Pretranscode uses dedicated bgDB to avoid SQLite connection starvation
- Admin playback setting is "direct_play" — forces direct play with audio transcode fallback
