
# HANDOVER DOCUMENT

**Date:** 2026-04-01

## Current State: Plan S designed (18 phases), subtitle bug fixed

### Done this session:

#### Plan S — Expo Mobile App (detailed design, 3 review rounds):
1. **Broke down Plan S** from 9 vague phases → 18 detailed phases in `plans/plan-s-expo-mobile-app/` folder
2. **Review round 1** (7 findings) — fixed: dev build vs Expo Go, API contracts (max_height not max_width, position not position_seconds), SecureStore for tokens, cleartext LAN, image URL helper, admin-only endpoints, unified search
3. **Review round 2** (5 findings) — fixed: test criteria MMKV→SecureStore, auth guard hydration gate (hasHydrated), resume source of truth (playbackInfo.position only), missing deps in phase-09, PiP concrete implementation
4. **Review round 3** (3 findings) — fixed: startsPictureInPictureAutomatically prop, stream URL response shape (api_key not token), app.json in Files to Modify

#### Bug Fix — Duplicate subtitle rendering (commit 50bff8f):
- **Problem:** Primary subtitle rendered twice — once by native `<track>` (browser) + once by DualSubtitleOverlay (custom JS)
- **Root cause:** `textTracks[i].mode = 'showing'` in WatchPage.tsx line 857 caused browser to render subtitle natively alongside custom overlay
- **Fix:** Changed `'showing'` → `'hidden'` — cues still available for iOS native fullscreen fallback, but browser doesn't render visually
- **File:** `webapp/src/pages/WatchPage.tsx` line 857

### Commits:
- `50bff8f` — Fix(playback): prevent double subtitle rendering on WatchPage

### Plan S Architecture Summary:
```
Phase 01-06: Monorepo Setup — pnpm workspaces + @velox/shared package
  - Extract: types (633 LOC), API client (222 LOC), store factories, hooks (1,608 LOC), libs
  - PlatformAdapter interface: storage/secureStorage/getDeviceName/getApiBaseUrl
  - Key decisions: SecureStore for tokens, MMKV for prefs, factory pattern for stores

Phase 07-08: Webapp Migration — re-export from shared + regression test

Phase 09-18: Expo Mobile App
  - Dev build (CNG, not Expo Go) — react-native-mmkv needs native code
  - expo-build-properties plugin for usesCleartextTraffic (LAN HTTP)
  - Auth: SecureStore + hydration gate (hasHydrated) to prevent flash-to-login
  - Player: ExoPlayer Direct Play, playbackInfo.position as resume truth
  - Subtitles: selected_subtitle_id in PlaybackInfoRequest
  - PiP: expo-video allowsPictureInPicture + startsPictureInPictureAutomatically
```

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

### Pending:
1. Plan M: Search, Filter & Folder Browser (designed, not coded)
2. Plan N: i18n remaining pages
3. Plan S: Expo Mobile App (18 phases designed, ready to build)
4. WatchPage.tsx still 1836 lines / 598KB chunk (monitoring item)

### Important notes for next session:
- NAS deploy via `docker compose up -d --force-recreate` (NOT docker restart)
- SSH alias: `thawnghoNas` (user: thawngminh, key: id_ed25519)
- Docker path on NAS: `/usr/local/bin/docker`
- group_add 937 is critical for VAAPI — without it, HW encode fails silently
- Pretranscode uses dedicated bgDB to avoid SQLite connection starvation
- Admin password: admin123 (not admin)
- All barrel re-exports maintain backward compatibility — no import migration needed
- Subtitle bug fix: track.mode 'hidden' not 'showing' (commit 50bff8f)
